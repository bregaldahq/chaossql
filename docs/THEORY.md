# Mathematical and Formal Foundations of ChaosSQL

This document establishes the mathematical foundations, formal concurrency models, conflict graph theory, formal inductive invariant definitions, Burckhardt-Musuvathi lower bounds, Black-Box linearizability (Elle), and the convergence proof of the **Delta-Debugging ($ddmin$)** algorithm.

---

## 1. Formal Transaction and History Models

### 1.1 Transaction Definition
A transaction $T_i$ is formally defined as a tuple:
$$T_i = (O_i, <_i)$$

Where:
* $O_i$ is a finite set of read operations $r_i[x]$, write operations $w_i[x]$, abort operations $a_i$, or commit operations $c_i$, accessing data entities $x \in \mathcal{D}$.
* $<_i$ is a strict partial order over $O_i$ that preserves the causality of program control flow.
* $T_i$ contains exactly one terminal operation: $a_i \in O_i$ or $c_i \in O_i$, which is the maximal element under $<_i$.

### 1.2 History (Schedule) Definition
A concurrent execution history $S$ over a set of transactions $\mathcal{T} = \{T_1, T_2, \dots, T_n\}$ is a partial order $S = (\bigcup_{i=1}^n O_i, <_S)$ satisfying:
1. $\bigcup_{i=1}^n <_i \;\subseteq\; <_S$ (preserves the internal order of each transaction).
2. For any two conflicting operations $o_i, o_j \in S$ (where $i \neq j$), either $o_i <_S o_j$ or $o_j <_S o_i$.

### 1.3 Bernstein Conflict Conditions
Two operations $o_i, o_j \in S$ ($i \neq j$) are in **direct conflict** if they access the same item $x \in \mathcal{D}$ and at least one of them is a write mutation:
$$\text{Conflict}(o_i, o_j) \iff (\text{target}(o_i) = \text{target}(o_j) = x) \;\land\; (o_i \text{ is } w \;\lor\; o_j \text{ is } w)$$

---

## 2. Serialization Graph and Conflict Serializability (CSR) Theorem

### 2.1 The Serialization Graph $SG(S)$
Given a schedule $S$ over transactions $\mathcal{T}$, the Serialization Graph (or Direct Serialization Graph $DSG$) is a directed graph $SG(S) = (V, E)$ where:
* $V = \mathcal{T}$ (vertices represent transactions).
* A directed edge $(T_i \to T_j) \in E$ exists if and only if there exist $o_i \in T_i$ and $o_j \in T_j$ such that:
  $$o_i <_S o_j \quad \land \quad \text{Conflict}(o_i, o_j)$$

### 2.2 Fundamental Theorem of Conflict Serializability (CSR)
$$\text{Schedule } S \text{ is Conflict Serializable (CSR)} \iff SG(S) \text{ is a Directed Acyclic Graph (DAG)}$$

---

## 3. Black-Box Linearizability & Adya Dependency Inference (Elle)

To verify transactional isolation without requiring white-box database engine instrumentation, ChaosSQL applies the black-box dependency inference methodology formalized by Adya (1999) and Kingsbury & Alvaro (VLDB 2020):

### 3.1 Direct Dependency Edge Types
Let $T_i, T_j \in \mathcal{T}$ with $i \neq j$:
1. **wr (Write-Read / Dependency):** $T_i \xrightarrow{wr} T_j$ when $T_i$ creates a version of record $x$ and $T_j$ observes that version.
2. **ww (Write-Write / Overwrite):** $T_i \xrightarrow{ww} T_j$ when $T_i$ writes a version of record $x$ and $T_j$ installs a subsequent version of $x$.
3. **rw (Read-Write / Anti-Dependency):** $T_i \xrightarrow{rw} T_j$ when $T_i$ reads a version of record $x$ and $T_j$ overwrites $x$ with a newer version.

### 3.2 Cycle-Based Anomaly Characterization
- **G0 (Dirty Write):** Cycle in $SG(S)$ containing exclusively $\xrightarrow{ww}$ edges.
- **G1a (Aborted Read):** Path $T_i \xrightarrow{wr} T_j$ where $T_i$ aborts ($a_i \in T_i$).
- **G1b (Intermediate Read):** Path $T_i \xrightarrow{wr} T_j$ where $T_j$ reads a version generated prior to $T_i$'s commit.
- **G1c (Circular Information Flow):** Cycle in $SG(S)$ containing only $\xrightarrow{wr}$ and $\xrightarrow{ww}$ edges.
- **G-single (Lost Update):** Cycle of length 2 involving one anti-dependency edge and one overwrite edge: $T_1 \xrightarrow{rw} T_2 \xrightarrow{ww} T_1$.
- **G2-item (Write Skew):** Cycle in $SG(S)$ containing anti-dependency edges $\xrightarrow{rw}$, notably $T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_1$.

---

## 4. Mathematical Formalization of MVCC Anomalies

### 4.1 Lost Update ($P4$)
Occurs when two transactions $T_1$ and $T_2$ concurrently read state $x_0$, compute independent transformations $f(x_0)$ and $g(x_0)$, and both write back their updates without mutual locking:

$$S_{P4} = r_1[x_0] \;\dots\; r_2[x_0] \;\dots\; w_1[f(x_0)] \;\dots\; c_1 \;\dots\; w_2[g(x_0)] \;\dots\; c_2$$

* **Expected state:** $x_{\text{final}} = g(f(x_0))$ or $f(g(x_0))$ (cumulative effect).
* **Observed state:** $x_{\text{final}} = g(x_0)$ (mutation $f(x_0)$ of $T_1$ is completely overwritten and lost).
* **In the Conflict Graph:** $T_1 \xrightarrow{r_1 \to w_2} T_2$ and $T_2 \xrightarrow{r_2 \to w_1} T_1 \implies \text{Cycle } T_1 \rightleftarrows T_2$.

### 4.2 Write Skew ($A5B$)
Let a global integrity invariant be $\mathcal{I}(x, y): x + y \ge 0$.
Initial state: $x = 100, y = 100$ ($x + y = 200 \ge 0$).
* $T_1$ seeks to debit 150 from $x$: verifies $x+y \ge 150$ (valid in snapshot: 200) and executes $w_1[x \leftarrow -50]$.
* $T_2$ concurrently seeks to debit 150 from $y$: verifies $x+y \ge 150$ (valid in snapshot: 200) and executes $w_2[y \leftarrow -50]$.

$$S_{A5B} = r_1[x, y] \;\dots\; r_2[x, y] \;\dots\; w_1[x] \;\dots\; c_1 \;\dots\; w_2[y] \;\dots\; c_2$$

* **Observed state:** $x = -50, y = -50 \implies x + y = -100 < 0 \implies \mathcal{I}(x, y) = \text{FALSE}$.
* Under `REPEATABLE READ`, both transactions succeed because $w_1$ touches only $x$ and $w_2$ touches only $y$ (disjoint write sets $\mathcal{W}_1 \cap \mathcal{W}_2 = \emptyset$). However, their read sets intersect each other's writes: $\mathcal{R}_1 \cap \mathcal{W}_2 \neq \emptyset$ and $\mathcal{R}_2 \cap \mathcal{W}_1 \neq \emptyset$.

---

## 5. State Invariants and Transitions

A database is modeled as a state transition system:
$$\mathcal{M} = (\Sigma, \sigma_0, \Delta, \mathcal{I})$$

Where:
* $\Sigma$ is the universe of database states (tables, tuples, indexes).
* $\sigma_0 \in \Sigma$ is the initial state defined by the DDL schema (`schema.sql`) and seed (`seed.sql`).
* $\Delta: \Sigma \times \mathcal{T} \to \Sigma$ is the transition function applying a transaction to the current state.
* $\mathcal{I}: \Sigma \to \{0, 1\}$ is the business invariant predicate.

### 5.1 Inductive Invariants
A predicate $\mathcal{I}$ is inductive for $\mathcal{M}$ if and only if:
1. **Base Case:** $\mathcal{I}(\sigma_0) = 1$
2. **Inductive Step:** $\forall \sigma \in \Sigma, \forall T \in \mathcal{T}: (\mathcal{I}(\sigma) = 1) \implies (\mathcal{I}(\Delta(\sigma, T)) = 1)$

### 5.2 Failure Under Non-Atomic Interleaving
Under chaotic scheduling, execution occurs not as an atomic transition $\Delta(\sigma, T)$, but across micro-steps $s_{i,k}$:
$$\sigma_{t+1} = \delta_{\text{step}}(\sigma_t, s_{i,k})$$

If the history $S$ is not conflict serializable ($S \not\in \text{CSR}$), the final state satisfies:
$$\sigma_{\text{final}} = \delta_{\text{step}}(\dots \delta_{\text{step}}(\sigma_0, s_{1,1}) \dots s_{n,m}) \quad \text{such that} \quad \mathcal{I}(\sigma_{\text{final}}) = 0$$

---

## 6. Burckhardt-Musuvathi Lower Bound Theorem

In Probabilistic Concurrency Testing (PCT), Burckhardt, Musuvathi et al. (ASPLOS 2010) proved a rigorous lower bound on bug detection probability in concurrent executions.

### 6.1 Theorem Statement
Let a concurrent program have $n$ threads executing at most $k$ steps each (total steps $\le n \cdot k$). If a concurrency bug has scheduling depth $d$ (requiring $d$ specific priority changes or context switches to be triggered), the PCT randomized priority scheduling algorithm detects the bug in a single execution run with probability:

$$\mathbb{P}(\text{Detection}) \ge \frac{1}{n \cdot k^{d-1}}$$

### 6.2 Empirical Implication for SQL Concurrency
Because empirical concurrency defects in transactional databases predominantly have bug depth $d \in \{1, 2\}$, the required iterations to guarantee detection with confidence $1 - \epsilon$ scale logarithmically:

$$R \ge \frac{\ln(1/\epsilon)}{\mathbb{P}(\text{Detection})}$$

For $d = 2$, $n = 5$, and $k = 40$:
$$\mathbb{P} \ge \frac{1}{200} \implies R_{95\%} \approx 600 \text{ runs}$$

---

## 7. Causal Delta-Debugging ($ddmin$) Theory and 1-Minimality Proofs

The ChaosSQL Shrinker isolates the minimal subsequence of transactions that reproduces the invariant violation.

### 7.1 Definition of 1-Minimality
Let $C = \langle op_1, op_2, \dots, op_N \rangle$ be an ordered sequence of scheduled operations.
Let the oracle function be $\text{test}: \mathcal{P}(C) \to \{\text{PASS}, \text{FAIL}\}$.
Assume $\text{test}(C) = \text{FAIL}$ and $\text{test}(\emptyset) = \text{PASS}$.

A subsequence $C^* \subseteq C$ is **$1$-Minimal** if:
$$\text{test}(C^*) = \text{FAIL} \quad \land \quad \forall op \in C^*: \text{test}(C^* \setminus \{op\}) = \text{PASS}$$

This guarantees that removing any single operation from $C^*$ causes the failure to vanish.

### 7.2 Causal Closure Constraint Proof
Classic Zeller $ddmin$ partitions $C$ into subsets $c_1, \dots, c_m$. However, SQL operations possess causal dependencies (foreign keys, table schemas, generated entity IDs). Removing an operation $op_i$ that creates an entity referenced by $op_j$ leads to invalid SQL errors (`FOREIGN KEY constraint failed`), which masks genuine isolation anomalies.

**Theorem (Causal Minimality):**
Let $G_{\text{cause}} = (C, E_{\text{cause}})$ be the DAG of causal parameter dependencies, where $(op_i, op_j) \in E_{\text{cause}}$ if $op_j$ consumes an output or entity produced by $op_i$. Define the causal closure of subset $C'$:
$$\text{Closure}(C') = C' \cup \{op_i \in C \mid \exists op_j \in C': op_i \xrightarrow{*} op_j \in G_{\text{cause}}\}$$

Executing $ddmin$ over causally closed subsets guarantees:
1. **Soundness:** Every candidate test schedule satisfies referential integrity constraints.
2. **Monotonic Convergence:** The oracle test function strictly evaluates isolation invariants $\mathcal{I}(\sigma)$, eliminating syntax and constraint false positives.

### 7.3 Complexity Bounds
1. **Best Case (Independent Binary Split):**
   $$\text{Complexity} = O(\log |C|)$$
2. **Worst Case (Linear Inter-operation Dependencies):**
   $$\text{Complexity} = O(|C|^2)$$

---

## 8. Combinatorial State Space and Micro-Jitter

For $W$ concurrent workers executing $M$ steps per transaction across $K$ total transactions, the total number of possible interleavings $\Omega$ is given by the multinomial coefficient:

$$|\Omega| = \frac{(K \cdot M)!}{(M!)^K}$$

For $K = 5$ transactions of $M = 3$ steps each:
$$|\Omega| = \frac{15!}{(3!)^5} = \frac{1,307,674,368,000}{7,776} \approx 168,168,000 \text{ possible interleavings}$$

By injecting stochastic micro-jitter $d_i \sim \mathcal{U}(\delta_{\min}, \delta_{\max})$ with $\delta_{\max} \approx 20\text{ms}$:
$$\mathbb{P}(\text{Critical Interleaving}) \to 1 - e^{-\lambda \cdot \tau_{\text{crit}}}$$

The probability of hitting the race window increases by orders of magnitude within short fuzzing runs ($N \le 50$ operations).
