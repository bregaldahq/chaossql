package cloud

import (
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestSanitizeSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean select",
			input:    "SELECT balance FROM accounts WHERE id = 1",
			expected: "SELECT balance FROM accounts WHERE id = 1",
		},
		{
			name:     "redact password in update",
			input:    "UPDATE users SET password = 'super_secret_123' WHERE id = 10",
			expected: "UPDATE users SET password = '[REDACTED]' WHERE id = 10",
		},
		{
			name:     "redact api token",
			input:    "UPDATE api_keys SET token = 'tok_live_99812491' WHERE org = 5",
			expected: "UPDATE api_keys SET token = '[REDACTED]' WHERE org = 5",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := SanitizeSQL(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func TestExtractTable(t *testing.T) {
	tests := []struct {
		sql      string
		expected string
	}{
		{"SELECT * FROM accounts WHERE id = 1", "accounts"},
		{"UPDATE schema_name.wallets SET bal = 100", "schema_name.wallets"},
		{"INSERT INTO transactions (id, val) VALUES (1, 10)", "transactions"},
		{"BEGIN", ""},
	}

	for _, tc := range tests {
		actual := ExtractTable(tc.sql)
		if actual != tc.expected {
			t.Errorf("sql: %s => expected table %q, got %q", tc.sql, tc.expected, actual)
		}
	}
}

func TestSanitizeTrace(t *testing.T) {
	trace := []domain.TraceEvent{
		{
			WorkerID:  1,
			Type:      domain.EventExec,
			SQL:       "SELECT balance FROM accounts WHERE id = 1",
			Timestamp: 150 * time.Microsecond,
		},
		{
			WorkerID:  2,
			Type:      domain.EventExec,
			SQL:       "UPDATE accounts SET password = 'my_secret' WHERE id = 1",
			Timestamp: 300 * time.Microsecond,
		},
		{
			WorkerID:  1,
			Type:      domain.EventCommit,
			SQL:       "COMMIT",
			Timestamp: 450 * time.Microsecond,
		},
	}

	sanitized := SanitizeTrace(trace)
	if len(sanitized) != 3 {
		t.Fatalf("expected 3 sanitized events, got %d", len(sanitized))
	}

	if sanitized[0].Worker != "T1" || sanitized[0].OpType != "read" || sanitized[0].Table != "accounts" {
		t.Errorf("unexpected event 0: %+v", sanitized[0])
	}
	if sanitized[1].Worker != "T2" || sanitized[1].OpType != "write" || sanitized[1].SQL != "UPDATE accounts SET password = '[REDACTED]' WHERE id = 1" {
		t.Errorf("unexpected event 1: %+v", sanitized[1])
	}
	if sanitized[2].OpType != "commit" {
		t.Errorf("unexpected event 2: %+v", sanitized[2])
	}
}
