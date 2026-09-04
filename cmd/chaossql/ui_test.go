package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestUICmd_Flags(t *testing.T) {
	cmd := newUICmd()

	portFlag := cmd.Flag("port")
	if portFlag == nil {
		t.Fatalf("expected --port flag on ui command")
	}
	if portFlag.DefValue != "8090" {
		t.Errorf("expected default port to be 8090, got %s", portFlag.DefValue)
	}

	noOpenFlag := cmd.Flag("no-open")
	if noOpenFlag == nil {
		t.Fatalf("expected --no-open flag on ui command")
	}
	if noOpenFlag.DefValue != "false" {
		t.Errorf("expected default no-open to be false, got %s", noOpenFlag.DefValue)
	}
}

func TestUICmd_MissingArg(t *testing.T) {
	cmd := newUICmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{}) // No args

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error when running ui command without arguments")
	}
}

func TestRunCmd_HasUIFlag(t *testing.T) {
	cmd := newRunCmd()
	uiFlag := cmd.Flag("ui")
	if uiFlag == nil {
		t.Fatalf("expected --ui flag on run command")
	}
	if uiFlag.DefValue != "false" {
		t.Errorf("expected default --ui to be false, got %s", uiFlag.DefValue)
	}
}

func TestTraceViewerServer_HTTP200(t *testing.T) {
	spec := domain.Spec{
		Name: "test_ui_server",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
	}
	trace := domain.ExecutionTrace{
		{
			Timestamp: 10 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "account_check",
			Type:      domain.EventBegin,
			SQL:       "BEGIN;",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "account_check",
			Type:      domain.EventExec,
			SQL:       "SELECT 1;",
		},
		{
			Timestamp: 30 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "account_check",
			Type:      domain.EventCommit,
			SQL:       "COMMIT;",
		},
	}
	graph := analyzer.NewAdyaGraph()
	htmlContent := reporter.GenerateEmbeddedTraceViewerHTML(trace, spec, graph, nil, nil)

	server, serverURL, err := startTraceViewerServer("127.0.0.1:0", htmlContent)
	if err != nil {
		t.Fatalf("failed to start trace viewer server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	resp, err := http.Get(serverURL + "/")
	if err != nil {
		t.Fatalf("failed to perform GET request to %s: %v", serverURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE in server response")
	}
	if !strings.Contains(bodyStr, "#120E1F") {
		t.Errorf("expected Bregalda canvas color token in response")
	}
	if !strings.Contains(bodyStr, "test_ui_server") {
		t.Errorf("expected spec name in response")
	}
	if !strings.Contains(bodyStr, "timeline-swimlane") && !strings.Contains(bodyStr, "timeline-container") {
		t.Errorf("expected timeline swimlane in response")
	}
	if !strings.Contains(bodyStr, "adya-graph") {
		t.Errorf("expected adya graph in response")
	}
	if !strings.Contains(bodyStr, "statement-inspector") {
		t.Errorf("expected statement inspector in response")
	}
}

func TestParseTracePayload_Formats(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Direct ExecutionTrace JSON
	traceFile := filepath.Join(tmpDir, "trace_array.json")
	traceJSON := `[
		{"timestamp_us": 1000, "worker_id": 1, "op_index": 1, "op_name": "op1", "type": "EXEC", "sql": "SELECT 1;"}
	]`
	if err := os.WriteFile(traceFile, []byte(traceJSON), 0644); err != nil {
		t.Fatalf("failed to write trace array: %v", err)
	}

	trace, spec, _, _, _, err := loadTraceData(traceFile)
	if err != nil {
		t.Fatalf("failed to load trace array: %v", err)
	}
	if len(trace) != 1 {
		t.Errorf("expected 1 event, got %d", len(trace))
	}
	if spec.Name == "" {
		t.Errorf("expected default spec name")
	}

	// 2. Structured ReplayPayload format
	payloadFile := filepath.Join(tmpDir, "trace_payload.json")
	payloadJSON := `{
		"spec": {
			"name": "structured_test",
			"database": {"driver": "postgres"}
		},
		"trace": [
			{"timestamp_us": 2000, "worker_id": 2, "op_index": 1, "op_name": "op2", "type": "EXEC", "sql": "UPDATE tbl SET v = 1;"}
		],
		"violation_detected": true,
		"failing_invariant": {
			"name": "inv1",
			"passed": false,
			"expression": "v == 0"
		}
	}`
	if err := os.WriteFile(payloadFile, []byte(payloadJSON), 0644); err != nil {
		t.Fatalf("failed to write payload json: %v", err)
	}

	trace2, spec2, _, _, invs, err := loadTraceData(payloadFile)
	if err != nil {
		t.Fatalf("failed to load payload json: %v", err)
	}
	if len(trace2) != 1 {
		t.Errorf("expected 1 event, got %d", len(trace2))
	}
	if spec2.Name != "structured_test" {
		t.Errorf("expected spec name structured_test, got %s", spec2.Name)
	}
	if len(invs) != 1 || invs[0].Name != "inv1" {
		t.Errorf("expected failing invariant inv1")
	}
}
