# Scenario 03: Hospital Write Skew (Anomaly A5B)

## Business Context
Hospital operational rule: **At least one doctor must remain on active on-call duty at all times**.
Initially, both Dr. Alice and Dr. Bob are on call (`is_on_call = 1`).

## Anomaly Breakdown
1. **Dr. Alice** attempts to leave on-call duty: queries active doctors (returns 2 active). Since $2 \ge 2$, she updates her status to 0 (`is_on_call = 0`) and commits.
2. **Dr. Bob** concurrently attempts to leave on-call duty: queries active doctors (returns 2 active in his snapshot). Since $2 \ge 2$, he updates his status to 0 (`is_on_call = 0`) and commits.
3. **Catastrophic Outcome:** **Zero doctors on call!** The hospital invariant is violated because the transactions modify disjoint rows ($\mathcal{W}_1 \cap \mathcal{W}_2 = \emptyset$), bypassing traditional write-write conflict detection under Snapshot Isolation.

## Hospital Invariant
$$\sum(\text{is\_on\_call}) \ge 1$$

## Formal Mitigation
* **SERIALIZABLE Isolation (SSI):**
  The database engine (e.g., PostgreSQL SSI) tracks anti-dependency cycles ($T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_1$) and aborts one of the conflicting transactions with `SQLSTATE 40001` (`serialization_failure`).
* **Pessimistic Locking (`SELECT ... FOR UPDATE`):**
  Lock the evaluated doctor rows to serialize concurrent reads and writes:
  ```sql
  SELECT is_on_call FROM doctors WHERE is_on_call = 1 FOR UPDATE;
  ```
* **Materialized Conflict:**
  Lock a common department or shift parent record to turn predicate-like anti-dependencies into explicit write-write conflicts.
