package cloud

import (
	"github.com/bregaldahq/chaossql/internal/domain"
	"testing"
)

func TestScenarioFingerprintTracksSemanticsWithoutConnectionIdentity(t *testing.T) {
	spec := domain.Spec{Version: "1", Name: "transfer", Description: "example", Database: domain.DatabaseConfig{Driver: "sqlite", Schema: "CREATE TABLE account(id int)", Seed: "INSERT INTO account VALUES(1)"}, Engine: domain.EngineConfig{Seed: 42, Workers: 2, Iterations: 10}, Operations: []domain.OperationConfig{{Name: "read", Params: map[string]string{"a": "1", "b": "2"}, Steps: []domain.StepConfig{{SQL: "SELECT * FROM account"}}}}, Invariants: []domain.InvariantConfig{{Name: "count", Query: "SELECT count(*) FROM account", Assert: "count == 1"}}}
	fingerprint, err := ScenarioFingerprint(spec)
	if err != nil || len(fingerprint) != 64 {
		t.Fatalf("fingerprint=%s error=%v", fingerprint, err)
	}
	renamed := spec
	renamed.Description = "new documentation"
	renamed.Database.DSN = "postgres://secret@local/db"
	renamed.Engine.Seed++
	renamed.Operations = append([]domain.OperationConfig(nil), spec.Operations...)
	renamed.Operations[0].Params = map[string]string{"b": "2", "a": "1"}
	same, _ := ScenarioFingerprint(renamed)
	if same != fingerprint {
		t.Fatal("connection credentials, execution seed, descriptions or map insertion order changed fingerprint")
	}
	for name, mutate := range map[string]func(*domain.Spec){
		"schema":    func(s *domain.Spec) { s.Database.Schema += "; SELECT 1" },
		"seed SQL":  func(s *domain.Spec) { s.Database.Seed += "; SELECT 2" },
		"driver":    func(s *domain.Spec) { s.Database.Driver = "postgres" },
		"isolation": func(s *domain.Spec) { s.Database.Isolation = domain.LevelSerializable },
		"engine":    func(s *domain.Spec) { s.Engine.Iterations++ },
		"operation": func(s *domain.Spec) { s.Operations = []domain.OperationConfig{{Name: "changed"}} },
		"invariant": func(s *domain.Spec) { s.Invariants = []domain.InvariantConfig{{Name: "changed"}} },
		"temporal":  func(s *domain.Spec) { s.TemporalInvariants = []domain.TemporalInvariantConfig{{Name: "changed"}} },
	} {
		t.Run(name, func(t *testing.T) {
			changed := spec
			mutate(&changed)
			got, err := ScenarioFingerprint(changed)
			if err != nil || got == fingerprint {
				t.Fatalf("changed semantics not distinguished: %s %v", got, err)
			}
		})
	}
}
