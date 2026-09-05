package mutator

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// MutationOptions configures which adversarial operators are applied to generate scenario variants.
type MutationOptions struct {
	NumVariants              int    `json:"num_variants"`
	Seed                     uint64 `json:"seed"`
	EnableJitterInjection    bool   `json:"enable_jitter_injection"`
	EnableSavepointInjection bool   `json:"enable_savepoint_injection"`
	EnableStepShuffle        bool   `json:"enable_step_shuffle"`
	EnableLockInversion      bool   `json:"enable_lock_inversion"`
}

// DefaultMutationOptions returns standard mutation settings with all operators enabled.
func DefaultMutationOptions() MutationOptions {
	return MutationOptions{
		NumVariants:              5,
		Seed:                     42,
		EnableJitterInjection:    true,
		EnableSavepointInjection: true,
		EnableStepShuffle:        true,
		EnableLockInversion:      true,
	}
}

// MutateScenario produces deterministic adversarial variants from a canonical chaos testing specification.
func MutateScenario(spec domain.Spec, opts MutationOptions) ([]domain.Spec, error) {
	if opts.NumVariants < 0 {
		return nil, errors.New("num variants must be non-negative")
	}
	if opts.NumVariants == 0 {
		return []domain.Spec{}, nil
	}

	variants := make([]domain.Spec, opts.NumVariants)

	for i := 0; i < opts.NumVariants; i++ {
		// Use deterministic PRNG seeded per variant
		variantSeed := opts.Seed + uint64(i)*7919 + 1
		rng := rand.New(rand.NewSource(int64(variantSeed)))

		v := cloneSpec(spec)
		v.Name = fmt.Sprintf("%s_variant_%d", spec.Name, i)
		v.Engine.Seed = variantSeed

		// Operator 1: Randomized jitter injection
		if opts.EnableJitterInjection {
			applyJitter(&v, rng)
		}

		// Operator 4: Lock Inversion (applied before savepoints so savepoint wrapping remains intact)
		if opts.EnableLockInversion {
			applyLockInversion(&v, rng)
		}

		// Operator 3: Step Shuffle (applied before savepoints so savepoint pairs are not split)
		if opts.EnableStepShuffle {
			applyStepShuffle(&v, rng)
		}

		// Operator 2: Nested Savepoints injection
		if opts.EnableSavepointInjection {
			applySavepoints(&v, rng)
		}

		variants[i] = v
	}

	return variants, nil
}

func cloneSpec(spec domain.Spec) domain.Spec {
	cloned := spec
	cloned.Database = spec.Database
	cloned.Engine = spec.Engine
	cloned.Engine.Faults = spec.Engine.Faults

	cloned.Invariants = make([]domain.InvariantConfig, len(spec.Invariants))
	copy(cloned.Invariants, spec.Invariants)

	if len(spec.TemporalInvariants) > 0 {
		cloned.TemporalInvariants = make([]domain.TemporalInvariantConfig, len(spec.TemporalInvariants))
		copy(cloned.TemporalInvariants, spec.TemporalInvariants)
	}

	cloned.Operations = make([]domain.OperationConfig, len(spec.Operations))
	for opIdx, op := range spec.Operations {
		clonedOp := op
		if op.Params != nil {
			clonedOp.Params = make(map[string]string, len(op.Params))
			for k, val := range op.Params {
				clonedOp.Params[k] = val
			}
		}
		clonedOp.Steps = make([]domain.StepConfig, len(op.Steps))
		copy(clonedOp.Steps, op.Steps)
		cloned.Operations[opIdx] = clonedOp
	}

	return cloned
}

// Operator 1 (Jitter): Injects randomized delays / micro-jitter configuration into spec.Engine.JitterMs.
func applyJitter(spec *domain.Spec, rng *rand.Rand) {
	minJitter := rng.Intn(10)
	maxJitter := minJitter + 2 + rng.Intn(25)
	if spec.Engine.JitterMs == [2]int{minJitter, maxJitter} {
		maxJitter++
	}
	spec.Engine.JitterMs = [2]int{minJitter, maxJitter}
}

// Operator 2 (Savepoints): Injects nested SAVEPOINT sp1; ...; RELEASE SAVEPOINT sp1; or ROLLBACK TO SAVEPOINT sp1;.
func applySavepoints(spec *domain.Spec, rng *rand.Rand) {
	for opIdx := range spec.Operations {
		op := &spec.Operations[opIdx]
		if len(op.Steps) == 0 {
			continue
		}

		// Probability check to avoid mutating every single operation in the same way
		if rng.Float64() < 0.2 && len(spec.Operations) > 1 {
			continue
		}

		if len(op.Steps) == 1 {
			newSteps := []domain.StepConfig{
				{SQL: "SAVEPOINT sp1;"},
				op.Steps[0],
				{SQL: "RELEASE SAVEPOINT sp1;"},
			}
			op.Steps = newSteps
		} else {
			newSteps := make([]domain.StepConfig, 0, len(op.Steps)+5)
			newSteps = append(newSteps, domain.StepConfig{SQL: "SAVEPOINT sp1;"})
			newSteps = append(newSteps, op.Steps[0])

			newSteps = append(newSteps, domain.StepConfig{SQL: "SAVEPOINT sp2;"})
			newSteps = append(newSteps, op.Steps[1])

			if rng.Float64() < 0.3 {
				newSteps = append(newSteps, domain.StepConfig{SQL: "ROLLBACK TO SAVEPOINT sp2;"})
			}
			newSteps = append(newSteps, domain.StepConfig{SQL: "RELEASE SAVEPOINT sp2;"})

			for k := 2; k < len(op.Steps); k++ {
				newSteps = append(newSteps, op.Steps[k])
			}

			newSteps = append(newSteps, domain.StepConfig{SQL: "RELEASE SAVEPOINT sp1;"})
			op.Steps = newSteps
		}
	}
}

var (
	tablePattern = regexp.MustCompile(`(?i)\b(?:FROM|INTO|UPDATE)\s+([a-zA-Z0-9_]+)`)
	idPattern    = regexp.MustCompile(`(?i)\b(?:id|account_id|order_id|item_id|doctor_id|user_id|book_id)\s*=\s*(\d+)`)
)

func extractTargetTable(sql string) string {
	matches := tablePattern.FindStringSubmatch(sql)
	if len(matches) >= 2 {
		return strings.ToLower(matches[1])
	}
	return ""
}

// Operator 3 (Step Shuffle): Permutes non-conflicting steps in multi-step operations.
func applyStepShuffle(spec *domain.Spec, rng *rand.Rand) {
	for opIdx := range spec.Operations {
		op := &spec.Operations[opIdx]
		n := len(op.Steps)
		if n < 2 {
			continue
		}

		// Build dependency DAG: edge i -> j means step i must execute before step j
		adj := make([][]int, n)
		inDegree := make([]int, n)

		for i := 0; i < n; i++ {
			stepI := op.Steps[i]
			upperI := strings.ToUpper(stepI.SQL)
			isSavepointI := strings.Contains(upperI, "SAVEPOINT")
			tableI := extractTargetTable(stepI.SQL)

			for j := i + 1; j < n; j++ {
				stepJ := op.Steps[j]
				upperJ := strings.ToUpper(stepJ.SQL)
				isSavepointJ := strings.Contains(upperJ, "SAVEPOINT")
				tableJ := extractTargetTable(stepJ.SQL)

				conflict := false

				// 1. Capture dependency: if step j references variable captured by step i
				if stepI.Capture != "" && (strings.Contains(stepJ.SQL, "{"+stepI.Capture) || strings.Contains(stepJ.SQL, stepI.Capture)) {
					conflict = true
				}

				// 2. Savepoint lifecycle ordering must be preserved
				if isSavepointI || isSavepointJ {
					conflict = true
				}

				// 3. Write-write conflict on same table: preserve ordering
				if tableI != "" && tableI == tableJ && (isMutation(upperI) || isMutation(upperJ)) {
					conflict = true
				}

				if conflict {
					adj[i] = append(adj[i], j)
					inDegree[j]++
				}
			}
		}

		// Randomized topological sort
		available := make([]int, 0, n)
		for i := 0; i < n; i++ {
			if inDegree[i] == 0 {
				available = append(available, i)
			}
		}

		result := make([]domain.StepConfig, 0, n)
		for len(available) > 0 {
			pickIdx := rng.Intn(len(available))
			curr := available[pickIdx]
			available = append(available[:pickIdx], available[pickIdx+1:]...)

			result = append(result, op.Steps[curr])

			for _, nextNode := range adj[curr] {
				inDegree[nextNode]--
				if inDegree[nextNode] == 0 {
					available = append(available, nextNode)
				}
			}
		}

		if len(result) == n {
			op.Steps = result
		}
	}
}

func isMutation(upperSQL string) bool {
	return strings.HasPrefix(upperSQL, "UPDATE") || strings.HasPrefix(upperSQL, "INSERT") || strings.HasPrefix(upperSQL, "DELETE")
}

// Operator 4 (Lock Inversion): Swaps update order of resource IDs where applicable.
func applyLockInversion(spec *domain.Spec, rng *rand.Rand) {
	for opIdx := range spec.Operations {
		op := &spec.Operations[opIdx]
		n := len(op.Steps)
		if n < 2 {
			continue
		}

		// Look for update/write steps with distinct resource IDs
		type stepResource struct {
			stepIndex int
			resID     string
		}

		var resources []stepResource
		for idx, st := range op.Steps {
			upper := strings.ToUpper(st.SQL)
			if strings.HasPrefix(upper, "UPDATE") || strings.HasPrefix(upper, "DELETE") || strings.Contains(upper, "FOR UPDATE") {
				matches := idPattern.FindStringSubmatch(st.SQL)
				if len(matches) >= 2 {
					resources = append(resources, stepResource{
						stepIndex: idx,
						resID:     matches[1],
					})
				}
			}
		}

		// If at least two write steps have different resource IDs (e.g. id=1 and id=2)
		if len(resources) >= 2 {
			first := resources[0]
			var second *stepResource
			for _, r := range resources[1:] {
				if r.resID != first.resID {
					second = &r
					break
				}
			}

			if second != nil {
				// Verify no capture dependency between first.stepIndex and second.stepIndex
				i1, i2 := first.stepIndex, second.stepIndex
				hasDep := false
				if op.Steps[i1].Capture != "" && strings.Contains(op.Steps[i2].SQL, op.Steps[i1].Capture) {
					hasDep = true
				}
				if !hasDep {
					// Swap steps to invert lock acquisition hierarchy
					op.Steps[i1], op.Steps[i2] = op.Steps[i2], op.Steps[i1]
				}
			}
		}
	}
}