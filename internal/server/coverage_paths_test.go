package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newAdminTestServer(t *testing.T) (http.Handler, *Store, string) {
	t.Helper()
	handler, store, _ := newTestServer(t)
	adminToken := "chaossql_admin_token_888"
	if err := store.CreateAPITokenWithRole("tok_admin_888", "org_cloud_test", adminToken, "Admin", RoleAdmin); err != nil {
		t.Fatal(err)
	}
	return handler, store, adminToken
}

// execWithoutForeignKeys applies a destructive fixture statement that the
// schema's foreign keys would otherwise refuse. The test store uses a single
// connection, so the pragma applies to the statement.
func execWithoutForeignKeys(t *testing.T, store *Store, statement string) {
	t.Helper()
	for _, stmt := range []string{`PRAGMA foreign_keys = OFF`, statement, `PRAGMA foreign_keys = ON`} {
		if _, err := store.db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
}

func doRequest(handler http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestValidateBootstrapToken(t *testing.T) {
	cases := map[string]string{
		"":                       "is required",
		"   ":                    "is required",
		"chaossql_dev_token":     "public example token",
		" chaossql_prod_secret ": "public example token",
		" unique-secret":         "surrounding whitespace",
		"unique-secret-value":    "",
	}
	for token, wantErr := range cases {
		err := ValidateBootstrapToken(token)
		if wantErr == "" && err != nil {
			t.Errorf("ValidateBootstrapToken(%q) = %v, want nil", token, err)
		}
		if wantErr != "" && (err == nil || !strings.Contains(err.Error(), wantErr)) {
			t.Errorf("ValidateBootstrapToken(%q) = %v, want error containing %q", token, err, wantErr)
		}
	}
}

func TestConfigureBootstrapOwnerRotatesOwnerToken(t *testing.T) {
	_, store, _ := newTestServer(t)
	if err := ConfigureBootstrapOwner(store, "chaossql_dev_token"); err == nil {
		t.Fatal("a public example token must be rejected")
	}
	if err := ConfigureBootstrapOwner(store, "first-owner-secret"); err != nil {
		t.Fatal(err)
	}
	if err := ConfigureBootstrapOwner(store, "second-owner-secret"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AuthenticateToken("first-owner-secret"); err == nil {
		t.Fatal("the rotated-out owner token must stop authenticating")
	}
	principal, err := store.AuthenticateToken("second-owner-secret")
	if err != nil || principal.OrgID != "org_default" || principal.Role != RoleOwner {
		t.Fatalf("principal = %+v, err = %v; want the default org owner", principal, err)
	}
}

func TestConfigureBootstrapOwnerReportsStorageFailure(t *testing.T) {
	_, store, _ := newTestServer(t)
	execWithoutForeignKeys(t, store, `DROP TABLE organizations`)
	if err := ConfigureBootstrapOwner(store, "owner-secret"); err == nil || !strings.Contains(err.Error(), "bootstrap organization") {
		t.Fatalf("err = %v, want an organization failure", err)
	}
}

func TestStartRetentionLogsFailedPass(t *testing.T) {
	_, store, _ := newTestServer(t)
	execWithoutForeignKeys(t, store, `DROP TABLE organizations`)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var mu sync.Mutex
	var logged []string
	done := make(chan struct{})
	StartRetention(ctx, store, func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		logged = append(logged, fmt.Sprintf(format, args...))
		if len(logged) == 1 {
			close(done)
		}
	})
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("retention never reported its first pass")
	}
	mu.Lock()
	defer mu.Unlock()
	if !strings.Contains(logged[0], "Retention pass failed") {
		t.Fatalf("log = %q, want a failed pass", logged[0])
	}
}

func TestHandlersReportStorageFailuresAsServerErrors(t *testing.T) {
	cases := []struct {
		name, method, path, body, dropTable string
	}{
		{"list all runs", http.MethodGet, "/v1/runs", "", "runs"},
		{"get run", http.MethodGet, "/v1/runs/run_1", "", "runs"},
		{"list repository runs", http.MethodGet, "/v1/repositories/acme/payments/runs", "", "repositories"},
		{"subscription", http.MethodGet, "/v1/organizations/me/subscription", "", "repositories"},
		{"list webhooks", http.MethodGet, "/v1/organizations/me/webhooks", "", "webhooks"},
		{"create webhook", http.MethodPost, "/v1/organizations/me/webhooks", `{"url":"https://hooks.slack.com/services/T/B/C"}`, "webhooks"},
		{"delete webhook", http.MethodDelete, "/v1/organizations/me/webhooks/wh_1", "", "webhooks"},
		{"test saved webhook", http.MethodPost, "/v1/organizations/me/webhooks/wh_1/test", "", "webhooks"},
		{"create member token", http.MethodPost, "/v1/organizations/me/tokens", `{"name":"ci"}`, "api_tokens_shadow"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler, store, adminToken := newAdminTestServer(t)
			if tc.dropTable == "api_tokens_shadow" {
				// Authentication reads api_tokens, so make only the INSERT fail.
				if _, err := store.db.Exec(`CREATE TRIGGER block_token_insert BEFORE INSERT ON api_tokens BEGIN SELECT RAISE(ABORT, 'read only'); END`); err != nil {
					t.Fatal(err)
				}
			} else {
				execWithoutForeignKeys(t, store, `DROP TABLE `+tc.dropTable)
			}
			w := doRequest(handler, tc.method, tc.path, adminToken, tc.body)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, body = %s; want 500", w.Code, w.Body.String())
			}
		})
	}
}

func TestSubscriptionForDeletedOrganizationIsNotFound(t *testing.T) {
	handler, store, token := newTestServer(t)
	execWithoutForeignKeys(t, store, `DELETE FROM organizations WHERE id = 'org_cloud_test'`)
	if w := doRequest(handler, http.MethodGet, "/v1/organizations/me/subscription", token, ""); w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestCreateMemberTokenValidatesBody(t *testing.T) {
	handler, store, adminToken := newAdminTestServer(t)
	for body, want := range map[string]int{
		`not json`:                                    http.StatusBadRequest,
		`{"name":"ci","role":"owner"}`:                http.StatusBadRequest,
		`{"name":"ci"} {"name":"again"}`:              http.StatusBadRequest,
		`{"name":"   "}`:                              http.StatusBadRequest,
		`{"name":"` + strings.Repeat("x", 129) + `"}`: http.StatusBadRequest,
	} {
		if w := doRequest(handler, http.MethodPost, "/v1/organizations/me/tokens", adminToken, body); w.Code != want {
			t.Errorf("body %.40q: status = %d, want %d", body, w.Code, want)
		}
	}

	w := doRequest(handler, http.MethodPost, "/v1/organizations/me/tokens", adminToken, `{"name":" nightly CI "}`)
	if w.Code != http.StatusCreated || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("status = %d, cache = %q", w.Code, w.Header().Get("Cache-Control"))
	}
	var created map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	principal, err := store.AuthenticateToken(created["token"])
	if err != nil || principal.Role != RoleMember || principal.OrgID != "org_cloud_test" {
		t.Fatalf("new token principal = %+v, err = %v", principal, err)
	}
}

func TestWebhookHandlersValidateAndDefault(t *testing.T) {
	handler, store, adminToken := newAdminTestServer(t)

	if w := doRequest(handler, http.MethodPost, "/v1/organizations/me/webhooks", adminToken, `{`); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid create body: status = %d", w.Code)
	}
	w := doRequest(handler, http.MethodPost, "/v1/organizations/me/webhooks", adminToken,
		`{"url":"https://hooks.slack.com/services/T/B/C","active":false}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status = %d, body = %s", w.Code, w.Body.String())
	}
	hooks, err := store.ListWebhooks(context.Background(), "org_cloud_test")
	if err != nil || len(hooks) != 1 {
		t.Fatalf("hooks = %+v, err = %v", hooks, err)
	}
	if hooks[0].TargetType != "generic" || hooks[0].Events != "regression,all" || hooks[0].Active {
		t.Fatalf("hook = %+v, want generic defaults and inactive", hooks[0])
	}

	if w := doRequest(handler, http.MethodDelete, "/v1/organizations/me/webhooks/wh_missing", adminToken, ""); w.Code != http.StatusNotFound {
		t.Fatalf("delete missing: status = %d", w.Code)
	}
	if w := doRequest(handler, http.MethodDelete, "/v1/organizations/me/webhooks/"+hooks[0].ID, adminToken, ""); w.Code != http.StatusNoContent {
		t.Fatalf("delete: status = %d", w.Code)
	}
	if w := doRequest(handler, http.MethodPost, "/v1/organizations/me/webhooks/wh_missing/test", adminToken, ""); w.Code != http.StatusNotFound {
		t.Fatalf("test missing saved webhook: status = %d", w.Code)
	}
	if w := doRequest(handler, http.MethodPost, "/v1/organizations/me/webhooks/test", adminToken, `{`); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid test body: status = %d", w.Code)
	}
	if w := doRequest(handler, http.MethodPost, "/v1/organizations/me/webhooks/test", adminToken, `{"url":"http://169.254.169.254/latest"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("disallowed test destination: status = %d", w.Code)
	}
}

func TestGetRunRejectsForeignOrganization(t *testing.T) {
	handler, _, token := newTestServer(t)
	if w := doRequest(handler, http.MethodGet, "/v1/organizations/org_other/subscription", token, ""); w.Code != http.StatusNotFound {
		t.Fatalf("foreign org: status = %d, want 404", w.Code)
	}
	if w := doRequest(handler, http.MethodGet, "/v1/runs/run_missing", token, ""); w.Code != http.StatusNotFound {
		t.Fatalf("missing run: status = %d, want 404", w.Code)
	}
}
