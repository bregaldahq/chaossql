# Scenario 04: Financial Audit Read Skew (Anomaly A5A)

## Business Context
In a financial institution, a customer holds two linked accounts: **Checking** (initial balance: $500) and **Savings** (initial balance: $500). Total customer net wealth is $1,000.
An audit or consolidated reporting process ($T_1$) computes total wealth by reading the balance of both accounts sequentially. Concurrently, fund transfer transactions ($T_2$) move capital between the accounts.

## Anomaly Breakdown
The **A5A (Read Skew)** anomaly occurs when an inconsistent read transaction observes a partial state of another concurrent transaction under `READ COMMITTED`:
1. $T_1$ (Audit) reads the **Checking** account ($x = 500$).
2. $T_2$ (Transfer) transfers $100 from Checking to Savings:
   - Decrements Checking: $x \leftarrow 400$
   - Increments Savings: $y \leftarrow 600$
   - Records transfer in `transfers`.
   - $T_2$ commits successfully.
3. $T_1$ reads the **Savings** account ($y = 600$), observing the write committed by $T_2$.
4. $T_1$ calculates total wealth: $500 + 600 = 1100 \ne 1000$.

### Mathematical Formulation (Adya / Berenson et al.)
In Adya's direct serialization graph (DSG), Read Skew is characterized by a cycle containing a read anti-dependency ($rw$) and a write-read dependency ($wr$):
$$T_1 \xrightarrow{rw} T_2 \xrightarrow{wr} T_1$$

Where:
- $T_1 \xrightarrow{rw} T_2$ on item $x$ (`accounts:1` / Checking): $T_1$ read the version of $x$ prior to the modification made by $T_2$.
- $T_2 \xrightarrow{wr} T_1$ on item $y$ (`accounts:2` / Savings): $T_1$ read the version of $y$ written and committed by $T_2$.

## Wealth Conservation Invariant
$$\text{total\_balance} == 1000$$

## Formal Mitigation
* **SERIALIZABLE or REPEATABLE READ / SNAPSHOT ISOLATION Level:**
  Ensures $T_1$ reads a temporally consistent, frozen-in-time snapshot of the database (snapshot established at the start of $T_1$), reading $x = 500$ and $y = 500$ ($\text{total} = 1000$).
* **Atomic Single-Statement Query:**
  Execute aggregate balance checks in a single SQL statement (`SELECT SUM(balance) FROM accounts;`), guaranteeing the database evaluates the sum across a single statement snapshot.
* **Atomic Transactions with Pessimistic Locking (`SELECT ... FOR UPDATE`):**
  Lock both records during audit evaluation to block concurrent modifications until $T_1$ completes:
  ```sql
  SELECT balance FROM accounts WHERE id IN (1, 2) FOR UPDATE;
  ```
