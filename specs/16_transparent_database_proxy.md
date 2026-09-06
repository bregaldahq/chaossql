# Spec 16: Layer-7 Transparent Database Reverse Proxy (`chaossql proxy`) (v1.4)

## 1. Domain Theory & Motivation
- **Context**: Enterprise test suites possess existing automated integration test suites without requiring manual authoring of scenario YAML files.
- **Transparent Middleware Architecture**:
  - Acts as a Layer-7 TCP reverse proxy between the application test suite and upstream database engines (PostgreSQL and MySQL).
  - Eliminates the need to rewrite application code or modify testing harnesses: tests merely point their database port to the proxy.
  - Intercepts wire protocol traffic, injects stochastic microsecond socket delays (PCT) and deliberate commit barriers ($50\mu\text{s}$ to $5\text{ms}$), and tracks read/write conflict edges in real time.

## 2. Layer-7 Wire Protocol Decoders (`pkg/proxy/`)
- **PostgreSQL Wire Protocol 3.0**:
  - Handles client StartupMessage and SSLRequest handshakes without requiring external TLS termination.
  - Decodes Simple Query (`'Q'`), Parse (`'P'`), Bind (`'B'`), Execute (`'E'`), and Terminate (`'X'`) messages.
  - Decodes backend RowDescription (`'T'`), DataRow (`'D'`), CommandComplete (`'C'`), and ReadyForQuery (`'Z'`) transaction state indicators (`'I'`, `'T'`, `'E'`).
- **MySQL Client/Server Protocol**:
  - Decodes 4-byte packet framing (3-byte length + sequence ID).
  - Intercepts commands: `0x03` (`COM_QUERY`), `0x16` (`COM_STMT_PREPARE`), `0x17` (`COM_STMT_EXECUTE`), and `0x01` (`COM_QUIT`).
  - Evaluates `OK_Packet` (`SERVER_STATUS_IN_TRANS`) and `ERR_Packet` (deadlock code `1213`).

## 3. Stochastic PCT Jitter & Commit Barrier Injector (`pkg/proxy/jitter.go`)
- **Probabilistic Concurrency Testing (PCT)**:
  - Injects stochastic microsecond delays into active transaction DML statements according to configured priority depths.
- **Commit Barrier Injection**:
  - Deliberately delays the dispatch of `COMMIT` packets ($50\mu\text{s}$ to $5\text{ms}$) to expand the concurrency race window between uncommitted writes and commit serialization.

## 4. Live Shadow Serialization Graph (Shadow DSG) (`pkg/proxy/shadow_graph.go`)
- Maintains real-time Read-Sets (RS) and Write-Sets (WS) per client connection session.
- Dynamically derives Adya dependency conflict edges:
  - Write-Read ($wr$), Write-Write ($ww$), and Read-Write anti-dependencies ($rw$).
- Detects and classifies concurrency isolation anomaly cycles without declarative invariants:
  - Lost Update ($P4$), Write Skew ($A5B$), Dirty Read ($G1a$), and Anti-Dependency Cycles ($G2$).
- Real-time terminal log alerts emitted upon cycle discovery.

## 5. CLI & Diagnostic Reporting (`cmd/chaossql/proxy.go`, `pkg/proxy/sarif.go`, `pkg/proxy/ui.go`)
- Command: `chaossql proxy --listen <host:port> --upstream <host:port> --protocol <postgres|mysql>`.
- Export: Generates OASIS SARIF 2.1.0 vulnerability reports compatible with GitHub Code Scanning.
- Web Visualizer: Optional live HTTP dashboard (`--ui-port <port>`) displaying live session telemetry, query counters, and interactive Adya anomaly graphs.
