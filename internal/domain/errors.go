package domain

import "errors"

var (
	ErrInvariantViolated     = errors.New("invariant violated")
	ErrDatabaseDriverFailed  = errors.New("database driver operation failed")
	ErrSpecValidationFailed  = errors.New("chaos spec validation failed")
	ErrNoFailingTraceFound   = errors.New("no invariant violation detected in execution")
	ErrSerializationFailure  = errors.New("database transaction serialization failure (SQLSTATE 40001)")
	ErrDeadlockDetected      = errors.New("database deadlock detected (SQLSTATE 40P01)")
)
