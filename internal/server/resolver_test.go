package server

import (
	"net"
	"testing"
)

// stubLookupIP replaces DNS resolution for the rest of the test and restores
// the previous resolver on cleanup. Always use it instead of swapping the
// resolver by hand: background dispatchers read it concurrently.
func stubLookupIP(t *testing.T, fn func(string) ([]net.IP, error)) {
	t.Helper()
	previous := resolver.Load()
	resolver.Store(&fn)
	t.Cleanup(func() { resolver.Store(previous) })
}
