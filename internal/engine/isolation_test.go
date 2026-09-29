package engine_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestSupportedIsolation(t *testing.T) {
	spec := domain.Spec{Database: domain.DatabaseConfig{Isolation: domain.LevelReadUncommitted}}

	kept, dropped := engine.SupportedIsolation(spec, drivers.NewSQLiteDriver(""))
	if dropped != "" || kept.Database.Isolation != domain.LevelReadUncommitted {
		t.Fatalf("SQLite supports READ_UNCOMMITTED; got %q dropped %q", kept.Database.Isolation, dropped)
	}

	fallback, dropped := engine.SupportedIsolation(spec, drivers.NewPostgresDriver("unused"))
	if dropped != domain.LevelReadUncommitted || fallback.Database.Isolation != "" {
		t.Fatalf("PostgreSQL must fall back to its default; got %q dropped %q", fallback.Database.Isolation, dropped)
	}
	if spec.Database.Isolation != domain.LevelReadUncommitted {
		t.Fatal("the caller's spec must not change")
	}

	unset, dropped := engine.SupportedIsolation(domain.Spec{}, drivers.NewPostgresDriver("unused"))
	if dropped != "" || unset.Database.Isolation != "" {
		t.Fatal("an unset level stays unset")
	}
}

// readCommittedOnly is a Mock that, like PostgreSQL, rejects READ_UNCOMMITTED.
type readCommittedOnly struct{ *drivers.MockDriver }

func (d readCommittedOnly) DriverName() string { return "rc-only" }

func (d readCommittedOnly) EffectiveIsolation(requested domain.IsolationLevel) (domain.IsolationLevel, error) {
	if requested == domain.LevelReadUncommitted {
		return "", fmt.Errorf("isolation %q is not supported by rc-only", requested)
	}
	if requested == "" {
		return domain.LevelReadCommitted, nil
	}
	return requested, nil
}

func TestDifferentialFuzzing_ReportsIsolationFallback(t *testing.T) {
	spec := domain.Spec{
		Name:       "fallback",
		Database:   domain.DatabaseConfig{Isolation: domain.LevelReadUncommitted},
		Engine:     domain.EngineConfig{Workers: 1, Iterations: 2, Seed: 7},
		Invariants: []domain.InvariantConfig{{Name: "zero", Query: "SELECT 0 AS n;", Assert: "n == 0"}},
		Operations: []domain.OperationConfig{{Name: "noop", Steps: []domain.StepConfig{{SQL: "SELECT 1;"}}}},
	}

	res, err := engine.RunDifferentialFuzzing(context.Background(), spec, drivers.NewMockDriver(), readCommittedOnly{drivers.NewMockDriver()}, 7)
	if err != nil {
		t.Fatalf("an unsupported level must fall back, not fail: %v", err)
	}
	if res.ResultA.Isolation != domain.LevelReadUncommitted || res.ResultB.Isolation != domain.LevelReadCommitted {
		t.Fatalf("expected A at READ_UNCOMMITTED and B at its default, got %q and %q", res.ResultA.Isolation, res.ResultB.Isolation)
	}
	if want := "rc-only does not support READ_UNCOMMITTED and ran at READ_COMMITTED."; !strings.Contains(res.DiffSummary, want) {
		t.Fatalf("summary %q does not name the fallback %q", res.DiffSummary, want)
	}
}
