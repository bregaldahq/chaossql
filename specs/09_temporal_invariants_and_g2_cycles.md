# Spec 09: Temporal Invariants, Multi-Transaction G2 Anti-Dependency Cycles & CI/CD Reporters

## 1. Temporal & Event Stream Invariants (`internal/evaluator/temporal.go`)
- Enables assertions over the chronological trace and intermediate database states:
  - `Monotonicity`: Asserts that an order of sequence IDs strictly increases.
  - `ConservationLaw`: Asserts that total quantity across accounts remains invariant throughout execution.
  - `NoIntermediateNegative`: Checks that no intermediate query observed negative stock or negative balances.

## 2. Multi-Transaction G2 Anti-Dependency Cycle Inference (`internal/analyzer/`)
- Formal Definition (Adya 1999):
  - Cycle of length $k \ge 2$ containing anti-dependency edges ($\xrightarrow{rw}$):
    $$T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} \dots \xrightarrow{rw} T_k \xrightarrow{rw} T_1$$
- Anomaly Classification: `AnomalyG2AntiDependency`.

## 3. JUnit XML & GitHub Actions Step Summary Reporters (`internal/reporter/`)
- `--export-junit <junit.xml>`: Emits standard JUnit test suite XML for Jenkins, GitLab CI, and GitHub Actions.
- `--export-summary <summary.md>`: Emits GitHub Actions Step Summary markdown with badge cards, invariant status, and $ddmin$ reduction.
