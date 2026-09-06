# Scenario 01: Fintech Lost Update (Anomaly P4)

## Business Context
In a banking system, two concurrent withdrawals occur on *Alice*'s account (initial balance: $1,000).

## Anomaly Breakdown
1. **Worker 1** reads the balance ($1,000) and prepares a withdrawal of $100 (expected final balance: $900).
2. **Worker 2** reads the same balance ($1,000) before Worker 1 commits, preparing a withdrawal of $150 (expected final balance: $850).
3. Both workers write their in-memory calculated balances back to the database.

* **Real-World Impact:** The balance becomes $850, but the audit ledger recorded $250 in total debits! The bank loses $100 due to a lost update anomaly ($P4$ / $G\text{-single}$).

## Consistency Invariant
$$\text{Current Balance} == 1000 - \sum(\text{Ledger Debits})$$

## Formal Mitigation
* **Atomic In-Database Update:**
  ```sql
  UPDATE accounts SET balance = balance - :amount WHERE id = 1 AND balance >= :amount;
  ```
* **Pessimistic Concurrency Control (2PL):**
  Acquire an exclusive row lock prior to evaluating balance:
  ```sql
  SELECT balance FROM accounts WHERE id = 1 FOR UPDATE;
  ```
* **Serializable Isolation:**
  Under `SERIALIZABLE` isolation, the database detects the read-write dependency cycle ($T_1 \xrightarrow{rw} T_2 \xrightarrow{ww} T_1$) and aborts one of the conflicting transactions with a serialization failure.
