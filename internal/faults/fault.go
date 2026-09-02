package faults

import (
	"math/rand"
	"sync"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// FaultInjector manages deterministic stochastic fault injection into transactions.
type FaultInjector struct {
	cfg  domain.FaultConfig
	rng  *rand.Rand
	lock sync.Mutex
}

// NewFaultInjector instantiates a new FaultInjector with given config and seed.
func NewFaultInjector(cfg domain.FaultConfig, seed uint64) *FaultInjector {
	return &FaultInjector{
		cfg: cfg,
		rng: rand.New(rand.NewSource(int64(seed + 99991))),
	}
}

// ShouldAbort returns true if a stochastic transaction abort fault should be injected.
func (f *FaultInjector) ShouldAbort() bool {
	if f == nil || f.cfg.AbortProbability <= 0 {
		return false
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.rng.Float64() < f.cfg.AbortProbability
}

// GetLatencySpike returns a latency duration if a latency spike fault is triggered.
func (f *FaultInjector) GetLatencySpike() time.Duration {
	if f == nil || f.cfg.LatencyProbability <= 0 {
		return 0
	}
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.rng.Float64() < f.cfg.LatencyProbability {
		minMs := f.cfg.LatencySpikeMs[0]
		maxMs := f.cfg.LatencySpikeMs[1]
		if maxMs <= minMs {
			maxMs = minMs + 10
		}
		if minMs < 0 {
			minMs = 0
		}
		delta := maxMs - minMs
		spike := minMs + f.rng.Intn(delta+1)
		return time.Duration(spike) * time.Millisecond
	}
	return 0
}

// ShouldDisconnect returns true if an abrupt client connection loss should be simulated.
func (f *FaultInjector) ShouldDisconnect() bool {
	if f == nil || f.cfg.DisconnectProbability <= 0 {
		return false
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.rng.Float64() < f.cfg.DisconnectProbability
}
