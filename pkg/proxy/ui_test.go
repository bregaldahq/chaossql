package proxy

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestProxy_LiveUIEndpoints(t *testing.T) {
	cfg := ProxyConfig{
		ListenAddr:   "127.0.0.1:0",
		UpstreamAddr: "127.0.0.1:5432",
		Protocol:     ProtocolPostgres,
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	httpServer, uiURL, err := StartLiveUI("127.0.0.1:0", server)
	if err != nil {
		t.Fatalf("failed to start live UI: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	}()

	// 1. Check HTML root endpoint
	resp, err := http.Get(uiURL + "/")
	if err != nil {
		t.Fatalf("failed to GET /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Fatalf("expected non-empty HTML body")
	}

	// 2. Check /api/status endpoint
	statusResp, err := http.Get(uiURL + "/api/status")
	if err != nil {
		t.Fatalf("failed to GET /api/status: %v", err)
	}
	defer statusResp.Body.Close()
	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/status, got %d", statusResp.StatusCode)
	}

	// 3. Check /api/sarif endpoint
	sarifResp, err := http.Get(uiURL + "/api/sarif")
	if err != nil {
		t.Fatalf("failed to GET /api/sarif: %v", err)
	}
	defer sarifResp.Body.Close()
	if sarifResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /api/sarif, got %d", sarifResp.StatusCode)
	}
}
