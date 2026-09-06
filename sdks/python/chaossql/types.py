"""
ChaosSQL Python SDK Data Types and Result Containers.
"""

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional
import os


@dataclass
class AnomalyType:
    UNKNOWN = "UNKNOWN"
    LOST_UPDATE = "P4_LOST_UPDATE"
    WRITE_SKEW = "A5B_WRITE_SKEW"
    READ_SKEW = "A5A_READ_SKEW"
    DIRTY_WRITE = "G0_DIRTY_WRITE"
    DIRTY_READ = "G1A_DIRTY_READ"
    CIRCULAR_INFO = "G1C_CIRCULAR_INFO"
    ANTI_DEPENDENCY = "G2_ANTI_DEPENDENCY"


@dataclass
class ChaosResult:
    """
    Execution summary and anomaly diagnostics returned by ChaosSQL engine.
    """
    success: bool
    violation_detected: bool
    anomaly_type: str = "UNKNOWN"
    failing_invariant: Optional[Dict[str, Any]] = None
    duration_ms: int = 0
    trace_events_count: int = 0
    minimal_operations: List[Dict[str, Any]] = field(default_factory=list)
    shrink_summary: Optional[Dict[str, Any]] = None
    mermaid: str = ""
    repro_go: str = ""
    repro_python: str = ""
    repro_typescript: str = ""
    error: Optional[str] = None

    @property
    def all_invariants_satisfied(self) -> bool:
        """True if execution concluded successfully with all invariant assertions intact."""
        return self.success and not self.violation_detected

    @property
    def is_clean(self) -> bool:
        """Synonym for all_invariants_satisfied (no anomalies discovered and execution succeeded)."""
        return self.success and not self.violation_detected

    @property
    def anomaly_detected(self) -> bool:
        """True if an isolation anomaly or invariant violation was found."""
        return self.violation_detected

    @property
    def anomaly_code(self) -> str:
        """
        Normalized short code for anomaly (e.g. 'A5B' for Write Skew, 'P4' for Lost Update).
        """
        t = (self.anomaly_type or "").upper()
        if "LOST_UPDATE" in t or "P4" in t:
            return "P4"
        if "WRITE_SKEW" in t or "A5B" in t:
            return "A5B"
        if "READ_SKEW" in t or "A5A" in t:
            return "A5A"
        if "DIRTY_WRITE" in t or "G0" in t:
            return "G0"
        if "DIRTY_READ" in t or "G1A" in t:
            return "G1a"
        if "CIRCULAR" in t or "G1C" in t:
            return "G1c"
        if "ANTI_DEPENDENCY" in t or "G2" in t:
            return "G2"
        return t

    def export_standalone_repro(self, file_path: str) -> str:
        """
        Exports minimal counterexample as standalone zero-dependency Python script.
        """
        code = self.repro_python
        if not code:
            raise ValueError("No reproduction code available in execution result.")

        os.makedirs(os.path.dirname(os.path.abspath(file_path)), exist_ok=True)
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(code)
        return file_path
