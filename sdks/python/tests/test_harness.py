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


def test_banking_lost_update_detected_and_shrunk():
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

    # Expect assert_no_anomalies to fail with AssertionError
    with pytest.raises(AssertionError) as exc_info:
        harness.assert_no_anomalies(workers=4, iterations=50, seed=42)

    msg = str(exc_info.value)
    assert "P4" in msg or "LOST_UPDATE" in msg
    assert "total_wealth_conserved" in msg

    # Direct run inspection
    res = harness.run(workers=4, iterations=50, seed=42)
    assert res.anomaly_detected
    assert res.anomaly_code == "P4"
    assert len(res.minimal_operations) > 0
    assert "sequenceDiagram" in res.mermaid


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
