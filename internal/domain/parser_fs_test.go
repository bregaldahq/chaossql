package domain_test

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestLoadSpecFS_ResolvesSQLFilesNextToTheSpec(t *testing.T) {
	fsys := fstest.MapFS{
		"bank/chaos.yaml": {Data: []byte(`
version: "1.0"
name: "bank"
database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"
invariants:
  - name: "positive"
    query: "SELECT balance FROM accounts;"
    assert: "balance >= 0"
operations:
  - name: "noop"
    steps:
      - sql: "SELECT 1;"
`)},
		"bank/schema.sql": {Data: []byte("CREATE TABLE accounts (balance INT);")},
		"bank/seed.sql":   {Data: []byte("INSERT INTO accounts VALUES (1);")},
	}

	spec, err := domain.LoadSpecFS(fsys, "bank/chaos.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Database.Schema != "CREATE TABLE accounts (balance INT);" || spec.Database.Seed != "INSERT INTO accounts VALUES (1);" {
		t.Fatalf("SQL files not resolved: %+v", spec.Database)
	}

	delete(fsys, "bank/seed.sql")
	if _, err := domain.LoadSpecFS(fsys, "bank/chaos.yaml"); err == nil {
		t.Fatal("a missing seed file must fail")
	}
}

func TestParseIsolationLevel(t *testing.T) {
	for input, want := range map[string]domain.IsolationLevel{
		"READ_UNCOMMITTED":  domain.LevelReadUncommitted,
		"read-committed":    domain.LevelReadCommitted,
		" repeatable read ": domain.LevelRepeatableRead,
		"Serializable":      domain.LevelSerializable,
	} {
		got, err := domain.ParseIsolationLevel(input)
		if err != nil || got != want {
			t.Errorf("ParseIsolationLevel(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"", "snapshot", "READ UNCOMMITTED;"} {
		if _, err := domain.ParseIsolationLevel(input); !errors.Is(err, domain.ErrSpecValidationFailed) {
			t.Errorf("ParseIsolationLevel(%q) must fail validation, got %v", input, err)
		}
	}
}
