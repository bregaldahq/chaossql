# Deadlock Cycle & Timeout Diagnostics

## Business Context
In financial ledger systems, concurrent bilateral fund transfers between Account 1 and Account 2 can lead to classic circular wait deadlocks:
- Transaction 1 decrements Account 1, then attempts to increment Account 2.
- Transaction 2 decrements Account 2, then attempts to increment Account 1.

Under row-level locking or table-level locks, both transactions wait indefinitely for each other's locks unless the database detects the cycle and aborts one transaction (`ErrDeadlockDetected` / `SQLSTATE 40P01` / MySQL 1213).

## Theoretical Formulation
A deadlock occurs when a directed wait-for graph $WFG = (V, E)$ contains a directed cycle:

$$T_1 \xrightarrow{\text{waits-for}} T_2 \xrightarrow{\text{waits-for}} T_1$$

ChaosSQL validates that database drivers and application invariants preserve total wealth even under continuous deadlock aborts and rollbacks.
