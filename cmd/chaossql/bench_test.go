package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBenchCmd_DefaultExecution(t *testing.T) {
	cmd := newBenchCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--duration", "200ms", "--workers", "2"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("bench command failed: %v", err)
	}

	out := buf.String()
	requiredKeywords := []string{
		"PRNG",
		"Adya Graph",
		"Delta-Debugging",
		"Database",
		"Component",
		"Metric",
		"Value",
		"Unit",
		"Status",
	}

	for _, kw := range requiredKeywords {
		if !strings.Contains(strings.ToLower(out), strings.ToLower(kw)) {
			t.Errorf("expected benchmark output to contain %q, output:\n%s", kw, out)
		}
	}
}

func TestBenchCmd_CustomScenario(t *testing.T) {
	scenarioPath := "../../examples/banking_lost_update/chaos.yaml"
	if _, err := os.Stat(scenarioPath); os.IsNotExist(err) {
		scenarioPath = "examples/banking_lost_update/chaos.yaml"
		if _, err := os.Stat(scenarioPath); os.IsNotExist(err) {
			scenarioPath = "/root/chaossql/examples/banking_lost_update/chaos.yaml"
		}
	}

	cmd := newBenchCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{scenarioPath, "--duration", "200ms", "--workers", "2"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("bench command with custom scenario failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(strings.ToLower(out), "banking") && !strings.Contains(strings.ToLower(out), "database") {
		t.Errorf("expected output to reference scenario or database, got:\n%s", out)
	}
}

func TestBenchCmd_JSONOutput(t *testing.T) {
	cmd := newBenchCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--duration", "200ms", "--workers", "2", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("bench command with --json failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON benchmark output: %v, raw:\n%s", err, buf.String())
	}

	if _, ok := parsed["metrics"]; !ok {
		t.Errorf("expected 'metrics' key in JSON output, got: %+v", parsed)
	}
	if _, ok := parsed["summary"]; !ok {
		t.Errorf("expected 'summary' key in JSON output, got: %+v", parsed)
	}
}

func TestBenchCmd_InvalidArgs(t *testing.T) {
	t.Run("InvalidDuration", func(t *testing.T) {
		cmd := newBenchCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--duration", "invalid-duration"})

		err := cmd.Execute()
		if err == nil {
			t.Errorf("expected error for invalid duration, got nil")
		}
	})

	t.Run("InvalidScenarioPath", func(t *testing.T) {
		cmd := newBenchCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"nonexistent_path_to_scenario.yaml", "--duration", "100ms"})

		err := cmd.Execute()
		if err == nil {
			t.Errorf("expected error for nonexistent scenario path, got nil")
		}
	})
}

func TestMicroBenchmarks_Direct(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("PRNGAndGenerators", func(t *testing.T) {
		res, err := benchmarkPRNGAndGenerators(ctx, 100*time.Millisecond)
		if err != nil {
			t.Fatalf("benchmarkPRNGAndGenerators failed: %v", err)
		}
		if res.ThroughputOpsPerSec <= 0 {
			t.Errorf("expected throughput > 0, got %f", res.ThroughputOpsPerSec)
		}
		if res.TotalOps <= 0 {
			t.Errorf("expected total ops > 0, got %d", res.TotalOps)
		}
	})

	t.Run("AdyaGraphLatency", func(t *testing.T) {
		for _, nodes := range []int{100, 500, 1000} {
			res, err := benchmarkAdyaGraph(ctx, nodes)
			if err != nil {
				t.Fatalf("benchmarkAdyaGraph(%d) failed: %v", nodes, err)
			}
			if res.Nodes != nodes {
				t.Errorf("expected %d nodes, got %d", nodes, res.Nodes)
			}
			if res.Duration <= 0 {
				t.Errorf("expected positive duration for %d nodes, got %v", nodes, res.Duration)
			}
		}
	})

	t.Run("DeltaDebuggingReduction", func(t *testing.T) {
		res, err := benchmarkDeltaDebugging(ctx, 100*time.Millisecond)
		if err != nil {
			t.Fatalf("benchmarkDeltaDebugging failed: %v", err)
		}
		if res.Iterations <= 0 {
			t.Errorf("expected iterations > 0, got %d", res.Iterations)
		}
		if res.IterationsPerSec <= 0 {
			t.Errorf("expected iterations/sec > 0, got %f", res.IterationsPerSec)
		}
	})

	t.Run("DatabaseConcurrency", func(t *testing.T) {
		cfg := BenchmarkConfig{
			Duration: 150 * time.Millisecond,
			Workers:  2,
		}
		res, err := benchmarkDatabaseConcurrency(ctx, cfg, nil)
		if err != nil {
			t.Fatalf("benchmarkDatabaseConcurrency failed: %v", err)
		}
		if res.TotalTx <= 0 {
			t.Errorf("expected total tx > 0, got %d", res.TotalTx)
		}
		if res.TPS <= 0 {
			t.Errorf("expected TPS > 0, got %f", res.TPS)
		}
	})
}