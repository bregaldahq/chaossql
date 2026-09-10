import os
import subprocess
import sys
import tempfile
import pytest
from chaossql import ChaosHarness


def test_passing_invariant():
    harness = ChaosHarness(driver="sqlite", dsn=":memory:")
    schema = "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);"
    seed = "INSERT INTO accounts VALUES (1, 1000);"

    res = (
        harness.with_schema(schema)
        .with_seed(seed)
        .with_invariant(
            name="balance_conserved",
            query="SELECT balance FROM accounts WHERE id = 1;",
            assertion="balance == 1000",
        )
        .add_operation("read_balance", [
            "SELECT balance FROM accounts WHERE id = 1",
        ])
        .assert_no_anomalies(workers=2, iterations=5, seed=42)
    )

    assert res.all_invariants_satisfied
    assert res.is_clean
    assert not res.violation_detected


def test_invariant_violation_is_reported_and_shrunk():
    harness = ChaosHarness(driver="sqlite", dsn=":memory:")

    harness.with_schema(
        "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);"
    ).with_seed(
        "INSERT INTO accounts VALUES (1, 1000);"
    ).with_invariant(
        name="expected_test_balance",
        query="SELECT balance FROM accounts WHERE id = 1;",
        assertion="balance == 1000",
    ).add_operation("change_balance", [
        "UPDATE accounts SET balance = 999 WHERE id = 1",
    ])

    res = harness.run(workers=1, iterations=1, seed=42)
    assert res.anomaly_detected
    assert len(res.minimal_operations) == 1
    assert "sequenceDiagram" in res.mermaid

    with pytest.raises(AssertionError) as exc_info:
        harness.assert_no_anomalies(workers=1, iterations=1, seed=42)

    msg = str(exc_info.value)
    assert "expected_test_balance" in msg


def test_export_standalone_repro():
    harness = ChaosHarness(driver="sqlite", dsn=":memory:")
    schema = "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);"
    seed = "INSERT INTO accounts VALUES (1, 1000);"

    harness.with_schema(schema) \
           .with_seed(seed) \
           .with_invariant(
               name="expected_balance",
               query="SELECT balance FROM accounts WHERE id = 1;",
               assertion="balance == 800",
           ) \
           .add_operation("withdraw", [
               "SELECT balance FROM accounts WHERE id = 1 -> cur",
               "UPDATE accounts SET balance = {cur - 100} WHERE id = 1",
           ])

    res = harness.run(workers=2, iterations=2, seed=42)
    assert res.violation_detected

    with tempfile.TemporaryDirectory() as tmp_dir:
        repro_path = os.path.join(tmp_dir, "repro_test.py")
        exported = res.export_standalone_repro(repro_path)
        assert os.path.isfile(exported)

        # Run generated python file to ensure it executes cleanly
        proc = subprocess.run(
            [sys.executable, repro_path],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        # In standalone reproducer, exit code 0 indicates anomaly was successfully reproduced
        assert proc.returncode == 0, f"Repro script failed: {proc.stderr}\nCode:\n{open(repro_path).read()}"
        assert "REPRODUCED" in proc.stdout or "Success" in proc.stdout
