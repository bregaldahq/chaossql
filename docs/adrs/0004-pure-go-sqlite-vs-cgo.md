# ADR 0004: Pure Go SQLite (modernc.org/sqlite) vs CGO

* **Status:** Accepted
* **Date:** 2026-09-01

## Context
The traditional `github.com/mattn/go-sqlite3` driver requires CGO, hindering seamless cross-compilation (Linux, macOS, Windows), preventing fully static binary creation, and introducing goroutine-to-C context switch overhead.

## Decision
Adopt `modernc.org/sqlite`, a direct transpilation of SQLite into pure Go.

## Consequences
* **Instant Cross-Compilation:** ChaosSQL binaries can be built for any target operating system using `CGO_ENABLED=0 go build`.
* **Memory Safety & Portability:** Managed directly by the Go runtime without external dynamic library linkages.
