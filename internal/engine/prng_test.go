package engine_test

import (
	"context"
	"math/rand/v2"
	"regexp"
	"strconv"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
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

func TestEvaluateGenerator_RandomInt(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(42, 0))
	rng2 := rand.New(rand.NewPCG(42, 0))

	// Test basic range
	val1, err := engine.EvaluateGenerator("$random_int(10, 500)", rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	intVal, err := strconv.Atoi(val1)
	if err != nil {
		t.Fatalf("expected numeric string, got %q: %v", val1, err)
	}
	if intVal < 10 || intVal > 500 {
		t.Errorf("expected value in [10, 500], got %d", intVal)
	}

	// Test determinism (same seed yields exact same value)
	val2, err := engine.EvaluateGenerator("$random_int(10, 500)", rng2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != val2 {
		t.Errorf("expected deterministic output, got %q and %q", val1, val2)
	}

	// Test boundary single value
	valSingle, err := engine.EvaluateGenerator("$random_int(42, 42)", rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valSingle != "42" {
		t.Errorf("expected '42', got %q", valSingle)
	}

	// Test negative range
	valNeg, err := engine.EvaluateGenerator("$random_int(-50, -10)", rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	negInt, err := strconv.Atoi(valNeg)
	if err != nil || negInt < -50 || negInt > -10 {
		t.Errorf("expected value in [-50, -10], got %v", valNeg)
	}

	// Test error cases
	if _, err := engine.EvaluateGenerator("$random_int(500, 10)", rng1); err == nil {
		t.Error("expected error when min > max")
	}
	if _, err := engine.EvaluateGenerator("$random_int(abc, 10)", rng1); err == nil {
		t.Error("expected error for non-integer min")
	}
	if _, err := engine.EvaluateGenerator("$random_int(10)", rng1); err == nil {
		t.Error("expected error for wrong number of arguments")
	}
}

func TestEvaluateGenerator_RandomChoice(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(100, 0))
	rng2 := rand.New(rand.NewPCG(100, 0))

	// Quoted choices
	expr := "$random_choice('DEPOSIT', 'WITHDRAW', 'TRANSFER')"
	val1, err := engine.EvaluateGenerator(expr, rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	validChoices := map[string]bool{"DEPOSIT": true, "WITHDRAW": true, "TRANSFER": true}
	if !validChoices[val1] {
		t.Errorf("unexpected choice: %q", val1)
	}

	// Determinism
	val2, err := engine.EvaluateGenerator(expr, rng2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != val2 {
		t.Errorf("expected deterministic choice: got %q and %q", val1, val2)
	}

	// Unquoted choices
	valUnquoted, err := engine.EvaluateGenerator("$random_choice(apple, banana, cherry)", rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	validFruits := map[string]bool{"apple": true, "banana": true, "cherry": true}
	if !validFruits[valUnquoted] {
		t.Errorf("unexpected unquoted choice: %q", valUnquoted)
	}

	// Choice with comma inside quotes
	valComma, err := engine.EvaluateGenerator(`$random_choice('foo, bar', 'baz')`, rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valComma != "foo, bar" && valComma != "baz" {
		t.Errorf("unexpected choice with comma: %q", valComma)
	}

	// Single choice
	valSingle, err := engine.EvaluateGenerator("$random_choice('ONLY_ONE')", rng1)
	if err != nil || valSingle != "ONLY_ONE" {
		t.Errorf("expected 'ONLY_ONE', got %q, err: %v", valSingle, err)
	}

	// Error cases
	if _, err := engine.EvaluateGenerator("$random_choice()", rng1); err == nil {
		t.Error("expected error for empty choices")
	}
	if _, err := engine.EvaluateGenerator("$random_choice('unclosed)", rng1); err == nil {
		t.Error("expected error for unclosed quote")
	}
}

func TestEvaluateGenerator_RandomString(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(200, 0))
	rng2 := rand.New(rand.NewPCG(200, 0))

	val1, err := engine.EvaluateGenerator("$random_string(8)", rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(val1) != 8 {
		t.Errorf("expected string of length 8, got length %d (%q)", len(val1), val1)
	}

	alphanumeric := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if !alphanumeric.MatchString(val1) {
		t.Errorf("expected alphanumeric string, got %q", val1)
	}

	// Determinism
	val2, err := engine.EvaluateGenerator("$random_string(8)", rng2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != val2 {
		t.Errorf("expected deterministic random string, got %q and %q", val1, val2)
	}

	// Length 0
	valZero, err := engine.EvaluateGenerator("$random_string(0)", rng1)
	if err != nil || valZero != "" {
		t.Errorf("expected empty string for length 0, got %q, err: %v", valZero, err)
	}

	// Error cases
	if _, err := engine.EvaluateGenerator("$random_string(-5)", rng1); err == nil {
		t.Error("expected error for negative length")
	}
	if _, err := engine.EvaluateGenerator("$random_string(xyz)", rng1); err == nil {
		t.Error("expected error for non-integer length")
	}
}

func TestEvaluateGenerator_UUID(t *testing.T) {
	rng1 := rand.New(rand.NewPCG(300, 0))
	rng2 := rand.New(rand.NewPCG(300, 0))

	val1, err := engine.EvaluateGenerator("$uuid()", rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(val1) != 36 {
		t.Errorf("expected UUID length 36, got %d (%q)", len(val1), val1)
	}

	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(val1) {
		t.Errorf("expected valid RFC 4122 v4 UUID, got %q", val1)
	}

	// Determinism
	val2, err := engine.EvaluateGenerator("$uuid()", rng2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != val2 {
		t.Errorf("expected deterministic UUID, got %q and %q", val1, val2)
	}

	// Consecutive generation should advance RNG and produce distinct valid UUIDs
	val3, err := engine.EvaluateGenerator("$uuid()", rng1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 == val3 {
		t.Errorf("consecutive UUID generations should differ: %q vs %q", val1, val3)
	}
	if !uuidRegex.MatchString(val3) {
		t.Errorf("expected valid RFC 4122 v4 UUID for second generation, got %q", val3)
	}

	// Error case: arguments provided to $uuid()
	if _, err := engine.EvaluateGenerator("$uuid(invalid)", rng1); err == nil {
		t.Error("expected error when arguments passed to $uuid()")
	}
}

func TestEvaluateGenerator_StaticAndLegacy(t *testing.T) {
	rng := rand.New(rand.NewPCG(400, 0))

	// Static string
	valStatic, err := engine.EvaluateGenerator("normal_string_value", rng)
	if err != nil || valStatic != "normal_string_value" {
		t.Errorf("expected untouched static string, got %q, err: %v", valStatic, err)
	}

	// Legacy int(min, max) format
	valLegacy, err := engine.EvaluateGenerator("int(10, 50)", rng)
	if err != nil {
		t.Fatalf("unexpected error for legacy int: %v", err)
	}
	intVal, err := strconv.Atoi(valLegacy)
	if err != nil || intVal < 10 || intVal > 50 {
		t.Errorf("expected integer in [10, 50], got %q", valLegacy)
	}

	// Unknown generator
	if _, err := engine.EvaluateGenerator("$unknown_func(1, 2)", rng); err == nil {
		t.Error("expected error for unknown generator function")
	}
}

func TestPRNG_EvaluateParam(t *testing.T) {
	prng := engine.NewPRNG(42)
	rng := rand.New(rand.NewPCG(42, 0))

	valInt := prng.EvaluateParam("$random_int(10, 50)", rng)
	if valInt == "$random_int(10, 50)" || valInt == "" {
		t.Errorf("failed to evaluate $random_int")
	}

	valChoice := prng.EvaluateParam("$random_choice('A', 'B')", rng)
	if valChoice != "A" && valChoice != "B" {
		t.Errorf("failed to evaluate $random_choice: %q", valChoice)
	}

	valStr := prng.EvaluateParam("$random_string(6)", rng)
	if len(valStr) != 6 {
		t.Errorf("failed to evaluate $random_string: %q", valStr)
	}

	valUUID := prng.EvaluateParam("$uuid()", rng)
	if len(valUUID) != 36 {
		t.Errorf("failed to evaluate $uuid: %q", valUUID)
	}

	valStatic := prng.EvaluateParam("static_string", rng)
	if valStatic != "static_string" {
		t.Errorf("expected static string to be untouched, got %q", valStatic)
	}
}

func TestRunner_DynamicGeneratorsIntegration(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()

	spec := domain.Spec{
		Name: "generators_integration_test",
		Database: domain.DatabaseConfig{
			Schema: "CREATE TABLE events (id TEXT PRIMARY KEY, type TEXT, amount INT, code TEXT);",
		},
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 5,
			Seed:       999,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "events_exist",
				Query:  "SELECT COUNT(*) AS total FROM events;",
				Assert: "total == 5",
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name:   "insert_event",
				Weight: 1.0,
				Params: map[string]string{
					"event_id": "$uuid()",
					"type":     "$random_choice('DEPOSIT', 'WITHDRAW', 'TRANSFER')",
					"amount":   "$random_int(10, 500)",
					"code":     "$random_string(8)",
				},
				Steps: []domain.StepConfig{
					{SQL: "INSERT INTO events (id, type, amount, code) VALUES ('{event_id}', '{type}', {amount}, '{code}');"},
				},
			},
		},
	}

	runner := engine.NewRunner(driver, spec.Engine.Seed)
	result, err := runner.Run(ctx, spec)
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success, invariant violation detected: %v", result.FailingInvariant)
	}

	if len(result.ScheduledOps) != 5 {
		t.Fatalf("expected 5 scheduled ops, got %d", len(result.ScheduledOps))
	}

	for i, op := range result.ScheduledOps {
		if op.Params["event_id"] == "$uuid()" || len(op.Params["event_id"]) != 36 {
			t.Errorf("op %d event_id not evaluated properly: %q", i, op.Params["event_id"])
		}
		if op.Params["type"] != "DEPOSIT" && op.Params["type"] != "WITHDRAW" && op.Params["type"] != "TRANSFER" {
			t.Errorf("op %d type not evaluated properly: %q", i, op.Params["type"])
		}
		amount, err := strconv.Atoi(op.Params["amount"])
		if err != nil || amount < 10 || amount > 500 {
			t.Errorf("op %d amount not evaluated properly: %q", i, op.Params["amount"])
		}
		if len(op.Params["code"]) != 8 {
			t.Errorf("op %d code not evaluated properly: %q", i, op.Params["code"])
		}
	}
}
