package engine

import (
	"fmt"
	"math/rand/v2"
	"sort"

	"github.com/bregaldahq/chaossql/internal/domain"
)

const schedulePlanVersion = 1

// BuildSchedulePlan derives all harness-controlled step decisions without
// depending on goroutine execution or database completion order.
func BuildSchedulePlan(spec domain.Spec, ops []domain.ScheduledOp, prng *PRNG) domain.SchedulePlan {
	workers := spec.Engine.Workers
	if workers <= 0 {
		workers = 4
	}

	plan := domain.SchedulePlan{
		Version:   schedulePlanVersion,
		Seed:      prng.MasterSeed(),
		Workers:   workers,
		Decisions: make([]domain.ScheduleDecision, 0),
	}
	orderedOps := append([]domain.ScheduledOp(nil), ops...)
	sort.SliceStable(orderedOps, func(i, j int) bool {
		return orderedOps[i].ID < orderedOps[j].ID
	})
	for _, op := range orderedOps {
		workerID := assignedWorker(op.ID, workers)
		for stepIndex := range op.Steps {
			stepNumber := stepIndex + 1
			plan.Decisions = append(plan.Decisions, domain.ScheduleDecision{
				Sequence:    len(plan.Decisions) + 1,
				OperationID: op.ID,
				WorkerID:    workerID,
				StepIndex:   stepNumber,
				JitterMs:    plannedJitter(spec.Engine.JitterMs, plan.Seed, op.ID, stepNumber),
				LatencyMs:   plannedLatency(spec.Engine.Faults, plan.Seed, op.ID, stepNumber),
				Abort:       plannedAbort(spec.Engine.Faults, plan.Seed, op.ID, stepNumber),
			})
		}
	}
	return plan
}

func validateScheduledOps(ops []domain.ScheduledOp) error {
	seen := make(map[int]struct{}, len(ops))
	for _, op := range ops {
		if op.ID <= 0 {
			return fmt.Errorf("scheduled operation ID must be positive: %d", op.ID)
		}
		if _, exists := seen[op.ID]; exists {
			return fmt.Errorf("scheduled operation ID must be unique: %d", op.ID)
		}
		seen[op.ID] = struct{}{}
	}
	return nil
}

func partitionScheduledOps(ops []domain.ScheduledOp, workers int) [][]domain.ScheduledOp {
	orderedOps := append([]domain.ScheduledOp(nil), ops...)
	sort.SliceStable(orderedOps, func(i, j int) bool {
		return orderedOps[i].ID < orderedOps[j].ID
	})
	queues := make([][]domain.ScheduledOp, workers)
	for _, op := range orderedOps {
		workerID := assignedWorker(op.ID, workers)
		queues[workerID-1] = append(queues[workerID-1], op)
	}
	return queues
}

func assignedWorker(operationID, workers int) int {
	index := (operationID - 1) % workers
	if index < 0 {
		index += workers
	}
	return index + 1
}

func plannedJitter(jitterRange [2]int, seed uint64, operationID, stepIndex int) int {
	minMs, maxMs := jitterRange[0], jitterRange[1]
	if maxMs <= 0 || maxMs < minMs {
		return 0
	}
	rng := decisionRand(seed, operationID, stepIndex, 1)
	return minMs + rng.IntN(maxMs-minMs+1)
}

func plannedLatency(cfg domain.FaultConfig, seed uint64, operationID, stepIndex int) int {
	if cfg.LatencyProbability <= 0 {
		return 0
	}
	rng := decisionRand(seed, operationID, stepIndex, 2)
	if rng.Float64() >= cfg.LatencyProbability {
		return 0
	}
	minMs, maxMs := cfg.LatencySpikeMs[0], cfg.LatencySpikeMs[1]
	if minMs < 0 {
		minMs = 0
	}
	if maxMs <= minMs {
		maxMs = minMs + 10
	}
	return minMs + rng.IntN(maxMs-minMs+1)
}

func plannedAbort(cfg domain.FaultConfig, seed uint64, operationID, stepIndex int) bool {
	if cfg.AbortProbability <= 0 {
		return false
	}
	return decisionRand(seed, operationID, stepIndex, 3).Float64() < cfg.AbortProbability
}

func decisionRand(seed uint64, operationID, stepIndex int, stream uint64) *rand.Rand {
	first := mix64(seed ^ uint64(operationID)*0x9e3779b97f4a7c15 ^ uint64(stepIndex)*0xbf58476d1ce4e5b9)
	second := mix64(first ^ stream*0x94d049bb133111eb)
	return rand.New(rand.NewPCG(first, second))
}

func mix64(value uint64) uint64 {
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}
