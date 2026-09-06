"""
IPC subprocess client for executing the ChaosSQL core engine.
"""

import json
import os
import shutil
import subprocess
import sys
from typing import Any, Dict, Optional
from .types import ChaosResult


def find_chaossql_binary(explicit_path: Optional[str] = None) -> str:
    """
    Locates the chaossql executable binary.
    Resolution precedence:
    1. explicit_path passed directly
    2. CHAOSSQL_BIN_PATH environment variable
    3. Traversal to repository or installed bin/chaossql
    4. PATH environment lookup
    """
    if explicit_path and os.path.isfile(explicit_path) and os.access(explicit_path, os.X_OK):
        return os.path.abspath(explicit_path)

    env_path = os.getenv("CHAOSSQL_BIN_PATH")
    if env_path and os.path.isfile(env_path) and os.access(env_path, os.X_OK):
        return os.path.abspath(env_path)

    binary_name = "chaossql.exe" if sys.platform == "win32" else "chaossql"

    # Check relative to SDK file location up to root bin/chaossql
    current_dir = os.path.dirname(os.path.abspath(__file__))
    for _ in range(5):
        candidate = os.path.join(current_dir, "bin", binary_name)
        if os.path.isfile(candidate) and os.access(candidate, os.X_OK):
            return candidate
        parent = os.path.dirname(current_dir)
        if parent == current_dir:
            break
        current_dir = parent

    # Check system PATH
    which_bin = shutil.which("chaossql") or shutil.which("chaossql.exe")
    if which_bin:
        return which_bin

    raise FileNotFoundError(
        "ChaosSQL core engine binary not found. Please compile it using `make build` "
        "or set the CHAOSSQL_BIN_PATH environment variable."
    )


def execute_ipc(
    payload: Dict[str, Any],
    binary_path: Optional[str] = None,
    timeout_sec: float = 60.0,
) -> ChaosResult:
    """
    Executes a scenario payload against chaossql engine via bidirectional JSON streaming over stdin/stdout.
    """
    bin_path = find_chaossql_binary(binary_path)

    json_input = json.dumps(payload)
    process = subprocess.Popen(
        [bin_path, "engine"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=0,
    )

    try:
        stdout_data, stderr_data = process.communicate(input=json_input, timeout=timeout_sec)
    except subprocess.TimeoutExpired:
        process.kill()
        stdout_data, stderr_data = process.communicate()
        raise TimeoutError(
            f"ChaosSQL engine process timed out after {timeout_sec}s. Stderr: {stderr_data}"
        )

    if process.returncode != 0 and not stdout_data:
        raise RuntimeError(
            f"ChaosSQL engine process exited with code {process.returncode}: {stderr_data.strip()}"
        )

    try:
        raw_resp = json.loads(stdout_data.strip())
    except json.JSONDecodeError as exc:
        raise RuntimeError(
            f"Failed to decode engine JSON response: {exc}\nStdout:\n{stdout_data}\nStderr:\n{stderr_data}"
        ) from exc

    return ChaosResult(
        success=raw_resp.get("success", False),
        violation_detected=raw_resp.get("violation_detected", False),
        anomaly_type=raw_resp.get("anomaly_type", "UNKNOWN"),
        failing_invariant=raw_resp.get("failing_invariant"),
        duration_ms=raw_resp.get("duration_ms", 0),
        trace_events_count=raw_resp.get("trace_events_count", 0),
        minimal_operations=raw_resp.get("minimal_operations", []),
        shrink_summary=raw_resp.get("shrink"),
        mermaid=raw_resp.get("mermaid", ""),
        repro_go=raw_resp.get("repro_go", ""),
        repro_python=raw_resp.get("repro_python", ""),
        repro_typescript=raw_resp.get("repro_typescript", ""),
        error=raw_resp.get("error"),
    )
