package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bregaldahq/chaossql/pkg/proxy"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	styleProxyHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#5B21B6")).
				Padding(0, 1)

	styleProxyCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2)

	styleProxyAnomaly = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#EF4444"))

	styleProxySuccess = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#10B981"))
)

func newProxyCmd() *cobra.Command {
	var (
		listenAddr       string
		upstreamAddr     string
		protocol         string
		jitterMin        time.Duration
		jitterMax        time.Duration
		commitBarrierMin time.Duration
		commitBarrierMax time.Duration
		pctDepth         int
		exportSARIF      string
		uiPort           int
		seed             uint64
		failOnAnomaly    bool
	)

	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Start transparent Layer-7 TCP database reverse proxy for runtime isolation fuzzing",
		Long: `Starts a transparent Layer-7 database reverse proxy that intercepts PostgreSQL Wire Protocol 3.0
or MySQL client/server wire traffic at runtime. It injects stochastic micro-jitter and commit barriers,
maintains a Live Shadow Serialization Graph (Shadow DSG), and surfaces concurrency isolation anomalies
(Lost Updates, Write Skew, Dirty Reads, Anti-Dependencies) in real application integration tests without
requiring a single line of application source code modification.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := proxy.ProxyConfig{
				ListenAddr:       listenAddr,
				UpstreamAddr:     upstreamAddr,
				Protocol:         proxy.ProtocolType(strings.ToLower(protocol)),
				MinJitter:        jitterMin,
				MaxJitter:        jitterMax,
				CommitBarrierMin: commitBarrierMin,
				CommitBarrierMax: commitBarrierMax,
				PCTDepth:         pctDepth,
				ExportSARIF:      exportSARIF,
				UIPort:           uiPort,
				Seed:             seed,
			}

			server, err := proxy.NewServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to initialize proxy: %w", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			stopCh := make(chan os.Signal, 1)
			signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

			go func() {
				if err := server.Start(ctx); err != nil {
					fmt.Fprintf(os.Stderr, "Proxy server stopped: %v\n", err)
				}
			}()

			var uiServerURL string
			if uiPort > 0 {
				uiAddr := fmt.Sprintf("127.0.0.1:%d", uiPort)
				uiServer, url, err := proxy.StartLiveUI(uiAddr, server)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to start live UI: %v\n", err)
				} else {
					uiServerURL = url
					defer func() {
						shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
						defer shutdownCancel()
						_ = uiServer.Shutdown(shutdownCtx)
					}()
				}
			}

			printProxyBanner(cfg, uiServerURL)

			<-stopCh
			fmt.Println("\n\nShutting down ChaosSQL Transparent Proxy...")

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer shutdownCancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				fmt.Fprintf(os.Stderr, "Proxy shutdown error: %v\n", err)
			}

			stats := server.Stats()
			anomalies := server.Engine().GetAnomalies()

			if exportSARIF != "" {
				if err := proxy.ExportSARIFToFile(anomalies, "proxy-session", exportSARIF); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to export SARIF report: %v\n", err)
				} else {
					fmt.Printf("✔ OASIS SARIF 2.1.0 report saved to: %s\n", exportSARIF)
				}
			}

			printProxySummary(stats, anomalies)

			if failOnAnomaly && len(anomalies) > 0 {
				return fmt.Errorf("concurrency isolation anomalies detected: %d", len(anomalies))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:5433", "Local address for proxy to listen on")
	cmd.Flags().StringVar(&upstreamAddr, "upstream", "127.0.0.1:5432", "Upstream target database address (host:port)")
	cmd.Flags().StringVar(&protocol, "protocol", "postgres", "Database wire protocol (postgres or mysql)")
	cmd.Flags().DurationVar(&jitterMin, "jitter-min", 10*time.Microsecond, "Minimum stochastic jitter delay for DML")
	cmd.Flags().DurationVar(&jitterMax, "jitter-max", 2*time.Millisecond, "Maximum stochastic jitter delay for DML")
	cmd.Flags().DurationVar(&commitBarrierMin, "commit-barrier-min", 50*time.Microsecond, "Minimum commit barrier delay")
	cmd.Flags().DurationVar(&commitBarrierMax, "commit-barrier-max", 5*time.Millisecond, "Maximum commit barrier delay")
	cmd.Flags().IntVar(&pctDepth, "pct-depth", 2, "Probabilistic Concurrency Testing (PCT) priority depth")
	cmd.Flags().StringVar(&exportSARIF, "export-sarif", "", "File path to write OASIS SARIF 2.1.0 anomaly report")
	cmd.Flags().IntVar(&uiPort, "ui-port", 0, "Port for optional live browser trace visualizer and metrics dashboard")
	cmd.Flags().Uint64Var(&seed, "seed", 42, "PRNG seed for deterministic scheduling")
	cmd.Flags().BoolVar(&failOnAnomaly, "fail-on-anomaly", true, "Exit with non-zero status code if anomalies are detected")

	return cmd
}

func printProxyBanner(cfg proxy.ProxyConfig, uiURL string) {
	fmt.Println()
	fmt.Println(styleProxyHeader.Render(" ChaosSQL Layer-7 Transparent Database Reverse Proxy "))
	fmt.Println()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("• Wire Protocol:    %s\n", cfg.Protocol))
	sb.WriteString(fmt.Sprintf("• Proxy Listening:  %s\n", cfg.ListenAddr))
	sb.WriteString(fmt.Sprintf("• Target Upstream:  %s\n", cfg.UpstreamAddr))
	sb.WriteString(fmt.Sprintf("• Stochastic Jitter: %v – %v (PCT depth: %d)\n", cfg.MinJitter, cfg.MaxJitter, cfg.PCTDepth))
	sb.WriteString(fmt.Sprintf("• Commit Barrier:   %v – %v\n", cfg.CommitBarrierMin, cfg.CommitBarrierMax))
	if uiURL != "" {
		sb.WriteString(fmt.Sprintf("• Live Web UI:      %s\n", uiURL))
	}
	sb.WriteString("\nReady for incoming client connections. Run your test suite now.\nPress Ctrl+C to stop.")

	fmt.Println(styleProxyCard.Render(sb.String()))
	fmt.Println()
}

func printProxySummary(stats proxy.ProxyStats, anomalies []proxy.DetectedAnomaly) {
	fmt.Println()
	fmt.Println(styleProxyHeader.Render(" Proxy Execution Telemetry & Anomaly Report "))
	fmt.Println()

	fmt.Printf("• Total Client Sessions:    %d\n", stats.TotalSessions)
	fmt.Printf("• Total Wire Queries:       %d\n", stats.TotalQueries)
	fmt.Printf("• Intercepted DML Actions:  %d\n", stats.InterceptedDML)
	fmt.Printf("• Jitter Delay Injections:  %d (Total Delay: %v)\n", stats.JitterInjectedCount, stats.TotalJitterDelay.Round(time.Millisecond))
	fmt.Println()

	if len(anomalies) == 0 {
		fmt.Println(styleProxySuccess.Render("✔ Zero Concurrency Isolation Anomalies Detected! All transactions serializable."))
	} else {
		fmt.Println(styleProxyAnomaly.Render(fmt.Sprintf("💥 %d Concurrency Isolation Anomalies Detected:", len(anomalies))))
		for i, a := range anomalies {
			fmt.Printf("\n  %d. [%s] %s\n", i+1, a.Type, a.Description)
			fmt.Printf("     Transactions: %s | Items: %s\n", strings.Join(a.TxIDs, ", "), strings.Join(a.Items, ", "))
			if len(a.Cycle) > 0 {
				for _, edge := range a.Cycle {
					fmt.Printf("     --> %s --(%s on %s)--> %s\n", edge.From, edge.Type, edge.Item, edge.To)
				}
			}
		}
	}
	fmt.Println()
}
