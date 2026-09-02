package engine_test

import (
	"math/rand/v2"
	"testing"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestPRNG_DeterministicWorkerSeed(t *testing.T) {
	prng := engine.NewPRNG(12345)
	
	seed1 := prng.WorkerSeed(1)
	seed1Again := prng.WorkerSeed(1)
	if seed1 != seed1Again {
		t.Errorf("expected deterministic worker seeds, got %v and %v", seed1, seed1Again)
	}

	seed2 := prng.WorkerSeed(2)
	if seed1 == seed2 {
		t.Errorf("expected different seeds for different workers")
	}

	prngDiff := engine.NewPRNG(54321)
	if prng.WorkerSeed(1) == prngDiff.WorkerSeed(1) {
		t.Errorf("expected different seeds for different master seeds")
	}
}

func TestPRNG_EvaluateParam(t *testing.T) {
	prng := engine.NewPRNG(42)
	rng := rand.New(rand.NewPCG(42, 0))

	val := prng.EvaluateParam("int(10, 50)", rng)
	if val == "int(10, 50)" || val == "" {
		t.Errorf("failed to evaluate parameter")
	}

	valSame := prng.EvaluateParam("int(10, 50)", rand.New(rand.NewPCG(42, 0)))
	if val != valSame {
		t.Errorf("expected deterministic evaluation, got %v and %v", val, valSame)
	}
	
	valStatic := prng.EvaluateParam("static_string", rng)
	if valStatic != "static_string" {
		t.Errorf("expected static string to be untouched")
	}
}
