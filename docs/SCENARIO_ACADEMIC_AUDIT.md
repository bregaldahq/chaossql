# Academic Audit & Formal Review of ChaosSQL Scenarios

This document provides the formal and epistemological audit of ChaosSQL demonstration and fuzzing scenarios, linking each scenario to seminal academic publications in concurrency control and database isolation (SIGMOD, VLDB, ACM TODS, CACM).

---

## 1. Academic Alignment Matrix

| Scenario | Anomaly Phenomenon | Adya Classification | Seminal Paper & Theorem | Formal Solution |
| :--- | :--- | :--- | :--- | :--- |
| **Banking** (`banking_lost_update`) | **P4** (Lost Update) | **G-single** (Cycle $rw + ww$) | Berenson et al. (*SIGMOD 1995*) | In-database Atomic Update or Pessimistic 2PL (`SELECT FOR UPDATE`) |
| **Inventory** (`inventory_oversell`) | **A3 / P4** (Predicate Depletion / Phantom) | **G-phantom** (Predicate Conflict) | Eswaran et al. (*CACM 1976*) | Guarded Decrement (`WHERE stock >= requested_qty`) or Predicate Locking |
| **Hospital** (`hospital_write_skew`) | **A5B** (Write Skew) | **G2-item** (Cycle $rw + rw$) | Fekete et al. (*TODS 2005*) & Ports (*VLDB 2012*) | Serializable Snapshot Isolation (SSI) or Materialized Conflict Locks |
| **Financial** (`read_skew_financial_audit`) | **A5A** (Read Skew) | **G1c** / $rw$-directed skew | Berenson et al. (*SIGMOD 1995*) & Gray (*IFIP 1975*) | Snapshot Isolation / Repeatable Read or Single-Statement Atomic Read |
| **Auction** (`dirty_write_auction`) | **G0** (Dirty Write) | **G0** (Cycle $ww + ww$) | Adya (*MIT 1999*) & Berenson et al. (*SIGMOD 1995*) | Two-Phase Write Locking (Strict PL-1 / Read Committed) |
| **Crypto Arbitrage** (`circular_info_crypto`) | **G1c** (Circular Information Flow) | **G1c** (Cycle of $wr$ and $ww$) | Adya (*PhD Thesis, MIT 1999*) | Snapshot Isolation / Repeatable Read |
| **Flash Crash** (`dirty_read_flash_crash`) | **A1 / G1a** (Aborted / Dirty Read) | **G1a** (Read Uncommitted Version) | Berenson et al. (*SIGMOD 1995*) | Read Committed (Prohibit uncommitted version reads) |
| **Ticket Booking** (`ticket_anti_dependency`) | **A5A / G2-item** (Anti-Dependency Skew) | **G2-item** ($rw$-conflict on availability) | Cahill, Röhm, Fekete (*SIGMOD 2008*) | Row Lock (`SELECT FOR UPDATE`) or Serializable Snapshot Isolation |
| **Deadlock Cycle** (`deadlock_cycle`) | **G-DL** (Resource Wait-For Cycle) | Cycle in Wait-For Graph ($WFG$) | Coffman et al. (*ACM Comp. Surveys 1971*) | Canonical Resource Ordering ($R_1 < R_2$) or WFG Detection with Abort |
| **FK Cascade Deadlock** (`fk_cascade_deadlock`) | **G-DL / Escalation** (Cascade Lock Conflict) | Implicit Dependency Lock Cycle | Gray & Reuter (*Transaction Processing 1993*) | Indexed Foreign Keys and Ordered Parent Deletions |

---

## 2. Detailed Scenario Formal Audits

### Scenario 01: Banking Lost Update ($P4$)
* **Seminal Citation:** Berenson, Bernstein, Gray, Melton, O'Neil, O'Neil (*A Critique of ANSI SQL Isolation Levels*, SIGMOD 1995).
* **Adya Formalism (1999):** $G\text{-single}$ anomaly. The concurrent schedule contains a directed cycle:
  $$T_1 \xrightarrow{rw} T_2 \xrightarrow{ww} T_1$$
* **Empirical Diagnostic:** `READ COMMITTED` (the default in PostgreSQL and Oracle) **does not protect** against this anomaly when applications perform in-memory Read-Modify-Write patterns.
* **Formal Solution:** Atomic in-database expressions (`UPDATE accounts SET balance = balance - 50 WHERE id = 1`) or explicit Two-Phase Locking (`SELECT balance FROM accounts WHERE id = 1 FOR UPDATE`).

---

### Scenario 02: Inventory Oversell ($A3$ / Predicate Depletion)
* **Seminal Citation:** Eswaran, Gray, Lorie, Traiger (*The Notions of Consistency and Predicate Locks in a Database System*, CACM 1976).
* **Predicate Formalism:** The business invariant asserts an aggregate capacity condition $\sum \text{quantity} \le \text{initial_stock}$. Predicate reads ($r_i[P]$) are invalidated by concurrent insertions ($w_j[y \in P]$).
* **Empirical Diagnostic:** A simple column constraint (`stock >= 0`) on the inventory table does not prevent concurrent purchasing transactions from accumulating more orders than available units if orders are recorded in a separate table.
* **Formal Solution:** Guarded atomic conditional decrement (`UPDATE items SET stock = stock - :qty WHERE id = :id AND stock >= :qty`) coupled with transaction rollback when row count equals 0.

---

### Scenario 03: Hospital Write Skew ($A5B$)
* **Seminal Citations:**
  1. Berenson et al. (*SIGMOD 1995*) — Definition of phenomenon $A5B$.
  2. Fekete, Li, O'Neil, O'Neil (*Making Snapshot Isolation Serializable*, ACM TODS 2005) — **Dangerous Structure Theorem**.
  3. Ports & Grittner (*Serializable Snapshot Isolation in PostgreSQL*, VLDB 2012).
* **Fekete Dangerous Structure Theorem (TODS 2005):**
  Every non-serializable execution history under Snapshot Isolation MUST contain two consecutive anti-dependency ($rw$) edges forming a cycle:
  $$T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_1$$
* **Empirical Diagnostic:** Because $\mathcal{W}_1 \cap \mathcal{W}_2 = \emptyset$ (Dr. Alice writes only to row 1 and Dr. Bob writes only to row 2), MVCC engines **detect zero write conflicts** under `REPEATABLE READ`, allowing both doctors to go off-call simultaneously and leaving zero doctors on duty.
* **Formal Solution:** True Serializable Snapshot Isolation (SSI) via SIREAD lock tracking or explicit lock promotion via `SELECT ... FOR UPDATE` on shared hospital department records.

---

### Scenario 04: Read Skew Financial Audit ($A5A$)
* **Seminal Citations:**
  1. Berenson et al. (*SIGMOD 1995*) — Formal definition of Read Skew ($A5A$).
  2. Gray, Lorie, Putzolu, Traiger (*Granularity of Locks and Degrees of Consistency in a Shared Data Base*, IFIP 1975).
* **Adya Formalism (1999):** An audit transaction $T_{\text{audit}}$ reads balance $x$, concurrently a transfer transaction $T_{\text{transfer}}$ moves funds $x \to y$ and commits, after which $T_{\text{audit}}$ reads updated balance $y$. The audit history produces a directed path $T_{\text{audit}} \xrightarrow{rw} T_{\text{transfer}} \xrightarrow{wr} T_{\text{audit}}$, observing inconsistent snapshots where the total balance sum violates the ledger invariant.
* **Empirical Diagnostic:** Under `READ COMMITTED`, each statement reads a newly committed snapshot rather than a transaction-consistent snapshot.
* **Formal Solution:** Elevate the transaction isolation level to `REPEATABLE READ` or execute the balance query as a single atomic SQL statement (`SELECT SUM(balance) FROM accounts`).

---

### Scenario 05: Dirty Write Auction ($G0$)
* **Seminal Citations:**
  1. Adya (*Weak Consistency: A Generalized Theory and Optimistic Protocols for Distributed Transactions*, MIT 1999) — Definition of Phenotype $G0$.
  2. Berenson et al. (*SIGMOD 1995*).
* **Adya Formalism (1999):** Anomaly $G0$ involves disjoint transactions overwriting uncommitted data written by each other:
  $$T_1 \xrightarrow{ww} T_2 \xrightarrow{ww} T_1$$
* **Empirical Diagnostic:** If transaction $T_1$ updates the auction high-bidder and high-bid amount, and transaction $T_2$ overwrites the high-bidder before $T_1$ commits, the resulting state may attribute the highest bid amount to the wrong bidder.
* **Formal Solution:** Strict PL-1 / Read Committed enforcement via long-duration exclusive write locks that prevent uncommitted data from being overwritten.

---

### Scenario 06: Circular Information Flow Crypto Arbitrage ($G1c$)
* **Seminal Citation:** Adya (*MIT 1999*).
* **Adya Formalism (1999):** Cycle formed strictly by write-read ($wr$) and write-write ($ww$) edges between two currency exchange transactions, violating causal consistency even when neither transaction reads uncommitted data.
* **Formal Solution:** Snapshot Isolation or Multi-Version Serialization Graph Checking.

---

### Scenario 07: Dirty Read Flash Crash ($G1a / A1$)
* **Seminal Citation:** Berenson et al. (*SIGMOD 1995*).
* **Adya Formalism (1999):** Transaction $T_2$ reads a transient price modification written by $T_1$ which subsequently aborts ($a_1 \in T_1$).
* **Formal Solution:** Enforcement of `READ COMMITTED` level, forbidding access to uncommitted tuple revisions.

---

### Scenario 08: Ticket Booking Anti-Dependency ($A5A / G2\text{-item}$)
* **Seminal Citation:** Cahill, Röhm, Fekete (*Serializable Isolation for Snapshot Databases*, SIGMOD 2008).
* **Adya Formalism (1999):** Anti-dependency cycle on inventory allocation where concurrent reservations read the same open seats and write distinct confirmation records without detecting write set overlap.
* **Formal Solution:** Pessimistic row locking (`SELECT ... FOR UPDATE`) or SSI predicate collision detection.

---

### Scenario 09: Deadlock Cycle ($G\text{-DL}$)
* **Seminal Citation:** Coffman, Elphick, Aoshani (*System Deadlocks*, ACM Computing Surveys 1971).
* **Formalism:** Circular dependency in the Wait-For Graph: $T_1 \xrightarrow{\text{waits-for}} T_2 \xrightarrow{\text{waits-for}} T_1$.
* **Formal Solution:** Strictly enforced lexicographical acquisition ordering on resources or automatic engine deadlock detection with youngest transaction abort.

---

### Scenario 10: Foreign Key Cascade Deadlock ($G\text{-DL} / \text{Cascade}$)
* **Seminal Citation:** Gray & Reuter (*Transaction Processing: Concepts and Techniques*, Morgan Kaufmann 1993).
* **Formalism:** Cascade deletion locks child foreign key ranges in orders differing from concurrent child insertions, producing an implicit deadlock cycle.
* **Formal Solution:** Foreign key indexing on referencing child tables and deterministic parent row locking before cascading mutations.
