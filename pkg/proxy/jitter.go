package proxy

import (
	"math/rand"
	"sync"
	"time"
)

// PCTScheduler calculates stochastic microsecond socket delays and commit barriers.
type PCTScheduler struct {
	mu               sync.Mutex
	minJitter        time.Duration
	maxJitter        time.Duration
	commitBarrierMin time.Duration
	commitBarrierMax time.Duration
	depth            int
	seed             uint64
	prng             *rand.Rand
}

// NewPCTScheduler creates an initialized scheduler instance.
func NewPCTScheduler(
	minJitter, maxJitter, commitBarrierMin, commitBarrierMax time.Duration,
	depth int,
	seed uint64,
) *PCTScheduler {
	if minJitter < 0 {
		minJitter = 0
	}
	if maxJitter < minJitter {
		maxJitter = minJitter
	}
	if commitBarrierMin < 0 {
		commitBarrierMin = 0
	}
	if commitBarrierMax < commitBarrierMin {
		commitBarrierMax = commitBarrierMin
	}
	if depth < 1 {
		depth = 1
	}
	if seed == 0 {
		seed = 42
	}

	src := rand.NewSource(int64(seed))
	return &PCTScheduler{
		minJitter:        minJitter,
		maxJitter:        maxJitter,
		commitBarrierMin: commitBarrierMin,
		commitBarrierMax: commitBarrierMax,
		depth:            depth,
		seed:             seed,
		prng:             rand.New(src),
	}
}

// CalculateJitter computes stochastic delay for DML or critical queries inside transactions.
func (s *PCTScheduler) CalculateJitter(sessionID uint64, isDML bool) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.maxJitter <= 0 || s.maxJitter <= s.minJitter {
		return s.minJitter
	}

	delta := int64(s.maxJitter - s.minJitter)
	if delta <= 0 {
		return s.minJitter
	}

	offset := s.prng.Int63n(delta + 1)
	delay := s.minJitter + time.Duration(offset)

	// Invert or scale priority if depth threshold applies
	if isDML && s.depth > 1 && (sessionID%uint64(s.depth) == 0) {
		delay = s.maxJitter - time.Duration(offset/2)
		if delay > s.maxJitter {
			delay = s.maxJitter
		}
		if delay < s.minJitter {
			delay = s.minJitter
		}
	}

	return delay
}

// CalculateCommitBarrier computes deliberate commit barrier delay (50us to 5ms).
func (s *PCTScheduler) CalculateCommitBarrier() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.commitBarrierMax <= 0 || s.commitBarrierMax <= s.commitBarrierMin {
		return s.commitBarrierMin
	}

	delta := int64(s.commitBarrierMax - s.commitBarrierMin)
	if delta <= 0 {
		return s.commitBarrierMin
	}

	offset := s.prng.Int63n(delta + 1)
	return s.commitBarrierMin + time.Duration(offset)
}
