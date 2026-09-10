//go:build !js || !wasm

package drivers_test

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
)

func TestPostgresEffectiveIsolation(t *testing.T) {
	driver := drivers.NewPostgresDriver("unused")
	assertEffectiveIsolation(t, driver, "", domain.LevelReadCommitted)
	assertEffectiveIsolation(t, driver, domain.LevelReadCommitted, domain.LevelReadCommitted)
	assertEffectiveIsolation(t, driver, domain.LevelRepeatableRead, domain.LevelRepeatableRead)
	assertEffectiveIsolation(t, driver, domain.LevelSerializable, domain.LevelSerializable)
	assertUnsupportedIsolation(t, driver, domain.LevelReadUncommitted)
}

func TestMySQLEffectiveIsolation(t *testing.T) {
	driver := drivers.NewMySQLDriver("unused")
	assertEffectiveIsolation(t, driver, "", domain.LevelRepeatableRead)
	for _, level := range allIsolationLevels() {
		assertEffectiveIsolation(t, driver, level, level)
	}
}

func TestSQLiteEffectiveIsolation(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	assertEffectiveIsolation(t, driver, "", domain.LevelSerializable)
	assertEffectiveIsolation(t, driver, domain.LevelSerializable, domain.LevelSerializable)
	assertEffectiveIsolation(t, driver, domain.LevelReadUncommitted, domain.LevelReadUncommitted)
	assertUnsupportedIsolation(t, driver, domain.LevelReadCommitted)
	assertUnsupportedIsolation(t, driver, domain.LevelRepeatableRead)
}

func TestMockEffectiveIsolation(t *testing.T) {
	driver := drivers.NewMockDriver()
	defer driver.Close()
	assertEffectiveIsolation(t, driver, "", domain.LevelSerializable)
	for _, level := range allIsolationLevels() {
		assertEffectiveIsolation(t, driver, level, level)
	}
}

type isolationResolver interface {
	EffectiveIsolation(domain.IsolationLevel) (domain.IsolationLevel, error)
}

func assertEffectiveIsolation(t *testing.T, resolver isolationResolver, requested, expected domain.IsolationLevel) {
	t.Helper()
	actual, err := resolver.EffectiveIsolation(requested)
	if err != nil {
		t.Fatalf("resolve isolation %q: %v", requested, err)
	}
	if actual != expected {
		t.Fatalf("isolation %q resolved to %q, want %q", requested, actual, expected)
	}
}

func assertUnsupportedIsolation(t *testing.T, resolver isolationResolver, requested domain.IsolationLevel) {
	t.Helper()
	if _, err := resolver.EffectiveIsolation(requested); err == nil {
		t.Fatalf("expected isolation %q to be rejected", requested)
	}
}

func allIsolationLevels() []domain.IsolationLevel {
	return []domain.IsolationLevel{
		domain.LevelReadUncommitted,
		domain.LevelReadCommitted,
		domain.LevelRepeatableRead,
		domain.LevelSerializable,
	}
}
