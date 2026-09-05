package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/shrinker"
)

type ValidationResult struct {
	Valid         bool   `json:"valid"`
	Error         string `json:"error,omitempty"`
	Name          string `json:"name,omitempty"`
	NumOperations int    `json:"numOperations,omitempty"`
	NumInvariants int    `json:"numInvariants,omitempty"`
}

type ProgressEvent struct {
	Type           string   `json:"type"`
	Iteration      int      `json:"iteration,omitempty"`
	Ops            int      `json:"ops,omitempty"`
	AnomaliesFound int      `json:"anomaliesFound,omitempty"`
	AnomalyType    string   `json:"anomalyType,omitempty"`
	Edges          []string `json:"edges,omitempty"`
	ShrinkStep     int      `json:"shrinkStep,omitempty"`
	OpsRemaining   int      `json:"opsRemaining,omitempty"`
}

type AdyaEdgeReport struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Item string `json:"item"`
}

type ExecutionReport struct {
	Success          bool                  `json:"success"`
	ViolationFound   bool                  `json:"violationFound"`
	FailingInvariant string                `json:"failingInvariant,omitempty"`
	AnomalyType      string                `json:"anomalyType,omitempty"`
	TotalOps         int                   `json:"totalOps"`
	ReducedOps       int                   `json:"reducedOps"`
	Trace            domain.ExecutionTrace `json:"trace"`
	ReducedTrace     domain.ExecutionTrace `json:"reducedTrace,omitempty"`
	AdyaEdges        []AdyaEdgeReport      `json:"adyaEdges"`
	DurationMs       int64                 `json:"durationMs"`
}

func ValidateScenarioYAML(yamlContent string) ValidationResult {
	spec, err := domain.ParseSpecBytes([]byte(yamlContent))
	if err != nil {
		return ValidationResult{Valid: false, Error: err.Error()}
	}
	return ValidationResult{
		Valid:         true,
		Name:          spec.Name,
		NumOperations: len(spec.Operations),
		NumInvariants: len(spec.Invariants),
	}
}

func ExecuteWasmScenario(ctx context.Context, configJSON string, onProgress func(ProgressEvent)) (*ExecutionReport, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var req struct {
		YAMLContent string `json:"yamlContent"`
		Workers     int    `json:"workers"`
		Iterations  int    `json:"iterations"`
		JitterMs    int    `json:"jitterMs"`
		Seed        uint64 `json:"seed"`
	}
	if err := json.Unmarshal([]byte(configJSON), &req); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}

	spec, err := domain.ParseSpecBytes([]byte(req.YAMLContent))
	if err != nil {
		return nil, fmt.Errorf("spec parse failed: %w", err)
	}

	if req.Workers > 0 {
		spec.Engine.Workers = req.Workers
	}
	if req.Iterations > 0 {
		spec.Engine.Iterations = req.Iterations
	}
	if req.JitterMs > 0 {
		spec.Engine.JitterMs = [2]int{0, req.JitterMs}
	}
	if req.Seed == 0 {
		req.Seed = uint64(time.Now().UnixNano())
	}

	startTime := time.Now()
	driver := drivers.NewMockDriver()
	runner := engine.NewRunner(driver, req.Seed)

	runRes, err := runner.Run(ctx, *spec)
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	graph := analyzer.BuildGraph(runRes.Trace)
	cycles := analyzer.FindCycles(graph)

	var anomalyType string
	if len(cycles) > 0 {
		firstCycle := cycles[0]
		anomaly := analyzer.ClassifyCycle(firstCycle)
		anomalyType = string(anomaly)
		var cycleEdges []string
		for _, e := range firstCycle {
			cycleEdges = append(cycleEdges, fmt.Sprintf("%s -[%s]-> %s", e.From, e.Type, e.To))
		}
		if onProgress != nil {
			onProgress(ProgressEvent{
				Type:        "CYCLE_DETECTED",
				AnomalyType: anomalyType,
				Edges:       cycleEdges,
			})
		}
	}

	var edgeReports []AdyaEdgeReport
	for _, edgeList := range graph.Edges {
		for _, e := range edgeList {
			edgeReports = append(edgeReports, AdyaEdgeReport{
				From: e.From,
				To:   e.To,
				Type: string(e.Type),
				Item: e.Item,
			})
		}
	}

	var reducedOps []domain.ScheduledOp
	var reducedTrace domain.ExecutionTrace
	if runRes.ViolationDetected {
		testFn := func(subset []domain.ScheduledOp) bool {
			subRes, subErr := runner.RunSchedule(ctx, *spec, subset)
			if subErr != nil {
				return true
			}
			return !subRes.ViolationDetected
		}
		shrinkRes, shrinkErr := shrinker.Shrink(ctx, testFn, runRes.ScheduledOps)
		if shrinkErr != nil && (errors.Is(shrinkErr, context.Canceled) || ctx.Err() != nil) {
			return nil, ctx.Err()
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if shrinkErr == nil && shrinkRes != nil && len(shrinkRes.MinimalOps) > 0 {
			reducedOps = shrinkRes.MinimalOps
			if minRun, minErr := runner.RunSchedule(ctx, *spec, reducedOps); minErr == nil {
				reducedTrace = minRun.Trace
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var failingInvName string
	if runRes.FailingInvariant != nil {
		failingInvName = runRes.FailingInvariant.Name
	}

	return &ExecutionReport{
		Success:          runRes.Success,
		ViolationFound:   runRes.ViolationDetected,
		FailingInvariant: failingInvName,
		AnomalyType:      anomalyType,
		TotalOps:         len(runRes.ScheduledOps),
		ReducedOps:       len(reducedOps),
		Trace:            runRes.Trace,
		ReducedTrace:     reducedTrace,
		AdyaEdges:        edgeReports,
		DurationMs:       time.Since(startTime).Milliseconds(),
	}, nil
}
