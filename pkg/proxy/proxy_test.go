package proxy

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// mockPGUpstreamServer simulates a simple PostgreSQL server for transparent proxy testing.
func startMockPGUpstream(t *testing.T) (string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock PG listener: %v", err)
	}

	stopCh := make(chan struct{})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stopCh:
					return
				default:
					return
				}
			}

			go handleMockPGConn(conn)
		}
	}()

	addr := ln.Addr().String()
	cleanup := func() {
		close(stopCh)
		_ = ln.Close()
	}
	return addr, cleanup
}

func handleMockPGConn(conn net.Conn) {
	defer conn.Close()

	// 1. Read startup message
	msg, err := ReadPGClientPacket(conn, true)
	if err != nil {
		return
	}

	if msg.IsSSLRequest {
		// Respond 'N' (no SSL)
		_, _ = conn.Write([]byte{'N'})
		// Next message will be regular startup
		msg, err = ReadPGClientPacket(conn, true)
		if err != nil {
			return
		}
	}

	// Send AuthOK ('R')
	authOK := buildPGPacket('R', []byte{0x00, 0x00, 0x00, 0x00})
	_, _ = conn.Write(authOK)

	// Send ReadyForQuery ('Z')
	ready := buildPGPacket('Z', []byte{'I'})
	_, _ = conn.Write(ready)

	// Command loop
	for {
		clientMsg, err := ReadPGClientPacket(conn, false)
		if err != nil {
			return
		}

		if clientMsg.Type == 'X' {
			return
		}

		if clientMsg.Type == 'Q' {
			sql, _ := ParsePGQuery(clientMsg)
			var status byte = 'I'
			if sql == "BEGIN" {
				status = 'T'
			}

			// Send CommandComplete 'C'
			cc := buildPGPacket('C', []byte("OK\x00"))
			_, _ = conn.Write(cc)

			// Send ReadyForQuery 'Z'
			rfq := buildPGPacket('Z', []byte{status})
			_, _ = conn.Write(rfq)
		}
	}
}

func TestProxy_EndToEnd_Postgres(t *testing.T) {
	upstreamAddr, cleanupUpstream := startMockPGUpstream(t)
	defer cleanupUpstream()

	cfg := ProxyConfig{
		ListenAddr:       "127.0.0.1:0",
		UpstreamAddr:     upstreamAddr,
		Protocol:         ProtocolPostgres,
		MinJitter:        10 * time.Microsecond,
		MaxJitter:        50 * time.Microsecond,
		CommitBarrierMin: 50 * time.Microsecond,
		CommitBarrierMax: 100 * time.Microsecond,
		PCTDepth:         2,
		Seed:             42,
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create proxy server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for proxy listener ready
	proxyAddr := server.ListenAddr()
	for i := 0; i < 20 && proxyAddr == ""; i++ {
		time.Sleep(10 * time.Millisecond)
		proxyAddr = server.ListenAddr()
	}
	if proxyAddr == "" {
		t.Fatalf("proxy failed to start listening")
	}

	// Connect Client 1 to proxy
	client1, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("failed to dial proxy: %v", err)
	}
	defer client1.Close()

	// Send StartupMessage
	startup := buildPGStartupPacket("app_user", "testdb")
	if _, err := client1.Write(startup); err != nil {
		t.Fatalf("failed to write startup: %v", err)
	}

	// Read AuthOK and ReadyForQuery
	authMsg, err := ReadPGServerPacket(client1)
	if err != nil || authMsg.Type != 'R' {
		t.Fatalf("failed to read AuthOK: %v", err)
	}
	rfqMsg, err := ReadPGServerPacket(client1)
	if err != nil || rfqMsg.Type != 'Z' {
		t.Fatalf("failed to read ReadyForQuery: %v", err)
	}

	// Execute BEGIN
	beginPkt := buildPGPacket('Q', []byte("BEGIN\x00"))
	_, _ = client1.Write(beginPkt)
	_, _ = ReadPGServerPacket(client1) // CommandComplete
	_, _ = ReadPGServerPacket(client1) // ReadyForQuery

	// Execute SELECT accounts:1
	selectPkt := buildPGPacket('Q', []byte("SELECT balance FROM accounts WHERE id = 1\x00"))
	_, _ = client1.Write(selectPkt)
	_, _ = ReadPGServerPacket(client1)
	_, _ = ReadPGServerPacket(client1)

	// Connect Client 2 to proxy
	client2, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("failed to dial proxy for client 2: %v", err)
	}
	defer client2.Close()

	_, _ = client2.Write(startup)
	_, _ = ReadPGServerPacket(client2)
	_, _ = ReadPGServerPacket(client2)

	// Client 2: BEGIN -> SELECT accounts:1 -> UPDATE accounts:1 -> COMMIT
	_, _ = client2.Write(buildPGPacket('Q', []byte("BEGIN\x00")))
	_, _ = ReadPGServerPacket(client2)
	_, _ = ReadPGServerPacket(client2)

	_, _ = client2.Write(buildPGPacket('Q', []byte("SELECT balance FROM accounts WHERE id = 1\x00")))
	_, _ = ReadPGServerPacket(client2)
	_, _ = ReadPGServerPacket(client2)

	_, _ = client2.Write(buildPGPacket('Q', []byte("UPDATE accounts SET balance = balance - 50 WHERE id = 1\x00")))
	_, _ = ReadPGServerPacket(client2)
	_, _ = ReadPGServerPacket(client2)

	_, _ = client2.Write(buildPGPacket('Q', []byte("COMMIT\x00")))
	_, _ = ReadPGServerPacket(client2)
	_, _ = ReadPGServerPacket(client2)

	// Client 1: UPDATE accounts:1 -> COMMIT (Lost Update!)
	_, _ = client1.Write(buildPGPacket('Q', []byte("UPDATE accounts SET balance = balance - 100 WHERE id = 1\x00")))
	_, _ = ReadPGServerPacket(client1)
	_, _ = ReadPGServerPacket(client1)

	_, _ = client1.Write(buildPGPacket('Q', []byte("COMMIT\x00")))
	_, _ = ReadPGServerPacket(client1)
	_, _ = ReadPGServerPacket(client1)

	time.Sleep(50 * time.Millisecond)

	stats := server.Stats()
	if stats.TotalQueries < 6 {
		t.Fatalf("expected at least 6 intercepted queries, got %d", stats.TotalQueries)
	}

	anomalies := server.Engine().GetAnomalies()
	if len(anomalies) == 0 {
		t.Fatalf("expected proxy to detect Lost Update (P4) anomaly, got 0")
	}

	foundLostUpdate := false
	for _, a := range anomalies {
		if a.Type == domain.AnomalyLostUpdate {
			foundLostUpdate = true
			break
		}
	}
	if !foundLostUpdate {
		t.Fatalf("expected Lost Update anomaly in results: %v", anomalies)
	}

	// Test graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func startMockMySQLUpstream(t *testing.T) (string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock MySQL listener: %v", err)
	}

	stopCh := make(chan struct{})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-stopCh:
					return
				default:
					return
				}
			}

			go handleMockMySQLConn(conn)
		}
	}()

	addr := ln.Addr().String()
	cleanup := func() {
		close(stopCh)
		_ = ln.Close()
	}
	return addr, cleanup
}

func handleMockMySQLConn(conn net.Conn) {
	defer conn.Close()

	inTx := false
	for {
		msg, err := ReadMySQLPacket(conn)
		if err != nil {
			return
		}

		if msg.Command == MySQLComQuit {
			return
		}

		if msg.Command == MySQLComQuery {
			sql, _ := ParseMySQLQuery(msg)
			if sql == "BEGIN" {
				inTx = true
			} else if sql == "COMMIT" || sql == "ROLLBACK" {
				inTx = false
			}

			var status uint16 = 0x0000
			if inTx {
				status = 0x0001 // SERVER_STATUS_IN_TRANS
			}

			// Respond with OK packet
			okPayload := []byte{0x00, 0x00, 0x00, byte(status & 0xFF), byte((status >> 8) & 0xFF), 0x00, 0x00}
			okPkt := buildMySQLPacket(msg.SeqID+1, okPayload)
			if _, err := conn.Write(okPkt); err != nil {
				return
			}
		}
	}
}

func TestProxy_EndToEnd_MySQL(t *testing.T) {
	upstreamAddr, cleanupUpstream := startMockMySQLUpstream(t)
	defer cleanupUpstream()

	cfg := ProxyConfig{
		ListenAddr:       "127.0.0.1:0",
		UpstreamAddr:     upstreamAddr,
		Protocol:         ProtocolMySQL,
		MinJitter:        10 * time.Microsecond,
		MaxJitter:        50 * time.Microsecond,
		CommitBarrierMin: 50 * time.Microsecond,
		CommitBarrierMax: 100 * time.Microsecond,
		PCTDepth:         2,
		Seed:             99,
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create MySQL proxy server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Start(ctx)
	}()

	// Wait for listener
	proxyAddr := server.ListenAddr()
	for i := 0; i < 20 && proxyAddr == ""; i++ {
		time.Sleep(10 * time.Millisecond)
		proxyAddr = server.ListenAddr()
	}
	if proxyAddr == "" {
		t.Fatalf("proxy failed to start listening")
	}

	// Client 1
	c1, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("failed to dial MySQL proxy: %v", err)
	}
	defer c1.Close()

	// Client 2
	c2, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("failed to dial MySQL proxy for c2: %v", err)
	}
	defer c2.Close()

	// C1: BEGIN -> SELECT accounts:1
	qBegin := buildMySQLPacket(0, append([]byte{MySQLComQuery}, []byte("BEGIN")...))
	_, _ = c1.Write(qBegin)
	_, _ = ReadMySQLPacket(c1)

	qSel1 := buildMySQLPacket(0, append([]byte{MySQLComQuery}, []byte("SELECT balance FROM accounts WHERE id = 1")...))
	_, _ = c1.Write(qSel1)
	_, _ = ReadMySQLPacket(c1)

	// C2: BEGIN -> SELECT accounts:1 -> UPDATE accounts:1 -> COMMIT
	_, _ = c2.Write(qBegin)
	_, _ = ReadMySQLPacket(c2)

	_, _ = c2.Write(qSel1)
	_, _ = ReadMySQLPacket(c2)

	qUpd2 := buildMySQLPacket(0, append([]byte{MySQLComQuery}, []byte("UPDATE accounts SET balance = balance - 50 WHERE id = 1")...))
	_, _ = c2.Write(qUpd2)
	_, _ = ReadMySQLPacket(c2)

	qCommit := buildMySQLPacket(0, append([]byte{MySQLComQuery}, []byte("COMMIT")...))
	_, _ = c2.Write(qCommit)
	_, _ = ReadMySQLPacket(c2)

	// C1: UPDATE accounts:1 -> COMMIT
	qUpd1 := buildMySQLPacket(0, append([]byte{MySQLComQuery}, []byte("UPDATE accounts SET balance = balance - 100 WHERE id = 1")...))
	_, _ = c1.Write(qUpd1)
	_, _ = ReadMySQLPacket(c1)

	_, _ = c1.Write(qCommit)
	_, _ = ReadMySQLPacket(c1)

	time.Sleep(50 * time.Millisecond)

	stats := server.Stats()
	if stats.TotalQueries < 6 {
		t.Fatalf("expected at least 6 queries in MySQL proxy, got %d", stats.TotalQueries)
	}

	anomalies := server.Engine().GetAnomalies()
	if len(anomalies) == 0 {
		t.Fatalf("expected anomaly in MySQL proxy, got 0")
	}

	foundLostUpdate := false
	for _, a := range anomalies {
		if a.Type == domain.AnomalyLostUpdate {
			foundLostUpdate = true
			break
		}
	}
	if !foundLostUpdate {
		t.Fatalf("expected Lost Update in MySQL anomalies: %v", anomalies)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}
