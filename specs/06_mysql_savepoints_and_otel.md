# Spec 06: MySQL Driver, Savepoints & OpenTelemetry Distributed Tracing

## 1. MySQL Driver Architecture
- Driver Name: `mysql`
- Pure Go implementation via `github.com/go-sql-driver/mysql`.
- Dynamic isolation level support: `READ UNCOMMITTED`, `READ COMMITTED`, `REPEATABLE READ`, `SERIALIZABLE`.
- Connection pooling with automatic reconnection and deadlock (`Error 1213: ER_LOCK_DEADLOCK`) error mapping.

## 2. Savepoints & Partial Rollback
- Syntax: `SAVEPOINT <id>`, `ROLLBACK TO <id>`, `RELEASE SAVEPOINT <id>`.
- Every savepoint statement executes through the operation's active transaction; it must never use an autocommit connection.
- A partial rollback may undo sub-operations without ending the transaction. The operation still performs one final driver commit or rollback.
- MySQL transaction fixtures must verify the effective isolation level and prove that a successful operation commits while a later statement error rolls back all earlier writes in that operation.

## 3. OpenTelemetry Distributed Tracing Format
- Format: OpenTelemetry JSON Trace Spec (OTLP HTTP/JSON span format).
- Spans:
  - Root Span: Chaos Scenario Execution.
  - Worker Spans: Worker goroutines.
  - Transaction Spans: Individual database transactions.
  - Statement Spans: Individual SQL queries with execution duration and DB state.
