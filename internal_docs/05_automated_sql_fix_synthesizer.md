# Technical Architecture: Automated SQL Fix Synthesizer (`chaossql fix`)

- **Project**: ChaosSQL v1.4 Self-Healing Engine
- **Module**: `cmd/chaossql/fix.go` & `internal/synthesizer/`
- **Status**: Internal Technical Specification

---

## 1. Overview & Motivation

Detecting a concurrency race condition in a database is only half the engineering challenge. Most application developers are not specialists in formal concurrency theory or relational isolation levels, and manual remediation often introduces fresh anomalies or catastrophic deadlocks.

The **`chaossql fix`** command operates as an automated SQL remediating compiler:
1. Analyzes the formal anomaly detected by the Adya Direct Serialization Graph.
2. Identifies conflicting transactions and statements via the minimal $ddmin$ counterexample.
3. Applies AST (Abstract Syntax Tree) transformation heuristics to the SQL statements or transaction isolation configuration.
4. **Executes a closed stochastic verification benchmark** with 100 deterministic iterations to mathematically prove the anomaly has been eradicated before suggesting the patch.

```
┌─────────────────────────┐     ┌───────────────────────┐     ┌────────────────────────┐
│ Failing Scenario        │ ──► │ AST Heuristics        │ ──► │ Fix Candidate          │
│ (e.g., P4 Lost Update)  │     │ (SQL Transformer)     │     │ (Atomic / Lock / SSI)  │
└─────────────────────────┘     └───────────────────────┘     └───────────┬────────────┘
                                                                          │
                                ┌───────────────────────┐                 │
                                │ Closed Verification   │ ◄───────────────┘
                                │ (100-seed fuzzer)     │
                                └───────────┬───────────┘
                                            │
                                            ▼
                                ┌───────────────────────┐
                                │ Certified Patch       │
                                │ (Unified Git Diff)    │
                                └───────────────────────┘
```

---

## 2. Transformation Catalog by Anomaly

| Code | Anomaly | Syntactic Pattern Detected | Automated Transformation Strategy |
| :--- | :--- | :--- | :--- |
| **$P4$** | **Lost Update** | `SELECT val FROM tbl WHERE id = X -> v`<br>`UPDATE tbl SET val = {v - delta}` | **Strategy 1 (Atomic In-Place Update)**:<br>`UPDATE tbl SET val = val - delta WHERE id = X AND val >= delta;`<br><br>**Strategy 2 (Pessimistic Lock)**:<br>`SELECT val FROM tbl WHERE id = X FOR UPDATE;` |
| **$A3$** | **Oversell** | Stock read followed by decrement without guard predicate. | **Unconditional Guard Predicate**:<br>`WHERE id = X AND stock >= delta`<br>+ assertion `rows_affected == 1`. |
| **$A5B$** | **Write Skew** | Disjoint reads with complementary writes under Snapshot Isolation. | **Strategy 1 (Isolation Elevation)**:<br>`SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;`<br><br>**Strategy 2 (Master Key Lock)**:<br>Inject `SELECT 1 FROM parent_lock WHERE id = X FOR UPDATE;` before evaluating invariant. |
| **$G0$** | **Dirty Write** | Multiple statements updating interdependent columns without versioning. | **OCC Injection (Optimistic Concurrency Control)**:<br>`UPDATE tbl SET val = new, version = version + 1 WHERE id = X AND version = old_ver;` |
| **$G1a$** | **Dirty Read** | Transactions reading uncommitted data under `READ UNCOMMITTED`. | **Minimum Isolation Floor**:<br>Rewrite connection configuration to enforce `READ COMMITTED` as minimum tolerable isolation level. |
| **$G\text{-DL}$** | **Deadlock** | Crossed locks acquired in inverse orders ($T_1$: A $\to$ B; $T_2$: B $\to$ A). | **Canonical Lock Ordering**:<br>Reorder SQL invocations to acquire locks in ascending primary key order (`ORDER BY id ASC`). |

---

## 3. Closed Verification Loop Algorithm

```go
// internal/synthesizer/engine.go
func (s *Synthesizer) SynthesizeAndVerify(ctx context.Context, sc *domain.Scenario) (*domain.FixResult, error) {
    // 1. Identify anomaly
    diag := s.analyzer.Analyze(sc.LastTrace)
    
    // 2. Select applicable transformation strategies
    candidates := s.ruleEngine.GenerateCandidates(sc, diag)
    
    for _, candidate := range candidates {
        // 3. Execute fuzzer benchmark across 100 deterministic seeds
        verifiedClean := true
        for seed := int64(1); seed <= 100; seed++ {
            trace, err := s.fuzzer.RunWithCandidate(ctx, candidate, seed)
            if err != nil || s.analyzer.HasCycle(trace) {
                verifiedClean = false
                break // Candidate failed verification
            }
        }
        
        // 4. If 100% clean, emit certified unified diff
        if verifiedClean {
            return &domain.FixResult{
                Success:      true,
                StrategyName: candidate.Name,
                UnifiedDiff:  s.generateDiff(sc.OriginalYAML, candidate.YAML),
                Explanation:  candidate.Explanation,
            }, nil
        }
    }
    
    return nil, errors.New("unable to certify automated fix with 100% confidence")
}
```

---

## 4. Command-Line Interface (CLI) Usage

```bash
# Analyze and suggest remediation for scenario
chaossql fix examples/banking_lost_update/chaos.yaml

# Expected Output:
# [✓] Diagnosed anomaly: P4 (Adya Lost Update)
# [✓] Testing strategy: In-Place Atomic Expression
# [✓] Verification complete: 100/100 clean seeds (0 invariant violations)
# 
# --- a/chaos.yaml
# +++ b/chaos.yaml
# @@ -15,4 +15,3 @@
# -      - "SELECT balance FROM accounts WHERE id = 1 -> bal1;"
# -      - "UPDATE accounts SET balance = {bal1 - 50} WHERE id = 1;"
# +      - "UPDATE accounts SET balance = balance - 50 WHERE id = 1 AND balance >= 50;"
# 
# Pass --apply to persist changes directly to disk.
```

This workflow closes the complete database reliability cycle: **Discover $\to$ Isolate ($ddmin$) $\to$ Explain (Adya DSG) $\to$ Fix and Certify (`chaossql fix`)**.
