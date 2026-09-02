package faults_test

import (
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/faults"
)

func TestFaultInjector_AbortProbability(t *testing.T) {
	cfg := domain.FaultConfig{
		AbortProbability: 1.0, // Always abort
	}
	inj := faults.NewFaultInjector(cfg, 42)
	for i := 0; i < 10; i++ {
		if !inj.ShouldAbort() {
			t.Errorf("expected ShouldAbort to return true when probability is 1.0")
		}
	}

	cfgZero := domain.FaultConfig{
		AbortProbability: 0.0,
	}
	injZero := faults.NewFaultInjector(cfgZero, 42)
	for i := 0; i < 10; i++ {
		if injZero.ShouldAbort() {
			t.Errorf("expected ShouldAbort to return false when probability is 0.0")
		}
	}
}

func TestFaultInjector_LatencySpike(t *testing.T) {
	cfg := domain.FaultConfig{
		LatencyProbability: 1.0,
		LatencySpikeMs:     [2]int{20, 50},
	}
	inj := faults.NewFaultInjector(cfg, 42)
	for i := 0; i < 10; i++ {
		spike := inj.GetLatencySpike()
		if spike < 20*time.Millisecond || spike > 50*time.Millisecond {
			t.Errorf("expected spike between 20ms and 50ms, got: %v", spike)
		}
	}

	cfgZero := domain.FaultConfig{
		LatencyProbability: 0.0,
	}
	injZero := faults.NewFaultInjector(cfgZero, 42)
	if injZero.GetLatencySpike() != 0 {
		t.Errorf("expected 0 latency spike when probability is 0.0")
	}
}

func TestFaultInjector_Disconnect(t *testing.T) {
	cfg := domain.FaultConfig{
		DisconnectProbability: 1.0,
	}
	inj := faults.NewFaultInjector(cfg, 42)
	if !inj.ShouldDisconnect() {
		t.Errorf("expected ShouldDisconnect to be true")
	}

	var nilInj *faults.FaultInjector
	if nilInj.ShouldAbort() || nilInj.GetLatencySpike() != 0 || nilInj.ShouldDisconnect() {
		t.Errorf("nil FaultInjector should return false/zero values safely")
	}
}
