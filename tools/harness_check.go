package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var RequiredFiles = []string{
	"AGENTS.md",
	"ARCHITECTURE.md",
	"go.mod",
	"Makefile",
	"docs/THEORY.md",
	"docs/ACADEMIC_FOUNDATIONS.md",
	"docs/SCENARIO_ACADEMIC_AUDIT.md",
	"docs/adrs/0001-deterministic-prng-and-replay.md",
	"docs/adrs/0002-async-step-interleaving.md",
	"docs/adrs/0003-causal-delta-debugging-shrinker.md",
	"docs/adrs/0004-pure-go-sqlite-vs-cgo.md",
	"docs/adrs/0005-repro-test-standalone-synthesis.md",
	"docs/adrs/0006-bubbletea-terminal-ux.md",
	"specs/01_invariant_evaluation.md",
	"specs/02_concurrency_interleaving.md",
	"specs/03_delta_debugging_shrinker.md",
	"specs/04_evidence_synthesis.md",
	"evals/01_shrinking_ratio.md",
	"evals/02_false_positive_rate.md",
	"evals/03_deterministic_replay.md",
	"examples/banking_lost_update/chaos.yaml",
	"examples/inventory_oversell/chaos.yaml",
	"examples/hospital_write_skew/chaos.yaml",
	"examples/read_skew_financial_audit/chaos.yaml",
}

func main() {
	root := "."
	var missing []string

	for _, rel := range RequiredFiles {
		path := filepath.Join(root, rel)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing = append(missing, rel)
		}
	}

	if len(missing) > 0 {
		fmt.Fprintln(os.Stderr, "[HARNESS ERROR] Arquivos obrigatórios do Harness ausentes:")
		for _, m := range missing {
			fmt.Fprintf(os.Stderr, " - %s\n", m)
		}
		os.Exit(1)
	}

	fmt.Println("[HARNESS OK] Todos os 24 ártefatos do Harness estão presentes e validados.")
}
