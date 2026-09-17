package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bregaldahq/chaossql/internal/server"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var (
	serverPortFlag      int
	serverDBPathFlag    string
	serverTokenFlag     string
	serverPublicURLFlag string
	serverStaticDirFlag string
)

func newServerCmd() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "ChaosSQL SaaS Control Plane, Concurrency Gate & Web Dashboard API",
		Long: `ChaosSQL Server is the enterprise self-hosted control plane for concurrency observability.
It ingests chaos test executions, stores baselines, evaluates regressions, dispatches
real-time multi-channel alerts (Discord, Slack, Generic Webhooks), and serves the visualizer dashboard.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runControlPlaneServer(cmd.OutOrStdout())
		},
	}

	serverCmd.PersistentFlags().IntVarP(&serverPortFlag, "port", "p", getServerEnvInt("PORT", 8080), "HTTP port to listen on")
	serverCmd.PersistentFlags().StringVar(&serverDBPathFlag, "db", getServerEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")
	serverCmd.PersistentFlags().StringVar(&serverTokenFlag, "token", getServerEnv("CHAOSSQL_ADMIN_TOKEN", "chaossql_dev_token"), "Admin/Default CI API token to seed")
	serverCmd.PersistentFlags().StringVar(&serverPublicURLFlag, "public-url", getServerEnv("PUBLIC_URL", "http://localhost:8080"), "Public URL for dashboard links")
	serverCmd.PersistentFlags().StringVar(&serverStaticDirFlag, "static-dir", getServerEnv("STATIC_DIR", ""), "Path to static directory to serve dashboard web assets")

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the ChaosSQL Cloud control plane server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runControlPlaneServer(cmd.OutOrStdout())
		},
	}

	var orgFlag string
	var nameFlag string
	tokenCmd := &cobra.Command{
		Use:   "create-token",
		Short: "Generate a new CI API token for an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := sql.Open("sqlite", serverDBPathFlag)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			store := server.NewStore(db)
			if err := store.AutoMigrate(); err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}

			if orgFlag == "" {
				orgFlag = "org_default"
			}
			_ = store.CreateOrganization(orgFlag, "Default Org", "pro")

			tokenBytes := make([]byte, 24)
			_, _ = rand.Read(tokenBytes)
			rawToken := "csql_" + hex.EncodeToString(tokenBytes)
			tokenID := fmt.Sprintf("tok_%d", time.Now().UnixNano())

			if err := store.CreateAPIToken(tokenID, orgFlag, rawToken, nameFlag); err != nil {
				return fmt.Errorf("failed to store token: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "=== ChaosSQL API Token Created ===")
			fmt.Fprintf(out, "Organization: %s\n", orgFlag)
			fmt.Fprintf(out, "Token Name:   %s\n", nameFlag)
			fmt.Fprintf(out, "API Token:    %s\n", rawToken)
			fmt.Fprintln(out, "===================================")
			return nil
		},
	}
	tokenCmd.Flags().StringVar(&orgFlag, "org", "org_default", "Organization ID")
	tokenCmd.Flags().StringVar(&nameFlag, "name", "CI Token", "Token descriptive name")

	serverCmd.AddCommand(startCmd, tokenCmd)
	return serverCmd
}

func runControlPlaneServer(out any) error {
	log.Printf("[ChaosSQL Cloud] Initializing database at %s...", serverDBPathFlag)
	db, err := sql.Open("sqlite", serverDBPathFlag)
	if err != nil {
		return fmt.Errorf("failed to open sqlite database: %w", err)
	}
	defer db.Close()

	store := server.NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	if serverTokenFlag != "" {
		_ = store.CreateOrganization("org_default", "Default Organization", "pro")
		_ = store.CreateAPIToken("tok_admin", "org_default", serverTokenFlag, "Initial Admin Token")
		log.Printf("[ChaosSQL Cloud] Admin token configured.")
	}

	engine := server.NewRegressionEngine(store)
	router := server.NewRouter(server.RouterConfig{
		Store:         store,
		Engine:        engine,
		PublicBaseURL: serverPublicURLFlag,
	})

	// If static directory is specified, mount static file server on fallback routes
	var handler http.Handler = router
	if serverStaticDirFlag != "" {
		if fi, err := os.Stat(serverStaticDirFlag); err == nil && fi.IsDir() {
			fs := http.FileServer(http.Dir(serverStaticDirFlag))
			mux := http.NewServeMux()
			mux.Handle("/v1/", router)
			mux.Handle("/health", router)
			mux.Handle("/", fs)
			handler = mux
			log.Printf("[ChaosSQL Cloud] Serving static dashboard from %s", serverStaticDirFlag)
		}
	}

	addr := fmt.Sprintf(":%d", serverPortFlag)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("[ChaosSQL Cloud] Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("[ChaosSQL Cloud] Shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("[ChaosSQL Cloud] Server listening on http://0.0.0.0:%d (Public URL: %s)", serverPortFlag, serverPublicURLFlag)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-idleConnsClosed
	log.Println("[ChaosSQL Cloud] Server stopped cleanly.")
	return nil
}

func getServerEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getServerEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
