# ADR 0001: Deterministic Pseudo-Random Generation and Seeding

* **Status:** Accepted
* **Date:** 2026-09-01

## Context
For a concurrency fuzzer to be practical and actionable in software engineering, any detected failure MUST be 100% reproducible across different machines and in CI/CD pipelines.

## Decision
Adopt a dedicated instance of `rand.Rand` (or `random.Random(seed)`) per execution run. All parameters, delays, and operation schedules are deterministically derived from this seed, ensuring that identical seeds invariably produce identical execution plans.

## Consequences
* Developers can share only the seed (e.g., `--seed 42`) to reproduce bugs deterministically.
* The shrinking algorithm can re-execute sub-schedules with the mathematical guarantee that parameters and values remain invariant.
