"""
Fluent developer test harness for ChaosSQL in Python.
"""

from typing import Any, Dict, List, Optional, Union
from .types import ChaosResult
from .engine import execute_ipc


class ChaosHarness:
    """
    Fluent builder matching the Go SDK (pkg/chaostest) ergonomics for orchestrating
    deterministic database concurrency experiments and isolating anomalies.
    """

    def __init__(
        self,
        driver: str = "sqlite",
        dsn: str = ":memory:",
        bin_path: Optional[str] = None,
    ):
        self.driver = driver
        self.dsn = dsn
        self.bin_path = bin_path
        self._schema_sql = ""
        self._seed_sql = ""
        self._invariants: List[Dict[str, str]] = []
        self._operations: List[Dict[str, Any]] = []
        self._last_result: Optional[ChaosResult] = None

    def with_schema(self, schema_sql: str) -> "ChaosHarness":
        """Sets DDL schema SQL to initialize the database."""
        self._schema_sql = schema_sql.strip()
        return self

    def with_seed(self, seed_sql: str) -> "ChaosHarness":
        """Sets initial DML SQL data seed."""
        self._seed_sql = seed_sql.strip()
        return self

    def with_invariant(self, name: str, query: str, assertion: str) -> "ChaosHarness":
        """Registers a consistency invariant assertion against the database state."""
        self._invariants.append({
            "name": name,
            "query": query.strip(),
            "assert": assertion.strip(),
        })
        return self

    def add_operation(
        self,
        name: str,
        steps: List[Union[str, Dict[str, str]]],
        weight: float = 1.0,
        params: Optional[Dict[str, str]] = None,
    ) -> "ChaosHarness":
        """
        Registers a transaction operation composed of sequential SQL steps.
        Steps can specify capture syntax (e.g. 'SELECT balance FROM accounts -> cur').
        """
        normalized_steps = []
        for step in steps:
            if isinstance(step, str):
                normalized_steps.append({"sql": step})
            elif isinstance(step, dict):
                normalized_steps.append(step)

        op_def = {
            "name": name,
            "weight": weight,
            "steps": normalized_steps,
        }
        if params:
            op_def["params"] = params

        self._operations.append(op_def)
        return self

    def _build_payload(self, workers: int, iterations: int, seed: int) -> Dict[str, Any]:
        return {
            "driver": self.driver,
            "dsn": self.dsn,
            "schema": self._schema_sql,
            "seed": self._seed_sql,
            "invariants": self._invariants,
            "operations": self._operations,
            "workers": workers,
            "iterations": iterations,
            "seed_value": seed,
        }

    def run(self, workers: int = 2, iterations: int = 20, seed: int = 42) -> ChaosResult:
        """
        Executes the concurrency schedule and applies causal Delta-Debugging (ddmin)
        if an invariant violation is discovered.
        """
        payload = self._build_payload(workers, iterations, seed)
        result = execute_ipc(payload, self.bin_path)
        self._last_result = result
        return result

    def run_and_shrink(self, workers: int = 2, iterations: int = 20, seed: int = 42) -> ChaosResult:
        """Alias for run, executing concurrency fuzzing and automatic ddmin reduction."""
        return self.run(workers=workers, iterations=iterations, seed=seed)

    def assert_no_anomalies(self, workers: int = 2, iterations: int = 20, seed: int = 42) -> ChaosResult:
        """
        Runs the chaos test and raises an AssertionError if any invariant violation occurs,
        formatting the isolated anomaly class, minimal counterexample, and Mermaid sequence.
        """
        result = self.run(workers=workers, iterations=iterations, seed=seed)
        if result.violation_detected:
            inv_name = "unknown"
            if result.failing_invariant:
                inv_name = result.failing_invariant.get("name", "unknown")

            lines = [
                "",
                f"🚨 ChaosSQL Isolation Anomaly Detected: {result.anomaly_type} [{result.anomaly_code}]",
                f"   Failing Invariant: {inv_name}",
            ]
            if result.failing_invariant and "observed" in result.failing_invariant:
                lines.append(f"   Observed DB State: {result.failing_invariant['observed']}")

            if result.minimal_operations:
                lines.append(f"   Minimal Counterexample ({len(result.minimal_operations)} operations isolated):")
                for op in result.minimal_operations:
                    lines.append(f"     - Op #{op.get('id', '?')} [{op.get('name', 'unnamed')}]")
                    for step in op.get("steps", []):
                        cap = f" (capture: {step['capture']})" if step.get("capture") else ""
                        lines.append(f"         SQL: {step.get('sql')}{cap}")

            if result.mermaid:
                lines.append("   Mermaid Sequence Diagram:")
                lines.append("   ```mermaid")
                for m_line in result.mermaid.splitlines():
                    lines.append(f"   {m_line}")
                lines.append("   ```")

            error_msg = "\n".join(lines)
            raise AssertionError(error_msg)

        return result

    def export_standalone_repro(self, file_path: str) -> str:
        """
        Exports minimal reproducing test as standalone Python script.
        """
        if not self._last_result:
            self.run()
        if not self._last_result or not self._last_result.repro_python:
            raise ValueError("No reproduction code available to export.")
        return self._last_result.export_standalone_repro(file_path)
