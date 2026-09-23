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

	var orgFlag string
	var nameFlag string
	var roleFlag string
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
			if err := store.EnsureOrganization(orgFlag, "Default Org", "pro"); err != nil {
				return fmt.Errorf("failed to configure organization: %w", err)
			}

			tokenBytes := make([]byte, 24)
			if _, err := rand.Read(tokenBytes); err != nil {
				return fmt.Errorf("failed to generate token: %w", err)
			}
			rawToken := "csql_" + hex.EncodeToString(tokenBytes)
			tokenID := fmt.Sprintf("tok_%d", time.Now().UnixNano())

			if err := store.CreateAPITokenWithRole(tokenID, orgFlag, rawToken, nameFlag, server.Role(roleFlag)); err != nil {
				return fmt.Errorf("failed to store token: %w", err)
			}

			fmt.Println("=== ChaosSQL API Token Created ===")
			fmt.Printf("Organization: %s\n", orgFlag)
			fmt.Printf("Token Name:   %s\n", nameFlag)
			fmt.Printf("Role:         %s\n", roleFlag)
			fmt.Printf("API Token:    %s\n", rawToken)
			fmt.Println("===================================")
			return nil
		},
	}
	tokenCmd.Flags().StringVar(&dbPathFlag, "db", getEnv("DB_PATH", "chaossql-cloud.db"), "Path to SQLite database file")
	tokenCmd.Flags().StringVar(&orgFlag, "org", "org_default", "Organization ID")
	tokenCmd.Flags().StringVar(&nameFlag, "name", "CI Token", "Token descriptive name")
	tokenCmd.Flags().StringVar(&roleFlag, "role", string(server.RoleMember), "Organization role: owner, admin, or member")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(tokenCmd)

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
