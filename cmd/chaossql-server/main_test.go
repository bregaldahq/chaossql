package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/server"
	_ "modernc.org/sqlite"
)

func TestGetEnvHelpers(t *testing.T) {
	t.Setenv("CHAOSSQL_TEST_STR", "value")
	t.Setenv("CHAOSSQL_TEST_INT", "9090")
	t.Setenv("CHAOSSQL_TEST_BAD_INT", "ninety")
	if got := getEnv("CHAOSSQL_TEST_STR", "fallback"); got != "value" {
		t.Errorf("getEnv set = %q", got)
	}
	if got := getEnv("CHAOSSQL_TEST_UNSET", "fallback"); got != "fallback" {
		t.Errorf("getEnv unset = %q", got)
	}
	if got := getEnvInt("CHAOSSQL_TEST_INT", 1); got != 9090 {
		t.Errorf("getEnvInt set = %d", got)
	}
	if got := getEnvInt("CHAOSSQL_TEST_BAD_INT", 1); got != 1 {
		t.Errorf("getEnvInt invalid = %d, want fallback", got)
	}
}

func TestServerCommandRejectsInvalidStartup(t *testing.T) {
	cases := map[string][]string{
		"missing token":        {"--db", filepath.Join(t.TempDir(), "a.db")},
		"public example token": {"--token", "chaossql_dev_token", "--db", filepath.Join(t.TempDir(), "b.db")},
		"unwritable database":  {"--token", "unique-owner-secret", "--db", filepath.Join(t.TempDir(), "missing", "dir", "c.db")},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := newRootCommand()
			cmd.SetArgs(append([]string{"start"}, args...))
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			if err := cmd.ExecuteContext(context.Background()); err == nil {
				t.Fatal("expected startup to fail")
			}
		})
	}
}

func TestServerCommandBootstrapsOwnerAndStopsOnCancel(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "cloud.db")
	port := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := newRootCommand()
	cmd.SetArgs([]string{"--token", "unique-owner-secret", "--db", dbPath, "--port", strconv.Itoa(port), "--enforce-retention"})
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	done := make(chan error, 1)
	go func() { done <- cmd.ExecuteContext(ctx) }()

	// Poll over HTTP only: opening the SQLite file here would contend with
	// the server's own migrations.
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/v1/health", port)
	deadline := time.Now().Add(10 * time.Second)
	for {
		resp, err := http.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		select {
		case err := <-done:
			t.Fatalf("server exited before serving: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("server never became healthy")
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("server returned %v, want a clean stop", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop after cancellation")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	principal, err := server.NewStore(db).AuthenticateToken("unique-owner-secret")
	if err != nil || principal.Role != server.RoleOwner {
		t.Fatalf("bootstrap owner = %+v, err = %v", principal, err)
	}
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}
