package proxy

import (
	"testing"
	"time"
)

func TestPCTScheduler_JitterBounds(t *testing.T) {
	minJ := 10 * time.Microsecond
	maxJ := 2 * time.Millisecond
	barrierMin := 50 * time.Microsecond
	barrierMax := 5 * time.Millisecond

	sched := NewPCTScheduler(minJ, maxJ, barrierMin, barrierMax, 2, 42)

	for i := 0; i < 100; i++ {
		j := sched.CalculateJitter(uint64(i%5), true)
		if j < minJ || j > maxJ {
			t.Fatalf("jitter %v out of bounds [%v, %v]", j, minJ, maxJ)
		}
	}

	for i := 0; i < 100; i++ {
		b := sched.CalculateCommitBarrier()
		if b < barrierMin || b > barrierMax {
			t.Fatalf("commit barrier %v out of bounds [%v, %v]", b, barrierMin, barrierMax)
		}
	}
}

func TestPCTScheduler_StrictDeterminism(t *testing.T) {
	sched1 := NewPCTScheduler(10*time.Microsecond, 500*time.Microsecond, 50*time.Microsecond, 1*time.Millisecond, 2, 12345)
	sched2 := NewPCTScheduler(10*time.Microsecond, 500*time.Microsecond, 50*time.Microsecond, 1*time.Millisecond, 2, 12345)

	for i := 0; i < 50; i++ {
		j1 := sched1.CalculateJitter(1, true)
		j2 := sched2.CalculateJitter(1, true)
		if j1 != j2 {
			t.Fatalf("iteration %d: non-deterministic jitter: %v vs %v", i, j1, j2)
		}

		b1 := sched1.CalculateCommitBarrier()
		b2 := sched2.CalculateCommitBarrier()
		if b1 != b2 {
			t.Fatalf("iteration %d: non-deterministic commit barrier: %v vs %v", i, b1, b2)
		}
	}
}

func TestPCTScheduler_ZeroJitter(t *testing.T) {
	sched := NewPCTScheduler(0, 0, 0, 0, 1, 0)
	if j := sched.CalculateJitter(1, true); j != 0 {
		t.Fatalf("expected 0 jitter, got %v", j)
	}
	if b := sched.CalculateCommitBarrier(); b != 0 {
		t.Fatalf("expected 0 barrier, got %v", b)
	}
}
