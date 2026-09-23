package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bregaldahq/chaossql/internal/server"
	"github.com/bregaldahq/chaossql/internal/serveradmin"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var (
	portFlag      int
	dbPathFlag    string
	tokenFlag     string
	publicURLFlag string
	retentionFlag bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chaossql-server",
		Short: "ChaosSQL SaaS Control Plane & Regression Detection API",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}

	rootCmd.Flags().IntVarP(&portFlag, "port", "p", getEnvInt("PORT", 8080), "HTTP port to listen on")
	rootCmd.Flags().StringVar(&dbPathFlag, "db", getEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")
	rootCmd.Flags().StringVar(&tokenFlag, "token", getEnv("CHAOSSQL_ADMIN_TOKEN", ""), "Required initial owner API token")
	rootCmd.Flags().StringVar(&publicURLFlag, "public-url", getEnv("PUBLIC_URL", "http://localhost:8080"), "Public URL for dashboard links")
	rootCmd.Flags().BoolVar(&retentionFlag, "enforce-retention", getEnv("CHAOSSQL_ENFORCE_RETENTION", "") == "true", "Delete run history older than each organization's plan retention window")

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the ChaosSQL Cloud control plane server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}
	startCmd.Flags().IntVarP(&portFlag, "port", "p", getEnvInt("PORT", 8080), "HTTP port to listen on")
	startCmd.Flags().StringVar(&dbPathFlag, "db", getEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")
	startCmd.Flags().StringVar(&tokenFlag, "token", getEnv("CHAOSSQL_ADMIN_TOKEN", ""), "Required initial owner API token")
	startCmd.Flags().StringVar(&publicURLFlag, "public-url", getEnv("PUBLIC_URL", "http://localhost:8080"), "Public URL for dashboard links")
	startCmd.Flags().BoolVar(&retentionFlag, "enforce-retention", getEnv("CHAOSSQL_ENFORCE_RETENTION", "") == "true", "Delete run history older than each organization's plan retention window")

	tokenCmd := serveradmin.NewCreateTokenCommand(&dbPathFlag)
	tokenCmd.Flags().StringVar(&dbPathFlag, "db", getEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")
	orgCmd := serveradmin.NewOrgCommand(&dbPathFlag)
	orgCmd.PersistentFlags().StringVar(&dbPathFlag, "db", getEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")

	rootCmd.AddCommand(startCmd, tokenCmd, orgCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer() error {
	if err := server.ValidateBootstrapToken(tokenFlag); err != nil {
		return err
	}
	log.Printf("[ChaosSQL Cloud] Initializing database at %s...", dbPathFlag)
	db, err := sql.Open("sqlite", dbPathFlag)
	if err != nil {
		return fmt.Errorf("failed to open sqlite database: %w", err)
	}
	defer db.Close()

	store := server.NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	if err := configureBootstrapOwner(store, tokenFlag); err != nil {
		return err
	}
	log.Printf("[ChaosSQL Cloud] Admin token configured.")

	engine := server.NewRegressionEngine(store)
	router := server.NewRouter(server.RouterConfig{
		Store:         store,
		Engine:        engine,
		PublicBaseURL: publicURLFlag,
	})

	addr := fmt.Sprintf(":%d", portFlag)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	idleConnsClosed := make(chan struct{})
	shutdownContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	go func() {
		<-shutdownContext.Done()

		log.Println("[ChaosSQL Cloud] Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("[ChaosSQL Cloud] Shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	if retentionFlag {
		server.StartRetention(shutdownContext, store, log.Printf)
		log.Printf("[ChaosSQL Cloud] Plan retention enforcement enabled.")
	}

	log.Printf("[ChaosSQL Cloud] Server listening on http://0.0.0.0:%d (Public URL: %s)", portFlag, publicURLFlag)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-idleConnsClosed
	log.Println("[ChaosSQL Cloud] Server stopped cleanly.")
	return nil
}

func configureBootstrapOwner(store *server.Store, token string) error {
	return server.ConfigureBootstrapOwner(store, token)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
