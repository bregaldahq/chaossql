package main

import (
	"testing"
)

func TestResolveDemoPath(t *testing.T) {
	tests := []struct {
		alias        string
		expectedPath string
		expectErr    bool
	}{
		{"banking", "examples/banking_lost_update/chaos.yaml", false},
		{"inventory", "examples/inventory_oversell/chaos.yaml", false},
		{"hospital", "examples/hospital_write_skew/chaos.yaml", false},
		{"financial", "examples/read_skew_financial_audit/chaos.yaml", false},
		{"auction", "examples/dirty_write_auction/chaos.yaml", false},
		{"crypto", "examples/circular_info_crypto_arbitrage/chaos.yaml", false},
		{"flash_crash", "examples/dirty_read_flash_crash/chaos.yaml", false},
		{"ticket", "examples/ticket_booking_anti_dependency/chaos.yaml", false},
		{"deadlock", "examples/deadlock_cycle/chaos.yaml", false},
		{"fk", "examples/foreign_key_cascade_deadlock/chaos.yaml", false},
		{"cascade", "examples/foreign_key_cascade_deadlock/chaos.yaml", false},
		{"foreign_key", "examples/foreign_key_cascade_deadlock/chaos.yaml", false},
		{"foreign_key_cascade_deadlock", "examples/foreign_key_cascade_deadlock/chaos.yaml", false},
		{"unknown_scenario", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			path, err := resolveDemoPath(tt.alias)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error for alias %q, got nil", tt.alias)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for alias %q: %v", tt.alias, err)
			}
			if path != tt.expectedPath {
				t.Errorf("expected %q, got %q", tt.expectedPath, path)
			}
		})
	}
}
