package analyzer_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestRegister_ValidLinearizableHistory(t *testing.T) {
	t.Run("clean read-after-write with single variable", func(t *testing.T) {
		events := []analyzer.RegisterEvent{
			{TxID: "T1", Var: "x", Op: analyzer.OpWrite, Val: "100"},
			{TxID: "T1", Op: analyzer.OpCommit},
			{TxID: "T2", Var: "x", Op: analyzer.OpRead, Val: "100", ReadFromTx: "T1"},
			{TxID: "T2", Op: analyzer.OpCommit},
		}

		res, err := analyzer.CheckRegisterLinearizability(events)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatal("expected non-nil result")
		}
		if !res.Linearizable {
			t.Errorf("expected history to be linearizable, got violations: %+v", res.Violations)
		}
		if len(res.Violations) != 0 {
			t.Errorf("expected 0 violations, got %d", len(res.Violations))
		}
	})

	t.Run("clean multi-variable history", func(t *testing.T) {
		events := []analyzer.RegisterEvent{
			{TxID: "T1", Var: "x", Op: analyzer.OpWrite, Val: "v1"},
			{TxID: "T1", Var: "y", Op: analyzer.OpWrite, Val: "w1"},
			{TxID: "T1", Op: analyzer.OpCommit},
			{TxID: "T2", Var: "x", Op: analyzer.OpRead, Val: "v1", ReadFromTx: "T1"},
			{TxID: "T2", Var: "y", Op: analyzer.OpRead, Val: "w1", ReadFromTx: "T1"},
			{TxID: "T2", Op: analyzer.OpCommit},
		}

		res, err := analyzer.CheckRegisterLinearizability(events)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Linearizable || len(res.Violations) != 0 {
			t.Errorf("expected linearizable history, got violations: %+v", res.Violations)
		}
	})

	t.Run("clean list append history", func(t *testing.T) {
		events := []analyzer.RegisterEvent{
			{TxID: "T1", Var: "list1", Op: analyzer.OpAppend, Val: "A"},
			{TxID: "T1", Var: "list1", Op: analyzer.OpAppend, Val: "B"},
			{TxID: "T1", Op: analyzer.OpCommit},
			{TxID: "T2", Var: "list1", Op: analyzer.OpRead, Val: "[A, B]"},
			{TxID: "T2", Op: analyzer.OpCommit},
		}

		res, err := analyzer.CheckRegisterLinearizability(events)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Linearizable || len(res.Violations) != 0 {
			t.Errorf("expected linearizable list append, got violations: %+v", res.Violations)
		}
	})
}

func TestRegister_G1bIntermediateRead(t *testing.T) {
	// T1 writes x=1, then T1 writes x=2 and commits; T2 reads x=1 (an intermediate uncommitted version).
	// Assert G1b detected with offending transactions and variable.
	events := []analyzer.RegisterEvent{
		{TxID: "T1", Var: "x", Op: analyzer.OpWrite, Val: "1"},
		{TxID: "T1", Var: "x", Op: analyzer.OpWrite, Val: "2"},
		{TxID: "T1", Op: analyzer.OpCommit},
		{TxID: "T2", Var: "x", Op: analyzer.OpRead, Val: "1", ReadFromTx: "T1"},
		{TxID: "T2", Op: analyzer.OpCommit},
	}

	res, err := analyzer.CheckRegisterLinearizability(events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if res.Linearizable {
		t.Fatal("expected history to NOT be linearizable due to G1b intermediate read")
	}

	var foundG1b bool
	for _, v := range res.Violations {
		if v.Type == domain.AnomalyG1bIntermediateRead {
			foundG1b = true
			if v.Var != "x" {
				t.Errorf("expected violation variable 'x', got %q", v.Var)
			}
			if v.TxID != "T2" {
				t.Errorf("expected reading transaction 'T2', got %q", v.TxID)
			}
			if v.OffendingTx != "T1" {
				t.Errorf("expected offending writer transaction 'T1', got %q", v.OffendingTx)
			}
			if v.Message == "" {
				t.Error("expected non-empty diagnostic message")
			}
		}
	}

	if !foundG1b {
		t.Errorf("expected AnomalyG1bIntermediateRead violation, got violations: %+v", res.Violations)
	}
}

func TestRegister_FracturedRead(t *testing.T) {
	// T1 appends [A, B] and commits; T2 reads [B] missing A.
	// Assert Fractured Read detected.
	t.Run("bracketed list append", func(t *testing.T) {
		events := []analyzer.RegisterEvent{
			{TxID: "T1", Var: "x", Op: analyzer.OpAppend, Val: "[A, B]"},
			{TxID: "T1", Op: analyzer.OpCommit},
			{TxID: "T2", Var: "x", Op: analyzer.OpRead, Val: "[B]"},
			{TxID: "T2", Op: analyzer.OpCommit},
		}

		res, err := analyzer.CheckRegisterLinearizability(events)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatal("expected non-nil result")
		}
		if res.Linearizable {
			t.Fatal("expected history to NOT be linearizable due to Fractured Read")
		}

		var foundFractured bool
		for _, v := range res.Violations {
			if v.Type == domain.AnomalyFracturedRead {
				foundFractured = true
				if v.Var != "x" {
					t.Errorf("expected violation variable 'x', got %q", v.Var)
				}
				if v.TxID != "T2" {
					t.Errorf("expected reading transaction 'T2', got %q", v.TxID)
				}
				if v.OffendingTx != "T1" {
					t.Errorf("expected offending transaction 'T1', got %q", v.OffendingTx)
				}
				if v.Message == "" {
					t.Error("expected non-empty diagnostic message")
				}
			}
		}

		if !foundFractured {
			t.Errorf("expected AnomalyFracturedRead violation, got violations: %+v", res.Violations)
		}
	})

	t.Run("sequential appends missing prefix", func(t *testing.T) {
		events := []analyzer.RegisterEvent{
			{TxID: "T1", Var: "x", Op: analyzer.OpAppend, Val: "A"},
			{TxID: "T1", Var: "x", Op: analyzer.OpAppend, Val: "B"},
			{TxID: "T1", Op: analyzer.OpCommit},
			{TxID: "T2", Var: "x", Op: analyzer.OpRead, Val: "[B]"},
		}

		res, err := analyzer.CheckRegisterLinearizability(events)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Linearizable {
			t.Fatal("expected history to NOT be linearizable due to Fractured Read")
		}

		var foundFractured bool
		for _, v := range res.Violations {
			if v.Type == domain.AnomalyFracturedRead {
				foundFractured = true
				if v.Var != "x" {
					t.Errorf("expected violation on 'x', got %q", v.Var)
				}
			}
		}
		if !foundFractured {
			t.Errorf("expected AnomalyFracturedRead, got: %+v", res.Violations)
		}
	})
}

func TestRegister_ThreadSafetyAndZeroDataRaces(t *testing.T) {
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			events := []analyzer.RegisterEvent{
				{TxID: fmt.Sprintf("T%d-1", id), Var: fmt.Sprintf("k%d", id), Op: analyzer.OpWrite, Val: "init"},
				{TxID: fmt.Sprintf("T%d-1", id), Var: fmt.Sprintf("k%d", id), Op: analyzer.OpWrite, Val: "final"},
				{TxID: fmt.Sprintf("T%d-1", id), Op: analyzer.OpCommit},
				{TxID: fmt.Sprintf("T%d-2", id), Var: fmt.Sprintf("k%d", id), Op: analyzer.OpRead, Val: "init", ReadFromTx: fmt.Sprintf("T%d-1", id)},
			}

			res, err := analyzer.CheckRegisterLinearizability(events)
			if err != nil {
				t.Errorf("worker %d failed: %v", id, err)
				return
			}
			if res.Linearizable {
				t.Errorf("worker %d expected violation, got none", id)
			}
		}(i)
	}

	wg.Wait()
}

func TestRegister_EmptyEvents(t *testing.T) {
	res, err := analyzer.CheckRegisterLinearizability(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Linearizable || len(res.Violations) != 0 {
		t.Errorf("expected empty events to be linearizable, got: %+v", res)
	}
}
