# ADR 0005: Self-Contained Reproduction Test Synthesis (repro_test.go)

* **Status:** Accepted
* **Date:** 2026-09-01

## Context
When a concurrency anomaly is discovered and minimized, developers and CI pipelines need to reproduce it without requiring a full ChaosSQL installation.

## Decision
Synthesize a standalone, self-contained **`repro_test.go`** file containing:
1. Embedded schema and seed definitions.
2. The exact concurrent goroutines representing the 2 or 3 transactions from the minimal trace.
3. Assertions verifying the violated invariant.

## Consequences
* Any developer can run `go test -v repro_test.go` and observe the bug in ~0.1s.
* Serves as irrefutable proof in Pull Requests and issue trackers.
