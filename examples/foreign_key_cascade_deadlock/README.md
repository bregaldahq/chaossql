# Foreign Key Cascade Deadlock & Referential Integrity ($G\text{-DL}$)

## Business Context
In relational e-commerce and order management systems (OMS), parent-child relationships (e.g., `parent_orders` and `child_items`) rely on foreign key constraints (`FOREIGN KEY (order_id) REFERENCES parent_orders(id) ON DELETE CASCADE`). 

During high-concurrency checkout and order lifecycle transitions, two business processes frequently interleave:
1. **Item Append (`add_order_item`)**: A customer adds an additional item to an open order, requiring inserting a new row into `child_items` and updating `parent_orders.total_cents`.
2. **Order Cancellation (`cancel_order_cascade`)**: An administrator or cancellation workflow terminates the order, marking `parent_orders.status = 'CANCELLED'`, deleting `parent_orders`, and cascading deletions to all associated `child_items`.

## Foreign Key Relational Lock Contention & Lock Hierarchy Inversion
When foreign key constraints and cascaded operations are executed without strict lock ordering, relational engines experience **Parent-Child Lock Hierarchy Inversion**:

1. **Transaction $T_1$ (`add_order_item`)**:
   - Step 1: Inserts into `child_items`, acquiring an exclusive write lock on the child record: $X\text{-lock}(\text{child\_items})$. To satisfy foreign key verification, the engine or application may also request or elevate a shared/intent lock on `parent_orders`.
   - Step 2: Updates `parent_orders`, attempting to acquire an exclusive lock: $X\text{-lock}(\text{parent\_orders})$.

2. **Transaction $T_2$ (`cancel_order_cascade`)**:
   - Step 1: Updates and deletes `parent_orders`, acquiring an exclusive lock on the parent record: $X\text{-lock}(\text{parent\_orders})$.
   - Step 2: Cascades delete to `child_items`, attempting to acquire an exclusive lock: $X\text{-lock}(\text{child\_items})$.

Because $T_1$ acquires locks in bottom-up order ($\text{Child} \to \text{Parent}$) while $T_2$ acquires locks in top-down order ($\text{Parent} \to \text{Child}$), the transactions experience classic cross-table lock contention.

## Mathematical Formulation: Deadlock Cycle ($G\text{-DL}$)
In database concurrency theory (Gray & Reuter 1993, Bernstein et al. 1987), lock contention creates a dynamic directed **Wait-For Graph** $WFG = (V, E)$, where vertices $V = \{T_1, T_2\}$ represent active concurrent transactions and directed edges $E = \{(T_i, T_j)\}$ represent blocking dependency where $T_i$ is blocked waiting for a lock held by $T_j$.

When lock hierarchy inversion occurs:
- $T_1$ holds $X(\text{child\_items})$ and waits for $X(\text{parent\_orders})$ held by $T_2$:
  $$T_1 \xrightarrow{\text{waits-for } X(\text{parent\_orders})} T_2$$
- $T_2$ holds $X(\text{parent\_orders})$ and waits for $X(\text{child\_items})$ held by $T_1$:
  $$T_2 \xrightarrow{\text{waits-for } X(\text{child\_items})} T_1$$

The union of dependency edges forms a closed directed cycle:
$$\text{Cycle}(WFG) = T_1 \to T_2 \to T_1 \neq \emptyset \iff G\text{-DL}$$

Under continuous lock contention without deterministic timeout aborts or under non-atomic execution, race conditions manifest as:
1. **Broken Referential Integrity (Orphan Items)**:
   $$\text{Orphans} = \{c \in \text{child\_items} \mid \not\exists p \in \text{parent\_orders} \text{ s.t. } c.\text{order\_id} = p.\text{id}\} \neq \emptyset$$
2. **Order Sum Inconsistency**:
   $$\exists p \in \text{parent\_orders} \text{ where } p.\text{status} = \text{'OPEN'} \land p.\text{total\_cents} \neq \sum_{c \in \text{child\_items}, c.\text{order\_id} = p.\text{id}} c.\text{price\_cents}$$

## Formal Invariants in ChaosSQL
ChaosSQL tests both invariants concurrently under stochastic worker interleavings:
- `referential_integrity`:
  ```sql
  SELECT count(*) AS orphan_items 
  FROM child_items 
  LEFT JOIN parent_orders ON child_items.order_id = parent_orders.id 
  WHERE parent_orders.id IS NULL;
  ```
  Assert: `orphan_items == 0`.
- `order_sum_consistency`:
  ```sql
  SELECT count(*) AS inconsistent_orders
  FROM parent_orders p
  LEFT JOIN (
    SELECT order_id, COALESCE(sum(price_cents), 0) AS items_sum
    FROM child_items
    GROUP BY order_id
  ) c ON p.id = c.order_id
  WHERE p.status = 'OPEN' AND p.total_cents != COALESCE(c.items_sum, 0);
  ```
  Assert: `inconsistent_orders == 0`.

## Remediation & Prevention
1. **Strict Hierarchical Lock Ordering**: Enforce a global partial order $\mathcal{O}$ over tables ($\text{parent\_orders} \prec \text{child\_items}$). Every transaction must acquire locks in increasing order according to $\mathcal{O}$.
2. **Explicit Parent Locking**: In `add_order_item`, acquire an explicit lock on the parent order before touching child items:
   ```sql
   SELECT id FROM parent_orders WHERE id = 1 FOR UPDATE;
   ```
3. **Atomic Serializable Transactions**: Run cascading operations within full `SERIALIZABLE` transactions with automatic retry on serialization/deadlock failure (`SQLSTATE 40P01` / `40001`).
