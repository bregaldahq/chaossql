package main

import (
	"bytes"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/server"
)

func TestServerRejectsMissingTokenBeforeCreatingDatabase(t *testing.T) {
	t.Setenv("CHAOSSQL_ADMIN_TOKEN", "")
	path := filepath.Join(t.TempDir(), "missing", "server.db")
	cmd := newServerCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"start", "--token=", "--db=" + path, "--port=-1"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "token is required") {
		t.Fatalf("expected missing token error before database access, got %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing credential created a database: %v", err)
	}
}

func TestServerBootstrapGrantsOwnerAndRotatesToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.db")
	for _, token := range []string{"original-review-owner", "rotated-review-owner"} {
		cmd := newServerCmd()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		// Fail the listener after startup to inspect the real bootstrap path.
		cmd.SetArgs([]string{"start", "--token=" + token, "--db=" + path, "--port=-1"})
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "invalid port") {
			t.Fatalf("expected listener failure after bootstrap, got %v", err)
		}
		db, err := sql.Open("sqlite", path)
		if err != nil {
			t.Fatal(err)
		}
		store := server.NewStore(db)
		principal, err := store.AuthenticateToken(token)
		if err != nil {
			db.Close()
			t.Fatalf("new token was not activated: %v", err)
		}
		if principal.Role != server.RoleOwner {
			db.Close()
			t.Fatalf("bootstrap role = %s; want owner", principal.Role)
		}
		if token == "rotated-review-owner" {
			if _, err := store.AuthenticateToken("original-review-owner"); !errors.Is(err, server.ErrUnauthorized) {
				db.Close()
				t.Fatalf("old token remains valid: %v", err)
			}
		}
		db.Close()
	}
}

func TestServerRejectsPublishedExampleTokens(t *testing.T) {
	for _, token := range []string{"chaossql_dev_token", "chaossql_prod_secret", "chaossql_prod_secret_token_change_me", "chaossql_enterprise_token_secret"} {
		t.Run(token, func(t *testing.T) {
			cmd := newServerCmd()
			cmd.SetOut(new(bytes.Buffer))
			cmd.SetErr(new(bytes.Buffer))
			cmd.SetArgs([]string{"start", "--token=" + token, "--db=" + filepath.Join(t.TempDir(), "missing", "server.db"), "--port=-1"})
			if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "public example token") {
				t.Fatalf("expected rejection of public example token, got %v", err)
			}
		})
	}
}

func TestServerAcceptsRetentionEnforcement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "retention.db")
	cmd := newServerCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"start", "--enforce-retention", "--token=retention-review-owner", "--db=" + path, "--port=-1"})
	// The listener fails after bootstrap and retention startup.
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "invalid port") {
		t.Fatalf("expected listener failure after retention startup, got %v", err)
	}
}
