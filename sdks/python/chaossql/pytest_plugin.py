"""
Native Pytest Plugin and Fixtures for ChaosSQL Concurrency Scenarios.
"""

import json
import os
import subprocess
from typing import Any, Optional
import pytest
from .engine import find_chaossql_binary
from .types import ChaosResult


class ChaosSQLRunner:
    """
    Test helper exposed via `chaossql_runner` pytest fixture.
    """

    def __init__(self, binary_path: Optional[str] = None, defaults: Optional[dict] = None):
        self.binary_path = binary_path or find_chaossql_binary()
        self.defaults = defaults or {}

    def run_scenario(
        self,
        scenario_path: str,
        workers: Optional[int] = None,
        iterations: Optional[int] = None,
        seed: Optional[int] = None,
    ) -> ChaosResult:
        """
        Executes a declarative YAML scenario file through the ChaosSQL CLI
        and returns structured report.
        """
        if not os.path.isabs(scenario_path):
            scenario_path = os.path.abspath(scenario_path)

        if not os.path.isfile(scenario_path):
            raise FileNotFoundError(f"ChaosSQL scenario file not found: {scenario_path}")

        eff_workers = workers if workers is not None else self.defaults.get("workers")
        eff_iterations = iterations if iterations is not None else self.defaults.get("iterations")
        eff_seed = seed if seed is not None else self.defaults.get("seed")

        cmd = [self.binary_path, "run", scenario_path, "--json"]
        if eff_workers is not None:
            cmd.extend(["--workers", str(eff_workers)])
        if eff_iterations is not None:
            cmd.extend(["--iterations", str(eff_iterations)])
        if eff_seed is not None:
            cmd.extend(["--seed", str(eff_seed)])

        res = subprocess.run(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            check=False,
        )

        try:
            raw_resp = json.loads(res.stdout.strip())
        except Exception as exc:
            raise RuntimeError(
                f"Failed to parse ChaosSQL output: {exc}\nStdout:\n{res.stdout}\nStderr:\n{res.stderr}"
            ) from exc

        return ChaosResult(
            success=raw_resp.get("success", False),
            violation_detected=raw_resp.get("violation_detected", False),
            anomaly_type=raw_resp.get("anomaly_type", "UNKNOWN"),
            failing_invariant=raw_resp.get("failing_invariant"),
            duration_ms=raw_resp.get("duration_ms", 0),
            trace_events_count=raw_resp.get("trace_events_count", 0),
            minimal_operations=(raw_resp.get("shrink") or {}).get("minimal_ops", []),
            shrink_summary=raw_resp.get("shrink"),
            mermaid=raw_resp.get("mermaid", ""),
            repro_go=raw_resp.get("repro_go", ""),
            repro_python=raw_resp.get("repro_python", ""),
            repro_typescript=raw_resp.get("repro_typescript", ""),
            error=raw_resp.get("error"),
        )


@pytest.fixture
def chaossql_runner(request: Any) -> ChaosSQLRunner:
    """
    Provides an active ChaosSQL runner configured according to optional @pytest.mark.chaossql parameters.
    """
    bin_path = os.getenv("CHAOSSQL_BIN_PATH")
    default_kwargs = {}

    marker = request.node.get_closest_marker("chaossql")
    if marker and marker.kwargs:
        default_kwargs = dict(marker.kwargs)

    return ChaosSQLRunner(binary_path=bin_path, defaults=default_kwargs)


# Direct alias as documented in technical architecture
chaossql_fixture = chaossql_runner


def pytest_configure(config: Any) -> None:
    """Register custom markers."""
    config.addinivalue_line(
        "markers",
        "chaossql(workers, duration, seed): Mark test as a ChaosSQL concurrency scenario fuzzer.",
    )
