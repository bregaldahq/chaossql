package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultWebhookDispatcherRejectsPrivateDestination(t *testing.T) {
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	defer endpoint.Close()
	dispatcher := NewWebhookDispatcher(nil)
	err := dispatcher.DispatchAlert(context.Background(), WebhookRecord{TargetType: "generic", URL: endpoint.URL}, &RegressionAlert{})
	if err == nil {
		t.Fatal("default dispatcher reached private network destination")
	}
}

func TestSafeWebhookDialPinsResolvedIPAddress(t *testing.T) {
	original := lookupIP
	lookupIP = func(string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil }
	defer func() { lookupIP = original }()
	client := NewSafeHTTPClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// The canceled dial error exposes the network address actually passed to the
	// dialer without contacting a remote host. Resolving the original hostname
	// again would report the hostname or a DNS error instead of the approved IP.
	_, err := client.Transport.(*http.Transport).DialContext(ctx, "tcp", "does-not-exist.invalid:443")
	if err == nil {
		t.Fatal("canceled dial succeeded")
	}
	op, ok := err.(*net.OpError)
	if !ok || op.Addr == nil || op.Addr.String() != "203.0.113.10:443" {
		t.Fatalf("dial was not pinned to approved IP: %T %v", err, err)
	}
}
