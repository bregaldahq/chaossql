package engine_test

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestDetectEventType(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected domain.TraceEventType
	}{
		{
			name:     "savepoint basic",
			sql:      "SAVEPOINT sp1;",
			expected: domain.EventSavepoint,
		},
		{
			name:     "savepoint lowercase with whitespace",
			sql:      "   savepoint my_sp   ",
			expected: domain.EventSavepoint,
		},
		{
			name:     "rollback to savepoint uppercase",
			sql:      "ROLLBACK TO SAVEPOINT sp1;",
			expected: domain.EventRollbackTo,
		},
		{
			name:     "rollback to savepoint lowercase",
			sql:      "rollback to savepoint sp1",
			expected: domain.EventRollbackTo,
		},
		{
			name:     "rollback to direct uppercase",
			sql:      "ROLLBACK TO sp1;",
			expected: domain.EventRollbackTo,
		},
		{
			name:     "rollback to direct lowercase with whitespace",
			sql:      "   rollback to sp1   ",
			expected: domain.EventRollbackTo,
		},
		{
			name:     "release savepoint uppercase",
			sql:      "RELEASE SAVEPOINT sp1;",
			expected: domain.EventReleaseSavepoint,
		},
		{
			name:     "release savepoint lowercase",
			sql:      "release savepoint sp1",
			expected: domain.EventReleaseSavepoint,
		},
		{
			name:     "release short uppercase",
			sql:      "RELEASE sp1;",
			expected: domain.EventReleaseSavepoint,
		},
		{
			name:     "release short lowercase",
			sql:      "release sp1",
			expected: domain.EventReleaseSavepoint,
		},
		{
			name:     "regular select",
			sql:      "SELECT * FROM accounts WHERE id = 1;",
			expected: domain.EventExec,
		},
		{
			name:     "regular insert",
			sql:      "INSERT INTO accounts (id, balance) VALUES (1, 100);",
			expected: domain.EventExec,
		},
		{
			name:     "regular update",
			sql:      "UPDATE accounts SET balance = 200 WHERE id = 1;",
			expected: domain.EventExec,
		},
		{
			name:     "regular delete",
			sql:      "DELETE FROM accounts WHERE id = 1;",
			expected: domain.EventExec,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := engine.DetectEventType(tc.sql)
			if actual != tc.expected {
				t.Errorf("DetectEventType(%q) = %q, expected %q", tc.sql, actual, tc.expected)
			}
		})
	}
}

func TestRunner_SavepointRollback(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()

	schema := "CREATE TABLE items (id INT PRIMARY KEY, name TEXT);"
	if err := driver.Reset(ctx, schema, ""); err != nil {
		t.Fatalf("driver reset failed: %v", err)
	}

	runner := engine.NewRunner(driver, 1)

	ops := []domain.ScheduledOp{
		{
			ID:   1,
			Name: "savepoint_rollback_op",
			Steps: []domain.StepConfig{
				{SQL: "INSERT INTO items (id, name) VALUES (1, 'row1');"},
				{SQL: "SAVEPOINT sp1;"},
				{SQL: "INSERT INTO items (id, name) VALUES (2, 'row2');"},
				{SQL: "ROLLBACK TO SAVEPOINT sp1;"},
			},
		},
	}

	spec := domain.Spec{
		Engine: domain.EngineConfig{
			Workers: 1,
		},
	}

	trace, err := runner.ExecuteSchedule(ctx, spec, ops)
	if err != nil {
		t.Fatalf("unexpected ExecuteSchedule error: %v", err)
	}

	// 1. Assert TraceEvents
	var foundSavepoint, foundRollbackTo, foundBegin, foundCommit bool
	for _, ev := range trace {
		switch ev.Type {
		case domain.EventBegin:
			foundBegin = true
		case domain.EventSavepoint:
			foundSavepoint = true
		case domain.EventRollbackTo:
			foundRollbackTo = true
		case domain.EventCommit:
			foundCommit = true
		}
	}

	if !foundBegin {
		t.Error("expected trace to contain EventBegin")
	}
	if !foundSavepoint {
		t.Error("expected trace to contain EventSavepoint")
	}
	if !foundRollbackTo {
		t.Error("expected trace to contain EventRollbackTo")
	}
	if !foundCommit {
		t.Error("expected trace to contain EventCommit")
	}

	// 2. Assert final DB state: contains row 1 but NOT row 2
	rows, err := driver.Query(ctx, "SELECT id, name FROM items ORDER BY id;")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	type Item struct {
		ID   int
		Name string
	}
	var items []Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			t.Fatalf("row scan failed: %v", err)
		}
		items = append(items, it)
	}

	if len(items) != 1 {
		t.Fatalf("expected exactly 1 item in db, got %d items: %+v", len(items), items)
	}

	if items[0].ID != 1 || items[0].Name != "row1" {
		t.Errorf("expected item 1 ('row1'), got: %+v", items[0])
	}

	for _, it := range items {
		if it.ID == 2 {
			t.Errorf("row 2 should have been rolled back, but was found in DB: %+v", it)
		}
	}
}

func TestRunner_SavepointRelease(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()

	schema := "CREATE TABLE items (id INT PRIMARY KEY, name TEXT);"
	if err := driver.Reset(ctx, schema, ""); err != nil {
		t.Fatalf("driver reset failed: %v", err)
	}

	runner := engine.NewRunner(driver, 1)

	ops := []domain.ScheduledOp{
		{
			ID:   1,
			Name: "savepoint_release_op",
			Steps: []domain.StepConfig{
				{SQL: "INSERT INTO items (id, name) VALUES (1, 'row1');"},
				{SQL: "SAVEPOINT sp1;"},
				{SQL: "INSERT INTO items (id, name) VALUES (2, 'row2');"},
				{SQL: "RELEASE SAVEPOINT sp1;"},
			},
		},
	}

	spec := domain.Spec{
		Engine: domain.EngineConfig{
			Workers: 1,
		},
	}

	trace, err := runner.ExecuteSchedule(ctx, spec, ops)
	if err != nil {
		t.Fatalf("unexpected ExecuteSchedule error: %v", err)
	}

	var foundSavepoint, foundRelease bool
	for _, ev := range trace {
		if ev.Type == domain.EventSavepoint {
			foundSavepoint = true
		}
		if ev.Type == domain.EventReleaseSavepoint {
			foundRelease = true
		}
	}

	if !foundSavepoint {
		t.Error("expected trace to contain EventSavepoint")
	}
	if !foundRelease {
		t.Error("expected trace to contain EventReleaseSavepoint")
	}

	// Final DB state should contain both row 1 and row 2
	rows, err := driver.Query(ctx, "SELECT id, name FROM items ORDER BY id;")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}

	if count != 2 {
		t.Errorf("expected 2 items in db after release savepoint, got %d", count)
	}
}
