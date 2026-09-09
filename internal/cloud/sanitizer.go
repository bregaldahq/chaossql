package cloud

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
)

var (
	tableRegex    = regexp.MustCompile(`(?i)(?:FROM|UPDATE|INTO|TABLE)\s+([a-zA-Z0-9_\.]+)`)
	passwordRegex = regexp.MustCompile(`(?i)(password|passwd|secret|api_key|token)\s*=\s*['"][^'"]*['"]`)
	bearerRegex   = regexp.MustCompile(`(?i)Bearer\s+[a-zA-Z0-9_\.\-]+`)
)

// SanitizeSQL strips credentials, tokens, and sensitive literals from SQL strings
func SanitizeSQL(sql string) string {
	cleaned := strings.TrimSpace(sql)
	cleaned = passwordRegex.ReplaceAllString(cleaned, `$1 = '[REDACTED]'`)
	cleaned = bearerRegex.ReplaceAllString(cleaned, `Bearer [REDACTED]`)
	return cleaned
}

// ExtractTable infers the primary table name from a SQL query
func ExtractTable(sql string) string {
	matches := tableRegex.FindStringSubmatch(sql)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

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

// SanitizeTrace transforms raw domain trace events into sanitized, cloud-safe events
func SanitizeTrace(trace []domain.TraceEvent) []SanitizedTraceEvent {
	if len(trace) == 0 {
		return []SanitizedTraceEvent{}
	}

	sanitized := make([]SanitizedTraceEvent, len(trace))
	for i, ev := range trace {
		sanitized[i] = SanitizedTraceEvent{
			Worker:   fmt.Sprintf("T%d", ev.WorkerID),
			OpType:   ClassifyOpType(ev),
			Table:    ExtractTable(ev.SQL),
			SQL:      SanitizeSQL(ev.SQL),
			Duration: ev.Timestamp.Microseconds(),
		}
	}
	return sanitized
}
