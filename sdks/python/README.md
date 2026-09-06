# ChaosSQL Python SDK (`chaossql`)

Deterministic Concurrency & Invariant Fuzzer for SQL Databases in Python.

## Installation

```bash
pip install chaossql
```

## Quickstart

```python
from chaossql import ChaosHarness

def test_banking_lost_update_prevented():
    harness = ChaosHarness(driver="sqlite", dsn=":memory:")

    result = (
        harness.with_schema("CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);")
               .with_seed("INSERT INTO accounts VALUES (1, 1000), (2, 1000);")
               .with_invariant(
                   name="total_wealth_conserved",
                   query="SELECT sum(balance) AS total FROM accounts;",
                   assertion="total == 2000"
               )
               .add_operation("transfer_1_to_2", [
                   "SELECT balance FROM accounts WHERE id = 1 -> cur",
                   "UPDATE accounts SET balance = {cur - 50} WHERE id = 1",
                   "UPDATE accounts SET balance = balance + 50 WHERE id = 2"
               ])
               .add_operation("transfer_2_to_1", [
                   "SELECT balance FROM accounts WHERE id = 2 -> cur",
                   "UPDATE accounts SET balance = {cur - 50} WHERE id = 2",
                   "UPDATE accounts SET balance = balance + 50 WHERE id = 1"
               ])
               .assert_no_anomalies(workers=4, iterations=50, seed=42)
    )
    assert result.all_invariants_satisfied
```

## Pytest Fixture

```python
import pytest

@pytest.mark.chaossql(workers=4, duration="5s", seed=100)
def test_inventory_under_concurrency(chaossql_runner):
    report = chaossql_runner.run_scenario("examples/inventory_oversell/chaos.yaml")
    assert report.is_clean
```
