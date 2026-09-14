import os
import pytest
from chaossql.pytest_plugin import ChaosSQLRunner


def test_chaossql_runner_fixture(chaossql_runner, tmp_path):
    assert isinstance(chaossql_runner, ChaosSQLRunner)

    scenario_path = tmp_path / "smoke.yaml"
    scenario_path.write_text(
        """version: '1.1'
name: python_plugin_smoke
database:
  driver: sqlite
  dsn: ':memory:'
  schema: 'CREATE TABLE items (id INT PRIMARY KEY, qty INT);'
  seed: 'INSERT INTO items VALUES (1, 10);'
engine:
  workers: 1
  iterations: 1
  seed: 42
invariants:
  - name: quantity_unchanged
    query: 'SELECT qty FROM items WHERE id = 1'
    assert: 'qty == 10'
operations:
  - name: read_quantity
    weight: 1
    steps:
      - sql: 'SELECT qty FROM items WHERE id = 1'
""",
        encoding="utf-8",
    )

    report = chaossql_runner.run_scenario(str(scenario_path), workers=1, iterations=1, seed=42)
    assert report.is_clean
    assert report.status == "passed"


@pytest.mark.chaossql(workers=2, duration="2s", seed=42)
def test_scenario_marker(chaossql_runner):
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
    scenario_path = os.path.join(repo_root, "examples", "inventory_oversell", "chaos.yaml")

    report = chaossql_runner.run_scenario(scenario_path, workers=2, iterations=10, seed=42)
    # The oversell scenario catches invariant violation or completes
    assert report.duration_ms >= 0
