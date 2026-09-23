package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/cloud"
)

func correctionPayload() cloud.RunIngestRequest {
	return cloud.RunIngestRequest{Version: "1.0", Timestamp: time.Date(2026, 9, 17, 15, 0, 0, 0, time.UTC), CI: &cloud.CIContext{Provider: "github-actions", Repository: "acme/payments", CommitSHA: "abc123", Branch: "main", RunID: "workflow123"}, Scenario: cloud.ScenarioMetadata{Name: "transfer", Driver: "sqlite", Seed: 42, Fingerprint: "fp1"}, Result: cloud.ExecutionSummary{Status: "passed", Success: true, TotalSchedules: 10}}
}

func TestIngestionUsesCommitChronologyAndTrustedFingerprint(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	p := correctionPayload()
	newer := p.Timestamp.Add(-time.Hour)
	older := newer.Add(-time.Hour)
	p.CI.CommitTimestamp = &newer
	w, first := correctionPost(h, token, "trusted", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	p.Timestamp = p.Timestamp.Add(time.Hour)
	p.CI.CommitTimestamp = &older
	p.CI.CommitSHA = "old"
	w, _ = correctionPost(h, token, "old", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	p.CI.CommitTimestamp = nil
	p.Scenario.Fingerprint = ""
	p.CI.CommitSHA = "legacy"
	w, _ = correctionPost(h, token, "legacy", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	p.Scenario.Fingerprint = "changed"
	p.CI.CommitTimestamp = &newer
	w, _ = correctionPost(h, token, "incompatible", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	var baseline string
	if err := s.db.QueryRow(`SELECT run_id FROM baselines`).Scan(&baseline); err != nil {
		t.Fatal(err)
	}
	if baseline != first.RunID {
		t.Fatalf("trusted baseline replaced by %s", baseline)
	}
	p.CI.Branch = "feature"
	p.CI.PullRequestNumber = 12
	p.Result.Status = "failed"
	p.Scenario.Fingerprint = ""
	w, response := correctionPost(h, token, "unknown-pr", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	if response.IsRegression || response.Baseline != nil {
		t.Fatalf("unknown fingerprint compared to trusted baseline: %+v", response)
	}
	p.Scenario.Fingerprint = "fp1"
	w, response = correctionPost(h, token, "matching-pr", p)
	if w.Code != 201 || !response.IsRegression || response.Baseline == nil || response.Baseline.RunID != first.RunID {
		t.Fatalf("matching regression lost: %d %+v", w.Code, response)
	}
}

func TestIngestionBaselineWriteFailureRollsBack(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	if _, err := s.db.Exec(`CREATE TRIGGER fail_baseline BEFORE INSERT ON baselines BEGIN SELECT RAISE(ABORT,'baseline unavailable'); END`); err != nil {
		t.Fatal(err)
	}
	w, _ := correctionPost(h, token, "baseline", correctionPayload())
	if w.Code != 500 {
		t.Fatalf("expected error: %d %s", w.Code, w.Body)
	}
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM runs`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("run survived failed baseline write")
	}
}

func TestIngestionKeyBodyHeaderAgreement(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	p := correctionPayload()
	p.IdempotencyKey = "body-key"
	w, a := correctionPost(h, token, "", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	w, b := correctionPost(h, token, "body-key", p)
	if w.Code != 200 || a.RunID != b.RunID {
		t.Fatalf("body key replay failed: %d %s", w.Code, w.Body)
	}
	w, _ = correctionPost(h, token, "different", p)
	if w.Code != 400 {
		t.Fatalf("conflicting header accepted: %d", w.Code)
	}
}

func TestPublicRunMetadataAndSavedDecision(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	p := correctionPayload()
	_, _ = correctionPost(h, token, "base", p)
	p.CI.Branch = "feature"
	p.CI.PullRequestNumber = 1
	p.Result.Status = "failed"
	p.Result.FailedSchedules = 3
	w, run := correctionPost(h, token, "pr", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	for _, path := range []string{"/v1/runs", "/v1/repositories/acme/payments/runs", "/v1/runs/" + run.RunID} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatal(w.Body)
		}
		var envelope struct {
			Run  map[string]any   `json:"run"`
			Runs []map[string]any `json:"runs"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		records := envelope.Runs
		if envelope.Run != nil {
			records = []map[string]any{envelope.Run}
		}
		found := false
		for _, record := range records {
			if record["id"] != run.RunID {
				continue
			}
			found = true
			for key, want := range map[string]any{"repo_full_name": "acme/payments", "scenario_name": "transfer", "driver": "sqlite", "is_regression": true, "total_schedules": float64(10), "failed_schedules": float64(3)} {
				if record[key] != want {
					t.Errorf("%s %s = %v, want %v", path, key, record[key], want)
				}
			}
		}
		if !found {
			t.Fatal("run missing")
		}
	}
}

func TestPublicRunMissingMeasurementsRemainUnknown(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	r := httptest.NewRequest("POST", "/v1/runs", strings.NewReader(`{"scenario":{"name":"legacy","driver":"sqlite"},"result":{"status":"passed"}}`))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 201 {
		t.Fatalf("legacy ingestion: %d %s", w.Code, w.Body)
	}
	r = httptest.NewRequest("GET", "/v1/runs", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var result struct {
		Runs []map[string]any `json:"runs"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Runs) != 1 {
		t.Fatal(w.Body)
	}
	if result.Runs[0]["total_schedules"] != nil || result.Runs[0]["failed_schedules"] != nil {
		t.Fatalf("unknown measurements fabricated: %s", w.Body)
	}
}

func TestIngestionExplicitZeroConflictsWithOmittedMeasurement(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	for i, body := range []string{`{"scenario":{"name":"s"},"result":{"status":"passed"}}`, `{"scenario":{"name":"s"},"result":{"status":"passed","total_schedules":0}}`} {
		r := httptest.NewRequest("POST", "/v1/runs", strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Idempotency-Key", "same-key")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		want := 201
		if i == 1 {
			want = 409
		}
		if w.Code != want {
			t.Fatalf("request %d: %d %s", i, w.Code, w.Body)
		}
	}
}

func TestTrustedBaselineCanReplaceLegacyIdentity(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	p := correctionPayload()
	p.Scenario.Fingerprint = ""
	if w, _ := correctionPost(h, token, "legacy", p); w.Code != 201 {
		t.Fatal(w.Body)
	}
	p.Scenario.Fingerprint = "fp1"
	p.CI.CommitTimestamp = &p.Timestamp
	w, trusted := correctionPost(h, token, "trusted", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	var baseline string
	if err := s.db.QueryRow(`SELECT run_id FROM baselines`).Scan(&baseline); err != nil {
		t.Fatal(err)
	}
	if baseline != trusted.RunID {
		t.Fatalf("legacy identity prevented establishing trusted baseline: %s", baseline)
	}
	if trusted.Baseline != nil {
		t.Fatal("incompatible legacy run used as comparison")
	}
}

func TestSavedWebhookTestAuthorizationAndDestination(t *testing.T) {
	h, s, member := newTestServer(t)
	defer s.db.Close()
	if err := s.CreateAPITokenWithRole("admin", "org_cloud_test", "admin", "Admin", RoleAdmin); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateOrganization("other", "Other", "enterprise"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateWebhook(context.Background(), &WebhookRecord{ID: "foreign", OrgID: "other", TargetType: "generic", URL: "https://example.com/secret", Events: "all", Active: true}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		token, path string
		want        int
	}{{"", "/v1/organizations/me/webhooks/foreign/test", 401}, {member, "/v1/organizations/me/webhooks/foreign/test", 403}, {"admin", "/v1/organizations/me/webhooks/foreign/test", 404}, {"admin", "/v1/organizations/other/webhooks/foreign/test", 404}, {"admin", "/v1/organizations/me/webhooks/missing/test", 404}} {
		r := httptest.NewRequest("POST", tc.path, nil)
		r.Header.Set("Authorization", "Bearer "+tc.token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Errorf("%s: %d %s", tc.path, w.Code, w.Body)
		}
	}
	received := make(chan string, 1)
	endpoint := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received <- r.URL.RequestURI(); w.WriteHeader(204) }))
	defer endpoint.Close()
	original := lookupIP
	lookupIP = func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil }
	defer func() { lookupIP = original }()
	transport := endpoint.Client().Transport.(*http.Transport).Clone()
	transport.TLSClientConfig = transport.TLSClientConfig.Clone()
	transport.TLSClientConfig.ServerName = "127.0.0.1"
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, strings.TrimPrefix(endpoint.URL, "https://"))
	}
	defer transport.CloseIdleConnections()
	server := &Server{cfg: RouterConfig{Store: s, PublicBaseURL: "https://cloud.example"}, dispatcher: NewWebhookDispatcher(&http.Client{Transport: transport})}
	for _, tc := range []struct {
		id, url string
		want    int
	}{{"saved", "https://public.example/SECRET?token=PRIVATE", 200}, {"private", "https://127.0.0.1/secret", 400}} {
		if err := s.CreateWebhook(context.Background(), &WebhookRecord{ID: tc.id, OrgID: "org_cloud_test", TargetType: "generic", URL: tc.url, Events: "all", Active: true}); err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest("POST", "/", nil)
		r.SetPathValue("id", "me")
		r.SetPathValue("wh_id", tc.id)
		r.Header.Set("Authorization", "Bearer admin")
		w := httptest.NewRecorder()
		server.authorize(RoleAdmin, server.handleTestSavedWebhook)(w, r)
		if w.Code != tc.want {
			t.Fatalf("saved test %s: %d %s", tc.id, w.Code, w.Body)
		}
		if strings.Contains(w.Body.String(), "SECRET") || strings.Contains(w.Body.String(), "PRIVATE") {
			t.Fatal("test response leaked credential URL")
		}
		if tc.want == 200 {
			if got := <-received; got != "/SECRET?token=PRIVATE" {
				t.Fatalf("wrong saved destination: %s", got)
			}
		}
	}
}

func TestMemberTokenIssuanceAuthorization(t *testing.T) {
	h, s, member := newTestServer(t)
	defer s.db.Close()
	if err := s.CreateAPITokenWithRole("admin", "org_cloud_test", "admin-secret", "Admin", RoleAdmin); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		token, body string
		code        int
	}{{member, `{"name":"CI"}`, 403}, {"", `{"name":"CI"}`, 401}, {"admin-secret", `{"name":"CI","role":"owner"}`, 400}} {
		r := httptest.NewRequest("POST", "/v1/organizations/me/tokens", strings.NewReader(tc.body))
		r.Header.Set("Authorization", "Bearer "+tc.token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.code {
			t.Errorf("token role request: %d %s", w.Code, w.Body)
		}
	}
	r := httptest.NewRequest("POST", "/v1/organizations/me/tokens", strings.NewReader(`{"name":"CI"}`))
	r.Header.Set("Authorization", "Bearer admin-secret")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 201 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("issuance response: %d %s %s", w.Code, w.Header(), w.Body)
	}
	var result struct {
		Token string `json:"token"`
		Role  Role   `json:"role"`
		OrgID string `json:"organization_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	principal, err := s.AuthenticateToken(result.Token)
	if err != nil || principal.Role != RoleMember || result.Role != RoleMember || result.OrgID != "org_cloud_test" {
		t.Fatalf("invalid member credential: %+v %+v %v", result, principal, err)
	}
}

func TestWebhookMetadataRedactsCredentialURL(t *testing.T) {
	original := lookupIP
	lookupIP = func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil }
	defer func() { lookupIP = original }()
	h, s, _ := newTestServer(t)
	defer s.db.Close()
	if err := s.CreateAPITokenWithRole("admin", "org_cloud_test", "admin-secret", "Admin", RoleAdmin); err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"POST", "GET"} {
		r := httptest.NewRequest(method, "/v1/organizations/me/webhooks", strings.NewReader(`{"target_type":"generic","url":"https://example.com/SECRET_PATH?token=SECRET_QUERY","events":"all"}`))
		r.Header.Set("Authorization", "Bearer admin-secret")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 && w.Code != 201 {
			t.Fatalf("webhook %s: %d %s", method, w.Code, w.Body)
		}
		if strings.Contains(w.Body.String(), "SECRET") {
			t.Errorf("webhook %s leaks destination credentials: %s", method, w.Body)
		}
	}
	hooks, err := s.ListWebhooks(context.Background(), "org_cloud_test")
	if err != nil || len(hooks) != 1 || !strings.Contains(hooks[0].URL, "SECRET_PATH") {
		t.Fatalf("saved destination lost: %+v %v", hooks, err)
	}
}

func TestDeletingWebhookWithQueuedAlerts(t *testing.T) {
	h, s, _ := newTestServer(t)
	defer s.db.Close()
	if err := s.CreateAPITokenWithRole("admin", "org_cloud_test", "admin", "Admin", RoleAdmin); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateWebhook(context.Background(), &WebhookRecord{ID: "queued", OrgID: "org_cloud_test", TargetType: "generic", URL: "https://example.com/secret", Events: "all", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.EnqueueOutbox(context.Background(), []*OutboxItem{{ID: "queued-alert", OrgID: "org_cloud_test", WebhookID: "queued", EventType: "failure", PayloadJSON: `{}`, NextRetryAt: time.Now().Add(time.Hour)}}); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("DELETE", "/v1/organizations/me/webhooks/queued", nil)
	r.Header.Set("Authorization", "Bearer admin")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 204 {
		t.Fatalf("delete destination with queued alerts: %d %s", w.Code, w.Body)
	}
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM webhook_outbox WHERE webhook_id='queued'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("deleted destination retained pending alerts")
	}
}

func TestIngestionRepositoryLimitResponse(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	if _, err := s.db.Exec(`UPDATE organizations SET plan='developer'`); err != nil {
		t.Fatal(err)
	}
	p := correctionPayload()
	w, _ := correctionPost(h, token, "first", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	p.CI.Repository = "acme/other"
	w, _ = correctionPost(h, token, "second", p)
	if w.Code != 403 || !strings.Contains(w.Body.String(), "repository plan limit reached") {
		t.Fatalf("limit response: %d %s", w.Code, w.Body)
	}
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM repositories`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("rejected repository persisted")
	}
}

func correctionPost(h http.Handler, token, key string, p cloud.RunIngestRequest) (*httptest.ResponseRecorder, cloud.RunIngestResponse) {
	b, _ := json.Marshal(p)
	r := httptest.NewRequest("POST", "/v1/runs", bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+token)
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var response cloud.RunIngestResponse
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	return w, response
}

func TestIngestionDistinctScenariosInSameWorkflow(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	p := correctionPayload()
	w, a := correctionPost(h, token, "", p)
	if w.Code != 201 {
		t.Fatalf("first: %d %s", w.Code, w.Body)
	}
	p.Scenario.Name = "other"
	p.Result.Status = "failed"
	w, b := correctionPost(h, token, "", p)
	if w.Code != 201 || a.RunID == b.RunID {
		t.Fatalf("distinct executions collapsed: %d %s", w.Code, w.Body)
	}
}

func TestIngestionOutboxFailureRollsBackAndRetryRecovers(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	if err := s.CreateWebhook(context.Background(), &WebhookRecord{ID: "wh", OrgID: "org_cloud_test", TargetType: "generic", URL: "http://127.0.0.1:1", Events: "all", Active: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`CREATE TRIGGER fail_outbox BEFORE INSERT ON webhook_outbox BEGIN SELECT RAISE(ABORT, 'injected failure'); END`); err != nil {
		t.Fatal(err)
	}
	p := correctionPayload()
	p.Result.Status = "failed"
	p.Result.ViolationDetected = true
	p.Result.AnomalyType = "P4"
	w, _ := correctionPost(h, token, "execution", p)
	if w.Code != 500 {
		t.Errorf("expected storage error, got %d %s", w.Code, w.Body)
	}
	for _, table := range []string{"runs", "findings", "baselines", "webhook_outbox"} {
		var count int
		if err := s.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("%s retained %d rows after rollback", table, count)
		}
	}
	if _, err := s.db.Exec(`DROP TRIGGER fail_outbox`); err != nil {
		t.Fatal(err)
	}
	w, _ = correctionPost(h, token, "execution", p)
	if w.Code != 201 {
		t.Fatalf("retry: %d %s", w.Code, w.Body)
	}
	w, _ = correctionPost(h, token, "execution", p)
	if w.Code != 200 {
		t.Fatalf("replay: %d %s", w.Code, w.Body)
	}
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM webhook_outbox`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected one durable notification, got %d", count)
	}
}

func TestIngestionIdempotencyConflictAndStableReplay(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	p := correctionPayload()
	w, first := correctionPost(h, token, "original", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	p.CI.CommitSHA = "newer"
	w, _ = correctionPost(h, token, "newer", p)
	if w.Code != 201 {
		t.Fatal(w.Body)
	}
	p.CI.CommitSHA = "abc123"
	w, replay := correctionPost(h, token, "original", p)
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(replay)
	if w.Code != 200 || string(a) != string(b) {
		t.Fatalf("replay changed response: original=%s replay=%s", a, b)
	}
	p.Scenario.Name = "changed"
	w, _ = correctionPost(h, token, "original", p)
	if w.Code != 409 {
		t.Fatalf("changed payload should conflict: %d %s", w.Code, w.Body)
	}
}

func TestIngestionConcurrentDuplicateRequests(t *testing.T) {
	h, s, token := newTestServer(t)
	defer s.db.Close()
	const n = 12
	results := make(chan *httptest.ResponseRecorder, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w, _ := correctionPost(h, token, "one-execution", correctionPayload())
			results <- w
		}()
	}
	wg.Wait()
	close(results)
	created := 0
	var id string
	for w := range results {
		if w.Code == 201 {
			created++
		} else if w.Code != 200 {
			t.Errorf("duplicate request: %d %s", w.Code, w.Body)
		}
		var r cloud.RunIngestResponse
		_ = json.Unmarshal(w.Body.Bytes(), &r)
		if id == "" {
			id = r.RunID
		}
		if r.RunID != id {
			t.Error("different replay IDs")
		}
	}
	if created != 1 {
		t.Fatalf("created %d runs", created)
	}
}

func TestRepositoryQuotaConcurrentCreation(t *testing.T) {
	_, s, _ := newTestServer(t)
	defer s.db.Close()
	if _, err := s.db.Exec(`UPDATE organizations SET plan='developer'`); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.GetOrCreateRepo("org_cloud_test", fmt.Sprintf("acme/repo%d", i), "main")
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if err != ErrPlanLimitReached {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if success != 1 {
		t.Fatalf("quota allowed %d repositories", success)
	}
	var name string
	if err := s.db.QueryRow(`SELECT full_name FROM repositories`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetOrCreateRepo("org_cloud_test", name, "main"); err != nil {
		t.Fatalf("existing repository blocked: %v", err)
	}
}
