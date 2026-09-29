package engine_test

import (
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
