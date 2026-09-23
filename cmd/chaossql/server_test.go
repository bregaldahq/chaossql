package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/server"
	_ "modernc.org/sqlite"
)

func TestServerCmd_Help(t *testing.T) {
	cmd := newRootCmd()
	outBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)
	cmd.SetArgs([]string{"server", "--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error executing server --help, got: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, "ChaosSQL Server is the enterprise self-hosted control plane") {
		t.Errorf("expected output to contain description, got: %s", out)
	}
	if !strings.Contains(out, "--port") {
		t.Errorf("expected output to contain --port flag, got: %s", out)
	}
	if !strings.Contains(out, "--db") {
		t.Errorf("expected output to contain --db flag, got: %s", out)
	}
}

func TestServerCmd_CreateToken(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-cloud.db")

	execute := func(args ...string) (string, error) {
		cmd := newRootCmd()
		outBuf := new(bytes.Buffer)
		cmd.SetOut(outBuf)
		cmd.SetErr(outBuf)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return outBuf.String(), err
	}

	// A mistyped organization must fail instead of provisioning a new tenant.
	if out, err := execute("server", "create-token", "--db", dbPath, "--org", "org_enterprise", "--name", "Production CI"); err == nil {
		t.Fatalf("expected create-token to reject an unknown organization, got:\n%s", out)
	}

	out, err := execute("server", "org", "create", "--db", dbPath, "--name", "Enterprise", "--plan", "enterprise")
	if err != nil {
		t.Fatalf("expected no error executing org create, got: %v\n%s", err, out)
	}
	match := regexp.MustCompile(`Organization: (org_[0-9a-f]+)`).FindStringSubmatch(out)
	if match == nil {
		t.Fatalf("expected organization ID in output, got: %s", out)
	}

	out, err = execute("server", "create-token", "--db", dbPath, "--org", match[1], "--name", "Production CI")
	if err != nil {
		t.Fatalf("expected no error executing create-token, got: %v", err)
	}
	if !strings.Contains(out, "ChaosSQL API Token Created") {
		t.Errorf("expected output to indicate token creation, got: %s", out)
	}
	if !strings.Contains(out, match[1]) {
		t.Errorf("expected output to contain %s, got: %s", match[1], out)
	}
	if !strings.Contains(out, "csql_") {
		t.Errorf("expected output to contain raw token with prefix csql_, got: %s", out)
	}
}

func TestServerCmd_LiveServerHealth(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-server-health.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	store := server.NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	engine := server.NewRegressionEngine(store)
	router := server.NewRouter(server.RouterConfig{
		Store:         store,
		Engine:        engine,
		PublicBaseURL: "http://localhost:18081",
	})

	srv := &http.Server{
		Addr:    "127.0.0.1:18081",
		Handler: router,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	// Wait for server readiness
	var resp *http.Response
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		r, err := http.Get("http://127.0.0.1:18081/v1/health")
		if err == nil && r.StatusCode == http.StatusOK {
			resp = r
			break
		}
	}

	if resp == nil {
		t.Fatal("timed out waiting for server /v1/health endpoint")
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode health body: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got: %v", body["status"])
	}
}
