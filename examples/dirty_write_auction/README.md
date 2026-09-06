# Scenario 05: Auction Bidding Dirty Write (Anomaly G0)

## Business Context
In a real-time online auction platform, multiple participants submit concurrent bids for the same collectible item (`auction_items:1`).
For every submitted bid, the system updates the highest offered bid (`highest_bid`), logs the transaction in `bids_log`, and assigns the provisional winner identifier (`winner_id`).

## Anomaly Breakdown
The **G0 (Dirty Write)** anomaly occurs when two transactions concurrently update the same items or columns without concurrency control or atomic serialization, resulting in interleaved writes and overwriting of intermediate or uncommitted data:
1. **Bidder 1 ($T_1$)** submits a bid of $200:
   - Updates `highest_bid = 200` in `auction_items`.
2. **Bidder 2 ($T_2$)** concurrently submits a bid of $350:
   - Updates `highest_bid = 350`.
   - Inserts the log entry in `bids_log (item_id=1, bidder_id=202, bid_amount=350)`.
   - Updates `winner_id = 202`.
3. **Bidder 1 ($T_1$)** completes its uncoordinated execution sequence:
   - Inserts the log entry in `bids_log (item_id=1, bidder_id=101, bid_amount=200)`.
   - Updates `winner_id = 101`.
4. **Inconsistent Final State:** The item records `highest_bid = 350`, but with `winner_id = 101`!
   Bidder 1 (who bid $200) is declared the provisional auction winner at the $350 valuation submitted by Bidder 2.

### Mathematical Formulation (Adya / Berenson et al.)
In Adya's direct serialization graph (DSG), anomaly $G0$ (Dirty Write) is characterized by a cycle composed purely of direct write-write ($ww$) dependencies:
$$T_1 \xrightarrow{ww} T_2 \xrightarrow{ww} T_1$$

Where:
- $T_1 \xrightarrow{ww} T_2$ on `auction_items:1`: $T_2$ overwrote the `highest_bid` column previously written by $T_1$.
- $T_2 \xrightarrow{ww} T_1$ on `auction_items:1`: $T_1$ overwrote the `winner_id` column previously written by $T_2$.

## Auction Consistency Invariant
The provisional winner recorded in `auction_items` must strictly match the participant who submitted the recorded `highest_bid`:
$$\text{is\_consistent} == 1 \lor \text{is\_consistent} == \text{true}$$

## Formal Mitigation
* **Atomic Update with Guard Predicate:**
  Execute the update of `highest_bid` and `winner_id` within a single atomic SQL statement with a validation guard:
  ```sql
  UPDATE auction_items 
  SET highest_bid = :new_bid, winner_id = :bidder_id 
  WHERE id = 1 AND :new_bid > highest_bid;
  ```
* **SERIALIZABLE Isolation or Pessimistic Locking (`SELECT ... FOR UPDATE`):**
  Lock the item row during bid validation and subsequent writing, preventing concurrent transactions from interleaving partial updates:
  ```sql
  SELECT id, highest_bid, winner_id FROM auction_items WHERE id = 1 FOR UPDATE;
  ```
* **Strict Two-Phase Locking (Strict 2PL / Strict PL-1):**
  Hold exclusive write locks until transaction completion, preventing intermediate or uncommitted state overwrites.
