//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteRestrictedAtomicUsesOwnerOnlyPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "finding.json")
	if err := writeRestrictedAtomic(path, []byte("{}\n")); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("permissions = %o, want 600", got)
	}
}
