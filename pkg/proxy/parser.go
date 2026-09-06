package proxy

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
)

var (
	reInsertID = regexp.MustCompile(`(?i)INSERT\s+INTO\s+([a-zA-Z0-9_]+)\s*\([^)]*?\bid\b[^)]*?\)\s*VALUES\s*\(\s*(\d+)`)
	reWhereID  = regexp.MustCompile(`(?i)\bWHERE\b.*?\bid\s*=\s*(\d+)`)
)

// InspectSQL analyzes an intercepted SQL string and determines:
// 1. isWrite: true if UPDATE, INSERT, DELETE
// 2. items: list of accessed logical keys (e.g. "accounts:1" or "doctors")
// 3. isBoundary: true if BEGIN, COMMIT, ROLLBACK, or SET ISOLATION
// 4. boundaryType: "BEGIN", "COMMIT", "ROLLBACK", "SET_ISOLATION", etc.
// 5. isolation: parsed IsolationLevel if SET TRANSACTION
func InspectSQL(rawSQL string) (isWrite bool, items []string, isBoundary bool, boundaryType string, isolation domain.IsolationLevel) {
	trimmed := strings.TrimSpace(rawSQL)
	trimmed = strings.TrimRight(trimmed, ";")
	trimmed = strings.TrimSpace(trimmed)
	upper := strings.ToUpper(trimmed)

	// Check transaction boundary statements
	switch {
	case upper == "BEGIN" || upper == "BEGIN TRANSACTION" || upper == "START TRANSACTION":
		return false, nil, true, "BEGIN", ""
	case upper == "COMMIT" || upper == "END":
		return false, nil, true, "COMMIT", ""
	case upper == "ROLLBACK" || upper == "ABORT":
		return false, nil, true, "ROLLBACK", ""
	case strings.HasPrefix(upper, "SET TRANSACTION ISOLATION LEVEL") || strings.HasPrefix(upper, "SET SESSION CHARACTERISTICS AS TRANSACTION ISOLATION LEVEL"):
		iso := parseIsolationLevel(upper)
		return false, nil, true, "SET_ISOLATION", iso
	}

	// Extract accessed items and read/write characteristics
	isWrite, item := extractItemFromSQL(trimmed, upper)
	if item != "" {
		items = append(items, item)
	}

	return isWrite, items, false, "", ""
}

// IsDMLOrSelect returns true for data access/modification statements.
func IsDMLOrSelect(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	return strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "UPDATE") ||
		strings.HasPrefix(upper, "INSERT") ||
		strings.HasPrefix(upper, "DELETE")
}

func parseIsolationLevel(upper string) domain.IsolationLevel {
	switch {
	case strings.Contains(upper, "SERIALIZABLE"):
		return domain.LevelSerializable
	case strings.Contains(upper, "REPEATABLE READ"):
		return domain.LevelRepeatableRead
	case strings.Contains(upper, "READ COMMITTED"):
		return domain.LevelReadCommitted
	case strings.Contains(upper, "READ UNCOMMITTED"):
		return domain.LevelReadUncommitted
	default:
		return domain.LevelReadCommitted
	}
}

func extractItemFromSQL(sql string, upper string) (isWrite bool, item string) {
	var table string
	var id string

	switch {
	case strings.HasPrefix(upper, "UPDATE "):
		isWrite = true
		rest := strings.TrimSpace(sql[7:])
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			table = fields[0]
		}
		if m := reWhereID.FindStringSubmatch(sql); len(m) > 1 {
			id = m[1]
		}

	case strings.HasPrefix(upper, "INSERT INTO "):
		isWrite = true
		if m := reInsertID.FindStringSubmatch(sql); len(m) > 2 {
			return true, m[1] + ":" + m[2]
		}
		rest := strings.TrimSpace(sql[12:])
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			table = strings.TrimRight(fields[0], " (")
		}

	case strings.HasPrefix(upper, "DELETE FROM "):
		isWrite = true
		rest := strings.TrimSpace(sql[12:])
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			table = fields[0]
		}
		if m := reWhereID.FindStringSubmatch(sql); len(m) > 1 {
			id = m[1]
		}

	case strings.HasPrefix(upper, "SELECT "):
		isWrite = false
		if fromIdx := strings.Index(upper, " FROM "); fromIdx != -1 {
			rest := strings.TrimSpace(sql[fromIdx+6:])
			fields := strings.Fields(rest)
			if len(fields) > 0 {
				table = strings.TrimRight(fields[0], ";,")
			}
			if m := reWhereID.FindStringSubmatch(sql); len(m) > 1 {
				id = m[1]
			}
		}
	}

	if table != "" {
		if id != "" {
			if _, err := strconv.Atoi(id); err == nil {
				return isWrite, table + ":" + id
			}
		}
		return isWrite, table
	}

	return false, ""
}
