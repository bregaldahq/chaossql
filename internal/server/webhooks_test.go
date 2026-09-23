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
		RunURL:         "https://chaossql.bregalda.com/visualizer?scenario=wallet_transfer&seed=184729",
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
		RunURL:         "https://chaossql.bregalda.com/dashboard",
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

	// Explicitly supply the controlled test transport: production dispatchers
	// reject this loopback destination through their default SSRF guard.
	dispatcher := NewWebhookDispatcher(mockWH.Client())
	outbox := NewOutboxDispatcher(s, dispatcher)
	defer outbox.Stop()
	server := &Server{cfg: RouterConfig{Store: s, Engine: NewRegressionEngine(s), PublicBaseURL: "https://cloud.example"}, dispatcher: dispatcher, outboxDispatcher: outbox}
	router := server.authorize(RoleMember, server.handleIngestRun)

	// 1. Ingest passing baseline run on main
	passReq := cloud.RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		CI: &cloud.CIContext{
			Provider:   "github-actions",
			Repository: "acme/checkout",
			CommitSHA:  "aaa1111",
			Branch:     "main",
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

func TestSaaSRouterDoesNotExposeShortWebhookRoutes(t *testing.T) {
	s := newTestStore(t)
	router := NewRouter(RouterConfig{Store: s, Engine: NewRegressionEngine(s)})

	// A stored webhook whose URL embeds a provider secret.
	_ = s.CreateWebhook(context.Background(), &WebhookRecord{
		ID: "wh_secret", OrgID: "org_default", TargetType: "discord",
		URL:    "https://discord.com/api/webhooks/123/SUPERSECRETTOKEN",
		Events: "all", Active: true, CreatedAt: time.Now().UTC(),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	req.RemoteAddr = "203.0.113.7:44444"
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for route absent from SaaS router, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "SUPERSECRETTOKEN") {
		t.Fatal("webhook secret leaked in an unauthenticated response")
	}
}

func TestSaaSRouterCannotEnableShortRoutesThroughEnvironment(t *testing.T) {
	s := newTestStore(t)
	router := NewRouter(RouterConfig{Store: s, Engine: NewRegressionEngine(s)})

	t.Setenv("CHAOSSQL_LOCAL_DASHBOARD", "1")

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	req.RemoteAddr = "203.0.113.7:44444"
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("environment must not alter SaaS routes; got %d", rec.Code)
	}
}

func TestExplicitLocalRouterAllowsShortWebhookRoutes(t *testing.T) {
	s := newTestStore(t)
	if err := s.CreateOrganization("org_default", "Local", "developer"); err != nil {
		t.Fatal(err)
	}
	router := NewLocalRouter(RouterConfig{Store: s, Engine: NewRegressionEngine(s)})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v1/webhooks", nil)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected explicit local dashboard access, got %d", rec.Code)
	}
}

func TestOutboxDispatcher_RetryAndFailure(t *testing.T) {
	s := newTestStore(t)
	orgID := "org_outbox_test"
	_ = s.CreateOrganization(orgID, "Outbox Org", "pro")

	var callCount int
	var returnStatus int = http.StatusInternalServerError

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(returnStatus)
	}))
	defer mockServer.Close()

	now := time.Now().UTC()
	_ = s.CreateWebhook(context.Background(), &WebhookRecord{
		ID:         "wh_retry_test",
		OrgID:      orgID,
		TargetType: "slack",
		URL:        mockServer.URL,
		Events:     "failure",
		Active:     true,
		CreatedAt:  now,
	})

	outboxItem := &OutboxItem{
		ID:          "outbox_retry_1",
		OrgID:       orgID,
		WebhookID:   "wh_retry_test",
		EventType:   "failure",
		PayloadJSON: `{"repo_full_name":"acme/retry","branch":"main","anomaly_type":"P4"}`,
		Status:      "pending",
		NextRetryAt: now.Add(-time.Second), // ready now
		CreatedAt:   now,
	}
	if err := s.EnqueueOutbox(context.Background(), []*OutboxItem{outboxItem}); err != nil {
		t.Fatalf("enqueue outbox item failed: %v", err)
	}

	// Create dispatcher with mock client that allows local test server
	dispatcher := NewWebhookDispatcher(mockServer.Client())
	od := &OutboxDispatcher{
		store:       s,
		dispatcher:  dispatcher,
		maxAttempts: 3,
	}

	// 1. First attempt fails (500)
	if err := od.ProcessPending(context.Background()); err != nil {
		t.Fatalf("process pending unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Verify attempt incremented and status is still pending with future retry
	var attempts int
	var status string
	var nextRetry time.Time
	_ = s.db.QueryRow(`SELECT attempts, status, next_retry_at FROM webhook_outbox WHERE id = 'outbox_retry_1'`).
		Scan(&attempts, &status, &nextRetry)
	if attempts != 1 || status != "pending" {
		t.Errorf("expected attempts=1, status=pending, got attempts=%d status=%s", attempts, status)
	}
	if !nextRetry.After(now) {
		t.Errorf("expected next_retry_at in future, got %v", nextRetry)
	}

	// 2. Fast forward time and fail attempt 2
	_, _ = s.db.Exec(`UPDATE webhook_outbox SET next_retry_at = ? WHERE id = 'outbox_retry_1'`, now.Add(-time.Second))
	_ = od.ProcessPending(context.Background())
	_ = s.db.QueryRow(`SELECT attempts, status FROM webhook_outbox WHERE id = 'outbox_retry_1'`).Scan(&attempts, &status)
	if attempts != 2 || status != "pending" {
		t.Errorf("expected attempts=2, status=pending, got attempts=%d status=%s", attempts, status)
	}

	// 3. Fast forward time and fail attempt 3 (should transition to 'failed')
	_, _ = s.db.Exec(`UPDATE webhook_outbox SET next_retry_at = ? WHERE id = 'outbox_retry_1'`, now.Add(-time.Second))
	_ = od.ProcessPending(context.Background())
	_ = s.db.QueryRow(`SELECT attempts, status FROM webhook_outbox WHERE id = 'outbox_retry_1'`).Scan(&attempts, &status)
	if attempts != 3 || status != "failed" {
		t.Errorf("expected attempts=3, status=failed, got attempts=%d status=%s", attempts, status)
	}

	// 4. Test success: new item with 200 OK transitions to 'delivered'
	returnStatus = http.StatusOK
	outboxSuccess := &OutboxItem{
		ID:          "outbox_success_1",
		OrgID:       orgID,
		WebhookID:   "wh_retry_test",
		EventType:   "failure",
		PayloadJSON: `{"repo_full_name":"acme/retry","branch":"main","anomaly_type":"P4"}`,
		Status:      "pending",
		NextRetryAt: now.Add(-time.Second),
		CreatedAt:   now,
	}
	_ = s.EnqueueOutbox(context.Background(), []*OutboxItem{outboxSuccess})
	_ = od.ProcessPending(context.Background())

	var deliveredStatus string
	var deliveredAt *time.Time
	_ = s.db.QueryRow(`SELECT status, delivered_at FROM webhook_outbox WHERE id = 'outbox_success_1'`).Scan(&deliveredStatus, &deliveredAt)
	if deliveredStatus != "delivered" || deliveredAt == nil {
		t.Errorf("expected delivered status and non-nil delivered_at, got status=%s delivered_at=%v", deliveredStatus, deliveredAt)
	}
}

func TestSafeHTTPClient_RejectsInternalDial(t *testing.T) {
	client := NewSafeHTTPClient()
	_, err := client.Get("http://127.0.0.1:9999/test")
	if err == nil {
		t.Fatalf("expected error dialing internal address, got nil")
	}
	if !strings.Contains(err.Error(), "ssrf protection") && !strings.Contains(err.Error(), "connection refused") {
		t.Logf("got error: %v", err)
	}
}
