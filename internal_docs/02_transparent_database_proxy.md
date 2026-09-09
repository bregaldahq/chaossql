# Technical Architecture: Transparent Proxy Mode (`chaossql proxy`)

- **Project**: ChaosSQL v1.4 Enterprise Roadmap
- **Module**: Layer-7 TCP Database Reverse Proxy (`pkg/proxy`, `cmd/chaossql/proxy.go`)
- **Status**: Internal Technical Specification
- **Supported Protocols**: PostgreSQL Wire Protocol 3.0 & MySQL Client/Server Protocol

---

## 1. Overview & Commercial Value Proposition

The primary friction point for enterprise adoption of ChaosSQL is the manual authoring of scenario definitions in `chaos.yaml`. Investment banks, e-commerce platforms, and SaaS systems already possess **thousands of automated integration tests** written in Django, Spring Boot, Ruby on Rails, NestJS, or Go.

**Transparent Proxy Mode** transforms ChaosSQL into a **Layer-7 network middleware**:

```
┌───────────────────────────────┐
│ Application Test Suite        │
│ (Django / Spring / Node / Go) │
└───────────────┬───────────────┘
                │  JDBC / pgx / psycopg Connection (Port 5433)
                ▼
┌────────────────────────────────────────────────────────────────────────┐
│ ChaosSQL L7 Transparent Proxy (:5433)                                  │
│                                                                        │
│  ┌────────────────────────┐  ┌──────────────────────────────────────┐  │
│  │ Wire Protocol Decoder  │  │ PCT & Jitter Delay Scheduler         │  │
│  │ (Postgres 3.0 / MySQL) │  │ • Microsecond Socket Stall           │  │
│  └───────────┬────────────┘  │ • Barrier Synchronization            │  │
│              │               └──────────────────┬───────────────────┘  │
│              ▼                                  ▼                      │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Live Shadow Serialization Graph Engine                           │  │
│  │ • Table/Row Read-Sets (RS) & Write-Sets (WS)                     │  │
│  │ • Dynamic Adya Conflict Graph ($wr, ww, rw$)                     │  │
│  │ • Transaction Isolation Cycle Detector                           │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────┬───────────────────────────────────────┘
                                 │
                                 │ Forwarded TCP with Jitter (Port 5432)
                                 ▼
                 ┌───────────────────────────────┐
                 │ PostgreSQL / MySQL Upstream   │
                 └───────────────────────────────┘
```

The target application simply points its connection string port to the proxy (`DATABASE_URL=postgres://app:pass@127.0.0.1:5433/testdb`). ChaosSQL intercepts wire traffic at runtime, injects stochastic jitter, and detects concurrency race conditions in real production application code without changing a single line of application source.

---

## 2. Layer-7 Wire Protocol Decoding

### 2.1 PostgreSQL Wire Protocol 3.0
The proxy intercepts and decodes packets conforming to the PostgreSQL canonical wire format (`[Byte: Type][Int32: Length][Payload]`):

- **Intercepted Client Messages (`Frontend`)**:
  - `'Q'` (*Simple Query*): Intercepts plaintext SQL commands (`BEGIN`, `COMMIT`, direct SQL statements).
  - `'P'` (*Parse*) / `'B'` (*Bind*) / `'E'` (*Execute*): Decodes prepared statements and parameterized values.
  - `'X'` (*Terminate*): Clean client connection termination.
- **Intercepted Server Messages (`Backend`)**:
  - `'T'` (*RowDescription*) / `'D'` (*DataRow*): Extracts column metadata and returned data for read-set inference.
  - `'C'` (*CommandComplete*): Extracts affected row count (`rows_affected`).
  - `'Z'` (*ReadyForQuery*): Marks formal transaction completion and transaction status (`I` = Idle, `T` = In Transaction, `E` = In Failed Transaction).

### 2.2 MySQL Client/Server Protocol
- **Intercepted Commands**:
  - `0x03` (`COM_QUERY`): Plaintext SQL query strings.
  - `0x16` (`COM_STMT_PREPARE`) and `0x17` (`COM_STMT_EXECUTE`): Binary prepared statements.
  - `OK_Packet` / `ERR_Packet`: Validates commit vs. rollback and deadlock error codes (`1213`).

---

## 3. Jitter Injection Engine & Socket-Level PCT Scheduling

The proxy maintains two socket channels (`net.Conn`) per client session: `clientConn` and `serverConn`.

```go
type ProxySession struct {
    ID         uint64
    ClientConn net.Conn
    ServerConn net.Conn
    InTx       bool
    Isolation  domain.IsolationLevel
    TxID       string
    ReadSet    map[string]struct{}
    WriteSet   map[string]struct{}
}

func (s *ProxySession) HandlePacket(pkt []byte) ([]byte, error) {
    opType, sqlText := parseQuery(pkt)

    // Inject stochastic micro-jitter if executing critical DML inside a transaction
    if s.InTx && isDMLOrSelect(opType) {
        delay := calculatePCTJitter(s.ID)
        time.Sleep(delay) // or cooperative goroutine scheduling
    }

    return pkt, nil
}
```

- **Commit Barrier Injection**:
  The peak window for $P4$ (Lost Update) or $A5B$ (Write Skew) occurs between the final `UPDATE` and the dispatch of the `COMMIT` packet. The proxy injects a deliberate delay of $50\mu\text{s}$ to $5\text{ms}$ immediately prior to forwarding the `COMMIT` packet for a session, enabling concurrent sessions to interleave pending writes.

---

## 4. Live Shadow Serialization Graph (Shadow DSG)

The proxy does not require explicit declarative invariants to detect the vast majority of critical anomalies:

1. The lexical parser extracts table names and primary keys from SQL statements (`FROM accounts WHERE id = 1` $\to$ Key: `accounts:1`).
2. Builds conflict edges in the Direct Serialization Graph (Adya):
   - If Session $S_1$ read `accounts:1` and Session $S_2$ committed a write to `accounts:1` before $S_1$ commits $\to$ anti-dependency edge $S_1 \xrightarrow{rw} S_2$.
   - If $S_2$ also read data previously written by $S_1$ $\to$ cycle detected ($S_1 \xrightarrow{rw} S_2 \xrightarrow{rw} S_1$).
3. Upon detecting a runtime dependency cycle:
   - Records the anomaly event into the execution trace.
   - Emits a real-time log alert and generates an **OASIS SARIF 2.1.0** report upon proxy session termination.

---

## 5. Command-Line Interface (CLI)

```bash
# Start transparent proxy for PostgreSQL
chaossql proxy \
  --listen 127.0.0.1:5433 \
  --upstream 127.0.0.1:5432 \
  --protocol postgres \
  --jitter-min 10us \
  --jitter-max 2ms \
  --pct-depth 2 \
  --export-sarif proxy-anomalies.sarif \
  --ui-port 8095
```

When application tests conclude (`Ctrl+C` or termination signal via CI), ChaosSQL prints a summary of captured anomalies and generates a SARIF artifact compatible with GitHub Code Scanning.
