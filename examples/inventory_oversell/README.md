# Scenario 02: E-Commerce Inventory Oversell (Anomaly A3 / P4)

## Business Context
During a flash sale promotion, an e-commerce platform makes available only **10 units** of a high-demand item (Super GPU). Dozens of concurrent buyers attempt checkout simultaneously.

## Anomaly Breakdown
1. **Workers read available stock** (e.g., `stock = 1`) and determine that the purchase is valid.
2. Multiple concurrent workers insert orders into `orders` and decrement the product inventory.
3. **Chaotic Outcome:** 14 units are sold even though only 10 were in stock!

## Business Invariant
$$\text{Remaining Stock} + \text{Total Sold} == 10 \quad \land \quad \text{Total Sold} \le 10$$

## Mitigation Strategies
* **Atomic Decrement with Guard Predicate:**
  ```sql
  UPDATE products SET stock = stock - 1 WHERE id = 1 AND stock >= 1;
  ```
  If no rows are affected (`RowsAffected == 0`), the purchase is rejected immediately.
* **Pessimistic Row Locking:**
  Acquire an exclusive lock before inspecting inventory:
  ```sql
  SELECT stock FROM products WHERE id = 1 FOR UPDATE;
  ```
* **Database Check Constraint:**
  Enforce integrity at the schema level (`CHECK (stock >= 0)`), ensuring any transaction attempting to decrement below zero triggers an immediate constraint violation and rolls back.
