package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/cloud"
)

func TestStoreWebhooksCRUD(t *testing.T) {
	s := newTestStore(t)
	orgID := "org_webhook_test"
	if err := s.CreateOrganization(orgID, "Webhook Org", "pro"); err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	// 1. Create Webhook
	wh := &WebhookRecord{
		ID:         "wh_01",
		OrgID:      orgID,
		TargetType: "discord",
		URL:        "https://discord.com/api/webhooks/123/abc",
		Events:     "regression",
		Active:     true,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.CreateWebhook(context.Background(), wh); err != nil {
		t.Fatalf("failed to create webhook: %v", err)
	}

	// 2. List Webhooks
	list, err := s.ListWebhooks(context.Background(), orgID)
	if err != nil {
		t.Fatalf("failed to list webhooks: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(list))
	}
	if list[0].URL != wh.URL {
		t.Errorf("expected url %s, got %s", wh.URL, list[0].URL)
	}
	if list[0].TargetType != "discord" {
		t.Errorf("expected target_type discord, got %s", list[0].TargetType)
	}

	// 3. Filter by Event
	active, err := s.GetActiveWebhooksForEvent(context.Background(), orgID, "regression")
	if err != nil {
		t.Fatalf("failed to get active webhooks: %v", err)
	}
	if len(active) != 1 {
		t.Errorf("expected 1 active webhook for regression, got %d", len(active))
	}

	activeNone, err := s.GetActiveWebhooksForEvent(context.Background(), orgID, "other_event")
	if err != nil {
		t.Fatalf("failed to get active webhooks: %v", err)
	}
	if len(activeNone) != 0 {
		t.Errorf("expected 0 active webhooks for other_event, got %d", len(activeNone))
	}

	// 4. Delete Webhook
	if err := s.DeleteWebhook(context.Background(), orgID, "wh_01"); err != nil {
		t.Fatalf("failed to delete webhook: %v", err)
	}

	listAfter, err := s.ListWebhooks(context.Background(), orgID)
	if err != nil {
		t.Fatalf("failed to list webhooks after delete: %v", err)
	}
	if len(listAfter) != 0 {
		t.Errorf("expected 0 webhooks after delete, got %d", len(listAfter))
	}
}

func TestDispatcherSendDiscordAlert(t *testing.T) {
	received := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(http.DefaultClient)

	alert := &RegressionAlert{
		RepoFullName:   "acme/payments",
		Branch:         "feat/wallet-withdraw",
		PRNumber:       382,
		CommitSHA:      "a8812ac",
		AnomalyType:    "P4",
		AnomalyName:    "Lost Update",
		Driver:         "PostgreSQL 16",
		Isolation:      "READ COMMITTED",
		Scenario:       "wallet_transfer",
		Seed:           184729,
		DurationMS:     340,
		BaselineStatus: "PASS",
		RunURL:         "https://chaossql.bregalda.com/#/visualizer?scenario=wallet_transfer&seed=184729",
		IsRegression:   true,
	}

	wh := WebhookRecord{
		ID:         "wh_test",
		TargetType: "discord",
		URL:        server.URL,
		Events:     "regression",
		Active:     true,
	}

	err := dispatcher.DispatchAlert(context.Background(), wh, alert)
	if err != nil {
		t.Fatalf("expected dispatch success, got: %v", err)
	}

	select {
	case body := <-received:
		embeds, ok := body["embeds"].([]interface{})
		if !ok || len(embeds) == 0 {
			t.Fatalf("expected embeds array in discord payload, got: %v", body)
		}
		embed := embeds[0].(map[string]interface{})
		title, _ := embed["title"].(string)
		if title == "" {
			t.Errorf("expected non-empty title in embed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for webhook to receive payload")
	}
}

func TestDispatcherSendSlackAlert(t *testing.T) {
	received := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatcher := NewWebhookDispatcher(http.DefaultClient)

	alert := &RegressionAlert{
		RepoFullName:   "acme/inventory",
		Branch:         "fix/stock-decrement",
		PRNumber:       120,
		CommitSHA:      "b9914cc",
		AnomalyType:    "A5A",
		AnomalyName:    "Write Skew",
		Driver:         "MySQL 8.0",
		Isolation:      "REPEATABLE READ",
		Scenario:       "flash_sale",
		Seed:           42,
		DurationMS:     210,
		BaselineStatus: "PASS",
		RunURL:         "https://chaossql.bregalda.com/#/dashboard",
		IsRegression:   true,
	}

	wh := WebhookRecord{
		ID:         "wh_slack",
		TargetType: "slack",
		URL:        server.URL,
		Events:     "regression",
		Active:     true,
	}

	err := dispatcher.DispatchAlert(context.Background(), wh, alert)
	if err != nil {
		t.Fatalf("expected dispatch success, got: %v", err)
	}

	select {
	case body := <-received:
		blocks, ok := body["blocks"].([]interface{})
		if !ok || len(blocks) == 0 {
			t.Fatalf("expected blocks array in slack payload, got: %v", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for webhook to receive payload")
	}
}

func TestIngestTriggersWebhookOnRegression(t *testing.T) {
	s := newTestStore(t)
	orgID := "org_ingest_wh"
	_ = s.CreateOrganization(orgID, "Wh Org", "pro")
	token := "tok_ingest_test"
	_ = s.CreateAPIToken("tok_id", orgID, token, "CI")

	received := make(chan bool, 1)
	mockWH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockWH.Close()

	_ = s.CreateWebhook(context.Background(), &WebhookRecord{
		ID:         "wh_live",
		OrgID:      orgID,
		TargetType: "discord",
		URL:        mockWH.URL,
		Events:     "regression,all",
		Active:     true,
		CreatedAt:  time.Now().UTC(),
	})

	router := NewRouter(RouterConfig{
		Store:  s,
		Engine: NewRegressionEngine(s),
	})

	// 1. Ingest passing baseline run on main
	passReq := cloud.RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		CI: &cloud.CIContext{
			Provider:   "github-actions",
			Repository: "acme/checkout",
			CommitSHA:  "aaa1111",
			Branch:     "main",
			Actor:      "octocat",
		},
		Scenario: cloud.ScenarioMetadata{
			Name:       "cart_checkout",
			Driver:     "sqlite",
			Workers:    4,
			Iterations: 100,
			Seed:       100,
		},
		Result: cloud.ExecutionSummary{
			Status:            "passed",
			Success:           true,
			ViolationDetected: false,
			AnomalyType:       "NONE",
			DurationMS:        410,
			TotalSchedules:    100,
			FailedSchedules:   0,
		},
	}
	body, _ := json.Marshal(passReq)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/runs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for baseline, got %d", rec.Code)
	}

	// 2. Ingest failing PR run (REGRESSION)
	failReq := cloud.RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		CI: &cloud.CIContext{
			Provider:          "github-actions",
			Repository:        "acme/checkout",
			CommitSHA:         "bbb2222",
			Branch:            "feat/discount",
			PullRequestNumber: 99,
			Actor:             "contributor",
		},
		Scenario: cloud.ScenarioMetadata{
			Name:       "cart_checkout",
			Driver:     "sqlite",
			Workers:    4,
			Iterations: 100,
			Seed:       101,
		},
		Result: cloud.ExecutionSummary{
			Status:            "failed",
			Success:           false,
			ViolationDetected: true,
			AnomalyType:       "P4",
			DurationMS:        550,
			TotalSchedules:    100,
			FailedSchedules:   18,
		},
	}
	bodyFail, _ := json.Marshal(failReq)
	recFail := httptest.NewRecorder()
	reqFail := httptest.NewRequest("POST", "/v1/runs", bytes.NewReader(bodyFail))
	reqFail.Header.Set("Authorization", "Bearer "+token)
	reqFail.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recFail, reqFail)

	if recFail.Code != http.StatusCreated {
		t.Fatalf("expected 201 for regression, got %d", recFail.Code)
	}

	// Verify webhook triggered
	select {
	case <-received:
		// success!
	case <-time.After(2 * time.Second):
		t.Fatal("webhook was not triggered on regression")
	}
}

func TestValidateWebhookTargetRejectsSSRF(t *testing.T) {
	rejected := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"plain http", "http://discord.com/api/webhooks/1/abc"},
		{"loopback", "https://127.0.0.1/webhook"},
		{"loopback by name", "https://localhost/webhook"},
		{"ipv6 loopback", "https://[::1]/webhook"},
		{"aws metadata", "https://169.254.169.254/latest/meta-data"},
		{"rfc1918 ten", "https://10.0.0.5/internal"},
		{"rfc1918 192", "https://192.168.1.1/admin"},
		{"rfc1918 172", "https://172.16.0.1/admin"},
		{"cgnat", "https://100.64.0.1/internal"},
		{"unspecified", "https://0.0.0.0/x"},
		{"credentials in url", "https://user:pass@discord.com/api/webhooks/1/abc"},
		{"no host", "https:///api/webhooks"},
		{"non-http scheme", "file:///etc/passwd"},
	}

	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateWebhookTarget(tc.url); err == nil {
				t.Fatalf("expected %q to be rejected, but it was allowed", tc.url)
			}
		})
	}
}

func TestValidateWebhookTargetAllowsPublicHTTPS(t *testing.T) {
	// Stub DNS so the test does not depend on network access.
	original := lookupIP
	lookupIP = func(host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("203.0.113.10")}, nil
	}
	defer func() { lookupIP = original }()

	allowed := []string{
		"https://discord.com/api/webhooks/123456789/AbC-dEf_123",
		"https://hooks.slack.com/services/T00000000/B00000000/xxxxxxxx",
		"https://alerts.example.com/custom/generic-endpoint",
	}

	for _, url := range allowed {
		if err := ValidateWebhookTarget(url); err != nil {
			t.Errorf("expected %q to be allowed, got: %v", url, err)
		}
	}
}

func TestShortWebhookRoutesRequireAuthByDefault(t *testing.T) {
	s := newTestStore(t)
	router := NewRouter(RouterConfig{Store: s, Engine: NewRegressionEngine(s)})

	// A stored webhook whose URL embeds a provider secret.
	_ = s.CreateWebhook(context.Background(), &WebhookRecord{
		ID: "wh_secret", OrgID: "org_default", TargetType: "discord",
		URL:    "https://discord.com/api/webhooks/123/SUPERSECRETTOKEN",
		Events: "all", Active: true, CreatedAt: time.Now().UTC(),
	})

	t.Setenv(localDashboardEnvVar, "")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	req.RemoteAddr = "203.0.113.7:44444"
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated remote caller, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "SUPERSECRETTOKEN") {
		t.Fatal("webhook secret leaked in an unauthenticated response")
	}
}

func TestShortWebhookRoutesRejectRemoteEvenWhenOptedIn(t *testing.T) {
	s := newTestStore(t)
	router := NewRouter(RouterConfig{Store: s, Engine: NewRegressionEngine(s)})

	t.Setenv(localDashboardEnvVar, "1")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	req.RemoteAddr = "203.0.113.7:44444"
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("opt-in must still require loopback; got %d for a remote caller", rec.Code)
	}
}

func TestShortWebhookRoutesAllowLoopbackWhenOptedIn(t *testing.T) {
	s := newTestStore(t)
	router := NewRouter(RouterConfig{Store: s, Engine: NewRegressionEngine(s)})

	t.Setenv(localDashboardEnvVar, "1")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	req.RemoteAddr = "127.0.0.1:55555"
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected loopback dashboard access to work when opted in, got %d", rec.Code)
	}
}
