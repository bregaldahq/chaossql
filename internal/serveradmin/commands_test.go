package serveradmin

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/server"
	"github.com/spf13/cobra"
)

func run(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func openStore(t *testing.T, path string) *server.Store {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return server.NewStore(db)
}

func TestOrgLifecycleCommands(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "admin.db")

	out, err := run(t, NewOrgCommand(&dbPath), "create", "--name", "Acme Payments", "--plan", "team")
	if err != nil {
		t.Fatalf("org create: %v\n%s", err, out)
	}
	orgID := regexp.MustCompile(`Organization: (org_[0-9a-f]+)`).FindStringSubmatch(out)
	owner := regexp.MustCompile(`Owner Token:  (csql_[0-9a-f]+)`).FindStringSubmatch(out)
	if orgID == nil || owner == nil {
		t.Fatalf("org create output missing identifiers:\n%s", out)
	}
	principal, err := openStore(t, dbPath).AuthenticateToken(owner[1])
	if err != nil || principal.OrgID != orgID[1] || principal.Role != server.RoleOwner {
		t.Fatalf("owner token principal=%+v err=%v", principal, err)
	}

	out, err = run(t, NewCreateTokenCommand(&dbPath), "--org", orgID[1], "--name", "CI")
	if err != nil || !strings.Contains(out, "Role:         member") {
		t.Fatalf("create-token: %v\n%s", err, out)
	}

	if out, err = run(t, NewOrgCommand(&dbPath), "set-plan", orgID[1], "pro"); err != nil {
		t.Fatalf("set-plan: %v\n%s", err, out)
	}
	out, err = run(t, NewOrgCommand(&dbPath), "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !regexp.MustCompile(orgID[1] + `\s+Acme Payments\s+pro\s+0\s+2\s`).MatchString(out) {
		t.Fatalf("list output does not show plan and usage:\n%s", out)
	}
	if strings.Contains(out, "csql_") {
		t.Fatalf("list output leaks credentials:\n%s", out)
	}
}

func TestCreateTokenRejectsUnknownOrganization(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "admin.db")
	out, err := run(t, NewCreateTokenCommand(&dbPath), "--org", "org_typo", "--name", "CI")
	if err == nil || strings.Contains(out, "csql_") {
		t.Fatalf("issued token for unknown organization: err=%v\n%s", err, out)
	}
	orgs, err := openStore(t, dbPath).ListOrganizations()
	if err != nil || len(orgs) != 0 {
		t.Fatalf("unknown organization was created: %+v err=%v", orgs, err)
	}
}

func TestOrgCommandsRejectInvalidInput(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "admin.db")
	for _, args := range [][]string{
		{"create"},
		{"create", "--name", "Acme", "--plan", "gold"},
		{"set-plan", "org_missing", "pro"},
		{"set-plan", "only-one-arg"},
	} {
		if out, err := run(t, NewOrgCommand(&dbPath), args...); err == nil {
			t.Fatalf("accepted %v:\n%s", args, out)
		}
	}
}
