package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestRunner_TraceTimestamps(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()

	if err := driver.Reset(ctx, "CREATE TABLE dummy (id INT);", ""); err != nil {
		t.Fatalf("driver reset failed: %v", err)
	}

	runner := engine.NewRunner(driver, 1)
	
	ops := []domain.ScheduledOp{
		{
			ID:   1,
			Name: "dummy_op",
			Steps: []domain.StepConfig{
				{SQL: "SELECT 1;"},
			},
		},
	}
	
	spec := domain.Spec{
		Engine: domain.EngineConfig{
			Workers: 1,
			JitterMs: [2]int{10, 20}, // Add intentional delay to test timestamps
		},
	}

	trace, err := runner.ExecuteSchedule(ctx, spec, ops)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(trace) < 3 {
		t.Fatalf("expected at least 3 events (BEGIN, EXEC, COMMIT), got %d", len(trace))
	}

	for i := 1; i < len(trace); i++ {
		if trace[i].Timestamp < trace[i-1].Timestamp {
			t.Errorf("timestamps should be monotonically increasing. event %d timestamp %v is less than event %d timestamp %v", i, trace[i].Timestamp, i-1, trace[i-1].Timestamp)
		}
	}
	
	totalDuration := trace[len(trace)-1].Timestamp - trace[0].Timestamp
	if totalDuration < 10*time.Millisecond {
		t.Errorf("expected total duration to be at least 10ms due to jitter, got %v", totalDuration)
	}
}
