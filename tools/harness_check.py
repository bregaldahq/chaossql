#!/usr/bin/env python3
"""Harness Integrity Checker for ChaosSQL."""
from pathlib import Path
import sys

REQUIRED_FILES = [
    "AGENTS.md",
    "ARCHITECTURE.md",
    "specs/01_invariant_evaluation.md",
    "specs/02_concurrency_interleaving.md",
    "specs/03_delta_debugging_shrinker.md",
    "specs/04_evidence_synthesis.md",
    "docs/adrs/0001-deterministic-prng-and-replay.md",
    "docs/adrs/0002-async-step-interleaving.md",
]

def main() -> int:
    root = Path(__file__).parent.parent
    missing = []
    for rel in REQUIRED_FILES:
        p = root / rel
        if not p.exists():
            missing.append(rel)

    if missing:
        print("[HARNESS ERROR] Faltando arquivos obrigatórios do harness:", file=sys.stderr)
        for m in missing:
            print(f" - {m}", file=sys.stderr)
        return 1

    print("[HARNESS OK] Todos os ártefatos do harness estão presentes e consistentes.")
    return 0

if __name__ == "__main__":
    sys.exit(main())
