/**
 * ChaosSQL Official Interactive Documentation Portal & Scenario Explorer
 * Bregalda Digital Identity • v1.1.0
 */

// ==========================================================================
// 1. SCENARIO DATA STORE (9 FLAGSHIP ANOMALIES)
// ==========================================================================
const SCENARIOS = [
  {
    id: "banking_lost_update",
    code: "P4",
    name: "Banking Lost Update",
    category: "Financial Ledger",
    severity: "CRITICAL (Silent Balance Loss)",
    adyaType: "G-single (Cycle: rw + ww)",
    vulnerableLevels: "READ COMMITTED (PostgreSQL, Oracle, SQLite)",
    safeLevels: "SERIALIZABLE (SSI), 2PL Pessimistic Locks",
    context: "In a fintech application, two concurrent withdrawals occur on Alice's account (initial balance: $1,000). Worker 1 reads balance $1,000 and prepares a $100 withdrawal. Concurrently, Worker 2 reads $1,000 and prepares another $100 withdrawal. Worker 1 commits balance $900. Worker 2 also commits balance $900 based on its stale snapshot. Alice successfully withdrew $200, but only $100 was deducted from her balance. The first debit was silently lost.",
    schemaSql: `-- Schema: Bank Accounts & Immutable Audit Ledger
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    holder TEXT NOT NULL,
    balance NUMERIC NOT NULL CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS ledger (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    amount NUMERIC NOT NULL,
    type TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initial State (Seed)
INSERT INTO accounts (id, holder, balance) VALUES (1, 'Alice', 1000.00);`,
    chaosYaml: `version: "1.0"
name: "banking_lost_update"
description: "Detects Lost Update (P4) anomaly on concurrent account withdrawals"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 4
  iterations: 20
  seed: 42
  jitter_ms: [1, 10]

invariants:
  - name: "ledger_balance_consistency"
    query: >
      SELECT 
        (SELECT balance FROM accounts WHERE id = 1) AS actual_balance,
        (1000 - COALESCE(SUM(amount), 0)) AS expected_balance
      FROM ledger WHERE account_id = 1;
    assert: "actual_balance == expected_balance and actual_balance >= 0"

operations:
  - name: "withdraw_vulnerable"
    weight: 1.0
    params:
      amount: "int(50, 150)"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
        capture: "current_bal"
      - sql: "UPDATE accounts SET balance = {current_bal - amount} WHERE id = 1;"
      - sql: "INSERT INTO ledger (account_id, amount, type) VALUES (1, {amount}, 'DEBIT');"`,
    invariantQuery: `SELECT 
  (SELECT balance FROM accounts WHERE id = 1) AS actual_balance,
  (1000 - COALESCE(SUM(amount), 0)) AS expected_balance
FROM ledger WHERE account_id = 1;`,
    invariantAssertion: `actual_balance == expected_balance and actual_balance >= 0`,
    invariantExplanation: "Formal inductive predicate: The balance in the accounts table must strictly match initial wealth minus total debited ledger rows. In a race condition, ledger records 2 debits ($200 total) while account balance reflects only 1 debit ($900), failing the assertion.",
    ddminStats: {
      initialOps: 20,
      shrunkOps: 2,
      reductionPct: "90.0%",
      shrinkTimeMs: 74,
      iterations: 4,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T1 (Worker 1)", "T2 (Worker 2)"],
      edges: [
        { from: "T1", to: "T2", label: "rw (read x₀ → write x₁)", type: "anti-dep" },
        { from: "T2", to: "T1", label: "ww (overwrite x₁)", type: "write-dep" }
      ],
      cycle: "T1 ➔ T2 ➔ T1 (G-single Cycle Detected)"
    }
  },
  {
    id: "inventory_oversell",
    code: "A3",
    name: "Inventory Oversell",
    category: "E-Commerce",
    severity: "HIGH (Phantom / Negative Stock)",
    adyaType: "G-phantom (Predicate Conflict on Stock)",
    vulnerableLevels: "READ COMMITTED, REPEATABLE READ (Non-locking)",
    safeLevels: "SERIALIZABLE, Guarded Decrement (WHERE stock >= qty)",
    context: "A flash sale promotion offers 10 units of Super GPU. Six concurrent checkout workers check available stock (returns > 0) and proceed to insert order records and decrement stock. Because reading available stock and decrementing are decoupled without atomic predicate locking, 14 orders are accepted for only 10 available physical units, resulting in an inventory deficit.",
    schemaSql: `-- Schema: Product Catalog & Order Registry
CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    stock INTEGER NOT NULL CHECK (stock >= 0)
);

CREATE TABLE IF NOT EXISTS orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL REFERENCES products(id),
    user_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initial State (Seed)
INSERT INTO products (id, name, stock) VALUES (1, 'Super GPU', 10);`,
    chaosYaml: `version: "1.0"
name: "inventory_oversell"
description: "Detects Overselling race condition on flash sale inventory"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 6
  iterations: 30
  seed: 101
  jitter_ms: [1, 8]

invariants:
  - name: "stock_never_negative_and_consistent"
    query: >
      SELECT 
        (SELECT stock FROM products WHERE id = 1) AS current_stock,
        (SELECT COALESCE(SUM(quantity), 0) FROM orders WHERE product_id = 1) AS total_sold;
    assert: "current_stock + total_sold == 10 and current_stock >= 0 and total_sold <= 10"

operations:
  - name: "purchase_item"
    weight: 1.0
    params:
      user_id: "int(100, 999)"
    steps:
      - sql: "SELECT stock FROM products WHERE id = 1;"
        capture: "available_stock"
      - sql: "INSERT INTO orders (product_id, user_id, quantity) VALUES (1, {user_id}, 1);"
      - sql: "UPDATE products SET stock = {available_stock - 1} WHERE id = 1;"`,
    invariantQuery: `SELECT 
  (SELECT stock FROM products WHERE id = 1) AS current_stock,
  (SELECT COALESCE(SUM(quantity), 0) FROM orders WHERE product_id = 1) AS total_sold;`,
    invariantAssertion: `current_stock + total_sold == 10 and current_stock >= 0 and total_sold <= 10`,
    invariantExplanation: "Conservation of Inventory: Initial stock (10) must equal current remaining stock plus total orders created. When race interleaves, orders exceed initial stock capacity.",
    ddminStats: {
      initialOps: 30,
      shrunkOps: 2,
      reductionPct: "93.3%",
      shrinkTimeMs: 88,
      iterations: 5,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T_Buyer1", "T_Buyer2"],
      edges: [
        { from: "T_Buyer1", to: "T_Buyer2", label: "rw (read stock → write order)", type: "anti-dep" },
        { from: "T_Buyer2", to: "T_Buyer1", label: "ww (overwrite stock)", type: "write-dep" }
      ],
      cycle: "Predicate Depletion Anti-Dependency Conflict"
    }
  },
  {
    id: "hospital_write_skew",
    code: "A5B",
    name: "Hospital Write Skew",
    category: "Healthcare Governance",
    severity: "CRITICAL (Zero Staff Violation)",
    adyaType: "G2-item (Dangerous Structure rw + rw)",
    vulnerableLevels: "SNAPSHOT ISOLATION / REPEATABLE READ (PostgreSQL, MySQL)",
    safeLevels: "SERIALIZABLE SNAPSHOT ISOLATION (SSI), SELECT FOR UPDATE",
    context: "Hospital safety bylaws require at least one doctor on active duty (is_on_call = 1). Dr. Alice and Dr. Bob are on duty. Dr. Alice requests leave: her transaction reads on-call count (2 >= 2) and sets her status to 0. Concurrently, Dr. Bob requests leave: his transaction reads on-call count from his snapshot (2 >= 2) and sets his status to 0. Since their write sets are disjoint (W₁ ∩ W₂ = ∅), both transactions commit under Snapshot Isolation, leaving 0 doctors on duty!",
    schemaSql: `-- Schema: Hospital Medical Staff & Audit
CREATE TABLE IF NOT EXISTS doctors (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    is_on_call INTEGER NOT NULL CHECK (is_on_call IN (0, 1))
);

CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    doctor_id INTEGER NOT NULL REFERENCES doctors(id),
    action TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initial State: 2 Doctors on Call
INSERT INTO doctors (id, name, is_on_call) VALUES (1, 'Dr. Alice', 1), (2, 'Dr. Bob', 1);`,
    chaosYaml: `version: "1.0"
name: "hospital_write_skew"
description: "Detects Write Skew (A5B) on on-call doctors schedule"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 2
  iterations: 10
  seed: 1
  jitter_ms: [2, 12]

invariants:
  - name: "at_least_one_doctor_on_call"
    query: "SELECT COALESCE(SUM(is_on_call), 0) AS active_doctors FROM doctors;"
    assert: "active_doctors >= 1"

operations:
  - name: "go_off_call_doctor_1"
    weight: 1.0
    steps:
      - sql: "SELECT SUM(is_on_call) FROM doctors;"
        capture: "count_on_call"
      - sql: "UPDATE doctors SET is_on_call = 0 WHERE id = 1 AND {count_on_call} >= 2;"
      - sql: "INSERT INTO audit_log (doctor_id, action) VALUES (1, 'LEAVE_CALL');"
  - name: "go_off_call_doctor_2"
    weight: 1.0
    steps:
      - sql: "SELECT SUM(is_on_call) FROM doctors;"
        capture: "count_on_call"
      - sql: "UPDATE doctors SET is_on_call = 0 WHERE id = 2 AND {count_on_call} >= 2;"
      - sql: "INSERT INTO audit_log (doctor_id, action) VALUES (2, 'LEAVE_CALL');"`,
    invariantQuery: `SELECT COALESCE(SUM(is_on_call), 0) AS active_doctors FROM doctors;`,
    invariantAssertion: `active_doctors >= 1`,
    invariantExplanation: "Fekete Dangerous Structure Theorem (TODS 2005): Non-serializable snapshot execution requires two consecutive anti-dependency (rw) edges. Alice reads Bob, writes Alice; Bob reads Alice, writes Bob. active_doctors drops to 0.",
    ddminStats: {
      initialOps: 10,
      shrunkOps: 2,
      reductionPct: "80.0%",
      shrinkTimeMs: 42,
      iterations: 3,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T_Alice", "T_Bob"],
      edges: [
        { from: "T_Alice", to: "T_Bob", label: "rw (reads Bob.is_on_call → writes Alice)", type: "anti-dep" },
        { from: "T_Bob", to: "T_Alice", label: "rw (reads Alice.is_on_call → writes Bob)", type: "anti-dep" }
      ],
      cycle: "T_Alice ⇄ T_Bob (Dangerous Structure G2-item Cycle)"
    }
  },
  {
    id: "read_skew_financial_audit",
    code: "A5A",
    name: "Financial Read Skew",
    category: "Banking & Accounting",
    severity: "HIGH (Audit Inconsistency)",
    adyaType: "G-skew (Cycle rw + wr)",
    vulnerableLevels: "READ COMMITTED (PostgreSQL, SQLite default)",
    safeLevels: "REPEATABLE READ, SNAPSHOT ISOLATION, SERIALIZABLE",
    context: "A customer holds $500 in Checking and $500 in Savings (Total: $1,000). A financial audit job (T1) reads Checking ($500). Concurrently, an automated transfer (T2) moves $100 from Checking to Savings ($400 / $600) and commits. T1 then reads Savings ($600). T1 calculates total wealth as $500 + $600 = $1,100, manufacturing $100 in phantom money in financial regulatory reports.",
    schemaSql: `-- Schema: Multi-Account Balances & Transfers
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    balance NUMERIC NOT NULL
);

CREATE TABLE IF NOT EXISTS transfers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_id INTEGER NOT NULL REFERENCES accounts(id),
    to_id INTEGER NOT NULL REFERENCES accounts(id),
    amount NUMERIC NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initial State: $500 in Checking, $500 in Savings
INSERT INTO accounts (id, name, balance) VALUES (1, 'Checking', 500.00), (2, 'Savings', 500.00);`,
    chaosYaml: `version: "1.0"
name: "read_skew_financial_audit"
description: "Detects Read Skew (A5A) anomaly during concurrent financial account transfers and audit"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 4
  iterations: 20
  seed: 42
  jitter_ms: [2, 10]

invariants:
  - name: "total_wealth_preservation"
    query: "SELECT COALESCE(SUM(balance), 0) AS total_balance FROM accounts;"
    assert: "total_balance == 1000"

operations:
  - name: "transfer_checking_to_savings"
    weight: 1.0
    params:
      amount: "int(10, 50)"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
        capture: "chk"
      - sql: "UPDATE accounts SET balance = {chk - amount} WHERE id = 1;"
      - sql: "SELECT balance FROM accounts WHERE id = 2;"
        capture: "sav"
      - sql: "UPDATE accounts SET balance = {sav + amount} WHERE id = 2;"
      - sql: "INSERT INTO transfers (from_id, to_id, amount) VALUES (1, 2, {amount});"
  - name: "read_total_wealth"
    weight: 1.0
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
        capture: "chk_balance"
      - sql: "SELECT balance FROM accounts WHERE id = 2;"
        capture: "sav_balance"`,
    invariantQuery: `SELECT COALESCE(SUM(balance), 0) AS total_balance FROM accounts;`,
    invariantAssertion: `total_balance == 1000`,
    invariantExplanation: "Total assets across all customer sub-accounts must equal exactly $1,000. Under Read Committed, non-repeatable row reads observe split state before and after concurrent commit.",
    ddminStats: {
      initialOps: 20,
      shrunkOps: 2,
      reductionPct: "90.0%",
      shrinkTimeMs: 65,
      iterations: 4,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T1 (Audit Job)", "T2 (Transfer)"],
      edges: [
        { from: "T1", to: "T2", label: "rw (read Checking pre-transfer)", type: "anti-dep" },
        { from: "T2", to: "T1", label: "wr (write Savings post-transfer)", type: "read-dep" }
      ],
      cycle: "T1 ➔ T2 ➔ T1 (G-skew Read Skew Cycle)"
    }
  },
  {
    id: "dirty_write_auction",
    code: "G0",
    name: "Auction Dirty Write",
    category: "Online Auction Platform",
    severity: "CRITICAL (Outcome Corruption)",
    adyaType: "G0 / G-write (Cycle ww + ww)",
    vulnerableLevels: "UNCOMMITTED / NON-LOCKING CONCURRENCY",
    safeLevels: "READ COMMITTED (Row Locks), 2PL",
    context: "In an online collector auction for item 1, Bidder 101 bids $200 and Bidder 202 bids $400. Without atomic transaction boundaries or row locks, Bidder 101 writes winner_id = 101, Bidder 202 writes highest_bid = 400, and Bidder 101 writes highest_bid = 200. The database reaches a corrupted hybrid state where Bidder 202 is declared the winner for only $200 instead of $400.",
    schemaSql: `-- Schema: Live Auction Items & Bid Stream
CREATE TABLE IF NOT EXISTS auction_items (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    highest_bid NUMERIC NOT NULL,
    winner_id INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS bids_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id INTEGER NOT NULL REFERENCES auction_items(id),
    bidder_id INTEGER NOT NULL,
    bid_amount NUMERIC NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initial State: Opening bid $100
INSERT INTO auction_items (id, title, highest_bid, winner_id) VALUES (1, 'Rare Vintage Watch', 100.00, 0);`,
    chaosYaml: `version: "1.0"
name: "dirty_write_auction"
description: "Detects Dirty Write (G0) anomaly on uncoordinated concurrent auction bids"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 4
  iterations: 20
  seed: 42
  jitter_ms: [2, 10]

invariants:
  - name: "winner_matches_highest_bid"
    query: >
      SELECT CASE 
        WHEN a.winner_id = 0 AND a.highest_bid = 100 THEN 1
        WHEN EXISTS (
          SELECT 1 FROM bids_log b 
          WHERE b.item_id = a.id 
            AND b.bidder_id = a.winner_id 
            AND b.bid_amount = a.highest_bid
        ) THEN 1
        ELSE 0 
      END AS is_consistent
      FROM auction_items a
      WHERE a.id = 1;
    assert: "is_consistent == 1"

operations:
  - name: "bid_bidder_1"
    weight: 1.0
    params:
      bid_amount: "int(150, 250)"
    steps:
      - sql: "UPDATE auction_items SET winner_id = 101 WHERE id = 1;"
      - sql: "INSERT INTO bids_log (item_id, bidder_id, bid_amount) VALUES (1, 101, {bid_amount});"
      - sql: "UPDATE auction_items SET highest_bid = {bid_amount} WHERE id = 1;"
  - name: "bid_bidder_2"
    weight: 1.0
    params:
      bid_amount: "int(300, 450)"
    steps:
      - sql: "UPDATE auction_items SET highest_bid = {bid_amount} WHERE id = 1;"
      - sql: "INSERT INTO bids_log (item_id, bidder_id, bid_amount) VALUES (1, 202, {bid_amount});"
      - sql: "UPDATE auction_items SET winner_id = 202 WHERE id = 1;"`,
    invariantQuery: `SELECT CASE 
  WHEN a.winner_id = 0 AND a.highest_bid = 100 THEN 1
  WHEN EXISTS (
    SELECT 1 FROM bids_log b 
    WHERE b.item_id = a.id AND b.bidder_id = a.winner_id AND b.bid_amount = a.highest_bid
  ) THEN 1
  ELSE 0 
END AS is_consistent
FROM auction_items a WHERE a.id = 1;`,
    invariantAssertion: `is_consistent == 1`,
    invariantExplanation: "Integrity Invariant: The recorded winner_id must strictly match the bidder who placed the recorded highest_bid in bids_log. When dirty writes interleave, winner_id and highest_bid belong to two different bids.",
    ddminStats: {
      initialOps: 20,
      shrunkOps: 2,
      reductionPct: "90.0%",
      shrinkTimeMs: 71,
      iterations: 4,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T_Bidder101", "T_Bidder202"],
      edges: [
        { from: "T_Bidder101", to: "T_Bidder202", label: "ww (write winner_id → overwrite highest_bid)", type: "write-dep" },
        { from: "T_Bidder202", to: "T_Bidder101", label: "ww (write highest_bid → overwrite winner_id)", type: "write-dep" }
      ],
      cycle: "T_Bidder101 ⇄ T_Bidder202 (G0 Pure Write-Write Cycle)"
    }
  },
  {
    id: "circular_info_crypto_arbitrage",
    code: "G1c",
    name: "Crypto Arbitrage Circular Flow",
    category: "Decentralized Finance (DeFi)",
    severity: "HIGH (Loss of Linearizability)",
    adyaType: "G1c (Circular Information Flow wr + wr)",
    vulnerableLevels: "READ COMMITTED (PostgreSQL, MySQL Read Committed)",
    safeLevels: "REPEATABLE READ, SNAPSHOT ISOLATION, STRICT SERIALIZABLE",
    context: "In decentralized finance (DeFi), automated market maker (AMM) arbitrage bots monitor relative pool prices between Uniswap and Sushiswap. Bot 1 updates Uniswap price to 3100 and reads Sushiswap price. Concurrently, Bot 2 updates Sushiswap price to 3200 and reads Uniswap price. Both execute cross-arbitrage trades based on mutual stale snapshots, violating strict external causal consistency.",
    schemaSql: `-- Schema: AMM DEX Liquidity Pools & Trade Orders
CREATE TABLE dex_pools (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    price_eth INT NOT NULL
);

CREATE TABLE trade_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dex_id INT NOT NULL,
    trader TEXT NOT NULL,
    amount INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initial State: Parity at 3000 ETH
INSERT INTO dex_pools (id, name, price_eth) VALUES (1, 'Uniswap_Pool', 3000);
INSERT INTO dex_pools (id, name, price_eth) VALUES (2, 'Sushiswap_Pool', 3000);`,
    chaosYaml: `version: "1.0"
name: "circular_info_crypto_arbitrage"
description: "Detects Circular Information Flow (G1c) during concurrent cross-DEX liquidity updates"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 2
  iterations: 20
  seed: 42
  jitter_ms: [0, 5]

invariants:
  - name: "arbitrage_price_consistency"
    query: "SELECT (SELECT price_eth FROM dex_pools WHERE id = 1) AS p1, (SELECT price_eth FROM dex_pools WHERE id = 2) AS p2;"
    assert: "p1 == 3100 or p2 == 3200"

operations:
  - name: "arb_bot_uniswap"
    weight: 1.0
    params:
      trade_vol: "$random_int(10, 100)"
    steps:
      - sql: "UPDATE dex_pools SET price_eth = 3100 WHERE id = 1;"
      - sql: "SELECT price_eth FROM dex_pools WHERE id = 2;"
        capture: "sushi_price"
      - sql: "INSERT INTO trade_orders (dex_id, trader, amount) VALUES (1, 'Bot_Uni', {sushi_price});"
  - name: "arb_bot_sushiswap"
    weight: 1.0
    params:
      trade_vol: "$random_int(10, 100)"
    steps:
      - sql: "UPDATE dex_pools SET price_eth = 3200 WHERE id = 2;"
      - sql: "SELECT price_eth FROM dex_pools WHERE id = 1;"
        capture: "uni_price"
      - sql: "INSERT INTO trade_orders (dex_id, trader, amount) VALUES (2, 'Bot_Sushi', {uni_price});"`,
    invariantQuery: `SELECT (SELECT price_eth FROM dex_pools WHERE id = 1) AS p1, (SELECT price_eth FROM dex_pools WHERE id = 2) AS p2;`,
    invariantAssertion: `p1 == 3100 or p2 == 3200`,
    invariantExplanation: "Linearizability Invariant: In any sequential execution, at least one price update becomes globally visible before the second transaction reads it. In weak isolation, both read stale initial states.",
    ddminStats: {
      initialOps: 20,
      shrunkOps: 2,
      reductionPct: "90.0%",
      shrinkTimeMs: 58,
      iterations: 4,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T_BotUniswap", "T_BotSushiswap"],
      edges: [
        { from: "T_BotUniswap", to: "T_BotSushiswap", label: "wr (write UniPrice → read UniPrice)", type: "read-dep" },
        { from: "T_BotSushiswap", to: "T_BotUniswap", label: "wr (write SushiPrice → read SushiPrice)", type: "read-dep" }
      ],
      cycle: "T_BotUni ⇄ T_BotSushi (G1c Circular Information Flow)"
    }
  },
  {
    id: "dirty_read_flash_crash",
    code: "G1a",
    name: "Flash Crash Dirty Read",
    category: "DeFi Lending & Liquidation",
    severity: "CRITICAL (Wrongful Liquidation)",
    adyaType: "G1a (Aborted / Dirty Read)",
    vulnerableLevels: "READ UNCOMMITTED",
    safeLevels: "READ COMMITTED, SNAPSHOT ISOLATION, SERIALIZABLE",
    context: "In a crypto collateralized lending protocol (MakerDAO/Aave), an oracle feed writes an erroneous flash crash price ($1,500 ETH) and immediately aborts/rolls back. A liquidation bot reading dirty uncommitted rows observes $1,500 ETH and liquidates a healthy whale collateral position (10 ETH collateral vs $20k debt), causing massive wrongful financial loss.",
    schemaSql: `-- Schema: Collateral Vault Positions & Price Oracles
CREATE TABLE vault_positions (
    id INT PRIMARY KEY,
    owner TEXT NOT NULL,
    collateral_eth INT NOT NULL,
    debt_usd INT NOT NULL,
    is_liquidated INT DEFAULT 0
);

CREATE TABLE oracle_prices (
    id INT PRIMARY KEY,
    asset TEXT NOT NULL,
    price_usd INT NOT NULL
);

-- Initial State: 10 ETH @ $3,000 = $30k collateral (Safe vs $20k debt)
INSERT INTO vault_positions VALUES (1, 'Alice_Whale', 10, 20000, 0);
INSERT INTO oracle_prices VALUES (1, 'ETH', 3000);`,
    chaosYaml: `version: "1.0"
name: "dirty_read_flash_crash"
description: "Detects Aborted / Dirty Read (G1a) triggering invalid collateral liquidation"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 2
  iterations: 20
  seed: 42
  jitter_ms: [0, 5]
  faults:
    abort_probability: 0.3

invariants:
  - name: "no_invalid_liquidation_on_reverted_price"
    query: "SELECT is_liquidated FROM vault_positions WHERE id = 1;"
    assert: "is_liquidated == 0"

operations:
  - name: "flash_crash_update"
    weight: 1.0
    steps:
      - sql: "UPDATE oracle_prices SET price_usd = 1500 WHERE id = 1;"
      - sql: "UPDATE oracle_prices SET price_usd = 3000 WHERE id = 1;"
  - name: "liquidation_bot"
    weight: 1.0
    steps:
      - sql: "SELECT price_usd FROM oracle_prices WHERE id = 1;"
        capture: "eth_price"
      - sql: "UPDATE vault_positions SET is_liquidated = 1 WHERE id = 1 AND ({eth_price} * 10) < 24000;"`,
    invariantQuery: `SELECT is_liquidated FROM vault_positions WHERE id = 1;`,
    invariantAssertion: `is_liquidated == 0`,
    invariantExplanation: "Safety Invariant: A vault position with valid market collateral must never be liquidated on an aborted or reverted temporary price drop.",
    ddminStats: {
      initialOps: 20,
      shrunkOps: 2,
      reductionPct: "90.0%",
      shrinkTimeMs: 53,
      iterations: 3,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T_OracleCrash (Aborted)", "T_LiquidationBot"],
      edges: [
        { from: "T_OracleCrash", to: "T_LiquidationBot", label: "wr (dirty read uncommitted ETH=$1500)", type: "read-dep" }
      ],
      cycle: "G1a Aborted Dirty Read Violation"
    }
  },
  {
    id: "ticket_booking_anti_dependency",
    code: "G2",
    name: "Ticket Booking Anti-Dependency",
    category: "Aviation & Event Ticketing",
    severity: "HIGH (Multi-Party Overbooking)",
    adyaType: "G2 (Tri-Partite Anti-Dependency Cycle rw + rw + rw)",
    vulnerableLevels: "SNAPSHOT ISOLATION / REPEATABLE READ",
    safeLevels: "SERIALIZABLE SNAPSHOT ISOLATION (SSI), Predicate Locks",
    context: "In airline and concert ticket reservation platforms, 3 concurrent users attempt to book adjacent seats. User 1 checks seat 2 is empty, then books seat 1. User 2 checks seat 3 is empty, then books seat 2. User 3 checks seat 1 is empty, then books seat 3. All 3 commit under weak snapshot isolation, violating the multi-transaction booking isolation constraint.",
    schemaSql: `-- Schema: Reserved Seats & Booking Audits
CREATE TABLE seats (
    id INT PRIMARY KEY,
    row_num INT NOT NULL,
    seat_num INT NOT NULL,
    reserved_by TEXT DEFAULT NULL
);

CREATE TABLE booking_audit (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    seat_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Initial State: 3 empty adjacent seats
INSERT INTO seats VALUES (1, 1, 1, NULL);
INSERT INTO seats VALUES (2, 1, 2, NULL);
INSERT INTO seats VALUES (3, 1, 3, NULL);`,
    chaosYaml: `version: "1.0"
name: "ticket_booking_anti_dependency"
description: "Detects 3-Transaction Anti-Dependency Cycles (G2) in adjacent seat reservations"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 3
  iterations: 15
  seed: 42
  jitter_ms: [0, 5]

invariants:
  - name: "adjacent_seat_isolation"
    query: "SELECT count(*) AS total_booked FROM seats WHERE reserved_by IS NOT NULL;"
    assert: "total_booked <= 2"

operations:
  - name: "book_seat_1_check_2"
    weight: 1.0
    steps:
      - sql: "SELECT reserved_by FROM seats WHERE id = 2;"
        capture: "s2"
      - sql: "UPDATE seats SET reserved_by = 'User_1' WHERE id = 1;"
  - name: "book_seat_2_check_3"
    weight: 1.0
    steps:
      - sql: "SELECT reserved_by FROM seats WHERE id = 3;"
        capture: "s3"
      - sql: "UPDATE seats SET reserved_by = 'User_2' WHERE id = 2;"
  - name: "book_seat_3_check_1"
    weight: 1.0
    steps:
      - sql: "SELECT reserved_by FROM seats WHERE id = 1;"
        capture: "s1"
      - sql: "UPDATE seats SET reserved_by = 'User_3' WHERE id = 3;"`,
    invariantQuery: `SELECT count(*) AS total_booked FROM seats WHERE reserved_by IS NOT NULL;`,
    invariantAssertion: `total_booked <= 2`,
    invariantExplanation: "Tri-Partite Anti-Dependency Cycle: T1 anti-depends on T2, T2 anti-depends on T3, and T3 anti-depends on T1. Each user sees the adjacent seat empty and books, exceeding the safety budget.",
    ddminStats: {
      initialOps: 15,
      shrunkOps: 3,
      reductionPct: "80.0%",
      shrinkTimeMs: 62,
      iterations: 4,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T1 (User 1)", "T2 (User 2)", "T3 (User 3)"],
      edges: [
        { from: "T1", to: "T2", label: "rw (reads Seat2 → writes Seat1)", type: "anti-dep" },
        { from: "T2", to: "T3", label: "rw (reads Seat3 → writes Seat2)", type: "anti-dep" },
        { from: "T3", to: "T1", label: "rw (reads Seat1 → writes Seat3)", type: "anti-dep" }
      ],
      cycle: "T1 ➔ T2 ➔ T3 ➔ T1 (3-Party G2 Anti-Dependency Cycle)"
    }
  },
  {
    id: "deadlock_cycle",
    code: "G-DL",
    name: "Deadlock Cycle & Recovery",
    category: "Lock Contention & Diagnostics",
    severity: "MEDIUM (System Stall / Abort)",
    adyaType: "G-DL (Waits-For Lock Dependency Cycle)",
    vulnerableLevels: "ALL CONCURRENT RDBMS ENGINES",
    safeLevels: "AUTOMATIC DEADLOCK DETECTOR + EXPONENTIAL RETRY",
    context: "In financial ledger systems, concurrent bilateral fund transfers between Account 1 and Account 2 cause classic circular wait deadlocks: T1 locks Account 1 and requests Account 2; T2 locks Account 2 and requests Account 1. ChaosSQL validates that the lock manager detects the deadlock cycle, terminates the victim transaction cleanly, and maintains total wealth preservation invariant ($2,000).",
    schemaSql: `-- Schema: Bank Ledgers with Lock Contention
CREATE TABLE bank_ledgers (
    id INT PRIMARY KEY,
    account_name TEXT NOT NULL,
    balance INT NOT NULL
);

-- Initial State: $1,000 in Checking A, $1,000 in Checking B
INSERT INTO bank_ledgers (id, account_name, balance) VALUES (1, 'Checking_A', 1000);
INSERT INTO bank_ledgers (id, account_name, balance) VALUES (2, 'Checking_B', 1000);`,
    chaosYaml: `version: "1.0"
name: "deadlock_cycle"
description: "Detects bidirectional concurrent lock contention and deadlock cycle recovery"

database:
  driver: "sqlite"
  schema: "schema.sql"
  seed: "seed.sql"

engine:
  workers: 2
  iterations: 20
  seed: 42
  jitter_ms: [0, 5]

invariants:
  - name: "total_ledger_preservation"
    query: "SELECT sum(balance) AS total_wealth FROM bank_ledgers;"
    assert: "total_wealth == 2000"

operations:
  - name: "lock_1_then_2"
    weight: 1.0
    steps:
      - sql: "UPDATE bank_ledgers SET balance = balance - 50 WHERE id = 1;"
      - sql: "UPDATE bank_ledgers SET balance = balance + 50 WHERE id = 2;"
  - name: "lock_2_then_1"
    weight: 1.0
    steps:
      - sql: "UPDATE bank_ledgers SET balance = balance - 50 WHERE id = 2;"
      - sql: "UPDATE bank_ledgers SET balance = balance + 50 WHERE id = 1;"`,
    invariantQuery: `SELECT sum(balance) AS total_wealth FROM bank_ledgers;`,
    invariantAssertion: `total_wealth == 2000`,
    invariantExplanation: "Deadlock Resilience: Even when SQLITE_BUSY or PG deadlock aborts one branch, the overall ledger total wealth ($2,000) must remain strictly intact.",
    ddminStats: {
      initialOps: 20,
      shrunkOps: 2,
      reductionPct: "90.0%",
      shrinkTimeMs: 49,
      iterations: 3,
      status: "1-MINIMAL VERIFIED"
    },
    graph: {
      nodes: ["T1 (Lock 1→2)", "T2 (Lock 2→1)"],
      edges: [
        { from: "T1", to: "T2", label: "waits-for (holds Lock 1, waits Lock 2)", type: "deadlock" },
        { from: "T2", to: "T1", label: "waits-for (holds Lock 2, waits Lock 1)", type: "deadlock" }
      ],
      cycle: "T1 ⇄ T2 (Deadlock Cycle Waits-For)"
    }
  }
];

// ==========================================================================
// 2. SDK PLAYGROUND CODE SAMPLES
// ==========================================================================
const SDK_SAMPLES = {
  banking: `package myapp_test

import (
    "context"
    "testing"
    "github.com/bregaldahq/chaossql/pkg/chaostest"
)

func TestAccountTransfer_LostUpdateIsolation(t *testing.T) {
    ctx := context.Background()

    schema := \`
    CREATE TABLE accounts (
        id INT PRIMARY KEY,
        holder TEXT NOT NULL,
        balance INT NOT NULL
    );
    CREATE TABLE ledger (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        account_id INT NOT NULL,
        amount INT NOT NULL
    );\`

    seed := \`
    INSERT INTO accounts VALUES (1, 'Alice', 1000);\`

    chaostest.New(t).
        WithSchema(schema).
        WithSeed(seed).
        WithJitter(1, 10).
        WithInvariant(
            "ledger_balance_consistency",
            "SELECT (SELECT balance FROM accounts WHERE id = 1) AS actual, (1000 - COALESCE(SUM(amount), 0)) AS expected FROM ledger WHERE account_id = 1;",
            "actual == expected and actual >= 0",
        ).
        AddOperation("withdraw_debit",
            "SELECT balance FROM accounts WHERE id = 1 -> cur",
            "UPDATE accounts SET balance = {cur - 100} WHERE id = 1",
            "INSERT INTO ledger (account_id, amount) VALUES (1, 100)",
        ).
        AssertNoAnomalies(ctx, 4, 25, 42) // workers=4, iterations=25, seed=42
}`,
  hospital: `package myapp_test

import (
    "context"
    "testing"
    "github.com/bregaldahq/chaossql/pkg/chaostest"
)

func TestHospitalDuty_WriteSkewPrevention(t *testing.T) {
    ctx := context.Background()

    schema := \`
    CREATE TABLE doctors (
        id INT PRIMARY KEY,
        name TEXT NOT NULL,
        is_on_call INT NOT NULL
    );\`

    seed := \`
    INSERT INTO doctors VALUES (1, 'Dr. Alice', 1), (2, 'Dr. Bob', 1);\`

    chaostest.New(t).
        WithSchema(schema).
        WithSeed(seed).
        WithInvariant(
            "at_least_one_doctor_on_call",
            "SELECT COALESCE(SUM(is_on_call), 0) AS active FROM doctors;",
            "active >= 1",
        ).
        AddOperation("leave_call_doc1",
            "SELECT SUM(is_on_call) FROM doctors -> active_docs",
            "UPDATE doctors SET is_on_call = 0 WHERE id = 1 AND {active_docs} >= 2",
        ).
        AddOperation("leave_call_doc2",
            "SELECT SUM(is_on_call) FROM doctors -> active_docs",
            "UPDATE doctors SET is_on_call = 0 WHERE id = 2 AND {active_docs} >= 2",
        ).
        AssertNoAnomalies(ctx, 2, 20, 1)
}`,
  inventory: `package myapp_test

import (
    "context"
    "testing"
    "github.com/bregaldahq/chaossql/pkg/chaostest"
)

func TestFlashSale_InventoryOversell(t *testing.T) {
    ctx := context.Background()

    schema := \`
    CREATE TABLE inventory (
        sku TEXT PRIMARY KEY,
        stock INT NOT NULL CHECK (stock >= 0)
    );
    CREATE TABLE orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sku TEXT NOT NULL,
        qty INT NOT NULL
    );\`

    seed := \`
    INSERT INTO inventory VALUES ('GPU-RTX-5090', 10);\`

    chaostest.New(t).
        WithSchema(schema).
        WithSeed(seed).
        WithInvariant(
            "stock_conservation",
            "SELECT (SELECT stock FROM inventory WHERE sku = 'GPU-RTX-5090') AS cur, (SELECT COALESCE(SUM(qty), 0) FROM orders) AS sold;",
            "cur + sold == 10 and cur >= 0 and sold <= 10",
        ).
        AddOperation("purchase_gpu",
            "SELECT stock FROM inventory WHERE sku = 'GPU-RTX-5090' -> s",
            "INSERT INTO orders (sku, qty) VALUES ('GPU-RTX-5090', 1)",
            "UPDATE inventory SET stock = {s - 1} WHERE sku = 'GPU-RTX-5090'",
        ).
        AssertNoAnomalies(ctx, 6, 30, 101)
}`,
  differential: `package myapp_test

import (
    "context"
    "testing"
    "github.com/bregaldahq/chaossql/pkg/chaostest"
    "github.com/bregaldahq/chaossql/internal/drivers"
)

func TestDifferential_SQLiteVsPostgres(t *testing.T) {
    ctx := context.Background()

    sqliteDriver, _ := drivers.GetDriver("sqlite", "file::memory:?cache=shared")
    postgresDriver, _ := drivers.GetDriver("postgres", "postgres://test:test@localhost:5432/test?sslmode=disable")

    tester := chaostest.New(t).
        WithSpecFile("examples/banking_lost_update/chaos.yaml")

    // Compare divergence between weak SQLite and SSI Postgres
    diffResult, err := tester.CompareDivergence(ctx, sqliteDriver, postgresDriver)
    if err != nil {
        t.Fatalf("Differential execution failed: %v", err)
    }

    t.Logf("Differential Semantic Divergence: %v", diffResult.HasDiverged)
}`
};

// ==========================================================================
// 3. CLI CODE SAMPLES
// ==========================================================================
const CLI_SAMPLES = {
  run: `# Run chaos stress test with delta-debugging active shrinking
chaossql run examples/banking_lost_update/chaos.yaml \
  --workers 4 \
  --iterations 25 \
  --seed 42 \
  --shrink \
  --export-repro ./repro_test.go \
  --export-html ./report.html`,
  init: `# Scaffold a new chaos testing scenario directory
chaossql init my_ecom_checkout \
  --driver sqlite \
  --template inventory`,
  validate: `# Statically validate schema DDL, SQL syntax, and invariant assertions
chaossql validate examples/hospital_write_skew/chaos.yaml`,
  matrix: `# Generate Hermitage empirical isolation matrix for target engine
chaossql matrix --driver sqlite --json
chaossql matrix --driver postgres --dsn "postgres://user:pass@localhost:5432/db"`,
  diff: `# Differential fuzzing between two database engines to detect semantic divergence
chaossql diff examples/banking_lost_update/chaos.yaml \
  --driver-a sqlite \
  --driver-b postgres \
  --dsn-b "postgres://user:pass@localhost:5432/test" \
  --export-diff-report ./diff_report.json`,
  bench: `# High-throughput engine stress benchmark (13.9M+ ops/s)
chaossql bench --iterations 50000 --concurrency 16`,
  github_action: `name: Concurrency Isolation Gate

on:
  pull_request:
    branches: [main]

jobs:
  chaos-fuzz:
    name: ChaosSQL Concurrency Gate
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Run ChaosSQL Scenario
        uses: bregaldahq/chaossql@v1.1
        with:
          spec-path: 'examples/banking_lost_update/chaos.yaml'
          workers: 4
          iterations: 30
          seed: 42
          export-summary: 'summary.md'
          export-junit: 'junit.xml'
          export-repro: 'repro_test.go'

      - name: Publish Test Results
        if: always()
        uses: EnricoMi/publish-unit-test-result-action@v2
        with:
          files: junit.xml`
};

// ==========================================================================
// 4. TERMINAL SIMULATOR SCRIPT (LIVE DDMIN RUN)
// ==========================================================================
const TERMINAL_LOGS = [
  { type: "prompt", cmd: "chaossql run examples/banking_lost_update/chaos.yaml --shrink --verbose" },
  { type: "info", text: "⚡ ChaosSQL v1.1.0 • Causal Concurrency Stress Tester & Anomaly Synthesizer" },
  { type: "dim", text: "• Loading schema.sql and seed.sql into modernc.org/sqlite (Zero CGO Engine)..." },
  { type: "dim", text: "• Initialized 4 worker threads with stochastic micro-jitter [1ms, 10ms] (Seed=42)..." },
  { type: "dim", text: "• Executing concurrent schedule S with 20 interleaved operations..." },
  { type: "error", text: "✘ ISOLATION ANOMALY DETECTED [P4_LOST_UPDATE]" },
  { type: "warn", text: "  ↳ Invariant Failed: 'ledger_balance_consistency'" },
  { type: "warn", text: "  ↳ Expected: $800.00 (2 debits of $100) | Actual: $900.00 (1 debit lost!)" },
  { type: "info", text: "• Building Serialization Graph DSG(S) = (V, E)..." },
  { type: "error", text: "  ↳ Directed Cycle Found: T₁ ──(rw)──► T₂ ──(ww)──► T₁" },
  { type: "info", text: "🔍 Activating Causal Delta-Debugging (ddmin Zeller 2002)..." },
  { type: "dim", text: "  [ddmin] Iteration 1: Trimming trace length 20 ➔ 10 operations... [FAIL: Reproduces]" },
  { type: "dim", text: "  [ddmin] Iteration 2: Trimming trace length 10 ➔ 4 operations...  [FAIL: Reproduces]" },
  { type: "dim", text: "  [ddmin] Iteration 3: Trimming trace length 4 ➔ 2 operations...   [FAIL: Reproduces]" },
  { type: "success", text: "  [ddmin] 1-Minimality Condition Met: Exactly 2 operations isolated in 74ms (-90.0% noise reduction)" },
  { type: "success", text: "✓ Synthesized standalone reproduction test: repro_test.go" },
  { type: "success", text: "✓ Exported dark-mode visual report: report.html" },
  { type: "prompt", cmd: "" }
];

// ==========================================================================
// 5. APPLICATION CONTROLLER & INITIALIZATION
// ==========================================================================
document.addEventListener("DOMContentLoaded", () => {
  initStickyHeader();
  initMobileMenu();
  initScenarioExplorer();
  initSdkPlayground();
  initCliHub();
  initMatrixFilter();
  initCopyButtons();
  initTerminalSimulator();
  initScrollSpy();
});

// --------------------------------------------------------------------------
// Sticky Header & Glassmorphism
// --------------------------------------------------------------------------
function initStickyHeader() {
  const header = document.querySelector(".site-header");
  if (!header) return;
  window.addEventListener("scroll", () => {
    if (window.scrollY > 20) {
      header.classList.add("scrolled");
    } else {
      header.classList.remove("scrolled");
    }
  }, { passive: true });
}

// --------------------------------------------------------------------------
// Mobile Menu Toggle
// --------------------------------------------------------------------------
function initMobileMenu() {
  const toggle = document.querySelector(".mobile-menu-toggle");
  const navLinks = document.querySelector(".nav-links");
  if (!toggle || !navLinks) return;

  toggle.addEventListener("click", () => {
    navLinks.classList.toggle("mobile-open");
    toggle.classList.toggle("active");
  });

  navLinks.querySelectorAll(".nav-link").forEach(link => {
    link.addEventListener("click", () => {
      navLinks.classList.remove("mobile-open");
      toggle.classList.remove("active");
    });
  });
}

// --------------------------------------------------------------------------
// Scenario Explorer Controller
// --------------------------------------------------------------------------
let activeScenarioId = "banking_lost_update";
let activeScenarioSubTab = "schema";

function initScenarioExplorer() {
  const navContainer = document.getElementById("scenario-nav-grid");
  if (!navContainer) return;

  // Render scenario navigation cards
  navContainer.innerHTML = SCENARIOS.map((s, idx) => `
    <button class="scenario-card-btn ${s.id === activeScenarioId ? "active" : ""}" data-scenario-id="${s.id}">
      <div class="scenario-code">${s.code} • ${s.category}</div>
      <div class="scenario-name">${s.name}</div>
    </button>
  `).join("");

  // Add click events to scenario cards
  navContainer.querySelectorAll(".scenario-card-btn").forEach(btn => {
    btn.addEventListener("click", () => {
      const id = btn.getAttribute("data-scenario-id");
      if (id === activeScenarioId) return;

      activeScenarioId = id;
      navContainer.querySelectorAll(".scenario-card-btn").forEach(b => b.classList.remove("active"));
      btn.classList.add("active");
      renderActiveScenario();
    });
  });

  renderActiveScenario();
}

function renderActiveScenario() {
  const scenario = SCENARIOS.find(s => s.id === activeScenarioId) || SCENARIOS[0];
  const detailPanel = document.getElementById("scenario-detail-panel");
  if (!detailPanel) return;

  detailPanel.innerHTML = `
    <div class="glass-card card-glow-top">
      <!-- Scenario Header -->
      <div class="flex flex-wrap items-center justify-between gap-md" style="margin-bottom: 1.5rem; border-bottom: 1px solid var(--border-subtle); padding-bottom: 1.25rem;">
        <div>
          <div class="flex items-center gap-sm" style="margin-bottom: 0.35rem;">
            <span class="badge badge-gold badge-lg">${scenario.code}</span>
            <span class="badge badge-purple">${scenario.category}</span>
            <span class="badge ${scenario.severity.includes("CRITICAL") ? "badge-red" : "badge-gold"}">${scenario.severity}</span>
          </div>
          <h3 style="font-size: 1.5rem; font-weight: 800; margin-top: 0.25rem;">${scenario.name}</h3>
        </div>
        <div class="text-right" style="font-size: 0.8125rem;">
          <div style="color: var(--text-muted); margin-bottom: 0.2rem;">Adya Classification:</div>
          <div class="font-mono text-gold" style="font-weight: 600;">${scenario.adyaType}</div>
        </div>
      </div>

      <!-- Context Box -->
      <div style="background: rgba(14, 10, 24, 0.7); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 1.25rem; margin-bottom: 2rem;">
        <h4 style="font-size: 0.875rem; font-weight: 700; color: var(--color-accent-gold); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.5rem;">
          Business Concurrency Context & Flaw
        </h4>
        <p style="color: var(--text-secondary); line-height: 1.6; font-size: 0.9375rem; margin-bottom: 0.75rem;">
          ${scenario.context}
        </p>
        <div class="flex flex-wrap gap-md" style="font-size: 0.8125rem; border-top: 1px dashed var(--border-subtle); padding-top: 0.75rem;">
          <div><strong class="text-red">Vulnerable Under:</strong> <span style="color: var(--text-primary);">${scenario.vulnerableLevels}</span></div>
          <div><strong class="text-green">Formally Protected By:</strong> <span style="color: var(--text-primary);">${scenario.safeLevels}</span></div>
        </div>
      </div>

      <!-- Scenario Sub-Tabs Nav -->
      <div class="tab-nav" style="margin-bottom: 1.5rem;">
        <button class="tab-btn ${activeScenarioSubTab === "schema" ? "active" : ""}" data-subtab="schema">
          <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16"><path d="M0 2a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2V2zm15 2h-4v3h4V4zm0 4h-4v3h4V8zm0 4h-4v3h3a1 1 0 0 0 1-1v-2zm-5 3v-3H6v3h4zm-5 0v-3H1v2a1 1 0 0 0 1 1h3zm-4-4h4V8H1v3zm0-4h4V4H1v3zm5-3v3h4V4H6zm4 4H6v3h4V8z"/></svg>
          Schema & Seed SQL
        </button>
        <button class="tab-btn ${activeScenarioSubTab === "chaos" ? "active" : ""}" data-subtab="chaos">
          <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16"><path d="M5.52.359A.5.5 0 0 1 6 0h4a.5.5 0 0 1 .474.658L8.694 6H12.5a.5.5 0 0 1 .395.807l-7 9a.5.5 0 0 1-.873-.454L6.82 9H3.5a.5.5 0 0 1-.48-.641l2.5-8z"/></svg>
          Chaos Operations YAML
        </button>
        <button class="tab-btn ${activeScenarioSubTab === "invariant" ? "active" : ""}" data-subtab="invariant">
          <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16"><path d="M9.05.435c-.58-.58-1.52-.58-2.1 0L.436 6.95c-.58.58-.58 1.519 0 2.098l6.516 6.516c.58.58 1.519.58 2.098 0l6.516-6.516c.58-.58.58-1.519 0-2.098L9.05.435zM5.495 6.033a.237.237 0 0 1-.24-.247C5.35 4.12 6.7 3.01 8.253 3.01c1.554 0 2.903 1.11 3.003 2.776a.237.237 0 0 1-.24.247h-.67a.237.237 0 0 1-.237-.225c-.08-.95-.87-1.638-1.856-1.638-1.026 0-1.82.688-1.856 1.638a.237.237 0 0 1-.237.225h-.667z"/></svg>
          Invariant Assertion
        </button>
        <button class="tab-btn ${activeScenarioSubTab === "ddmin" ? "active" : ""}" data-subtab="ddmin">
          <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16"><path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"/><path d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3.5a.5.5 0 0 1-.5-.5v-4A.5.5 0 0 1 8 4z"/></svg>
          Causal ddmin Reduction
        </button>
        <button class="tab-btn ${activeScenarioSubTab === "graph" ? "active" : ""}" data-subtab="graph">
          <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16"><path d="M1 2.5A1.5 1.5 0 0 1 2.5 1h3A1.5 1.5 0 0 1 7 2.5v3A1.5 1.5 0 0 1 5.5 7h-3A1.5 1.5 0 0 1 1 5.5v-3zM2.5 2a.5.5 0 0 0-.5.5v3a.5.5 0 0 0 .5.5h3a.5.5 0 0 0 .5-.5v-3a.5.5 0 0 0-.5-.5h-3zm6.5.5A1.5 1.5 0 0 1 10.5 1h3A1.5 1.5 0 0 1 15 2.5v3A1.5 1.5 0 0 1 13.5 7h-3A1.5 1.5 0 0 1 9 5.5v-3zm1.5-.5a.5.5 0 0 0-.5.5v3a.5.5 0 0 0 .5.5h3a.5.5 0 0 0 .5-.5v-3a.5.5 0 0 0-.5-.5h-3zM1 10.5A1.5 1.5 0 0 1 2.5 9h3A1.5 1.5 0 0 1 7 10.5v3A1.5 1.5 0 0 1 5.5 15h-3A1.5 1.5 0 0 1 1 13.5v-3zm1.5-.5a.5.5 0 0 0-.5.5v3a.5.5 0 0 0 .5.5h3a.5.5 0 0 0 .5-.5v-3a.5.5 0 0 0-.5-.5h-3z"/></svg>
          Adya Conflict Graph
        </button>
      </div>

      <!-- Sub-Tab Content -->
      <div id="scenario-subtab-content">
        ${renderScenarioSubTabContent(scenario, activeScenarioSubTab)}
      </div>
    </div>
  `;

  // Attach event listener to subtabs
  detailPanel.querySelectorAll(".tab-nav .tab-btn").forEach(btn => {
    btn.addEventListener("click", () => {
      activeScenarioSubTab = btn.getAttribute("data-subtab");
      detailPanel.querySelectorAll(".tab-nav .tab-btn").forEach(b => b.classList.remove("active"));
      btn.classList.add("active");
      const contentEl = document.getElementById("scenario-subtab-content");
      if (contentEl) {
        contentEl.innerHTML = renderScenarioSubTabContent(scenario, activeScenarioSubTab);
        initCopyButtons();
      }
    });
  });

  initCopyButtons();
}

function renderScenarioSubTabContent(scenario, tab) {
  if (tab === "schema") {
    return `
      <div class="code-block-wrapper">
        <div class="code-header">
          <span class="code-lang-label">SQL Schema & Seed</span>
          <span class="code-filename">schema.sql • seed.sql</span>
          <button class="copy-btn" data-copy-target="#scenario-sql-code">
            <svg width="14" height="14" fill="currentColor" viewBox="0 0 16 16"><path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/><path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/></svg>
            Copy SQL
          </button>
        </div>
        <pre><code id="scenario-sql-code" class="language-sql">${escapeHtml(scenario.schemaSql)}</code></pre>
      </div>
    `;
  }

  if (tab === "chaos") {
    return `
      <div class="code-block-wrapper">
        <div class="code-header">
          <span class="code-lang-label">YAML Chaos Spec</span>
          <span class="code-filename">chaos.yaml</span>
          <button class="copy-btn" data-copy-target="#scenario-yaml-code">
            <svg width="14" height="14" fill="currentColor" viewBox="0 0 16 16"><path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/><path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/></svg>
            Copy YAML
          </button>
        </div>
        <pre><code id="scenario-yaml-code" class="language-yaml">${escapeHtml(scenario.chaosYaml)}</code></pre>
      </div>
    `;
  }

  if (tab === "invariant") {
    return `
      <div style="display: flex; flex-direction: column; gap: 1.5rem;">
        <div class="code-block-wrapper" style="margin: 0;">
          <div class="code-header">
            <span class="code-lang-label">Verification Query (SQL)</span>
            <span class="code-filename">Invariant Probe</span>
            <button class="copy-btn" data-copy-target="#scenario-inv-query">
              <svg width="14" height="14" fill="currentColor" viewBox="0 0 16 16"><path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/><path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/></svg>
              Copy
            </button>
          </div>
          <pre><code id="scenario-inv-query" class="language-sql">${escapeHtml(scenario.invariantQuery)}</code></pre>
        </div>

        <div style="background: rgba(18, 14, 30, 0.9); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 1.25rem;">
          <div class="flex items-center justify-between" style="margin-bottom: 0.75rem;">
            <span style="font-size: 0.8125rem; font-weight: 700; color: var(--color-accent-gold); text-transform: uppercase;">Boolean Assertion Logic</span>
            <span class="badge badge-green">Inductive Integrity</span>
          </div>
          <div class="font-mono text-gold" style="font-size: 1rem; font-weight: 600; padding: 0.5rem 0.75rem; background: rgba(0,0,0,0.3); border-radius: var(--radius-xs); margin-bottom: 0.75rem;">
            ${escapeHtml(scenario.invariantAssertion)}
          </div>
          <p style="color: var(--text-secondary); font-size: 0.875rem; line-height: 1.5;">
            ${scenario.invariantExplanation}
          </p>
        </div>
      </div>
    `;
  }

  if (tab === "ddmin") {
    const s = scenario.ddminStats;
    return `
      <div style="display: flex; flex-direction: column; gap: 1.5rem;">
        <div class="grid grid-4 gap-md">
          <div style="background: rgba(18, 14, 30, 0.8); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 1.25rem; text-align: center;">
            <div style="font-size: 0.75rem; color: var(--text-muted); text-transform: uppercase; margin-bottom: 0.25rem;">Initial Noisy Trace</div>
            <div style="font-size: 1.75rem; font-weight: 800; color: var(--color-accent-gold);">${s.initialOps} <span style="font-size: 0.9rem; font-weight: 500; color: var(--text-secondary);">ops</span></div>
          </div>
          <div style="background: rgba(18, 14, 30, 0.8); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 1.25rem; text-align: center;">
            <div style="font-size: 0.75rem; color: var(--text-muted); text-transform: uppercase; margin-bottom: 0.25rem;">Minimal Repro Trace</div>
            <div style="font-size: 1.75rem; font-weight: 800; color: var(--color-accent-green);">${s.shrunkOps} <span style="font-size: 0.9rem; font-weight: 500; color: var(--text-secondary);">ops</span></div>
          </div>
          <div style="background: rgba(18, 14, 30, 0.8); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 1.25rem; text-align: center;">
            <div style="font-size: 0.75rem; color: var(--text-muted); text-transform: uppercase; margin-bottom: 0.25rem;">Noise Reduction</div>
            <div style="font-size: 1.75rem; font-weight: 800; color: var(--color-accent-cyan);">${s.reductionPct}</div>
          </div>
          <div style="background: rgba(18, 14, 30, 0.8); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 1.25rem; text-align: center;">
            <div style="font-size: 0.75rem; color: var(--text-muted); text-transform: uppercase; margin-bottom: 0.25rem;">Convergence Time</div>
            <div style="font-size: 1.75rem; font-weight: 800; color: var(--text-purple);">${s.shrinkTimeMs} <span style="font-size: 0.9rem; font-weight: 500; color: var(--text-secondary);">ms</span></div>
          </div>
        </div>

        <div style="background: rgba(14, 10, 24, 0.8); border: 1px solid var(--border-subtle); border-radius: var(--radius-md); padding: 1.25rem;">
          <h4 style="font-size: 0.875rem; font-weight: 700; color: var(--color-accent-gold); text-transform: uppercase; margin-bottom: 0.75rem;">
            1-Minimal Delta-Debugging Convergence Proof (Andreas Zeller 2002)
          </h4>
          <p style="color: var(--text-secondary); font-size: 0.875rem; line-height: 1.6; margin-bottom: 0.75rem;">
            ChaosSQL automatically bisects the scheduled transaction trace using the causal oracle function <code>test(C*) = FAIL</code>.
            Because <code>∀ op ∈ C*: test(C* \\ {op}) = PASS</code>, the reduced sequence contains zero extraneous noise operations.
          </p>
          <div class="flex items-center gap-sm">
            <span class="badge badge-green"><span class="badge-dot"></span> ${s.status}</span>
            <span class="badge badge-neutral">Synthesized in &lt; 200ms</span>
          </div>
        </div>
      </div>
    `;
  }

  if (tab === "graph") {
    return renderAdyaGraphSvg(scenario.graph);
  }

  return "";
}

function renderAdyaGraphSvg(graph) {
  const isTriParty = graph.nodes.length === 3;
  return `
    <div class="adya-graph-box">
      <div class="graph-cycle-indicator">
        <span class="badge badge-red"><span class="badge-dot pulse"></span> CYCLE DETECTED</span>
      </div>
      
      <svg width="580" height="240" viewBox="0 0 580 240" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient id="purpleGrad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#6E44B8"/>
            <stop offset="100%" stop-color="#4B2E83"/>
          </linearGradient>
          <linearGradient id="goldGrad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#FFD02E"/>
            <stop offset="100%" stop-color="#F5C400"/>
          </linearGradient>
          <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
            <feGaussianBlur stdDeviation="6" result="blur"/>
            <feComposite in="SourceGraphic" in2="blur" operator="over"/>
          </filter>
          <marker id="arrowRed" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
            <path d="M 0 0 L 8 4 L 0 8 z" fill="#EF4444"/>
          </marker>
          <marker id="arrowGold" markerWidth="8" markerHeight="8" refX="6" refY="4" orient="auto">
            <path d="M 0 0 L 8 4 L 0 8 z" fill="#F5C400"/>
          </marker>
        </defs>

        ${isTriParty ? `
          <!-- Tri-Party Triangle Nodes (T1, T2, T3) -->
          <g filter="url(#glow)">
            <!-- Edges -->
            <path d="M 160 70 L 420 70" stroke="#EF4444" stroke-width="2.5" stroke-dasharray="6,4" marker-end="url(#arrowRed)"/>
            <path d="M 450 100 L 300 200" stroke="#EF4444" stroke-width="2.5" stroke-dasharray="6,4" marker-end="url(#arrowRed)"/>
            <path d="M 280 200 L 130 100" stroke="#EF4444" stroke-width="2.5" stroke-dasharray="6,4" marker-end="url(#arrowRed)"/>
          </g>

          <!-- Edge Labels -->
          <text x="290" y="55" fill="#EF4444" font-family="JetBrains Mono" font-size="11" font-weight="700" text-anchor="middle">rw (Seat 2 → 1)</text>
          <text x="410" y="160" fill="#EF4444" font-family="JetBrains Mono" font-size="11" font-weight="700" text-anchor="middle">rw (Seat 3 → 2)</text>
          <text x="170" y="160" fill="#EF4444" font-family="JetBrains Mono" font-size="11" font-weight="700" text-anchor="middle">rw (Seat 1 → 3)</text>

          <!-- Node T1 -->
          <circle cx="120" cy="70" r="32" fill="url(#purpleGrad)" stroke="#6E44B8" stroke-width="2"/>
          <text x="120" y="75" fill="#FCFBF8" font-family="Inter" font-weight="700" font-size="13" text-anchor="middle">T₁</text>

          <!-- Node T2 -->
          <circle cx="460" cy="70" r="32" fill="url(#purpleGrad)" stroke="#6E44B8" stroke-width="2"/>
          <text x="460" y="75" fill="#FCFBF8" font-family="Inter" font-weight="700" font-size="13" text-anchor="middle">T₂</text>

          <!-- Node T3 -->
          <circle cx="290" cy="210" r="32" fill="url(#purpleGrad)" stroke="#6E44B8" stroke-width="2"/>
          <text x="290" y="215" fill="#FCFBF8" font-family="Inter" font-weight="700" font-size="13" text-anchor="middle">T₃</text>
        ` : `
          <!-- Bi-Party 2-Node Graph (T1 <--> T2) -->
          <g filter="url(#glow)">
            <!-- Top Edge: T1 -> T2 -->
            <path d="M 160 90 C 260 40, 320 40, 420 90" fill="none" stroke="#EF4444" stroke-width="2.5" stroke-dasharray="6,4" marker-end="url(#arrowRed)"/>
            <!-- Bottom Edge: T2 -> T1 -->
            <path d="M 420 150 C 320 200, 260 200, 160 150" fill="none" stroke="#F5C400" stroke-width="2.5" stroke-dasharray="6,4" marker-end="url(#arrowGold)"/>
          </g>

          <!-- Edge Labels -->
          <text x="290" y="55" fill="#EF4444" font-family="JetBrains Mono" font-size="12" font-weight="700" text-anchor="middle">${graph.edges[0]?.label || "rw (anti-dependency)"}</text>
          <text x="290" y="195" fill="#F5C400" font-family="JetBrains Mono" font-size="12" font-weight="700" text-anchor="middle">${graph.edges[1]?.label || "ww (write-dependency)"}</text>

          <!-- Node T1 -->
          <circle cx="120" cy="120" r="36" fill="url(#purpleGrad)" stroke="#6E44B8" stroke-width="2"/>
          <text x="120" y="125" fill="#FCFBF8" font-family="Inter" font-weight="700" font-size="14" text-anchor="middle">${graph.nodes[0] ? graph.nodes[0].split(" ")[0] : "T₁"}</text>

          <!-- Node T2 -->
          <circle cx="460" cy="120" r="36" fill="url(#purpleGrad)" stroke="#6E44B8" stroke-width="2"/>
          <text x="460" y="125" fill="#FCFBF8" font-family="Inter" font-weight="700" font-size="14" text-anchor="middle">${graph.nodes[1] ? graph.nodes[1].split(" ")[0] : "T₂"}</text>
        `}
      </svg>
      <div style="font-family: var(--font-mono); font-size: 0.8125rem; color: var(--color-accent-gold); margin-top: 1rem; font-weight: 600;">
        Formal Cycle: ${graph.cycle}
      </div>
    </div>
  `;
}

// --------------------------------------------------------------------------
// SDK Playground Controller
// --------------------------------------------------------------------------
let activeSdkTab = "banking";

function initSdkPlayground() {
  const container = document.getElementById("sdk-code-container");
  const nav = document.getElementById("sdk-tab-nav");
  if (!container || !nav) return;

  function updateSdkCode() {
    container.innerHTML = `
      <div class="code-block-wrapper" style="margin: 0;">
        <div class="code-header">
          <span class="code-lang-label">Go Testing SDK (*testing.T)</span>
          <span class="code-filename">myapp_test.go</span>
          <button class="copy-btn" data-copy-target="#sdk-go-code">
            <svg width="14" height="14" fill="currentColor" viewBox="0 0 16 16"><path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/><path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/></svg>
            Copy Test
          </button>
        </div>
        <pre><code id="sdk-go-code" class="language-go">${escapeHtml(SDK_SAMPLES[activeSdkTab] || "")}</code></pre>
      </div>
    `;
    initCopyButtons();
  }

  nav.querySelectorAll(".tab-btn").forEach(btn => {
    btn.addEventListener("click", () => {
      activeSdkTab = btn.getAttribute("data-sdk-tab");
      nav.querySelectorAll(".tab-btn").forEach(b => b.classList.remove("active"));
      btn.classList.add("active");
      updateSdkCode();
    });
  });

  updateSdkCode();
}

// --------------------------------------------------------------------------
// CLI Hub Controller
// --------------------------------------------------------------------------
let activeCliTab = "run";

function initCliHub() {
  const container = document.getElementById("cli-code-container");
  const nav = document.getElementById("cli-tab-nav");
  if (!container || !nav) return;

  function updateCliCode() {
    const isYaml = activeCliTab === "github_action";
    container.innerHTML = `
      <div class="code-block-wrapper" style="margin: 0;">
        <div class="code-header">
          <span class="code-lang-label">${isYaml ? "GitHub Actions CI (.github/workflows)" : "Terminal Shell Command"}</span>
          <span class="code-filename">${isYaml ? "ci.yml" : "bash"}</span>
          <button class="copy-btn" data-copy-target="#cli-snippet-code">
            <svg width="14" height="14" fill="currentColor" viewBox="0 0 16 16"><path d="M4 1.5H3a2 2 0 0 0-2 2V14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V3.5a2 2 0 0 0-2-2h-1v1h1a1 1 0 0 1 1 1V14a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V3.5a1 1 0 0 1 1-1h1v-1z"/><path d="M9.5 1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5h-3a.5.5 0 0 1-.5-.5v-1a.5.5 0 0 1 .5-.5h3zm-3-1A1.5 1.5 0 0 0 5 1.5v1A1.5 1.5 0 0 0 6.5 4h3A1.5 1.5 0 0 0 11 2.5v-1A1.5 1.5 0 0 0 9.5 0h-3z"/></svg>
            Copy Snippet
          </button>
        </div>
        <pre><code id="cli-snippet-code" class="language-${isYaml ? "yaml" : "bash"}">${escapeHtml(CLI_SAMPLES[activeCliTab] || "")}</code></pre>
      </div>
    `;
    initCopyButtons();
  }

  nav.querySelectorAll(".tab-btn").forEach(btn => {
    btn.addEventListener("click", () => {
      activeCliTab = btn.getAttribute("data-cli-tab");
      nav.querySelectorAll(".tab-btn").forEach(b => b.classList.remove("active"));
      btn.classList.add("active");
      updateCliCode();
    });
  });

  updateCliCode();
}

// --------------------------------------------------------------------------
// Hermitage Matrix Filter Controller
// --------------------------------------------------------------------------
function initMatrixFilter() {
  const filterBtns = document.querySelectorAll("[data-matrix-filter]");
  const rows = document.querySelectorAll(".matrix-table tbody tr");
  if (!filterBtns.length || !rows.length) return;

  filterBtns.forEach(btn => {
    btn.addEventListener("click", () => {
      const filter = btn.getAttribute("data-matrix-filter");
      filterBtns.forEach(b => b.classList.remove("active"));
      btn.classList.add("active");

      rows.forEach(row => {
        if (filter === "all") {
          row.style.display = "";
        } else {
          const category = row.getAttribute("data-category");
          if (category === filter) {
            row.style.display = "";
          } else {
            row.style.display = "none";
          }
        }
      });
    });
  });
}

// --------------------------------------------------------------------------
// Copy-to-Clipboard Controller with Tooltip Feedback
// --------------------------------------------------------------------------
function initCopyButtons() {
  document.querySelectorAll(".copy-btn").forEach(btn => {
    // Avoid double attaching
    if (btn.hasAttribute("data-copy-bound")) return;
    btn.setAttribute("data-copy-bound", "true");

    btn.addEventListener("click", async () => {
      const targetSelector = btn.getAttribute("data-copy-target");
      let textToCopy = "";

      if (targetSelector) {
        const el = document.querySelector(targetSelector);
        if (el) textToCopy = el.innerText;
      } else if (btn.getAttribute("data-copy-text")) {
        textToCopy = btn.getAttribute("data-copy-text");
      }

      if (!textToCopy) return;

      try {
        await navigator.clipboard.writeText(textToCopy);
        const originalHtml = btn.innerHTML;
        btn.classList.add("copied");
        btn.innerHTML = `
          <svg width="14" height="14" fill="currentColor" viewBox="0 0 16 16"><path d="M13.854 3.646a.5.5 0 0 1 0 .708l-7 7a.5.5 0 0 1-.708 0l-3.5-3.5a.5.5 0 1 1 .708-.708L6.5 10.293l6.646-6.647a.5.5 0 0 1 .708 0z"/></svg>
          Copied!
        `;
        setTimeout(() => {
          btn.classList.remove("copied");
          btn.innerHTML = originalHtml;
        }, 2000);
      } catch (err) {
        console.error("Clipboard copy failed:", err);
      }
    });
  });

  // Hero install copy box
  const installBox = document.querySelector(".hero-install-box");
  if (installBox && !installBox.hasAttribute("data-copy-bound")) {
    installBox.setAttribute("data-copy-bound", "true");
    installBox.addEventListener("click", async () => {
      const cmd = "go install github.com/bregaldahq/chaossql/cmd/chaossql@latest";
      try {
        await navigator.clipboard.writeText(cmd);
        const promptEl = installBox.querySelector(".prompt");
        if (promptEl) {
          const orig = promptEl.textContent;
          promptEl.textContent = "✓ Copied!";
          promptEl.style.color = "var(--color-accent-green)";
          setTimeout(() => {
            promptEl.textContent = orig;
            promptEl.style.color = "var(--color-accent-gold)";
          }, 2000);
        }
      } catch (e) {
        console.error("Failed to copy install command", e);
      }
    });
  }
}

// --------------------------------------------------------------------------
// Terminal Simulator Controller (Hero Live Animation)
// --------------------------------------------------------------------------
function initTerminalSimulator() {
  const terminalBody = document.getElementById("terminal-screen");
  const replayBtn = document.getElementById("terminal-replay-btn");
  if (!terminalBody) return;

  let timerId = null;
  let lineIdx = 0;
  let isRunning = false;

  function runSimulation() {
    if (timerId) clearTimeout(timerId);
    terminalBody.innerHTML = "";
    lineIdx = 0;
    isRunning = true;
    stepNextLine();
  }

  function stepNextLine() {
    if (lineIdx >= TERMINAL_LOGS.length) {
      isRunning = false;
      return;
    }

    const log = TERMINAL_LOGS[lineIdx];
    lineIdx++;

    let lineHtml = "";
    if (log.type === "prompt") {
      if (log.cmd) {
        lineHtml = `
          <div class="terminal-line">
            <span class="terminal-prompt">$</span>
            <span class="terminal-cmd">${escapeHtml(log.cmd)}</span>
          </div>
        `;
      } else {
        lineHtml = `
          <div class="terminal-line">
            <span class="terminal-prompt">$</span>
            <span class="terminal-cursor"></span>
          </div>
        `;
      }
    } else {
      const cls = "terminal-" + log.type;
      lineHtml = `<div class="terminal-out ${cls}">${escapeHtml(log.text)}</div>`;
    }

    // Append line
    terminalBody.insertAdjacentHTML("beforeend", lineHtml);
    terminalBody.scrollTop = terminalBody.scrollHeight;

    // Delay before next line
    let delay = 180;
    if (log.type === "prompt") delay = 350;
    if (log.type === "error") delay = 300;
    if (log.text && log.text.includes("[ddmin]")) delay = 250;

    timerId = setTimeout(stepNextLine, delay);
  }

  if (replayBtn) {
    replayBtn.addEventListener("click", () => {
      runSimulation();
    });
  }

  // Auto-start with subtle delay on page load
  setTimeout(runSimulation, 600);
}

// --------------------------------------------------------------------------
// ScrollSpy & Navigation Active State
// --------------------------------------------------------------------------
function initScrollSpy() {
  const sections = document.querySelectorAll("section[id]");
  const navLinks = document.querySelectorAll(".nav-link[href^='#']");
  if (!sections.length || !navLinks.length) return;

  const observer = new IntersectionObserver(entries => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const id = entry.target.getAttribute("id");
        navLinks.forEach(link => {
          if (link.getAttribute("href") === "#" + id) {
            link.classList.add("active");
          } else {
            link.classList.remove("active");
          }
        });
      }
    });
  }, { threshold: 0.25, rootMargin: "-80px 0px -40% 0px" });

  sections.forEach(s => observer.observe(s));
}

// Helper: Escape HTML
function escapeHtml(str) {
  if (!str) return "";
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}
