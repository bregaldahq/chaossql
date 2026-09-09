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
	portFlag      int
	dbPathFlag    string
	tokenFlag     string
	publicURLFlag string
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
	rootCmd.Flags().StringVar(&tokenFlag, "token", getEnv("CHAOSSQL_ADMIN_TOKEN", "chaossql_dev_token"), "Admin/Default CI API token to seed")
	rootCmd.Flags().StringVar(&publicURLFlag, "public-url", getEnv("PUBLIC_URL", "http://localhost:8080"), "Public URL for dashboard links")

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the ChaosSQL Cloud control plane server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}
	startCmd.Flags().IntVarP(&portFlag, "port", "p", getEnvInt("PORT", 8080), "HTTP port to listen on")
	startCmd.Flags().StringVar(&dbPathFlag, "db", getEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")
	startCmd.Flags().StringVar(&tokenFlag, "token", getEnv("CHAOSSQL_ADMIN_TOKEN", "chaossql_dev_token"), "Admin/Default CI API token to seed")
	startCmd.Flags().StringVar(&publicURLFlag, "public-url", getEnv("PUBLIC_URL", "http://localhost:8080"), "Public URL for dashboard links")

	var orgFlag string
	var nameFlag string
	tokenCmd := &cobra.Command{
		Use:   "create-token",
		Short: "Generate a new CI API token for an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := sql.Open("sqlite", dbPathFlag)
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

			fmt.Println("=== ChaosSQL API Token Created ===")
			fmt.Printf("Organization: %s\n", orgFlag)
			fmt.Printf("Token Name:   %s\n", nameFlag)
			fmt.Printf("API Token:    %s\n", rawToken)
			fmt.Println("===================================")
			return nil
		},
	}
	tokenCmd.Flags().StringVar(&dbPathFlag, "db", getEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")
	tokenCmd.Flags().StringVar(&orgFlag, "org", "org_default", "Organization ID")
	tokenCmd.Flags().StringVar(&nameFlag, "name", "CI Token", "Token descriptive name")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(tokenCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer() error {
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

	// Seed default token if provided
	if tokenFlag != "" {
		_ = store.CreateOrganization("org_default", "Default Organization", "pro")
		_ = store.CreateAPIToken("tok_admin", "org_default", tokenFlag, "Initial Admin Token")
		log.Printf("[ChaosSQL Cloud] Admin token configured.")
	}

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

	log.Printf("[ChaosSQL Cloud] Server listening on http://0.0.0.0:%d (Public URL: %s)", portFlag, publicURLFlag)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-idleConnsClosed
	log.Println("[ChaosSQL Cloud] Server stopped cleanly.")
	return nil
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
