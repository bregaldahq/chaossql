package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		driverName   string
		scenarioName string
		force        bool
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
    balance INT NOT NULL
);

CREATE TABLE withdrawals (
    account_id INT NOT NULL,
    amount INT NOT NULL
);
`, scenarioName)

			seedContent := `-- Seed data
INSERT INTO accounts (id, balance) VALUES (1, 1000);
`

			yamlContent := fmt.Sprintf(`version: "1.0"
name: "%s"
description: "Concurrent withdrawals that read, then write, the balance (lost update)"

database:
  driver: "%s"
%s  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 4
  iterations: 20
  seed: 42
  jitter_ms: [1, 10]

invariants:
  # Every recorded withdrawal must have left the balance. int() keeps the
  # comparison numeric on drivers that return text values (MySQL).
  - name: "balance_matches_withdrawals"
    query: >
      SELECT
        (SELECT balance FROM accounts WHERE id = 1) AS balance,
        (SELECT COALESCE(SUM(amount), 0) FROM withdrawals WHERE account_id = 1) AS withdrawn;
    assert: "int(balance) + int(withdrawn) == 1000"

operations:
  # Reads the balance, computes the new value in the application, writes it
  # back. Two withdrawals that read the same balance erase each other.
  # The fix: UPDATE accounts SET balance = balance - {amount} WHERE id = 1;
  - name: "withdraw"
    weight: 1.0
    params:
      amount: "$random_int(5, 25)"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
        capture: "current_balance"
      - sql: "UPDATE accounts SET balance = {current_balance - amount} WHERE id = 1;"
      - sql: "INSERT INTO withdrawals (account_id, amount) VALUES (1, {amount});"
`, scenarioName, driverName, initIsolationBlock(driverName))

			readmeContent := fmt.Sprintf(`# %s

## Business Context
Withdrawals read the account balance, subtract the amount in the application
and write the result back. Every withdrawal is also recorded in the
withdrawals table, so the balance plus everything withdrawn must stay 1000.

## Anomaly
Lost update (P4). Two withdrawals read the same balance, both write, and the
second write erases the first. Run it and ChaosSQL shrinks the failure to the
two transactions that collide:

    chaossql run chaos.yaml

## Fix
Let the database compute the new balance in one statement, so no write is
based on a stale read:

    UPDATE accounts SET balance = balance - {amount} WHERE id = 1;

Replace the SELECT and UPDATE steps in chaos.yaml with that single step and
run again: the invariant holds. SELECT ... FOR UPDATE (PostgreSQL, MySQL) is
the other common fix.

Now replace the tables, the invariant and the operations with your own.
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
			cmd.Printf("\nNext: chaossql run %s\n", yamlPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&driverName, "driver", "sqlite", "Target database driver (sqlite, postgres, mysql)")
	cmd.Flags().StringVar(&scenarioName, "name", "", "Scenario name (defaults to directory base name)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing files if scenario already exists")

	return cmd
}

// initIsolationBlock sets READ_UNCOMMITTED for SQLite, whose default level
// runs transactions one at a time and would hide the lost update. PostgreSQL
// shows it at its default, READ COMMITTED.
func initIsolationBlock(driverName string) string {
	if driverName != "sqlite" && driverName != "sqlite3" {
		return ""
	}
	return "  # SQLite's default level (SERIALIZABLE) runs transactions one at a time.\n" +
		"  isolation: \"READ_UNCOMMITTED\"\n"
}
