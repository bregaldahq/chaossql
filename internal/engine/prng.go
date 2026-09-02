package engine

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

// PRNG provides deterministic pseudo-random generation.
type PRNG struct {
	masterSeed uint64
}

// NewPRNG initializes a new PRNG with a master seed.
func NewPRNG(seed uint64) *PRNG {
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	return &PRNG{masterSeed: seed}
}

// MasterSeed returns the seed used.
func (p *PRNG) MasterSeed() uint64 {
	return p.masterSeed
}

// WorkerSeed derives an isolated sub-seed for a specific worker.
func (p *PRNG) WorkerSeed(workerID int) uint64 {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d-%d", p.masterSeed, workerID)))
	return h.Sum64()
}

// EvaluateParam parses declarative parameter generators like "int(10, 50)".
func (p *PRNG) EvaluateParam(expr string, rng *rand.Rand) string {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "int(") && strings.HasSuffix(expr, ")") {
		inside := expr[4 : len(expr)-1]
		parts := strings.Split(inside, ",")
		if len(parts) == 2 {
			minVal, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			maxVal, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && maxVal >= minVal {
				val := minVal + rng.IntN(maxVal-minVal+1)
				return strconv.Itoa(val)
			}
		}
	}
	return expr
}

// Jitter returns a uniform delay between jitterRange[0] and jitterRange[1] milliseconds.
func (p *PRNG) Jitter(jitterRange [2]int, rng *rand.Rand) time.Duration {
	minMs := jitterRange[0]
	maxMs := jitterRange[1]
	if maxMs <= 0 || maxMs < minMs {
		return 0
	}
	delayMs := minMs + rng.IntN(maxMs-minMs+1)
	return time.Duration(delayMs) * time.Millisecond
}
