# ADR 0002: Async Step Interleaving and Jitter Injection

* **Status:** Accepted
* **Date:** 2026-09-01

## Context
Simply executing transactions in parallel rarely triggers narrow race condition windows if queries execute too rapidly without temporal friction.

## Decision
Decompose each operation into discrete steps (`steps`) and inject micro-jitter (`time.Sleep` / `asyncio.sleep(delay_ms / 1000.0)`) between steps within the same transaction.

## Consequences
* Drastically increases the probability of collision between concurrent reads and writes.
* Enables testing and verification of high-latency and bursty database behavior.
