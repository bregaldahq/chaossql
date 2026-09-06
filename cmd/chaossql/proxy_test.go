package main

import (
	"testing"
	"time"
)

func TestProxyCmd_Flags(t *testing.T) {
	cmd := newProxyCmd()
	if cmd.Use != "proxy" {
		t.Fatalf("expected Use 'proxy', got %s", cmd.Use)
	}

	listenFlag := cmd.Flag("listen")
	if listenFlag == nil || listenFlag.DefValue != "127.0.0.1:5433" {
		t.Fatalf("expected default listen flag 127.0.0.1:5433, got %v", listenFlag)
	}

	upstreamFlag := cmd.Flag("upstream")
	if upstreamFlag == nil || upstreamFlag.DefValue != "127.0.0.1:5432" {
		t.Fatalf("expected default upstream flag 127.0.0.1:5432, got %v", upstreamFlag)
	}

	protocolFlag := cmd.Flag("protocol")
	if protocolFlag == nil || protocolFlag.DefValue != "postgres" {
		t.Fatalf("expected default protocol flag postgres, got %v", protocolFlag)
	}

	jMinFlag := cmd.Flag("jitter-min")
	if jMinFlag == nil || jMinFlag.DefValue != (10*time.Microsecond).String() {
		t.Fatalf("expected jitter-min 10us, got %v", jMinFlag)
	}

	sarifFlag := cmd.Flag("export-sarif")
	if sarifFlag == nil {
		t.Fatalf("expected export-sarif flag")
	}

	uiPortFlag := cmd.Flag("ui-port")
	if uiPortFlag == nil {
		t.Fatalf("expected ui-port flag")
	}
}
