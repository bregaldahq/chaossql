package proxy

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

type txState struct {
	SessionID  uint64
	TxID       string
	Isolation  domain.IsolationLevel
	ReadSet    map[string]struct{}
	WriteSet   map[string]struct{}
	Statements []string
}

// ShadowGraphEngine tracks active sessions, construct Adya conflict edges, and finds runtime anomaly cycles.
type ShadowGraphEngine struct {
	mu                 sync.RWMutex
	graph              *analyzer.AdyaGraph
	activeTx           map[string]*txState
	itemReaders        map[string]map[string]bool // item -> map[txID]bool
	itemLastWriter     map[string]string          // item -> committed txID
	uncommittedWriters map[string]string          // item -> active txID
	abortedTx          map[string]bool
	detectedAnomalies  []DetectedAnomaly
	reportedCycles     map[string]bool
	trace              domain.ExecutionTrace
	onAnomaly          func(DetectedAnomaly)
	startTime          time.Time
}

// NewShadowGraphEngine constructs a thread-safe Shadow Serialization Graph engine.
func NewShadowGraphEngine() *ShadowGraphEngine {
	return &ShadowGraphEngine{
		graph:              analyzer.NewAdyaGraph(),
		activeTx:           make(map[string]*txState),
		itemReaders:        make(map[string]map[string]bool),
		itemLastWriter:     make(map[string]string),
		uncommittedWriters: make(map[string]string),
		abortedTx:          make(map[string]bool),
		detectedAnomalies:  make([]DetectedAnomaly, 0),
		reportedCycles:     make(map[string]bool),
		trace:              make(domain.ExecutionTrace, 0),
		startTime:          time.Now(),
	}
}

// SetAnomalyCallback sets a hook triggered when a new anomaly is detected.
func (e *ShadowGraphEngine) SetAnomalyCallback(cb func(DetectedAnomaly)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onAnomaly = cb
}

// StartTx begins tracking an active transaction for the session.
func (e *ShadowGraphEngine) StartTx(session *ProxySession, iso domain.IsolationLevel) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session.StartTransaction(iso)
	txID := session.TxID
	e.graph.AddNode(txID)

	e.activeTx[txID] = &txState{
		SessionID:  session.ID,
		TxID:       txID,
		Isolation:  session.Isolation,
		ReadSet:    make(map[string]struct{}),
		WriteSet:   make(map[string]struct{}),
		Statements: make([]string, 0),
	}

	e.recordTraceEvent(session.ID, txID, domain.EventBegin, "BEGIN", "")
}

// RecordRead registers read keys and discovers dependencies.
func (e *ShadowGraphEngine) RecordRead(session *ProxySession, items []string, sql string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	txID := session.TxID
	if txID == "" {
		txID = fmt.Sprintf("S%d-autocommit", session.ID)
		e.graph.AddNode(txID)
	}

	session.AddRead(items...)
	session.AddStatement(sql)

	if st, ok := e.activeTx[txID]; ok {
		for _, it := range items {
			st.ReadSet[it] = struct{}{}
		}
		st.Statements = append(st.Statements, sql)
	}

	for _, item := range items {
		table := getTableName(item)

		if e.itemReaders[item] == nil {
			e.itemReaders[item] = make(map[string]bool)
		}
		e.itemReaders[item][txID] = true

		if item != table {
			if e.itemReaders[table] == nil {
				e.itemReaders[table] = make(map[string]bool)
			}
			e.itemReaders[table][txID] = true
		}

		// Write-Read (WR) conflict from uncommitted active write
		if writerTx, exists := e.uncommittedWriters[item]; exists && writerTx != txID {
			e.graph.AddEdge(writerTx, txID, analyzer.DepWR, item)
		}

		// Write-Read (WR) dependency from committed write
		if lastWriter, exists := e.itemLastWriter[item]; exists && lastWriter != txID {
			e.graph.AddEdge(lastWriter, txID, analyzer.DepWR, item)
		}
	}

	e.recordTraceEvent(session.ID, txID, domain.EventExec, sql, "")
}

// RecordWrite registers written keys and identifies anti-dependencies (RW) and write conflicts (WW).
func (e *ShadowGraphEngine) RecordWrite(session *ProxySession, items []string, sql string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	txID := session.TxID
	if txID == "" {
		txID = fmt.Sprintf("S%d-autocommit", session.ID)
		e.graph.AddNode(txID)
	}

	session.AddWrite(items...)
	session.AddStatement(sql)

	if st, ok := e.activeTx[txID]; ok {
		for _, it := range items {
			st.WriteSet[it] = struct{}{}
		}
		st.Statements = append(st.Statements, sql)
	}

	for _, item := range items {
		table := getTableName(item)
		e.uncommittedWriters[item] = txID
		e.uncommittedWriters[table] = txID

		// Write-Write (WW) conflict
		hasWWWriter := false
		if lastWriter, exists := e.itemLastWriter[item]; exists && lastWriter != txID {
			e.graph.AddEdge(lastWriter, txID, analyzer.DepWW, item)
			hasWWWriter = true
		}
		if uncommitted, exists := e.uncommittedWriters[item]; exists && uncommitted != txID {
			e.graph.AddEdge(uncommitted, txID, analyzer.DepWW, item)
			hasWWWriter = true
		}

		// Read-Write (RW) anti-dependency: previous readers conflict with this write
		if readers, exists := e.itemReaders[item]; exists {
			for readerTx := range readers {
				if readerTx != txID {
					// If this reader is the exact previous writer on the same item, WW takes precedence
					if !hasWWWriter || readerTx != e.itemLastWriter[item] {
						e.graph.AddEdge(readerTx, txID, analyzer.DepRW, item)
					}
				}
			}
		}

		if item != table {
			if tableReaders, exists := e.itemReaders[table]; exists {
				for readerTx := range tableReaders {
					if readerTx != txID {
						if !hasWWWriter || readerTx != e.itemLastWriter[table] {
							e.graph.AddEdge(readerTx, txID, analyzer.DepRW, item)
						}
					}
				}
			}
		}
	}

	e.recordTraceEvent(session.ID, txID, domain.EventExec, sql, "")
}

// RecordCommit commits the transaction, checks for Adya cycles, and emits anomaly alerts.
func (e *ShadowGraphEngine) RecordCommit(session *ProxySession) []DetectedAnomaly {
	e.mu.Lock()
	defer e.mu.Unlock()

	txID := session.TxID
	if txID == "" {
		txID = fmt.Sprintf("S%d-autocommit", session.ID)
	}

	e.recordTraceEvent(session.ID, txID, domain.EventCommit, "COMMIT", "")

	// Promote uncommitted writes to committed
	if st, ok := e.activeTx[txID]; ok {
		for item := range st.WriteSet {
			e.itemLastWriter[item] = txID
			table := getTableName(item)
			e.itemLastWriter[table] = txID
			delete(e.uncommittedWriters, item)
			delete(e.uncommittedWriters, table)
		}
	}

	newAnomalies := e.detectCycles(session.ID, txID)

	session.ResetTransaction()
	delete(e.activeTx, txID)

	return newAnomalies
}

// RecordRollback records an abort and checks for dirty read (G1a) anomalies.
func (e *ShadowGraphEngine) RecordRollback(session *ProxySession) []DetectedAnomaly {
	e.mu.Lock()
	defer e.mu.Unlock()

	txID := session.TxID
	if txID == "" {
		txID = fmt.Sprintf("S%d-autocommit", session.ID)
	}

	e.abortedTx[txID] = true
	e.recordTraceEvent(session.ID, txID, domain.EventRollback, "ROLLBACK", "")

	var anomalies []DetectedAnomaly

	// If another transaction read an uncommitted value from this aborted transaction -> G1a Dirty Read
	if st, ok := e.activeTx[txID]; ok {
		for item := range st.WriteSet {
			delete(e.uncommittedWriters, item)
			table := getTableName(item)
			delete(e.uncommittedWriters, table)

			if readers, exists := e.itemReaders[item]; exists {
				for readerTx := range readers {
					if readerTx != txID && !e.abortedTx[readerTx] {
						cycleKey := fmt.Sprintf("G1A-%s-%s-%s", txID, readerTx, item)
						if !e.reportedCycles[cycleKey] {
							e.reportedCycles[cycleKey] = true
							anomaly := DetectedAnomaly{
								ID:          fmt.Sprintf("anomaly-%d", len(e.detectedAnomalies)+1),
								Type:        domain.AnomalyG1aDirtyRead,
								SessionIDs:  []uint64{session.ID},
								TxIDs:       []string{txID, readerTx},
								Items:       []string{item},
								DetectedAt:  time.Now(),
								Description: fmt.Sprintf("Transaction %s read uncommitted data for %s from aborted transaction %s", readerTx, item, txID),
							}
							e.detectedAnomalies = append(e.detectedAnomalies, anomaly)
							anomalies = append(anomalies, anomaly)
							e.emitAlert(anomaly)
						}
					}
				}
			}
		}
	}

	session.ResetTransaction()
	delete(e.activeTx, txID)

	return anomalies
}

func (e *ShadowGraphEngine) detectCycles(sessionID uint64, txID string) []DetectedAnomaly {
	cycles := analyzer.FindCycles(e.graph)
	var newAnomalies []DetectedAnomaly

	for _, cycle := range cycles {
		if len(cycle) == 0 {
			continue
		}

		key := cycleFingerprint(cycle)
		if e.reportedCycles[key] {
			continue
		}
		e.reportedCycles[key] = true

		anomalyType := analyzer.ClassifyCycle(cycle)
		txSet := make(map[string]bool)
		itemSet := make(map[string]bool)
		for _, edge := range cycle {
			txSet[edge.From] = true
			txSet[edge.To] = true
			itemSet[edge.Item] = true
		}

		txList := make([]string, 0, len(txSet))
		for tx := range txSet {
			txList = append(txList, tx)
		}
		itemList := make([]string, 0, len(itemSet))
		for it := range itemSet {
			itemList = append(itemList, it)
		}

		anomaly := DetectedAnomaly{
			ID:          fmt.Sprintf("anomaly-%d", len(e.detectedAnomalies)+1),
			Type:        anomalyType,
			SessionIDs:  []uint64{sessionID},
			TxIDs:       txList,
			Items:       itemList,
			Cycle:       cycle,
			DetectedAt:  time.Now(),
			Description: formatCycleDescription(anomalyType, cycle),
		}

		e.detectedAnomalies = append(e.detectedAnomalies, anomaly)
		newAnomalies = append(newAnomalies, anomaly)
		e.emitAlert(anomaly)
	}

	return newAnomalies
}

func (e *ShadowGraphEngine) emitAlert(a DetectedAnomaly) {
	fmt.Fprintf(os.Stderr, "\n[CHAOSSQL PROXY] 💥 Live Concurrency Anomaly Detected: %s\n", a.Type)
	fmt.Fprintf(os.Stderr, "  • Description: %s\n", a.Description)
	fmt.Fprintf(os.Stderr, "  • Transactions: %s\n", strings.Join(a.TxIDs, ", "))
	fmt.Fprintf(os.Stderr, "  • Items: %s\n\n", strings.Join(a.Items, ", "))

	if e.onAnomaly != nil {
		e.onAnomaly(a)
	}
}

func (e *ShadowGraphEngine) recordTraceEvent(sessionID uint64, txID string, evType domain.TraceEventType, sql, errStr string) {
	elapsed := time.Since(e.startTime)
	e.trace = append(e.trace, domain.TraceEvent{
		Timestamp: elapsed,
		WorkerID:  int(sessionID),
		OpName:    txID,
		Type:      evType,
		SQL:       sql,
		Error:     errStr,
	})
}

// GetTrace returns a copy of the execution trace log.
func (e *ShadowGraphEngine) GetTrace() domain.ExecutionTrace {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make(domain.ExecutionTrace, len(e.trace))
	copy(res, e.trace)
	return res
}

// GetAnomalies returns all detected anomalies.
func (e *ShadowGraphEngine) GetAnomalies() []DetectedAnomaly {
	e.mu.RLock()
	defer e.mu.RUnlock()
	res := make([]DetectedAnomaly, len(e.detectedAnomalies))
	copy(res, e.detectedAnomalies)
	return res
}

// GetGraph returns the internal Adya conflict graph.
func (e *ShadowGraphEngine) GetGraph() *analyzer.AdyaGraph {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.graph
}

func getTableName(item string) string {
	if idx := strings.IndexByte(item, ':'); idx != -1 {
		return item[:idx]
	}
	return item
}

func cycleFingerprint(c analyzer.Cycle) string {
	var sb strings.Builder
	for _, e := range c {
		sb.WriteString(fmt.Sprintf("%s->%s:%s:%s|", e.From, e.To, e.Type, e.Item))
	}
	return sb.String()
}

func formatCycleDescription(anomalyType domain.AnomalyType, cycle analyzer.Cycle) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s cycle: ", anomalyType))
	for i, edge := range cycle {
		if i > 0 {
			sb.WriteString(" -> ")
		}
		sb.WriteString(fmt.Sprintf("%s --(%s on %s)--> %s", edge.From, edge.Type, edge.Item, edge.To))
	}
	return sb.String()
}
