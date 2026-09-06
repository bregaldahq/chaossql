package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Server represents the active Layer-7 Transparent Database Reverse Proxy.
type Server struct {
	mu             sync.RWMutex
	cfg            ProxyConfig
	listener       net.Listener
	scheduler      *PCTScheduler
	engine         *ShadowGraphEngine
	sessions       sync.Map
	sessionCounter atomic.Uint64
	listenAddr     string
	closed         atomic.Bool
	stopCh         chan struct{}
	wg             sync.WaitGroup

	// Telemetry counters
	totalQueries   atomic.Uint64
	interceptedDML atomic.Uint64
	jitterCount    atomic.Uint64
	totalDelayNs   atomic.Int64
}

// NewServer constructs an unstarted proxy server.
func NewServer(cfg ProxyConfig) (*Server, error) {
	if cfg.Protocol == "" {
		cfg.Protocol = ProtocolPostgres
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:5433"
	}
	if cfg.UpstreamAddr == "" {
		cfg.UpstreamAddr = "127.0.0.1:5432"
	}

	sched := NewPCTScheduler(
		cfg.MinJitter,
		cfg.MaxJitter,
		cfg.CommitBarrierMin,
		cfg.CommitBarrierMax,
		cfg.PCTDepth,
		cfg.Seed,
	)

	return &Server{
		cfg:       cfg,
		scheduler: sched,
		engine:    NewShadowGraphEngine(),
		stopCh:    make(chan struct{}),
	}, nil
}

// Engine returns the underlying ShadowGraphEngine.
func (s *Server) Engine() *ShadowGraphEngine {
	return s.engine
}

// ListenAddr returns the actual bound network address.
func (s *Server) ListenAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.listenAddr
}

// Stats returns a snapshot of proxy operational metrics.
func (s *Server) Stats() ProxyStats {
	var activeCount int
	s.sessions.Range(func(_, _ interface{}) bool {
		activeCount++
		return true
	})

	return ProxyStats{
		ActiveSessions:      activeCount,
		TotalSessions:       s.sessionCounter.Load(),
		TotalQueries:        s.totalQueries.Load(),
		InterceptedDML:      s.interceptedDML.Load(),
		JitterInjectedCount: s.jitterCount.Load(),
		TotalJitterDelay:    time.Duration(s.totalDelayNs.Load()),
		AnomaliesDetected:   len(s.engine.GetAnomalies()),
	}
}

// Start opens the TCP listener and accepts client connections until context is cancelled or Shutdown is invoked.
func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to bind proxy listener on %s: %w", s.cfg.ListenAddr, err)
	}
	s.mu.Lock()
	s.listener = ln
	s.listenAddr = ln.Addr().String()
	s.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			_ = s.Shutdown(context.Background())
		case <-s.stopCh:
		}
	}()

	for {
		clientConn, err := ln.Accept()
		if err != nil {
			if s.closed.Load() {
				return nil
			}
			return err
		}

		s.wg.Add(1)
		go s.handleConnection(clientConn)
	}
}

// Shutdown gracefully closes the listener and active client connections.
func (s *Server) Shutdown(ctx context.Context) error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(s.stopCh)
	s.mu.Lock()
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.mu.Unlock()

	// Close all active sessions
	s.sessions.Range(func(key, value interface{}) bool {
		sess := value.(*ProxySession)
		if sess.ClientConn != nil {
			_ = sess.ClientConn.Close()
		}
		if sess.ServerConn != nil {
			_ = sess.ServerConn.Close()
		}
		return true
	})

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) handleConnection(clientConn net.Conn) {
	defer s.wg.Done()
	defer clientConn.Close()

	serverConn, err := net.Dial("tcp", s.cfg.UpstreamAddr)
	if err != nil {
		return
	}
	defer serverConn.Close()

	sessionID := s.sessionCounter.Add(1)
	session := NewProxySession(sessionID, clientConn, serverConn)
	s.sessions.Store(sessionID, session)
	defer s.sessions.Delete(sessionID)

	var sessionWg sync.WaitGroup
	sessionWg.Add(2)

	if s.cfg.Protocol == ProtocolMySQL {
		go func() {
			defer sessionWg.Done()
			s.forwardMySQLClient(session)
		}()
		go func() {
			defer sessionWg.Done()
			s.forwardMySQLServer(session)
		}()
	} else {
		// PostgreSQL default: complete initial handshake (SSLRequest negotiation + StartupMessage)
		if err := s.handlePGHandshake(session); err != nil {
			return
		}

		go func() {
			defer sessionWg.Done()
			s.forwardPGClient(session)
		}()
		go func() {
			defer sessionWg.Done()
			s.forwardPGServer(session)
		}()
	}

	sessionWg.Wait()
}

func (s *Server) handlePGHandshake(session *ProxySession) error {
	for {
		msg, err := ReadPGClientPacket(session.ClientConn, true)
		if err != nil {
			return err
		}

		if msg.IsSSLRequest {
			if _, err := session.ServerConn.Write(msg.Raw); err != nil {
				return err
			}
			var sslResp [1]byte
			if _, err := io.ReadFull(session.ServerConn, sslResp[:]); err != nil {
				return err
			}
			if _, err := session.ClientConn.Write(sslResp[:]); err != nil {
				return err
			}
			continue
		}

		// Regular StartupMessage
		if _, err := session.ServerConn.Write(msg.Raw); err != nil {
			return err
		}
		return nil
	}
}

func (s *Server) forwardPGClient(session *ProxySession) {
	for {
		msg, err := ReadPGClientPacket(session.ClientConn, false)
		if err != nil {
			return
		}

		// Non-startup packets
		if msg.Type == 'X' {
			// Terminate connection
			_, _ = session.ServerConn.Write(msg.Raw)
			return
		}

		if msg.Type == 'Q' || msg.Type == 'P' {
			s.totalQueries.Add(1)
			sql, ok := ParsePGQuery(msg)
			if ok {
				isWrite, items, isBoundary, bType, iso := InspectSQL(sql)

				if isBoundary {
					switch bType {
					case "BEGIN":
						s.engine.StartTx(session, iso)
					case "COMMIT":
						barrierDelay := s.scheduler.CalculateCommitBarrier()
						if barrierDelay > 0 {
							s.jitterCount.Add(1)
							s.totalDelayNs.Add(int64(barrierDelay))
							time.Sleep(barrierDelay)
						}
						if _, err := session.ServerConn.Write(msg.Raw); err != nil {
							return
						}
						s.engine.RecordCommit(session)
						continue

					case "ROLLBACK":
						if _, err := session.ServerConn.Write(msg.Raw); err != nil {
							return
						}
						s.engine.RecordRollback(session)
						continue

					case "SET_ISOLATION":
						if iso != "" {
							session.SetIsolation(iso)
						}
					}
				} else {
					if isWrite {
						s.interceptedDML.Add(1)
					}
					if session.IsInTx() && (isWrite || IsDMLOrSelect(sql)) {
						delay := s.scheduler.CalculateJitter(session.ID, isWrite)
						if delay > 0 {
							s.jitterCount.Add(1)
							s.totalDelayNs.Add(int64(delay))
							time.Sleep(delay)
						}
					}

					if isWrite {
						s.engine.RecordWrite(session, items, sql)
					} else {
						s.engine.RecordRead(session, items, sql)
					}
				}
			}
		}

		if _, err := session.ServerConn.Write(msg.Raw); err != nil {
			return
		}
	}
}

func (s *Server) forwardPGServer(session *ProxySession) {
	for {
		msg, err := ReadPGServerPacket(session.ServerConn)
		if err != nil {
			return
		}

		if msg.Type == 'Z' {
			status, ok := ParsePGReadyForQuery(msg)
			if ok {
				switch status {
				case 'I':
					session.SetInTx(false)
				case 'T':
					session.SetInTx(true)
				}
			}
		}

		if _, err := session.ClientConn.Write(msg.Raw); err != nil {
			return
		}
	}
}

func (s *Server) forwardMySQLClient(session *ProxySession) {
	for {
		msg, err := ReadMySQLPacket(session.ClientConn)
		if err != nil {
			return
		}

		if msg.Command == MySQLComQuit {
			_, _ = session.ServerConn.Write(msg.Raw)
			return
		}

		if msg.Command == MySQLComQuery || msg.Command == MySQLComStmtPrepare {
			s.totalQueries.Add(1)
			sql, ok := ParseMySQLQuery(msg)
			if ok {
				isWrite, items, isBoundary, bType, iso := InspectSQL(sql)

				if isBoundary {
					switch bType {
					case "BEGIN":
						s.engine.StartTx(session, iso)
					case "COMMIT":
						barrierDelay := s.scheduler.CalculateCommitBarrier()
						if barrierDelay > 0 {
							s.jitterCount.Add(1)
							s.totalDelayNs.Add(int64(barrierDelay))
							time.Sleep(barrierDelay)
						}
						if _, err := session.ServerConn.Write(msg.Raw); err != nil {
							return
						}
						s.engine.RecordCommit(session)
						continue

					case "ROLLBACK":
						if _, err := session.ServerConn.Write(msg.Raw); err != nil {
							return
						}
						s.engine.RecordRollback(session)
						continue

					case "SET_ISOLATION":
						if iso != "" {
							session.SetIsolation(iso)
						}
					}
				} else {
					if isWrite {
						s.interceptedDML.Add(1)
					}
					if session.IsInTx() && (isWrite || IsDMLOrSelect(sql)) {
						delay := s.scheduler.CalculateJitter(session.ID, isWrite)
						if delay > 0 {
							s.jitterCount.Add(1)
							s.totalDelayNs.Add(int64(delay))
							time.Sleep(delay)
						}
					}

					if isWrite {
						s.engine.RecordWrite(session, items, sql)
					} else {
						s.engine.RecordRead(session, items, sql)
					}
				}
			}
		}

		if _, err := session.ServerConn.Write(msg.Raw); err != nil {
			return
		}
	}
}

func (s *Server) forwardMySQLServer(session *ProxySession) {
	for {
		msg, err := ReadMySQLPacket(session.ServerConn)
		if err != nil {
			return
		}

		if isErr, errCode, _ := ParseMySQLErrPacket(msg); isErr {
			if IsMySQLDeadlock(errCode) {
				s.engine.RecordRollback(session)
			}
		} else if isOK, inTrans := ParseMySQLOKPacket(msg); isOK {
			if !inTrans && session.IsInTx() {
				s.engine.RecordCommit(session)
			}
		}

		if _, err := session.ClientConn.Write(msg.Raw); err != nil {
			return
		}
	}
}
