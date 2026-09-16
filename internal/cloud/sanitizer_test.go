package cloud

import (
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
)

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

	if sanitized[0].Worker != "T1" || sanitized[0].OpType != "read" || sanitized[0].Table != "" || sanitized[0].SQL != "" {
		t.Errorf("unexpected event 0: %+v", sanitized[0])
	}
	if sanitized[1].Worker != "T2" || sanitized[1].OpType != "write" || sanitized[1].Table != "" || sanitized[1].SQL != "" {
		t.Errorf("unexpected event 1: %+v", sanitized[1])
	}
	if sanitized[2].OpType != "commit" {
		t.Errorf("unexpected event 2: %+v", sanitized[2])
	}
}
