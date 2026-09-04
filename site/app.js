// ChaosSQL — Studio Bregalda Interactive Controller
const SCENARIOS = [
  {
    id: "banking",
    name: "Banking Lost Update",
    code: "P4",
    summary: "Concurrent withdrawals under READ COMMITTED overwrite balances without transactional serialization.",
    schema: `-- Schema
CREATE TABLE accounts (
    id INT PRIMARY KEY,
    owner TEXT NOT NULL,
    balance INT NOT NULL
);

-- Seed
INSERT INTO accounts VALUES (1, 'Alice', 1000);`,
    chaos: `version: "1.0"
name: "banking_lost_update"
database:
  driver: "sqlite"
operations:
  - name: "withdraw_100"
    steps:
      - "SELECT balance FROM accounts WHERE id = 1 -> cur"
      - "UPDATE accounts SET balance = {cur - 100} WHERE id = 1"
invariants:
  - name: "total_balance_check"
    query: "SELECT balance FROM accounts WHERE id = 1;"
    assert: "balance == 1000 - (total_completed * 100)"`,
    reduction: {
      originalOps: 20,
      minimalOps: 2,
      reductionPct: "90.0%",
      elapsed: "68ms",
      cycle: "T1 ──(rw)──► T2 ──(ww)──► T1",
      explanation: "Two concurrent read-modify-write transactions read balance 1000 simultaneously. T1 writes 900, but T2 overwrites with 900 based on its stale read, silently losing $100."
    }
  },
  {
    id: "inventory",
    name: "Inventory Oversell",
    code: "A3",
    summary: "Concurrent checkouts reserve the final available item due to predicate anti-dependencies (phantom reads).",
    schema: `-- Schema
CREATE TABLE products (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    stock INT NOT NULL
);

-- Seed
INSERT INTO products VALUES (1, 'RTX 5090 GPU', 1);`,
    chaos: `version: "1.0"
name: "inventory_oversell"
operations:
  - name: "checkout"
    steps:
      - "SELECT stock FROM products WHERE id = 1 -> avail"
      - "UPDATE products SET stock = {avail - 1} WHERE id = 1 AND {avail > 0}"
invariants:
  - name: "stock_non_negative"
    query: "SELECT stock FROM products WHERE id = 1;"
    assert: "stock >= 0"`,
    reduction: {
      originalOps: 30,
      minimalOps: 2,
      reductionPct: "93.3%",
      elapsed: "74ms",
      cycle: "T1 ──(rw)──► T2 ──(ww)──► T1",
      explanation: "Both workers evaluate {avail > 0} as true before either commit updates the physical stock row, driving inventory to -1."
    }
  },
  {
    id: "hospital",
    name: "Hospital Write Skew",
    code: "A5B",
    summary: "Two doctors independently sign off duty under Snapshot Isolation because their write sets do not overlap.",
    schema: `-- Schema
CREATE TABLE doctors (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    on_call BOOLEAN NOT NULL
);

-- Seed
INSERT INTO doctors VALUES (1, 'Dr. Smith', 1), (2, 'Dr. Patel', 1);`,
    chaos: `version: "1.0"
name: "hospital_write_skew"
operations:
  - name: "leave_duty_1"
    steps:
      - "SELECT count(*) AS active FROM doctors WHERE on_call = 1 -> count"
      - "UPDATE doctors SET on_call = 0 WHERE id = 1 AND {count >= 2}"
  - name: "leave_duty_2"
    steps:
      - "SELECT count(*) AS active FROM doctors WHERE on_call = 1 -> count"
      - "UPDATE doctors SET on_call = 0 WHERE id = 2 AND {count >= 2}"
invariants:
  - name: "at_least_one_active"
    query: "SELECT count(*) AS active FROM doctors WHERE on_call = 1;"
    assert: "active >= 1"`,
    reduction: {
      originalOps: 16,
      minimalOps: 2,
      reductionPct: "87.5%",
      elapsed: "54ms",
      cycle: "T1 ──(rw)──► T2 ──(rw)──► T1",
      explanation: "Classic A5B Write Skew. Both transactions observe active=2. T1 updates Doctor 1, T2 updates Doctor 2. Both commit cleanly under Snapshot Isolation, leaving 0 doctors on duty."
    }
  },
  {
    id: "financial",
    name: "Financial Read Skew",
    code: "A5A",
    summary: "An audit query interleaves between the debit and credit legs of a transfer, calculating an incorrect balance total.",
    schema: `-- Schema
CREATE TABLE accounts (
    id INT PRIMARY KEY,
    type TEXT NOT NULL,
    balance INT NOT NULL
);

-- Seed
INSERT INTO accounts VALUES (1, 'Checking', 500), (2, 'Savings', 500);`,
    chaos: `version: "1.0"
name: "read_skew_financial_audit"
operations:
  - name: "transfer"
    steps:
      - "UPDATE accounts SET balance = balance - 200 WHERE id = 1"
      - "UPDATE accounts SET balance = balance + 200 WHERE id = 2"
  - name: "audit"
    steps:
      - "SELECT balance FROM accounts WHERE id = 1 -> c"
      - "SELECT balance FROM accounts WHERE id = 2 -> s"
invariants:
  - name: "total_wealth_preservation"
    query: "SELECT sum(balance) AS total FROM accounts;"
    assert: "total == 1000"`,
    reduction: {
      originalOps: 24,
      minimalOps: 2,
      reductionPct: "91.6%",
      elapsed: "62ms",
      cycle: "T1 ──(rw)──► T2 ──(wr)──► T1",
      explanation: "Audit transaction T1 reads Checking after debit (300), but reads Savings before credit (500), recording an apparent total wealth of 800 instead of 1000."
    }
  },
  {
    id: "auction",
    name: "Auction Dirty Write",
    code: "G0",
    summary: "Concurrent uncommitted updates interleave on the same item, resulting in winner ID mismatched from final bid price.",
    schema: `-- Schema
CREATE TABLE auction (
    id INT PRIMARY KEY,
    highest_bid INT NOT NULL,
    winner_id INT NOT NULL
);

-- Seed
INSERT INTO auction VALUES (1, 100, 0);`,
    chaos: `version: "1.0"
name: "auction_dirty_write"
operations:
  - name: "bid_alice"
    steps:
      - "UPDATE auction SET highest_bid = 200 WHERE id = 1"
      - "UPDATE auction SET winner_id = 101 WHERE id = 1"
  - name: "bid_bob"
    steps:
      - "UPDATE auction SET highest_bid = 300 WHERE id = 1"
      - "UPDATE auction SET winner_id = 102 WHERE id = 1"
invariants:
  - name: "winner_price_consistency"
    query: "SELECT (highest_bid == 200 AND winner_id == 101) OR (highest_bid == 300 AND winner_id == 102) AS ok FROM auction WHERE id = 1;"
    assert: "ok == 1"`,
    reduction: {
      originalOps: 12,
      minimalOps: 2,
      reductionPct: "83.3%",
      elapsed: "48ms",
      cycle: "T1 ──(ww)──► T2 ──(ww)──► T1",
      explanation: "T1 updates highest_bid to 200; T2 updates highest_bid to 300 and winner_id to 102; T1 then overwrites winner_id to 101, yielding price 300 won by bidder 101."
    }
  },
  {
    id: "crypto",
    name: "Crypto Arbitrage Circular Info",
    code: "G1c",
    summary: "Two automated liquidity pool transactions observe each other's intermediate state, producing circular dependency.",
    schema: `-- Schema
CREATE TABLE pool (
    id INT PRIMARY KEY,
    reserve_a INT NOT NULL,
    reserve_b INT NOT NULL
);

-- Seed
INSERT INTO pool VALUES (1, 1000, 1000);`,
    chaos: `version: "1.0"
name: "crypto_arbitrage"
operations:
  - name: "swap_a_for_b"
    steps:
      - "UPDATE pool SET reserve_a = reserve_a + 50 WHERE id = 1"
      - "UPDATE pool SET reserve_b = reserve_b - 50 WHERE id = 1"
  - name: "swap_b_for_a"
    steps:
      - "UPDATE pool SET reserve_b = reserve_b + 50 WHERE id = 1"
      - "UPDATE pool SET reserve_a = reserve_a - 50 WHERE id = 1"
invariants:
  - name: "constant_product"
    query: "SELECT (reserve_a + reserve_b) AS total FROM pool WHERE id = 1;"
    assert: "total == 2000"`,
    reduction: {
      originalOps: 18,
      minimalOps: 2,
      reductionPct: "88.8%",
      elapsed: "56ms",
      cycle: "T1 ──(wr)──► T2 ──(wr)──► T1",
      explanation: "Circular information flow G1c: T1 observes write of T2 while T2 observes write of T1, creating an impossible serial order."
    }
  },
  {
    id: "flashcrash",
    name: "Flash Crash Dirty Read",
    code: "G1a",
    summary: "A liquidation engine reads a temporary price drop from an aborted transaction and triggers erroneous liquidations.",
    schema: `-- Schema
CREATE TABLE oracle (
    id INT PRIMARY KEY,
    asset TEXT NOT NULL,
    price INT NOT NULL
);

-- Seed
INSERT INTO oracle VALUES (1, 'ETH', 3000);`,
    chaos: `version: "1.0"
name: "flash_crash_dirty_read"
faults:
  abort_probability: 0.5
operations:
  - name: "flash_crash"
    steps:
      - "UPDATE oracle SET price = 500 WHERE id = 1"
      - "SELECT pg_sleep(0.01)"
      - "ROLLBACK"
  - name: "liquidator"
    steps:
      - "SELECT price FROM oracle WHERE id = 1 -> p"
invariants:
  - name: "price_validity"
    query: "SELECT price FROM oracle WHERE id = 1;"
    assert: "price >= 2000"`,
    reduction: {
      originalOps: 14,
      minimalOps: 2,
      reductionPct: "85.7%",
      elapsed: "52ms",
      cycle: "w1(price) ... r2(price) ... a1",
      explanation: "T1 temporarily writes price=500 and subsequently aborts. T2 reads price=500 under READ UNCOMMITTED before the abort occurs."
    }
  },
  {
    id: "ticket",
    name: "Ticket Anti-Dependency Cycle",
    code: "G2",
    summary: "Three concurrent reservation transactions each read partial seat maps, creating a 3-way anti-dependency cycle.",
    schema: `-- Schema
CREATE TABLE seats (
    id INT PRIMARY KEY,
    reserved_by INT NOT NULL
);

-- Seed
INSERT INTO seats VALUES (1, 0), (2, 0), (3, 0);`,
    chaos: `version: "1.0"
name: "ticket_anti_dependency"
operations:
  - name: "reserve_adjacent"
    steps:
      - "SELECT count(*) AS free FROM seats WHERE reserved_by = 0 -> count"
      - "UPDATE seats SET reserved_by = $worker_id WHERE reserved_by = 0"
invariants:
  - name: "no_double_booking"
    query: "SELECT count(*) AS valid FROM seats GROUP BY reserved_by;"
    assert: "valid <= 3"`,
    reduction: {
      originalOps: 30,
      minimalOps: 3,
      reductionPct: "90.0%",
      elapsed: "82ms",
      cycle: "T1 ──(rw)──► T2 ──(rw)──► T3 ──(rw)──► T1",
      explanation: "A 3-node directed cycle containing only read-write anti-dependency edges (G2 cycle), violating serializability."
    }
  },
  {
    id: "deadlock",
    name: "Deadlock Cycle & Recovery",
    code: "G-DL",
    summary: "Mutual lock contention (T1 waits for T2, T2 waits for T1) triggers timeout aborts and invariant verification.",
    schema: `-- Schema
CREATE TABLE balances (
    id INT PRIMARY KEY,
    amount INT NOT NULL
);

-- Seed
INSERT INTO balances VALUES (1, 500), (2, 500);`,
    chaos: `version: "1.0"
name: "deadlock_cycle"
operations:
  - name: "lock_1_then_2"
    steps:
      - "UPDATE balances SET amount = amount - 10 WHERE id = 1"
      - "UPDATE balances SET amount = amount + 10 WHERE id = 2"
  - name: "lock_2_then_1"
    steps:
      - "UPDATE balances SET amount = amount - 20 WHERE id = 2"
      - "UPDATE balances SET amount = amount + 20 WHERE id = 1"
invariants:
  - name: "wealth_preservation"
    query: "SELECT sum(amount) AS total FROM balances;"
    assert: "total == 1000"`,
    reduction: {
      originalOps: 20,
      minimalOps: 2,
      reductionPct: "90.0%",
      elapsed: "65ms",
      cycle: "T1 ──(waits)──► T2 ──(waits)──► T1",
      explanation: "T1 holds lock on row 1 and requests row 2; T2 holds lock on row 2 and requests row 1. Engine aborts one transaction (SQLSTATE 40P01), rolling back cleanly."
    }
  }
];

let currentScenarioIndex = 0;
let currentTab = "schema";

function initApp() {
  renderScenarioNav();
  renderScenarioStage();
  setupEventListeners();
}

function renderScenarioNav() {
  const navContainer = document.getElementById("scenarioNavList");
  if (!navContainer) return;

  navContainer.innerHTML = SCENARIOS.map((sc, index) => `
    <button class="scenario-nav-item ${index === currentScenarioIndex ? 'active' : ''}" data-index="${index}">
      <span class="scenario-nav-name">${sc.name}</span>
      <span class="scenario-nav-code">${sc.code}</span>
    </button>
  `).join('');

  navContainer.querySelectorAll(".scenario-nav-item").forEach(btn => {
    btn.addEventListener("click", () => {
      currentScenarioIndex = parseInt(btn.getAttribute("data-index"), 10);
      renderScenarioNav();
      renderScenarioStage();
    });
  });
}

function renderScenarioStage() {
  const sc = SCENARIOS[currentScenarioIndex];
  if (!sc) return;

  const titleEl = document.getElementById("stageTitle");
  const summaryEl = document.getElementById("stageSummary");
  const contentEl = document.getElementById("stageTabContent");

  if (titleEl) titleEl.textContent = `${sc.name} (${sc.code})`;
  if (summaryEl) summaryEl.textContent = sc.summary;

  if (!contentEl) return;

  if (currentTab === "schema") {
    contentEl.innerHTML = `
      <div class="code-container">
        <button class="copy-code-btn" onclick="copySnippet(this)">Copy SQL</button>
        <pre><code>${escapeHtml(sc.schema)}</code></pre>
      </div>
    `;
  } else if (currentTab === "chaos") {
    contentEl.innerHTML = `
      <div class="code-container">
        <button class="copy-code-btn" onclick="copySnippet(this)">Copy YAML</button>
        <pre><code>${escapeHtml(sc.chaos)}</code></pre>
      </div>
    `;
  } else if (currentTab === "invariant") {
    contentEl.innerHTML = `
      <div style="margin-bottom: 16px;">
        <div style="font-family: var(--font-mono); font-size: 0.8rem; color: var(--color-yellow); margin-bottom: 4px;">ANOMALY CYCLE</div>
        <div style="font-family: var(--font-mono); font-size: 0.95rem; color: var(--color-cream);">${sc.reduction.cycle}</div>
      </div>
      <p style="font-size: 0.95rem; color: var(--text-secondary); line-height: 1.6; margin-bottom: 20px;">${sc.reduction.explanation}</p>
      <div class="metrics-row">
        <div class="metric-box">
          <div class="metric-val">${sc.reduction.originalOps} → ${sc.reduction.minimalOps}</div>
          <div class="metric-lbl">1-Minimal Ops Shrunk</div>
        </div>
        <div class="metric-box">
          <div class="metric-val">${sc.reduction.reductionPct}</div>
          <div class="metric-lbl">Causal Noise Removed</div>
        </div>
        <div class="metric-box">
          <div class="metric-val">${sc.reduction.elapsed}</div>
          <div class="metric-lbl">Convergence Time</div>
        </div>
      </div>
    `;
  }
}

function setupEventListeners() {
  document.querySelectorAll(".stage-tab-btn").forEach(tabBtn => {
    tabBtn.addEventListener("click", () => {
      document.querySelectorAll(".stage-tab-btn").forEach(b => b.classList.remove("active"));
      tabBtn.classList.add("active");
      currentTab = tabBtn.getAttribute("data-tab");
      renderScenarioStage();
    });
  });

  const copyInstallBtn = document.getElementById("copyInstallBtn");
  if (copyInstallBtn) {
    copyInstallBtn.addEventListener("click", () => {
      navigator.clipboard.writeText("go install github.com/bregaldahq/chaossql/cmd/chaossql@latest");
      copyInstallBtn.style.color = "var(--color-green)";
      setTimeout(() => copyInstallBtn.style.color = "", 1500);
    });
  }

  const copySdkBtn = document.getElementById("copySdkBtn");
  if (copySdkBtn) {
    copySdkBtn.addEventListener("click", () => {
      const code = document.querySelector("#sdk pre code")?.textContent;
      if (code) {
        navigator.clipboard.writeText(code);
        copySdkBtn.textContent = "Copied!";
        setTimeout(() => copySdkBtn.textContent = "Copy", 1500);
      }
    });
  }
}

function copySnippet(button) {
  const code = button.parentElement.querySelector("code")?.textContent;
  if (code) {
    navigator.clipboard.writeText(code);
    const orig = button.textContent;
    button.textContent = "Copied!";
    setTimeout(() => button.textContent = orig, 1500);
  }
}

function escapeHtml(str) {
  return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

document.addEventListener("DOMContentLoaded", initApp);
