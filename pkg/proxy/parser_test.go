package proxy

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestExtractSQLKeys_TransactionBoundaries(t *testing.T) {
	tests := []struct {
		sql          string
		isTxBoundary bool
		boundaryType string
		isolation    domain.IsolationLevel
	}{
		{"BEGIN", true, "BEGIN", ""},
		{"BEGIN TRANSACTION", true, "BEGIN", ""},
		{"START TRANSACTION", true, "BEGIN", ""},
		{"COMMIT", true, "COMMIT", ""},
		{"END", true, "COMMIT", ""},
		{"ROLLBACK", true, "ROLLBACK", ""},
		{"ABORT", true, "ROLLBACK", ""},
		{"SET TRANSACTION ISOLATION LEVEL SERIALIZABLE", true, "SET_ISOLATION", domain.LevelSerializable},
		{"SET TRANSACTION ISOLATION LEVEL REPEATABLE READ", true, "SET_ISOLATION", domain.LevelRepeatableRead},
		{"SET TRANSACTION ISOLATION LEVEL READ COMMITTED", true, "SET_ISOLATION", domain.LevelReadCommitted},
	}

	for _, tt := range tests {
		t.Run(tt.sql, func(t *testing.T) {
			_, _, isBoundary, bType, iso := InspectSQL(tt.sql)
			if isBoundary != tt.isTxBoundary {
				t.Fatalf("expected isBoundary=%v, got %v for %q", tt.isTxBoundary, isBoundary, tt.sql)
			}
			if bType != tt.boundaryType {
				t.Fatalf("expected boundaryType=%s, got %s for %q", tt.boundaryType, bType, tt.sql)
			}
			if tt.isolation != "" && iso != tt.isolation {
				t.Fatalf("expected isolation=%s, got %s for %q", tt.isolation, iso, tt.sql)
			}
		})
	}
}

func TestExtractSQLKeys_ItemExtraction(t *testing.T) {
	tests := []struct {
		sql           string
		expectedWrite bool
		expectedItems []string
	}{
		{
			sql:           "SELECT balance FROM accounts WHERE id = 1",
			expectedWrite: false,
			expectedItems: []string{"accounts:1"},
		},
		{
			sql:           "SELECT balance FROM accounts WHERE id = 42;",
			expectedWrite: false,
			expectedItems: []string{"accounts:42"},
		},
		{
			sql:           "UPDATE accounts SET balance = balance - 100 WHERE id = 1",
			expectedWrite: true,
			expectedItems: []string{"accounts:1"},
		},
		{
			sql:           "INSERT INTO accounts (id, balance) VALUES (5, 500)",
			expectedWrite: true,
			expectedItems: []string{"accounts:5"},
		},
		{
			sql:           "DELETE FROM accounts WHERE id = 7",
			expectedWrite: true,
			expectedItems: []string{"accounts:7"},
		},
		{
			sql:           "SELECT * FROM doctors WHERE on_call = true",
			expectedWrite: false,
			expectedItems: []string{"doctors"},
		},
		{
			sql:           "UPDATE doctors SET on_call = false WHERE id = 2",
			expectedWrite: true,
			expectedItems: []string{"doctors:2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.sql, func(t *testing.T) {
			isWrite, items, isBoundary, _, _ := InspectSQL(tt.sql)
			if isBoundary {
				t.Fatalf("did not expect boundary for %q", tt.sql)
			}
			if isWrite != tt.expectedWrite {
				t.Fatalf("expected isWrite=%v, got %v for %q", tt.expectedWrite, isWrite, tt.sql)
			}
			if len(items) != len(tt.expectedItems) {
				t.Fatalf("expected items %v, got %v for %q", tt.expectedItems, items, tt.sql)
			}
			for i, it := range tt.expectedItems {
				if items[i] != it {
					t.Errorf("item[%d]: expected %s, got %s", i, it, items[i])
				}
			}
		})
	}
}
