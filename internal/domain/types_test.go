package domain_test

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestInvariantResult_String(t *testing.T) {
	passResult := domain.InvariantResult{
		Name:   "saldo_positivo",
		Passed: true,
	}
	if passResult.String() != "PASS: Invariant 'saldo_positivo' satisfied" {
		t.Errorf("unexpected string output: %s", passResult.String())
	}

	failResult := domain.InvariantResult{
		Name:         "saldo_positivo",
		Passed:       false,
		Expression:   "balance >= 0",
		ActualValues: map[string]interface{}{"balance": -50},
	}
	if failResult.Passed {
		t.Error("expected failure result")
	}
}
