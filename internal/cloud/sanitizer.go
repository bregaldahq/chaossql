package cloud

import (
	"fmt"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// ClassifyOpType classifies the high-level operation category (read, write, commit, etc.)
func ClassifyOpType(event domain.TraceEvent) string {
	switch event.Type {
	case domain.EventBegin:
		return "begin"
	case domain.EventCommit:
		return "commit"
	case domain.EventRollback, domain.EventRollbackTo:
		return "rollback"
	case domain.EventSavepoint:
		return "savepoint"
	case domain.EventError:
		return "error"
	}

	upper := strings.ToUpper(event.SQL)
	if strings.HasPrefix(upper, "SELECT") {
		return "read"
	}
	if strings.HasPrefix(upper, "UPDATE") || strings.HasPrefix(upper, "INSERT") || strings.HasPrefix(upper, "DELETE") {
		return "write"
	}
	return "exec"
}

// SanitizeTrace transforms raw trace events into structural metadata without
// SQL text or schema identifiers.
func SanitizeTrace(trace []domain.TraceEvent) []SanitizedTraceEvent {
	if len(trace) == 0 {
		return []SanitizedTraceEvent{}
	}

	sanitized := make([]SanitizedTraceEvent, len(trace))
	for i, ev := range trace {
		sanitized[i] = SanitizedTraceEvent{
			Worker:   fmt.Sprintf("T%d", ev.WorkerID),
			OpType:   ClassifyOpType(ev),
			Duration: ev.Timestamp.Microseconds(),
		}
	}
	return sanitized
}
