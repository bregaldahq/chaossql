package drivers

import (
	"strings"
)

// MaskDSN sanitizes a database connection string, redacting plain-text passwords.
func MaskDSN(dsn string) string {
	if dsn == "" || dsn == ":memory:" {
		return dsn
	}

	// URL format (e.g. postgres://user:pass@host:port/db)
	if strings.Contains(dsn, "://") {
		parts := strings.SplitN(dsn, "://", 2)
		scheme := parts[0]
		rest := parts[1]
		if atIdx := strings.LastIndex(rest, "@"); atIdx != -1 {
			userPass := rest[:atIdx]
			hostPart := rest[atIdx:]
			if colIdx := strings.Index(userPass, ":"); colIdx != -1 {
				user := userPass[:colIdx]
				return scheme + "://" + user + ":******" + hostPart
			}
		}
		return dsn
	}

	// Standard MySQL DSN format (e.g. user:pass@tcp(host:port)/db)
	if atIdx := strings.LastIndex(dsn, "@"); atIdx != -1 {
		userPass := dsn[:atIdx]
		hostPart := dsn[atIdx:]
		if colIdx := strings.Index(userPass, ":"); colIdx != -1 {
			user := userPass[:colIdx]
			return user + ":******" + hostPart
		}
	}

	return dsn
}
