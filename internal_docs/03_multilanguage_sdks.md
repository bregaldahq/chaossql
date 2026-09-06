# Technical Architecture: Multi-Language SDKs (Python, TypeScript & Node.js)

- **Project**: ChaosSQL v1.3 / v1.4 Ecosystem Roadmap
- **Modules**: `chaossql-py` (PyPI) & `@chaossql/test` (npm)
- **Status**: Internal Technical Specification

---

## 1. Overview & Architectural Distribution Strategy

The vast majority of enterprise systems accessing relational databases and suffering from concurrency anomalies are written in **Python (Django, FastAPI, SQLAlchemy)** and **TypeScript/Node.js (NestJS, Express, Prisma, Drizzle)**.

To drive widespread developer adoption without requiring teams to write Go code, we provide native SDKs matching the ergonomics of the native Go SDK (`pkg/chaostest`).

### Distribution Strategy: "Embedded Engine Binary" Approach
Following the proven packaging model of **Prisma** and **esbuild**:
- The Python package (`pip install chaossql`) and npm package (`npm install @chaossql/test`) bundle lightweight, zero-CGO static ChaosSQL binaries for each target architecture (`darwin-arm64`, `darwin-x64`, `linux-x64`, `linux-arm64`, `windows-x64`).
- Communication between host language SDKs and the ChaosSQL engine occurs via **IPC with bidirectional JSON streaming over stdin/stdout**, guaranteeing complete memory isolation, zero interpreter hangs (bypassing the Python GIL), and absolute cross-platform stability.

---

## 2. Python SDK Specification (`chaossql-py`)

### 2.1 Installation & Requirements
- `pip install chaossql`
- Python 3.9+ support
- Native integration with `pytest` and `unittest`.

### 2.2 Fluent API
```python
from chaossql import ChaosHarness

def test_banking_lost_update_prevented():
    harness = ChaosHarness(driver="sqlite", dsn=":memory:")
    
    schema = """
    CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);
    """
    seed = """
    INSERT INTO accounts VALUES (1, 1000), (2, 1000);
    """

    harness.with_schema(schema) \
           .with_seed(seed) \
           .with_invariant(
               name="total_wealth_conserved",
               query="SELECT sum(balance) AS total FROM accounts;",
               assertion="total == 2000"
           ) \
           .add_operation("transfer_1_to_2", [
               "SELECT balance FROM accounts WHERE id = 1 -> cur",
               "UPDATE accounts SET balance = {cur - 50} WHERE id = 1",
               "UPDATE accounts SET balance = balance + 50 WHERE id = 2"
           ]) \
           .add_operation("transfer_2_to_1", [
               "SELECT balance FROM accounts WHERE id = 2 -> cur",
               "UPDATE accounts SET balance = {cur - 50} WHERE id = 2",
               "UPDATE accounts SET balance = balance + 50 WHERE id = 1"
           ])

    # Run with 4 concurrent workers across 50 deterministic iterations
    result = harness.assert_no_anomalies(workers=4, iterations=50, seed=42)
    assert result.all_invariants_satisfied
```

### 2.3 Native Pytest Fixture
```python
# conftest.py
import pytest
from chaossql.pytest_plugin import chaossql_fixture

@pytest.mark.chaossql(workers=4, duration="5s", seed=100)
def test_inventory_under_concurrency(chaossql_runner):
    report = chaossql_runner.run_scenario("tests/scenarios/inventory_oversell.yaml")
    assert report.is_clean
```

---

## 3. TypeScript / Node.js SDK Specification (`@chaossql/test`)

### 3.1 Installation & Requirements
- `npm install --save-dev @chaossql/test`
- Comprehensive TypeScript typings (`.d.ts`) compatible with TypeScript 5.0+
- Support for Jest, Vitest, Node Test Runner, and Playwright.

### 3.2 Fluent TypeScript API
```typescript
import { describe, it, expect } from 'vitest';
import { ChaosHarness } from '@chaossql/test';

describe('Concurrency Suite', () => {
  it('should detect write skew in hospital on-call roster', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite' });

    const result = await harness
      .withSchema(`
        CREATE TABLE doctors (id INT PRIMARY KEY, name TEXT, on_call INT);
        INSERT INTO doctors VALUES (1, 'Alice', 1), (2, 'Bob', 1);
      `)
      .withInvariant(
        'at_least_one_doctor_on_call',
        'SELECT count(*) as active FROM doctors WHERE on_call = 1;',
        'active >= 1'
      )
      .addOperation('alice_leaves', [
        'SELECT count(*) as cnt FROM doctors WHERE on_call = 1 -> active',
        'UPDATE doctors SET on_call = 0 WHERE id = 1 AND {active > 1}'
      ])
      .addOperation('bob_leaves', [
        'SELECT count(*) as cnt FROM doctors WHERE on_call = 1 -> active',
        'UPDATE doctors SET on_call = 0 WHERE id = 2 AND {active > 1}'
      ])
      .runAndShrink({ workers: 2, iterations: 20, seed: 1337 });

    // Expect anomaly A5B (Write Skew) to be isolated
    expect(result.anomalyDetected).toBe(true);
    expect(result.anomalyCode).toBe('A5B');
    expect(result.minimalOperations.length).toBe(2);
  });
});
```

---

## 4. Automatic Synthesis of Native Regression Tests

When a test fails in Python or TypeScript and $ddmin$ isolates a 2-operation counterexample, the SDK exposes:
`harness.export_standalone_repro("tests/generated_repro_test.py")` (or `.ts`).

The generated artifact is a pure Python (`asyncio`) or TypeScript (`Promise.all`) test script utilizing standard database drivers without external ChaosSQL runtime dependencies, allowing any developer to instantly reproduce and inspect the race condition within their native IDE environment.
