package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

// ProtocolType identifies the database wire protocol.
type ProtocolType string

const (
	ProtocolPostgres ProtocolType = "postgres"
	ProtocolMySQL    ProtocolType = "mysql"
)

// ProxyConfig defines runtime parameters for the transparent database reverse proxy.
type ProxyConfig struct {
	ListenAddr       string        `json:"listen_addr"`
	UpstreamAddr     string        `json:"upstream_addr"`
	Protocol         ProtocolType  `json:"protocol"`
	MinJitter        time.Duration `json:"min_jitter"`
	MaxJitter        time.Duration `json:"max_jitter"`
	PCTDepth         int           `json:"pct_depth"`
	CommitBarrierMin time.Duration `json:"commit_barrier_min"`
	CommitBarrierMax time.Duration `json:"commit_barrier_max"`
	ExportSARIF      string        `json:"export_sarif,omitempty"`
	UIPort           int           `json:"ui_port,omitempty"`
	Seed             uint64        `json:"seed"`
}

// ProxySession tracks individual client connection state, active transactions, and read/write sets.
type ProxySession struct {
	mu         sync.RWMutex
	ID         uint64
	ClientConn net.Conn
	ServerConn net.Conn
	InTx       bool
	Isolation  domain.IsolationLevel
	TxID       string
	TxCounter  uint64
	ReadSet    map[string]struct{}
	WriteSet   map[string]struct{}
	Statements []string
	CreatedAt  time.Time
}

// NewProxySession constructs an initialized ProxySession instance.
func NewProxySession(id uint64, clientConn, serverConn net.Conn) *ProxySession {
	return &ProxySession{
		ID:         id,
		ClientConn: clientConn,
		ServerConn: serverConn,
		InTx:       false,
		Isolation:  domain.LevelReadCommitted,
		ReadSet:    make(map[string]struct{}),
		WriteSet:   make(map[string]struct{}),
		Statements: make([]string, 0),
		CreatedAt:  time.Now(),
	}
}

// StartTransaction transitions the session into an active transaction block.
func (s *ProxySession) StartTransaction(isolation domain.IsolationLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TxCounter++
	s.InTx = true
	if isolation != "" {
		s.Isolation = isolation
	}
	s.TxID = fmt.Sprintf("S%d-tx%d", s.ID, s.TxCounter)
	s.ReadSet = make(map[string]struct{})
	s.WriteSet = make(map[string]struct{})
	s.Statements = make([]string, 0)
}

// ResetTransaction clears active transaction state upon commit or rollback.
func (s *ProxySession) ResetTransaction() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InTx = false
	s.TxID = ""
	s.ReadSet = make(map[string]struct{})
	s.WriteSet = make(map[string]struct{})
}

// AddRead records read items in the session's current read set.
func (s *ProxySession) AddRead(items ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		if item != "" {
			s.ReadSet[item] = struct{}{}
		}
	}
}

// AddWrite records written items in the session's current write set.
func (s *ProxySession) AddWrite(items ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range items {
		if item != "" {
			s.WriteSet[item] = struct{}{}
		}
	}
}

// AddStatement logs an executed statement for causal diagnosis.
func (s *ProxySession) AddStatement(stmt string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Statements = append(s.Statements, stmt)
}

// IsInTx returns whether the session is currently in a transaction block.
func (s *ProxySession) IsInTx() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.InTx
}

// SetInTx updates the in-transaction status in a thread-safe manner.
func (s *ProxySession) SetInTx(inTx bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.InTx = inTx
}

// SetIsolation updates the session isolation level in a thread-safe manner.
func (s *ProxySession) SetIsolation(iso domain.IsolationLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Isolation = iso
}

// GetSnapshot returns safe copies of active transaction sets.
func (s *ProxySession) GetSnapshot() (inTx bool, txID string, readSet, writeSet map[string]struct{}, stmts []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rs := make(map[string]struct{}, len(s.ReadSet))
	for k := range s.ReadSet {
		rs[k] = struct{}{}
	}
	ws := make(map[string]struct{}, len(s.WriteSet))
	for k := range s.WriteSet {
		ws[k] = struct{}{}
	}
	st := make([]string, len(s.Statements))
	copy(st, s.Statements)
	return s.InTx, s.TxID, rs, ws, st
}

// DetectedAnomaly encapsulates an isolation anomaly identified in real-time by the shadow graph.
type DetectedAnomaly struct {
	ID          string             `json:"id"`
	Type        domain.AnomalyType `json:"type"`
	SessionIDs  []uint64           `json:"session_ids"`
	TxIDs       []string           `json:"tx_ids"`
	Items       []string           `json:"items"`
	Cycle       []analyzer.Edge    `json:"cycle,omitempty"`
	Statements  []string           `json:"statements,omitempty"`
	DetectedAt  time.Time          `json:"detected_at"`
	Description string             `json:"description"`
}

// ProxyStats aggregates runtime telemetry and anomaly counters.
type ProxyStats struct {
	ActiveSessions      int           `json:"active_sessions"`
	TotalSessions       uint64        `json:"total_sessions"`
	TotalQueries        uint64        `json:"total_queries"`
	InterceptedDML      uint64        `json:"intercepted_dml"`
	JitterInjectedCount uint64        `json:"jitter_injected_count"`
	TotalJitterDelay    time.Duration `json:"total_jitter_delay"`
	AnomaliesDetected   int           `json:"anomalies_detected"`
}
