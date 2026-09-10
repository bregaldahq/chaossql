package drivers

import (
	"database/sql"
	"fmt"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// TransactionOptions configures one operation transaction.
type TransactionOptions struct {
	Isolation domain.IsolationLevel
}

func resolveIsolation(driver string, requested, fallback domain.IsolationLevel, supported ...domain.IsolationLevel) (domain.IsolationLevel, error) {
	level := requested
	if level == "" {
		level = fallback
	}
	for _, candidate := range supported {
		if level == candidate {
			return level, nil
		}
	}
	return "", fmt.Errorf("isolation %q is not supported by %s", level, driver)
}

func toSQLIsolation(level domain.IsolationLevel) (sql.IsolationLevel, error) {
	switch level {
	case domain.LevelReadUncommitted:
		return sql.LevelReadUncommitted, nil
	case domain.LevelReadCommitted:
		return sql.LevelReadCommitted, nil
	case domain.LevelRepeatableRead:
		return sql.LevelRepeatableRead, nil
	case domain.LevelSerializable:
		return sql.LevelSerializable, nil
	default:
		return sql.LevelDefault, fmt.Errorf("unknown isolation level %q", level)
	}
}

func fromSQLIsolation(level sql.IsolationLevel, fallback domain.IsolationLevel) (domain.IsolationLevel, error) {
	switch level {
	case sql.LevelDefault:
		return fallback, nil
	case sql.LevelReadUncommitted:
		return domain.LevelReadUncommitted, nil
	case sql.LevelReadCommitted:
		return domain.LevelReadCommitted, nil
	case sql.LevelRepeatableRead:
		return domain.LevelRepeatableRead, nil
	case sql.LevelSerializable:
		return domain.LevelSerializable, nil
	default:
		return "", fmt.Errorf("unsupported database/sql isolation level %d", level)
	}
}
