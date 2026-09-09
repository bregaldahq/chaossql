# Technical Architecture: GitHub PR Commenter Bot & Automated Code Review

- **Project**: ChaosSQL v1.3 Enterprise Developer Experience
- **Module**: `chaossql comment` & GitHub Actions Integration (`.github/actions/commenter`)
- **Status**: Internal Technical Specification

---

## 1. Overview & Developer Experience

While ChaosSQL outputs OASIS SARIF 2.1.0 reports for GitHub Code Scanning, developers rarely inspect the "Security" tab during routine pull request reviews.

The **GitHub PR Commenter Bot** delivers visual feedback directly in the Pull Request conversation:

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│  PR #42: Add concurrent withdrawal endpoint                                      │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  🤖 ChaosSQL Bot commented 2 minutes ago                                         │
│                                                                                  │
│  ### ⚠️ ChaosSQL Isolation Anomaly Detected                                      │
│                                                                                  │
│  A race condition violating serializability was discovered in scenario           │
│  `banking_withdraw.yaml`.                                                        │
│                                                                                  │
│  **Classification**: `P4 (Adya Lost Update)` • **Driver**: PostgreSQL (pgx)      │
│                                                                                  │
│  #### Direct Serialization Graph (DSG)                                           │
│  ```mermaid                                                                      │
│  graph LR                                                                        │
│    T1["Tx 1 (Worker 0)"] -- "rw (accounts.balance)" --> T2["Tx 2 (Worker 2)"]    │
│    T2 -- "ww (accounts.balance)" --> T1                                          │
│    style T1 fill:#1F1934,stroke:#F5C400,stroke-width:2px,color:#FCFBF8          │
│    style T2 fill:#1F1934,stroke:#EF4444,stroke-width:2px,color:#FCFBF8          │
│  ```                                                                             │
│                                                                                  │
│  #### 1-Minimal Causal Counterexample (ddmin reduced from 48 ops to 2 ops):     │
│  | Step | Worker | Operation | Dynamic Parameters |                              │
│  |---|---|---|---|                                                              │
│  | 1 | W0 | `SELECT balance FROM accounts WHERE id = 1;` | `bal1 = 1000` |       │
│  | 2 | W2 | `UPDATE accounts SET balance = 900 WHERE id = 1;` | `cur = 1000` |   │
│                                                                                  │
│  #### 💡 Suggested Production Fix:                                               │
│  ```suggestion                                                                   │
│  UPDATE accounts SET balance = balance - 100 WHERE id = 1 AND balance >= 100;    │
│  ```                                                                             │
│                                                                                  │
│  [🔍 Open Interactive Web Trace Visualizer](#)                                   │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Technical Architecture & Comment Lifecycle

### 2.1 "Sticky In-Place Update" Pattern (Zero Notification Noise)
To avoid cluttering the Pull Request discussion with repetitive comments on every author `git push`:
1. The bot queries existing PR comments via the GitHub REST API (`GET /repos/{owner}/{repo}/issues/{issue_number}/comments`).
2. Scans for the hidden HTML marker: `<!-- chaossql-audit-marker -->`.
3. **If previous comment exists**: Performs a `PATCH` request updating comment contents in-place, updating status from `FAIL` to `PASS` if the subsequent commit resolved the race condition.
4. **If no comment exists**: Performs a `POST` request creating the initial comment.

### 2.2 Automated Mermaid Diagram Generation
The `chaossql comment` command inspects the trace artifact and transforms the internal Direct Serialization Graph (`internal/analyzer/adya.go`) into native GitHub Flavored Markdown Mermaid syntax:

```go
func GenerateMermaidFromAdya(dsg *analyzer.SerializationGraph) string {
    var sb strings.Builder
    sb.WriteString("```mermaid\ngraph LR\n")
    for _, edge := range dsg.CycleEdges {
        sb.WriteString(fmt.Sprintf("  %s -- \"%s (%s)\" --> %s\n",
            edge.FromTx, edge.Type, edge.TargetKey, edge.ToTx))
    }
    sb.WriteString("```\n")
    return sb.String()
}
```

---

## 3. GitHub Actions Configuration

```yaml
# .github/workflows/concurrency-bot.yml
name: Concurrency Reviewer
on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  review:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
      contents: read
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run ChaosSQL
        id: chaos
        run: |
          go install github.com/bregaldahq/chaossql/cmd/chaossql@latest
          chaossql run tests/concurrency/chaos.yaml --export-json chaos_result.json || true

      - name: Publish PR Concurrency Review
        run: |
          chaossql comment \
            --report chaos_result.json \
            --repo ${{ github.repository }} \
            --pr ${{ github.event.pull_request.number }} \
            --token ${{ secrets.GITHUB_TOKEN }}
```

This workflow equips every development Pull Request with an automated **virtual database reliability reviewer**.
