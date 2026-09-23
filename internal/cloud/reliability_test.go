package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaultRetriesKeepExecutionIdentity(t *testing.T) {
	var keys []string
	var bodies []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		encoded, _ := json.Marshal(body)
		bodies = append(bodies, string(encoded))
		if len(keys) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if body["idempotency_key"] != keys[len(keys)-1] {
			t.Error("body and header identity differ")
		}
		_, _ = w.Write([]byte(`{"success":true,"run_id":"recovered"}`))
	}))
	defer ts.Close()
	client := NewClient(Config{BaseURL: ts.URL, Token: "test"})
	req := &RunIngestRequest{}
	if _, err := client.PublishRun(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] == "" || keys[0] != keys[1] || bodies[0] != bodies[1] {
		t.Fatalf("retry changed identity or payload: %v", keys)
	}
	if _, err := client.PublishRun(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if keys[2] == keys[0] {
		t.Fatal("separate execution reused identity")
	}
}

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type brokenResponseBody struct{}

func (brokenResponseBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (brokenResponseBody) Close() error             { return nil }

func TestTerminalStatusDoesNotRetryTruncatedResponse(t *testing.T) {
	attempts := 0
	transport := testTransport(func(req *http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{StatusCode: http.StatusConflict, Body: brokenResponseBody{}, Header: make(http.Header)}, nil
	})
	_, err := NewClient(Config{Token: "test", HTTPClient: &http.Client{Transport: transport}}).PublishRun(context.Background(), &RunIngestRequest{})
	if err == nil || attempts != 1 {
		t.Fatalf("terminal conflict retried: err=%v attempts=%d", err, attempts)
	}
}

func TestExplicitIdentitySurvivesNetworkRetry(t *testing.T) {
	attempts := 0
	transport := testTransport(func(req *http.Request) (*http.Response, error) {
		attempts++
		if req.Header.Get("Idempotency-Key") != "explicit-execution" {
			t.Error("explicit key lost")
		}
		var payload RunIngestRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload.IdempotencyKey != "explicit-execution" {
			t.Error("explicit body key lost")
		}
		if attempts == 1 {
			return nil, io.ErrUnexpectedEOF
		}
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(`{"success":true,"run_id":"ok"}`)), Header: make(http.Header)}, nil
	})
	_, err := NewClient(Config{Token: "test", HTTPClient: &http.Client{Transport: transport}}).PublishRun(context.Background(), &RunIngestRequest{IdempotencyKey: "explicit-execution"})
	if err != nil || attempts != 2 {
		t.Fatalf("err=%v attempts=%d", err, attempts)
	}
}

func TestCancellationDuringRequestPreservesContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	transport := testTransport(func(req *http.Request) (*http.Response, error) { cancel(); return nil, context.Canceled })
	_, err := NewClient(Config{Token: "test", MaxRetries: -1, HTTPClient: &http.Client{Transport: transport}}).PublishRun(ctx, &RunIngestRequest{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation became transport error: %v", err)
	}
}

func TestPublishRejectsUnacknowledgedAndOversizedResponses(t *testing.T) {
	for _, body := range []string{`{"success":false}`, strings.Repeat("x", MaxPayloadBytes+1)} {
		transport := testTransport(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})
		_, err := NewClient(Config{Token: "test", HTTPClient: &http.Client{Transport: transport}}).PublishRun(context.Background(), &RunIngestRequest{})
		if err == nil {
			t.Fatal("invalid response accepted")
		}
		if len(body) > MaxPayloadBytes && !errors.Is(err, ErrResponseTooLarge) {
			t.Fatalf("wrong error: %v", err)
		}
	}
}

func TestIdentityMetadataRejectsUnsafeAndMalformedValues(t *testing.T) {
	for _, raw := range []string{`{"idempotency_key":"x\nInjected: value"}`, `{"idempotency_key":"postgres://secret@db"}`, `{"ci":{"commit_timestamp":"not-a-timestamp"}}`} {
		if _, err := DecodeMetadataOnlyRequest([]byte(raw)); err == nil {
			t.Fatalf("unsafe metadata accepted: %s", raw)
		}
	}
}

func TestTerminalClientErrorsDoNotRetry(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 409, 413, 422} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			attempts := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { attempts++; w.WriteHeader(status) }))
			defer ts.Close()
			_, err := NewClient(Config{BaseURL: ts.URL, Token: "test", MaxRetries: 2}).PublishRun(context.Background(), &RunIngestRequest{})
			if err == nil || attempts != 1 {
				t.Fatalf("err=%v attempts=%d", err, attempts)
			}
		})
	}
}

func TestPublishCancellationPreservesContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewClient(Config{Token: "test", MaxRetries: 1}).PublishRun(ctx, &RunIngestRequest{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestMetadataPreservesExecutionIdentityAndCommitTime(t *testing.T) {
	raw := []byte(`{"idempotency_key":"execution-1","ci":{"commit_timestamp":"2026-09-01T12:00:00Z"}}`)
	req, err := DecodeMetadataOnlyRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(projectMetadataPayload(req, time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	_ = json.Unmarshal(encoded, &payload)
	if payload["idempotency_key"] != "execution-1" || payload["ci"].(map[string]any)["commit_timestamp"] != "2026-09-01T12:00:00Z" {
		t.Fatalf("lost identity metadata: %s", encoded)
	}
}
