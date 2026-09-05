package mutator_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/pkg/mutator"
)

func loadTestScenario(t *testing.T, name string) domain.Spec {
	t.Helper()
	path := filepath.Join("..", "..", "examples", name, "chaos.yaml")
	spec, err := domain.LoadSpec(path)
	if err != nil {
		t.Fatalf("failed to load canonical scenario %q: %v", name, err)
	}
	return *spec
}

func TestMutateScenario_OptionValidation(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")

	// Negative variant count should return error
	opts := mutator.MutationOptions{
		NumVariants: -1,
		Seed:        42,
	}
	_, err := mutator.MutateScenario(spec, opts)
	if err == nil {
		t.Fatal("expected error for negative NumVariants, got nil")
	}

	// Zero variant count should return empty slice and nil error
	opts.NumVariants = 0
	variants, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("unexpected error for NumVariants=0: %v", err)
	}
	if len(variants) != 0 {
		t.Fatalf("expected 0 variants, got %d", len(variants))
	}
}

func TestMutateScenario_VariantCount(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")
	opts := mutator.MutationOptions{
		NumVariants:              5,
		Seed:                     42,
		EnableJitterInjection:    true,
		EnableSavepointInjection: true,
		EnableStepShuffle:        true,
		EnableLockInversion:      true,
	}

	variants, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(variants) != 5 {
		t.Fatalf("expected 5 variants, got %d", len(variants))
	}
}

func TestMutateScenario_PreservesSchemaAndSeed(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")
	origSchema := spec.Database.Schema
	origSeed := spec.Database.Seed

	if origSchema == "" || origSeed == "" {
		t.Fatal("test scenario schema or seed is empty")
	}

	opts := mutator.MutationOptions{
		NumVariants:              10,
		Seed:                     999,
		EnableJitterInjection:    true,
		EnableSavepointInjection: true,
		EnableStepShuffle:        true,
		EnableLockInversion:      true,
	}

	variants, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, v := range variants {
		if v.Database.Schema != origSchema {
			t.Errorf("variant %d did not preserve schema intact", i)
		}
		if v.Database.Seed != origSeed {
			t.Errorf("variant %d did not preserve seed intact", i)
		}
	}
}

func TestMutateScenario_Determinism(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")
	opts := mutator.MutationOptions{
		NumVariants:              4,
		Seed:                     12345,
		EnableJitterInjection:    true,
		EnableSavepointInjection: true,
		EnableStepShuffle:        true,
		EnableLockInversion:      true,
	}

	run1, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("run 1 failed: %v", err)
	}

	run2, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("run 2 failed: %v", err)
	}

	if len(run1) != len(run2) {
		t.Fatalf("mismatched variant counts: %d vs %d", len(run1), len(run2))
	}

	for i := range run1 {
		if !reflect.DeepEqual(run1[i], run2[i]) {
			t.Fatalf("variant %d is not deterministic between runs with same seed %d", i, opts.Seed)
		}
	}
}

func TestMutateScenario_DifferentSeeds(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")
	opts1 := mutator.MutationOptions{
		NumVariants:              3,
		Seed:                     111,
		EnableJitterInjection:    true,
		EnableSavepointInjection: true,
		EnableStepShuffle:        true,
		EnableLockInversion:      true,
	}
	opts2 := mutator.MutationOptions{
		NumVariants:              3,
		Seed:                     222,
		EnableJitterInjection:    true,
		EnableSavepointInjection: true,
		EnableStepShuffle:        true,
		EnableLockInversion:      true,
	}

	variants1, err := mutator.MutateScenario(spec, opts1)
	if err != nil {
		t.Fatalf("run 1 failed: %v", err)
	}
	variants2, err := mutator.MutateScenario(spec, opts2)
	if err != nil {
		t.Fatalf("run 2 failed: %v", err)
	}

	allIdentical := true
	for i := range variants1 {
		if !reflect.DeepEqual(variants1[i], variants2[i]) {
			allIdentical = false
			break
		}
	}
	if allIdentical {
		t.Error("expected different seeds to produce different mutations")
	}
}

func TestMutateScenario_Validity(t *testing.T) {
	scenarios := []string{
		"banking_lost_update",
		"inventory_oversell",
		"hospital_write_skew",
		"deadlock_cycle",
		"foreign_key_cascade_deadlock",
	}

	for _, sc := range scenarios {
		t.Run(sc, func(t *testing.T) {
			spec := loadTestScenario(t, sc)
			opts := mutator.MutationOptions{
				NumVariants:              5,
				Seed:                     777,
				EnableJitterInjection:    true,
				EnableSavepointInjection: true,
				EnableStepShuffle:        true,
				EnableLockInversion:      true,
			}

			variants, err := mutator.MutateScenario(spec, opts)
			if err != nil {
				t.Fatalf("failed to mutate %s: %v", sc, err)
			}

			for i, v := range variants {
				if err := v.Validate(); err != nil {
					t.Fatalf("variant %d of scenario %s failed Validate(): %v", i, sc, err)
				}
			}
		})
	}
}

func TestMutateScenario_JitterInjection(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")
	origJitter := spec.Engine.JitterMs

	opts := mutator.MutationOptions{
		NumVariants:           5,
		Seed:                  42,
		EnableJitterInjection: true,
	}

	variants, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("failed to mutate: %v", err)
	}

	foundDifferent := false
	for i, v := range variants {
		j := v.Engine.JitterMs
		if j[0] < 0 || j[1] < j[0] {
			t.Errorf("variant %d has invalid jitter [%d, %d]", i, j[0], j[1])
		}
		if j != origJitter {
			foundDifferent = true
		}
	}
	if !foundDifferent {
		t.Error("expected at least one variant to have mutated jitter_ms")
	}
}

func TestMutateScenario_SavepointInjection(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")

	opts := mutator.MutationOptions{
		NumVariants:              5,
		Seed:                     101,
		EnableSavepointInjection: true,
	}

	variants, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("failed to mutate: %v", err)
	}

	foundSavepoint := false
	for _, v := range variants {
		for _, op := range v.Operations {
			var savepoints []string
			for _, st := range op.Steps {
				upper := strings.ToUpper(st.SQL)
				if strings.HasPrefix(upper, "SAVEPOINT ") {
					foundSavepoint = true
					parts := strings.Fields(upper)
					if len(parts) >= 2 {
						spName := strings.TrimRight(parts[1], ";")
						savepoints = append(savepoints, spName)
					}
				}
				if strings.HasPrefix(upper, "ROLLBACK TO SAVEPOINT ") {
					parts := strings.Fields(upper)
					spName := strings.TrimRight(parts[len(parts)-1], ";")
					if len(savepoints) == 0 || savepoints[len(savepoints)-1] != spName {
						t.Errorf("mismatched savepoint rollback: expected top %v, got %s", savepoints, spName)
					}
				}
				if strings.HasPrefix(upper, "RELEASE SAVEPOINT ") {
					parts := strings.Fields(upper)
					spName := strings.TrimRight(parts[len(parts)-1], ";")
					if len(savepoints) == 0 || savepoints[len(savepoints)-1] != spName {
						t.Errorf("mismatched savepoint release: expected top %v, got %s", savepoints, spName)
					} else {
						savepoints = savepoints[:len(savepoints)-1]
					}
				}
			}
			if len(savepoints) != 0 {
				t.Errorf("unclosed savepoints in operation %s: %v", op.Name, savepoints)
			}
		}
	}
	if !foundSavepoint {
		t.Error("expected at least one variant to have injected savepoints")
	}
}

func TestMutateScenario_StepShuffle(t *testing.T) {
	spec := loadTestScenario(t, "banking_lost_update")

	opts := mutator.MutationOptions{
		NumVariants:       5,
		Seed:              42,
		EnableStepShuffle: true,
	}

	variants, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("failed to mutate: %v", err)
	}

	for i, v := range variants {
		for _, op := range v.Operations {
			captured := make(map[string]bool)
			for stepIdx, st := range op.Steps {
				if strings.Contains(st.SQL, "{current_bal") && !captured["current_bal"] {
					t.Fatalf("variant %d operation %s step %d uses {current_bal} before it was captured", i, op.Name, stepIdx)
				}
				if st.Capture != "" {
					captured[st.Capture] = true
				}
			}
		}
	}
}

func TestMutateScenario_LockInversion(t *testing.T) {
	spec := loadTestScenario(t, "deadlock_cycle")

	opts := mutator.MutationOptions{
		NumVariants:         5,
		Seed:                42,
		EnableLockInversion: true,
	}

	variants, err := mutator.MutateScenario(spec, opts)
	if err != nil {
		t.Fatalf("failed to mutate: %v", err)
	}

	foundInverted := false
	for _, v := range variants {
		for _, op := range v.Operations {
			if op.Name == "lock_1_then_2" && len(op.Steps) >= 2 {
				if strings.Contains(op.Steps[0].SQL, "id = 2") && strings.Contains(op.Steps[1].SQL, "id = 1") {
					foundInverted = true
				}
			}
		}
	}

	if !foundInverted {
		t.Error("expected at least one variant of deadlock_cycle to have inverted lock order (id = 2 before id = 1)")
	}
}