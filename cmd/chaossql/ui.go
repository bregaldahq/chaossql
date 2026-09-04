package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/spf13/cobra"
)

type uiTraceFilePayload struct {
	Spec              *domain.Spec             `json:"spec,omitempty"`
	Trace             domain.ExecutionTrace    `json:"trace,omitempty"`
	ScheduledOps      []domain.ScheduledOp     `json:"scheduled_ops,omitempty"`
	Shrink            *domain.ShrinkResult     `json:"shrink,omitempty"`
	Invariants        []domain.InvariantResult `json:"invariants,omitempty"`
	FailingInvariant  *domain.InvariantResult  `json:"failing_invariant,omitempty"`
	AnomalyType       domain.AnomalyType       `json:"anomaly_type,omitempty"`
	ViolationDetected bool                     `json:"violation_detected"`
}

func newUICmd() *cobra.Command {
	var port int
	var noOpen bool

	cmd := &cobra.Command{
		Use:   "ui <trace.json>",
		Short: "Launch the local editorial trace visualizer web server for inspecting execution traces",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tracePath := args[0]
			trace, spec, graph, shrink, invResults, err := loadTraceData(tracePath)
			if err != nil {
				return fmt.Errorf("failed to load trace data from %s: %w", tracePath, err)
			}

			htmlContent := reporter.GenerateEmbeddedTraceViewerHTML(trace, spec, graph, shrink, invResults)
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			return serveTraceViewer(addr, htmlContent, noOpen)
		},
	}

	cmd.Flags().IntVar(&port, "port", 8090, "Local HTTP server port")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not automatically open browser")

	return cmd
}

func loadTraceData(filePath string) (
	trace domain.ExecutionTrace,
	spec domain.Spec,
	graph *analyzer.AdyaGraph,
	shrink *domain.ShrinkResult,
	invResults []domain.InvariantResult,
	err error,
) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, domain.Spec{}, nil, nil, nil, fmt.Errorf("read file error: %w", err)
	}

	var payload uiTraceFilePayload
	if err := json.Unmarshal(data, &payload); err == nil && len(payload.Trace) > 0 {
		trace = payload.Trace
		if payload.Spec != nil {
			spec = *payload.Spec
		} else {
			spec = domain.Spec{Name: stringsTrimExt(filepath.Base(filePath))}
		}
		shrink = payload.Shrink
		invResults = payload.Invariants
		if payload.FailingInvariant != nil {
			invResults = append(invResults, *payload.FailingInvariant)
		}
	} else {
		// Try parsing direct ExecutionTrace array
		var directTrace domain.ExecutionTrace
		if traceErr := json.Unmarshal(data, &directTrace); traceErr == nil && len(directTrace) > 0 {
			trace = directTrace
			spec = domain.Spec{Name: stringsTrimExt(filepath.Base(filePath))}
		} else {
			// Try parsing ExecutionResult
			var execResult domain.ExecutionResult
			if resErr := json.Unmarshal(data, &execResult); resErr == nil && len(execResult.Trace) > 0 {
				trace = execResult.Trace
				spec = domain.Spec{Name: stringsTrimExt(filepath.Base(filePath))}
				if execResult.FailingInvariant != nil {
					invResults = append(invResults, *execResult.FailingInvariant)
				}
			} else {
				return nil, domain.Spec{}, nil, nil, nil, fmt.Errorf("unrecognized trace JSON format: %w", err)
			}
		}
	}

	if len(trace) > 0 {
		graph = analyzer.BuildGraph(trace)
	} else {
		graph = analyzer.NewAdyaGraph()
	}

	return trace, spec, graph, shrink, invResults, nil
}

func stringsTrimExt(fileName string) string {
	ext := filepath.Ext(fileName)
	return fileName[:len(fileName)-len(ext)]
}

func startTraceViewerServer(addr string, htmlContent string) (*http.Server, string, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(htmlContent))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		_ = server.Serve(ln)
	}()

	actualURL := fmt.Sprintf("http://%s", ln.Addr().String())
	return server, actualURL, nil
}

func serveTraceViewer(addr string, htmlContent string, noOpen bool) error {
	server, serverURL, err := startTraceViewerServer(addr, htmlContent)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("  │  ChaosSQL Embedded Trace Viewer                                 │")
	fmt.Printf("  │  URL: %-58s│\n", serverURL)
	fmt.Println("  │  Studio Bregalda Design System (Calm, editorial, zero-deps)     │")
	fmt.Println("  │  Press Ctrl+C to terminate                                      │")
	fmt.Println("  └─────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	if !noOpen {
		_ = openBrowser(serverURL)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down ChaosSQL Trace Viewer server...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		// Check for WSL wslview first
		if _, err := exec.LookPath("wslview"); err == nil {
			cmd = exec.Command("wslview", url)
		} else if _, err := exec.LookPath("xdg-open"); err == nil {
			cmd = exec.Command("xdg-open", url)
		}
	}
	if cmd != nil {
		return cmd.Start()
	}
	return nil
}
