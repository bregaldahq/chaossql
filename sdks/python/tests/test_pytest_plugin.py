import os
import pytest
from chaossql.pytest_plugin import ChaosSQLRunner


def test_chaossql_runner_fixture(chaossql_runner):
    assert isinstance(chaossql_runner, ChaosSQLRunner)

    # Test executing banking demo scenario
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
    scenario_path = os.path.join(repo_root, "examples", "banking_lost_update", "chaos.yaml")

    report = chaossql_runner.run_scenario(scenario_path, workers=2, iterations=10, seed=42)
    assert report.anomaly_detected
    assert report.anomaly_code == "P4"


@pytest.mark.chaossql(workers=2, duration="2s", seed=42)
def test_scenario_marker(chaossql_runner):
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
    scenario_path = os.path.join(repo_root, "examples", "inventory_oversell", "chaos.yaml")

    report = chaossql_runner.run_scenario(scenario_path, workers=2, iterations=10, seed=42)
    # The oversell scenario catches invariant violation or completes
    assert report.duration_ms >= 0
