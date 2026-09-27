package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/pkg/proxy"
)

func TestProxyCmd_Flags(t *testing.T) {
	cmd := newProxyCmd()
	if cmd.Use != "proxy" {
		t.Fatalf("expected Use 'proxy', got %s", cmd.Use)
	}

	listenFlag := cmd.Flag("listen")
	if listenFlag == nil || listenFlag.DefValue != "127.0.0.1:5433" {
		t.Fatalf("expected default listen flag 127.0.0.1:5433, got %v", listenFlag)
	}

	upstreamFlag := cmd.Flag("upstream")
	if upstreamFlag == nil || upstreamFlag.DefValue != "127.0.0.1:5432" {
		t.Fatalf("expected default upstream flag 127.0.0.1:5432, got %v", upstreamFlag)
	}

	protocolFlag := cmd.Flag("protocol")
	if protocolFlag == nil || protocolFlag.DefValue != "postgres" {
		t.Fatalf("expected default protocol flag postgres, got %v", protocolFlag)
	}

	jMinFlag := cmd.Flag("jitter-min")
	if jMinFlag == nil || jMinFlag.DefValue != (10*time.Microsecond).String() {
		t.Fatalf("expected jitter-min 10us, got %v", jMinFlag)
	}

	sarifFlag := cmd.Flag("export-sarif")
	if sarifFlag == nil {
		t.Fatalf("expected export-sarif flag")
	}

	uiPortFlag := cmd.Flag("ui-port")
	if uiPortFlag == nil {
		t.Fatalf("expected ui-port flag")
	}
}

func TestProxyCmd_ServesUntilContextCancelled(t *testing.T) {
	listenAddr, _ := freeLocalAddr(t)
	_, uiPort := freeLocalAddr(t)
	sarifPath := filepath.Join(t.TempDir(), "proxy.sarif")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := newProxyCmd()
	cmd.SetArgs([]string{"--listen", listenAddr, "--upstream", "127.0.0.1:1", "--ui-port", strconv.Itoa(uiPort), "--export-sarif", sarifPath})
	cmd.SilenceUsage = true

	var runErr error
	output := captureStdout(t, func() {
		done := make(chan error, 1)
		go func() { done <- cmd.ExecuteContext(ctx) }()

		statusURL := fmt.Sprintf("http://127.0.0.1:%d/api/status", uiPort)
		deadline := time.Now().Add(10 * time.Second)
		for {
			resp, err := http.Get(statusURL)
			if err == nil {
				resp.Body.Close()
				if conn, err := net.Dial("tcp", listenAddr); err == nil {
					conn.Close()
					break
				}
			}
			if time.Now().After(deadline) {
				t.Error("proxy and live UI never became reachable")
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		cancel()
		runErr = <-done
	})

	if runErr != nil {
		t.Fatalf("proxy returned %v, want a clean shutdown with no anomalies", runErr)
	}
	for _, want := range []string{"Transparent Database Reverse Proxy", "Live Web UI", "Zero Concurrency Isolation Anomalies", "SARIF 2.1.0 report saved"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q:\n%s", want, output)
		}
	}
	if data, err := os.ReadFile(sarifPath); err != nil || !json.Valid(data) {
		t.Fatalf("SARIF file = %s, err = %v", data, err)
	}
}

func TestPrintProxySummary_ListsAnomalyCycles(t *testing.T) {
	output := captureStdout(t, func() {
		printProxySummary(proxy.ProxyStats{TotalSessions: 2}, []proxy.DetectedAnomaly{{
			Type:        domain.AnomalyLostUpdate,
			Description: "lost update on accounts:1",
			TxIDs:       []string{"S1-tx1", "S2-tx1"},
			Items:       []string{"accounts:1"},
			Cycle:       []analyzer.Edge{{From: "S1-tx1", To: "S2-tx1", Type: analyzer.DepWW, Item: "accounts:1"}},
		}})
	})
	for _, want := range []string{"1 Concurrency Isolation Anomalies Detected", "lost update on accounts:1", "S1-tx1 --(WW on accounts:1)--> S2-tx1"} {
		if !strings.Contains(output, want) {
			t.Errorf("summary missing %q:\n%s", want, output)
		}
	}
}
