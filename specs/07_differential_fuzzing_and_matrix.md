# Spec 07: Differential Fuzzing & Hermitage Isolation Matrix

## 1. Differential Fuzzing Architecture (`chaossql diff`)
- Objective: Detect isolation semantic divergence between two database engines (or isolation levels).
- Execution:
  1. Instantiates Driver A (e.g. SQLite) and Driver B (e.g. Postgres / SQLite-serializable).
  2. Executes identical schedules with synchronized PRNG sub-seeds.
  3. Compares:
     - Invariant evaluation results (Divergence = True if Driver A fails while Driver B passes).
     - Execution traces and Serialization Graphs.

## 2. Hermitage Isolation Matrix (`chaossql matrix`)
- Runs the canonical Hermitage test suite (P1 Dirty Read, P2 Non-Repeatable Read, P3 Phantom Read, P4 Lost Update, A5A Read Skew, A5B Write Skew, G0 Dirty Write, G1c Circular Info) against target engine.
- Generates a markdown and terminal compatibility matrix.

## 3. G1c Circular Information Flow
- Occurs when transaction $T_1$ reads $x$ written by $T_2$, and $T_2$ reads $y$ written by $T_1$:
  $$T_1 \xrightarrow{wr} T_2 \xrightarrow{wr} T_1$$
- Permitted under weak isolation, violates Strict Serializability.
