package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrUnauthorized     = errors.New("invalid or unauthorized ChaosSQL Cloud token (HTTP 401)")
	ErrBadRequest       = errors.New("invalid run ingest payload (HTTP 400)")
	ErrCloudUnavailable = errors.New("chaossql cloud api unavailable after retries")
)

// Config configures the Cloud HTTP API client
type Config struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	MaxRetries int
	FailFast   bool
	HTTPClient *http.Client
}

// Client interacts with the ChaosSQL Cloud control plane
type Client struct {
	cfg Config
}

// NewClient initializes a Cloud API client with sensible defaults
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.chaossql.bregalda.com"
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")

	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 3
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	return &Client{cfg: cfg}
}

// PublishRun serializes and publishes an execution report to POST /v1/runs with exponential retries
func (c *Client) PublishRun(ctx context.Context, req *RunIngestRequest) (*RunIngestResponse, error) {
	if c.cfg.Token == "" {
		return nil, errors.New("missing cloud token")
	}

	if req.Version == "" {
		req.Version = "1.0"
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode run ingest payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/runs", c.cfg.BaseURL)
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<attempt) * 50 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create http request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.Token)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("User-Agent", "ChaosSQL-CLI/v1.4.0 (pure-go)")

		resp, err := c.cfg.HTTPClient.Do(httpReq)
		if err != nil {
			lastErr = err
			continue // retry network error
		}

		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			var ingestResp RunIngestResponse
			if err := json.Unmarshal(respBody, &ingestResp); err != nil {
				return nil, fmt.Errorf("failed to parse cloud response: %w", err)
			}
			return &ingestResp, nil
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, ErrUnauthorized
		}

		if resp.StatusCode == http.StatusBadRequest {
			return nil, fmt.Errorf("%w: %s", ErrBadRequest, string(respBody))
		}

		// 5xx Server Error: retry
		lastErr = fmt.Errorf("server error HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil, fmt.Errorf("%w: %v", ErrCloudUnavailable, lastErr)
}
