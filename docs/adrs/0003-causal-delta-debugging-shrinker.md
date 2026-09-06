# ADR 0003: Causal Delta-Debugging (ddmin) for SQL Transactions

* **Status:** Accepted
* **Date:** 2026-09-01

## Context
The classic Delta-Debugging algorithm (Zeller '99) assumes elements in the candidate set are mutually independent. In relational SQL databases, a transaction $T_j$ may depend on the existence of a foreign key or entity created by transaction $T_i$. Naively removing $T_i$ triggers *Foreign Key Constraint* failures, corrupting the oracle.

## Decision
Implement **Causal Delta-Debugging**:
1. Construct an acyclic graph of causal dependencies ($T_i \to T_j$) based on generated runtime parameters.
2. When testing a candidate subset $C'$, compute the transitive closure $\text{Closure}(C')$ to ensure all causal prerequisites are retained.

## Consequences
* Total elimination of false positives caused by SQL referential integrity errors.
* Reduction of shrinker reduction steps by up to 75%.
