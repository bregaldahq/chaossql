package proxy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestGenerateProxySARIF(t *testing.T) {
	anomalies := []DetectedAnomaly{
		{
			ID:         "anomaly-1",
			Type:       domain.AnomalyLostUpdate,
			SessionIDs: []uint64{1, 2},
			TxIDs:      []string{"S1-tx1", "S2-tx1"},
			Items:      []string{"accounts:1"},
			Cycle: []analyzer.Edge{
				{From: "S1-tx1", To: "S2-tx1", Type: analyzer.DepRW, Item: "accounts:1"},
				{From: "S2-tx1", To: "S1-tx1", Type: analyzer.DepWW, Item: "accounts:1"},
			},
			DetectedAt:  time.Now(),
			Description: "P4_LOST_UPDATE cycle detected on accounts:1",
		},
	}

	sarifJSON, err := GenerateProxySARIF(anomalies, "test-proxy-session")
	if err != nil {
		t.Fatalf("failed to generate SARIF: %v", err)
	}

	var report reporter.SarifReport
	if err := json.Unmarshal([]byte(sarifJSON), &report); err != nil {
		t.Fatalf("failed to unmarshal generated SARIF: %v", err)
	}

	if report.Version != "2.1.0" {
		t.Fatalf("expected SARIF version 2.1.0, got %s", report.Version)
	}
	if len(report.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(report.Runs))
	}
	if len(report.Runs[0].Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Runs[0].Results))
	}
	if report.Runs[0].Results[0].RuleID != "chaossql/P4-lost-update" {
		t.Fatalf("expected rule chaossql/P4-lost-update, got %s", report.Runs[0].Results[0].RuleID)
	}

	// Test ExportToFile
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "proxy-anomalies.sarif")
	if err := ExportSARIFToFile(anomalies, "test-proxy", outPath); err != nil {
		t.Fatalf("failed to export SARIF to file: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil || len(data) == 0 {
		t.Fatalf("failed to read written SARIF file: %v", err)
	}
}
