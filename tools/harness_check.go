package main

import (
	"fmt"
	"os"
)

func main() {
	requiredFiles := []string{
		"AGENTS.md",
		"ARCHITECTURE.md",
		"CONTRIBUTING.md",
		"LICENSE",
		"SECURITY.md",
		"Makefile",
		"README.md",
		"action.yml",
		".github/workflows/ci.yml",
		"docs/THEORY.md",
		"docs/ACADEMIC_FOUNDATIONS.md",
		"docs/SCENARIO_ACADEMIC_AUDIT.md",
		"docs/adrs/0001-deterministic-prng-and-replay.md",
		"docs/adrs/0002-async-step-interleaving.md",
		"docs/adrs/0003-causal-delta-debugging-shrinker.md",
		"docs/adrs/0004-pure-go-sqlite-vs-cgo.md",
		"docs/adrs/0005-repro-test-standalone-synthesis.md",
		"docs/adrs/0006-bubbletea-terminal-ux.md",
		"evals/01_shrinking_ratio.md",
		"evals/02_false_positive_rate.md",
		"evals/03_deterministic_replay.md",
		"specs/01_invariant_evaluation.md",
		"specs/02_concurrency_interleaving.md",
		"specs/03_delta_debugging_shrinker.md",
		"specs/04_evidence_synthesis.md",
		"specs/05_advanced_anomaly_taxonomy.md",
		"specs/06_mysql_savepoints_and_otel.md",
		"specs/07_differential_fuzzing_and_matrix.md",
		"specs/08_fault_injection_and_dirty_reads.md",
		"specs/09_temporal_invariants_and_g2_cycles.md",
		"specs/10_developer_tooling_and_static_validator.md",
		"specs/11_version_1_1_developer_sdk_and_smart_generators.md",
	}

	missing := 0
	for _, f := range requiredFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			fmt.Printf("❌ [FALTANDO] Artefato obrigatorio do Harness: %s\n", f)
			missing++
		}
	}

	if missing > 0 {
		fmt.Printf("\n[ERRO] %d artefatos ausentes no Harness.\n", missing)
		os.Exit(1)
	}

	fmt.Printf("[HARNESS OK] Todos os %d artefatos do Harness estao presentes e validados.\n", len(requiredFiles))
}
