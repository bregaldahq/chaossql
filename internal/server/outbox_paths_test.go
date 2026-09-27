package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOutboxDispatcherFailsUndeliverableItems(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	orgID := "org_outbox_paths"
	if err := s.CreateOrganization(orgID, "Outbox Paths", "pro"); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := s.CreateWebhook(ctx, &WebhookRecord{ID: "wh_deleted", OrgID: orgID, TargetType: "slack", URL: "https://hooks.slack.com/services/T/B/C", Events: "all", Active: true, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateWebhook(ctx, &WebhookRecord{ID: "wh_inactive", OrgID: orgID, TargetType: "slack", URL: "https://hooks.slack.com/services/T/B/C", Events: "all", Active: false, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateWebhook(ctx, &WebhookRecord{ID: "wh_active", OrgID: orgID, TargetType: "slack", URL: "https://hooks.slack.com/services/T/B/C", Events: "all", Active: true, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	items := []*OutboxItem{
		{ID: "ob_missing", OrgID: orgID, WebhookID: "wh_deleted", EventType: "failure", PayloadJSON: `{}`},
		{ID: "ob_inactive", OrgID: orgID, WebhookID: "wh_inactive", EventType: "failure", PayloadJSON: `{}`},
		{ID: "ob_bad_json", OrgID: orgID, WebhookID: "wh_active", EventType: "failure", PayloadJSON: `{not json`},
	}
	for _, item := range items {
		item.Status, item.NextRetryAt, item.CreatedAt = "pending", now.Add(-time.Second), now
	}
	if err := s.EnqueueOutbox(ctx, items); err != nil {
		t.Fatal(err)
	}
	// Simulate a webhook deleted after its alert was queued.
	execWithoutForeignKeys(t, s, `DELETE FROM webhooks WHERE id = 'wh_deleted'`)

	od := &OutboxDispatcher{store: s, dispatcher: NewWebhookDispatcher(http.DefaultClient), maxAttempts: 1}
	if err := od.ProcessPending(ctx); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{"ob_missing": "webhook not found", "ob_inactive": "webhook inactive", "ob_bad_json": "invalid payload json"}
	for id, wantErr := range want {
		var status, lastErr string
		if err := s.db.QueryRow(`SELECT status, last_error FROM webhook_outbox WHERE id = ?`, id).Scan(&status, &lastErr); err != nil {
			t.Fatal(err)
		}
		if status != "failed" || lastErr != wantErr {
			t.Errorf("%s: status=%q last_error=%q, want failed/%q", id, status, lastErr, wantErr)
		}
	}

	if err := (&OutboxDispatcher{}).ProcessPending(ctx); err != nil {
		t.Fatalf("dispatcher without a store must be a no-op, got %v", err)
	}
}

func TestOutboxDispatcherTriggerAndStop(t *testing.T) {
	od := NewOutboxDispatcher(nil, nil)
	for i := 0; i < 200; i++ {
		od.Trigger() // must never block even when the trigger buffer is full
	}
	done := make(chan struct{})
	go func() { od.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return")
	}
}

func TestDiscordAnomalyAlertFormatting(t *testing.T) {
	received := make(chan map[string]any, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		received <- body
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	alert := &RegressionAlert{RepoFullName: "acme/payments", CommitSHA: "0123456789abcdef", AnomalyType: "A5B", IsRegression: false}
	if err := NewWebhookDispatcher(http.DefaultClient).DispatchAlert(context.Background(), WebhookRecord{TargetType: "Discord", URL: srv.URL}, alert); err != nil {
		t.Fatal(err)
	}
	body := <-received
	raw, _ := json.Marshal(body)
	payload := string(raw)
	for _, want := range []string{"ISOLATION ANOMALY DETECTED", "0123456", "Status: A5B"} {
		if !strings.Contains(payload, want) {
			t.Errorf("discord payload missing %q: %s", want, payload)
		}
	}
	if strings.Contains(payload, "0123456789abcdef") {
		t.Errorf("commit SHA must be shortened: %s", payload)
	}
}

func TestSafeHTTPClientBlocksRedirectToInternalAddress(t *testing.T) {
	client := NewSafeHTTPClient()
	req := httptest.NewRequest(http.MethodGet, "http://169.254.169.254/latest/meta-data", nil)
	if err := client.CheckRedirect(req, []*http.Request{{}}); err == nil || !strings.Contains(err.Error(), "ssrf") {
		t.Fatalf("redirect to metadata service: err = %v", err)
	}
	if err := client.CheckRedirect(req, make([]*http.Request, 5)); err == nil || !strings.Contains(err.Error(), "5 redirects") {
		t.Fatalf("redirect limit: err = %v", err)
	}
	allowed := httptest.NewRequest(http.MethodGet, "https://hooks.slack.com/services/T/B/C", nil)
	if err := client.CheckRedirect(allowed, []*http.Request{{}}); err != nil {
		t.Fatalf("redirect to an allowed host: err = %v", err)
	}
}
