package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		driverName string
		scenarioName string
		force bool
	)

	cmd := &cobra.Command{
		Use:   "init <scenario_path>",
		Short: "Scaffold a new chaos testing scenario template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := args[0]
			if scenarioName == "" {
				scenarioName = filepath.Base(targetDir)
			}

			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}

			schemaPath := filepath.Join(targetDir, "schema.sql")
			seedPath := filepath.Join(targetDir, "seed.sql")
			yamlPath := filepath.Join(targetDir, "chaos.yaml")
			readmePath := filepath.Join(targetDir, "README.md")

			if !force {
				if _, err := os.Stat(yamlPath); err == nil {
					return fmt.Errorf("scenario already exists at %s (use --force to overwrite)", targetDir)
				}
			}

			schemaContent := fmt.Sprintf(`-- Schema for %s
CREATE TABLE accounts (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    balance INT NOT NULL
);
`, scenarioName)

			seedContent := `-- Seed data
INSERT INTO accounts (id, name, balance) VALUES (1, 'Alice', 1000);
INSERT INTO accounts (id, name, balance) VALUES (2, 'Bob', 1000);
`

			yamlContent := fmt.Sprintf(`version: "1.0"
name: "%s"
description: "Generated chaos test scenario for %s"

database:
  driver: "%s"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 4
  iterations: 20
  seed: 42
  jitter_ms: [0, 5]

invariants:
  - name: "total_wealth_preservation"
    query: "SELECT sum(balance) AS total_balance FROM accounts;"
    assert: "total_balance == 2000"

operations:
  - name: "transfer_alice_to_bob"
    weight: 1.0
    steps:
      - sql: "UPDATE accounts SET balance = balance - 100 WHERE id = 1;"
      - sql: "UPDATE accounts SET balance = balance + 100 WHERE id = 2;"

  - name: "transfer_bob_to_alice"
    weight: 1.0
    steps:
      - sql: "UPDATE accounts SET balance = balance - 100 WHERE id = 2;"
      - sql: "UPDATE accounts SET balance = balance + 100 WHERE id = 1;"
`, scenarioName, scenarioName, driverName)

			readmeContent := fmt.Sprintf(`# %s

## Business Context
Describe the domain logic, invariants, and expected transactional isolation semantics.

## Theoretical Anomaly
Document the Adya/Berenson isolation anomaly phenomena targeted by this scenario.
`, scenarioName)

			if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
				return fmt.Errorf("failed to write schema.sql: %w", err)
			}
			if err := os.WriteFile(seedPath, []byte(seedContent), 0644); err != nil {
				return fmt.Errorf("failed to write seed.sql: %w", err)
			}
			if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
				return fmt.Errorf("failed to write chaos.yaml: %w", err)
			}
			if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
				return fmt.Errorf("failed to write README.md: %w", err)
			}

			cmd.Printf("✔ Successfully scaffolded new scenario %q in %s\n", scenarioName, targetDir)
			cmd.Printf("  • %s\n", yamlPath)
			cmd.Printf("  • %s\n", schemaPath)
			cmd.Printf("  • %s\n", seedPath)
			cmd.Printf("  • %s\n", readmePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&driverName, "driver", "sqlite", "Target database driver (sqlite, postgres, mysql)")
	cmd.Flags().StringVar(&scenarioName, "name", "", "Scenario name (defaults to directory base name)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files if scenario already exists")

	return cmd
}
