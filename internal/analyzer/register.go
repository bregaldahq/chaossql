package analyzer

import (
	"fmt"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// RegisterOpType represents the type of operation performed on a register.
type RegisterOpType string

const (
	OpRead     RegisterOpType = "READ"
	OpWrite    RegisterOpType = "WRITE"
	OpAppend   RegisterOpType = "APPEND"
	OpCommit   RegisterOpType = "COMMIT"
	OpRollback RegisterOpType = "ROLLBACK"
)

// RegisterEvent models a single register or list-append operation in a transaction history.
type RegisterEvent struct {
	TxID           string         `json:"tx_id"`
	Var            string         `json:"var"`
	Op             RegisterOpType `json:"op"`
	Val            string         `json:"val"`
	ReadFromTx     string         `json:"read_from_tx,omitempty"`
	MonotonicIndex int64          `json:"monotonic_index,omitempty"`
	Committed      bool           `json:"committed,omitempty"`
}

// RegisterViolation captures a specific consistency or linearizability anomaly.
type RegisterViolation struct {
	Type        domain.AnomalyType `json:"type"`
	TxID        string             `json:"tx_id"`
	OffendingTx string             `json:"offending_tx,omitempty"`
	Var         string             `json:"var"`
	Expected    string             `json:"expected,omitempty"`
	Actual      string             `json:"actual,omitempty"`
	Message     string             `json:"message"`
}

// RegisterAnalysisResult holds the outcome of the register linearizability analysis.
type RegisterAnalysisResult struct {
	Linearizable bool                `json:"linearizable"`
	Violations   []RegisterViolation `json:"violations,omitempty"`
}

// CheckRegisterLinearizability analyzes a history of register and list-append operations
// for consistency violations inspired by Kyle Kingsbury's Elle / Jepsen linearizability auditing,
// including G1b (Intermediate Reads) and Fractured Reads.
func CheckRegisterLinearizability(events []RegisterEvent) (*RegisterAnalysisResult, error) {
	if len(events) == 0 {
		return &RegisterAnalysisResult{
			Linearizable: true,
			Violations:   nil,
		}, nil
	}

	abortedTxs := make(map[string]bool)
	committedTxs := make(map[string]bool)

	// writesByTx maps variable -> txID -> list of values written in order
	writesByTx := make(map[string]map[string][]string)

	// appendsByTx maps variable -> txID -> list of elements appended in order
	appendsByTx := make(map[string]map[string][]string)

	// globalAppends maps variable -> all committed appended elements in order
	globalAppends := make(map[string][]string)

	// Validate ops and record tx outcomes and mutations
	for _, e := range events {
		op := RegisterOpType(strings.ToUpper(strings.TrimSpace(string(e.Op))))
		switch op {
		case OpRead, OpWrite, OpAppend, OpCommit, OpRollback:
			// valid
		default:
			return nil, fmt.Errorf("unknown register op type: %q", e.Op)
		}

		if op == OpRollback {
			abortedTxs[e.TxID] = true
			continue
		}
		if op == OpCommit || e.Committed {
			committedTxs[e.TxID] = true
		}

		if op == OpWrite {
			if writesByTx[e.Var] == nil {
				writesByTx[e.Var] = make(map[string][]string)
			}
			writesByTx[e.Var][e.TxID] = append(writesByTx[e.Var][e.TxID], e.Val)
		} else if op == OpAppend {
			if appendsByTx[e.Var] == nil {
				appendsByTx[e.Var] = make(map[string][]string)
			}
			elems := parseElements(e.Val)
			appendsByTx[e.Var][e.TxID] = append(appendsByTx[e.Var][e.TxID], elems...)
			globalAppends[e.Var] = append(globalAppends[e.Var], elems...)
		}
	}

	var violations []RegisterViolation

	// Check read operations for anomalies
	for _, e := range events {
		op := RegisterOpType(strings.ToUpper(strings.TrimSpace(string(e.Op))))
		if op != OpRead {
			continue
		}

		// 1. Check G1a Dirty Read: Reading from an aborted transaction
		if e.ReadFromTx != "" && abortedTxs[e.ReadFromTx] {
			violations = append(violations, RegisterViolation{
				Type:        domain.AnomalyG1aDirtyRead,
				TxID:        e.TxID,
				OffendingTx: e.ReadFromTx,
				Var:         e.Var,
				Actual:      e.Val,
				Message:     fmt.Sprintf("G1a Dirty Read: transaction %s read value %q of variable %s from aborted transaction %s", e.TxID, e.Val, e.Var, e.ReadFromTx),
			})
			continue
		}

		// 2. Check G1b Intermediate Read: Reading an overwritten uncommitted version
		if txWrites, ok := writesByTx[e.Var]; ok {
			if e.ReadFromTx != "" {
				writerWrites := txWrites[e.ReadFromTx]
				if len(writerWrites) > 1 {
					finalVal := writerWrites[len(writerWrites)-1]
					// If the read value is not the final value written by the transaction,
					// check if it matches an intermediate write of that transaction
					if e.Val != finalVal {
						for _, intermediate := range writerWrites[:len(writerWrites)-1] {
							if e.Val == intermediate {
								violations = append(violations, RegisterViolation{
									Type:        domain.AnomalyG1bIntermediateRead,
									TxID:        e.TxID,
									OffendingTx: e.ReadFromTx,
									Var:         e.Var,
									Expected:    finalVal,
									Actual:      e.Val,
									Message:     fmt.Sprintf("G1b Intermediate Read: transaction %s read intermediate version %q of variable %s from transaction %s (final committed version was %q)", e.TxID, e.Val, e.Var, e.ReadFromTx, finalVal),
								})
								break
							}
						}
					}
				}
			} else {
				// ReadFromTx not explicitly provided: detect if e.Val is an intermediate write of any transaction
				for txID, writerWrites := range txWrites {
					if len(writerWrites) > 1 {
						finalVal := writerWrites[len(writerWrites)-1]
						if e.Val != finalVal {
							for _, intermediate := range writerWrites[:len(writerWrites)-1] {
								if e.Val == intermediate {
									violations = append(violations, RegisterViolation{
										Type:        domain.AnomalyG1bIntermediateRead,
										TxID:        e.TxID,
										OffendingTx: txID,
										Var:         e.Var,
										Expected:    finalVal,
										Actual:      e.Val,
										Message:     fmt.Sprintf("G1b Intermediate Read: transaction %s read intermediate version %q of variable %s from transaction %s (final committed version was %q)", e.TxID, e.Val, e.Var, txID, finalVal),
									})
									break
								}
							}
						}
					}
				}
			}
		}

		// 3. Check Fractured Read for list-append registers
		observedElems := parseElements(e.Val)
		if len(observedElems) > 0 {
			observedSet := make(map[string]bool, len(observedElems))
			for _, elem := range observedElems {
				observedSet[elem] = true
			}

			// Check intra-transaction fractures first
			if txAppends, ok := appendsByTx[e.Var]; ok {
				for txID, elems := range txAppends {
					if len(elems) < 2 {
						continue
					}
					// For any pair (i < j), elem i was appended before elem j by txID
					for i := 0; i < len(elems); i++ {
						for j := i + 1; j < len(elems); j++ {
							earlier := elems[i]
							later := elems[j]
							// If later element was observed but earlier element was missed: Fractured Read!
							if observedSet[later] && !observedSet[earlier] {
								violations = append(violations, RegisterViolation{
									Type:        domain.AnomalyFracturedRead,
									TxID:        e.TxID,
									OffendingTx: txID,
									Var:         e.Var,
									Expected:    fmt.Sprintf("preceding element %q appended by %s", earlier, txID),
									Actual:      fmt.Sprintf("observed later element %q missing %q", later, earlier),
									Message:     fmt.Sprintf("Fractured Read on variable %s: transaction %s observed element %q but missed preceding element %q appended by transaction %s", e.Var, e.TxID, later, earlier, txID),
								})
							}
						}
					}
				}
			}

			// Check cross-transaction global append ordering
			allAppends := globalAppends[e.Var]
			if len(allAppends) >= 2 {
				for i := 0; i < len(allAppends); i++ {
					for j := i + 1; j < len(allAppends); j++ {
						earlier := allAppends[i]
						later := allAppends[j]
						if observedSet[later] && !observedSet[earlier] {
							// Avoid duplicate if already reported in intra-transaction
							alreadyReported := false
							for _, v := range violations {
								if v.Type == domain.AnomalyFracturedRead && v.TxID == e.TxID && v.Var == e.Var && v.Actual == fmt.Sprintf("observed later element %q missing %q", later, earlier) {
									alreadyReported = true
									break
								}
							}
							if !alreadyReported {
								violations = append(violations, RegisterViolation{
									Type:     domain.AnomalyFracturedRead,
									TxID:     e.TxID,
									Var:      e.Var,
									Expected: fmt.Sprintf("preceding element %q", earlier),
									Actual:   fmt.Sprintf("observed later element %q missing %q", later, earlier),
									Message:  fmt.Sprintf("Fractured Read on variable %s: transaction %s observed element %q but missed preceding element %q", e.Var, e.TxID, later, earlier),
								})
							}
						}
					}
				}
			}
		}
	}

	return &RegisterAnalysisResult{
		Linearizable: len(violations) == 0,
		Violations:   violations,
	}, nil
}

// parseElements extracts string elements from representations like "[A, B]", "A, B", "A", or "[]".
func parseElements(val string) []string {
	s := strings.TrimSpace(val)
	if s == "" || s == "[]" {
		return nil
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = strings.TrimSpace(s[1 : len(s)-1])
		if s == "" {
			return nil
		}
	}

	parts := strings.Split(s, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		item := strings.TrimSpace(p)
		item = strings.Trim(item, "\"'")
		if item != "" {
			res = append(res, item)
		}
	}
	return res
}
