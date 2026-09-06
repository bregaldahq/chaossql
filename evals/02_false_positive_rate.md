# Eval 02: False Positive Rate ($0\%$)

## Objective
Guarantee that ChaosSQL never flags a false positive invariant violation on systems correctly protected by pessimistic locking (`FOR UPDATE`) or strict `SERIALIZABLE` isolation.

## Acceptance Criteria
1. **False Positive Rate:** Across 100 test benchmark iterations running corrected SQL, exactly 0 violations must be reported.
2. **Driver Error Handling & Isolation:** Serialization errors (e.g., PostgreSQL `40001`) and deadlock detection codes (`40P01`) must be caught, aborted, and accounted for without treating expected driver-level rollbacks as invariant violations.
