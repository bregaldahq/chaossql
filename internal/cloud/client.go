package cloud

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bregaldahq/chaossql/internal/version"
)

var (
	ErrUnauthorized     = errors.New("invalid or unauthorized ChaosSQL Cloud token (HTTP 401)")
	ErrBadRequest       = errors.New("invalid run ingest payload (HTTP 400)")
	ErrCloudUnavailable = errors.New("chaossql cloud api unavailable after retries")
	ErrPayloadTooLarge  = errors.New("cloud metadata payload exceeds 64 KiB limit")
	ErrResponseTooLarge = errors.New("cloud response exceeds 64 KiB limit")
)

// Config configures the Cloud HTTP API client
type Config struct {
	BaseURL    string
	Token      string
	Timeout    time.Duration
	MaxRetries int // Zero uses three retries; a negative value disables retries.
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
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	} else if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("missing run ingest payload")
	}
	if c.cfg.Token == "" {
		return nil, errors.New("missing cloud token")
	}

	payload := projectMetadataPayload(req, time.Now())
	if payload.IdempotencyKey == "" {
		var key [16]byte
		if _, err := rand.Read(key[:]); err != nil {
			return nil, fmt.Errorf("generate ingestion identity: %w", err)
		}
		payload.IdempotencyKey = hex.EncodeToString(key[:])
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode run ingest payload: %w", err)
	}
	if len(bodyBytes) > MaxPayloadBytes {
		return nil, fmt.Errorf("%w: %d bytes", ErrPayloadTooLarge, len(bodyBytes))
	}
	if err := validateMetadataPayload(&payload); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/v1/runs", c.cfg.BaseURL)
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
		httpReq.Header.Set("Idempotency-Key", payload.IdempotencyKey)
		httpReq.Header.Set("User-Agent", "ChaosSQL-CLI/v"+version.Version+" (pure-go)")

		resp, err := c.cfg.HTTPClient.Do(httpReq)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = err
			continue // retry network error
		}

		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, MaxPayloadBytes+1))
		_ = resp.Body.Close()
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if len(respBody) > MaxPayloadBytes {
			return nil, ErrResponseTooLarge
		}
		retryable := resp.StatusCode >= 500 || resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode == http.StatusTooManyRequests
		if readErr != nil && (retryable || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
			lastErr = readErr
			continue
		}

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			var ingestResp RunIngestResponse
			if err := json.Unmarshal(respBody, &ingestResp); err != nil {
				return nil, fmt.Errorf("failed to parse cloud response: %w", err)
			}
			if !ingestResp.Success {
				return nil, errors.New("cloud did not acknowledge ingestion")
			}
			return &ingestResp, nil
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, ErrUnauthorized
		}

		if resp.StatusCode == http.StatusBadRequest {
			return nil, ErrBadRequest
		}
		if !retryable {
			return nil, fmt.Errorf("cloud rejected ingestion: HTTP %d", resp.StatusCode)
		}

		// 5xx Server Error: retry
		lastErr = fmt.Errorf("server error HTTP %d", resp.StatusCode)
	}

	return nil, fmt.Errorf("%w: %v", ErrCloudUnavailable, lastErr)
}

// BaseURL returns the configured base URL
func (c *Client) BaseURL() string {
	return c.cfg.BaseURL
}
