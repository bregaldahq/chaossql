/**
 * ChaosSQL WebAssembly Engine Bridge
 * Provides typed communication with the Go WebAssembly runtime via dedicated Web Worker.
 */

export interface TraceOp {
  id: string;
  worker: number;
  tx: string;
  type: 'read' | 'write' | 'conflict' | 'exec';
  name: string;
  startUs: number;
  durationUs: number;
  vars: string;
  status: string;
}

export interface AdyaEdge {
  from: string;
  to: string;
  type: 'rw' | 'ww' | 'wr' | 'waits' | 'cascade' | string;
  item?: string;
  label?: string;
}

export interface AdyaNodeDef {
  id: string;
  label: string;
  rolePt: string;
  roleEn: string;
}

export interface AdyaEdgeDef {
  from: string;
  to: string;
  type: 'rw' | 'ww' | 'wr' | 'waits' | 'cascade' | string;
  item?: string;
  label?: string;
}

export interface WasmExecutionReport {
  totalOps: number;
  reducedOps: number;
  durationMs: number;
  anomalyType: string | null;
  adyaEdges?: AdyaEdge[];
  trace?: Array<{
    worker_id?: number;
    WorkerID?: number;
    type?: string;
    Type?: string;
    sql?: string;
    SQL?: string;
    duration_ns?: number;
  }>;
  reducedTrace?: Array<{
    worker_id?: number;
    WorkerID?: number;
    type?: string;
    Type?: string;
    sql?: string;
    SQL?: string;
  }>;
}

export interface PlaygroundConfig {
  yamlContent: string;
  workers: number;
  iterations: number;
  jitterMs: number;
  seed: number;
}

export interface ValidationResult {
  valid: boolean;
  error?: string;
  name?: string;
  operations?: number;
  invariants?: number;
}

export interface LogItem {
  id: string;
  type: 'info' | 'tick' | 'violation' | 'shrunk' | 'error' | 'success';
  message: string;
  timestamp: string;
  meta?: Record<string, unknown>;
}

// 20 realistic operations interleaved across 4 workers (W0, W1, W2, W3)
export const RAW_TRACE_OPS: TraceOp[] = [
  { id: 'op_0', worker: 0, tx: 'T1', type: 'read', name: 'SELECT balance FROM accounts WHERE id = 1 -> cur', startUs: 5, durationUs: 25, vars: 'cur = 1000', status: 'OK' },
  { id: 'op_1', worker: 1, tx: 'T2', type: 'read', name: 'SELECT balance FROM accounts WHERE id = 1 -> cur', startUs: 12, durationUs: 28, vars: 'cur = 1000', status: 'OK' },
  { id: 'op_2', worker: 2, tx: 'T3', type: 'read', name: 'SELECT count(*) FROM accounts -> cnt', startUs: 20, durationUs: 22, vars: 'cnt = 2', status: 'OK' },
  { id: 'op_3', worker: 3, tx: 'T4', type: 'read', name: "SELECT id, status FROM ledger WHERE id = 10", startUs: 32, durationUs: 24, vars: "status = 'ACTIVE'", status: 'OK' },
  { id: 'op_4', worker: 0, tx: 'T1', type: 'read', name: 'SELECT balance FROM accounts WHERE id = 2', startUs: 45, durationUs: 26, vars: 'balance = 500', status: 'OK' },
  { id: 'op_5', worker: 1, tx: 'T2', type: 'read', name: 'SELECT min(balance) FROM accounts', startUs: 58, durationUs: 25, vars: 'min = 500', status: 'OK' },
  { id: 'op_6', worker: 2, tx: 'T3', type: 'write', name: 'UPDATE audit_heartbeat SET last_ping = 6500', startUs: 70, durationUs: 30, vars: 'ping = 6500', status: 'COMMITTED' },
  { id: 'op_7', worker: 3, tx: 'T4', type: 'read', name: 'SELECT max(id) FROM audit_heartbeat', startUs: 85, durationUs: 20, vars: 'max = 1', status: 'OK' },
  { id: 'op_8', worker: 0, tx: 'T1', type: 'read', name: 'BEGIN; -- Worker 0 critical section', startUs: 98, durationUs: 15, vars: 'tx = T1', status: 'OK' },
  { id: 'op_9', worker: 1, tx: 'T2', type: 'read', name: 'BEGIN; -- Worker 1 critical section', startUs: 105, durationUs: 15, vars: 'tx = T2', status: 'OK' },
  { id: 'op_10', worker: 2, tx: 'T3', type: 'read', name: 'SELECT balance FROM accounts WHERE id = 1', startUs: 115, durationUs: 20, vars: 'cur = 1000', status: 'OK' },
  { id: 'op_11', worker: 3, tx: 'T4', type: 'write', name: "INSERT INTO trace_events VALUES ('SCHED_TICK')", startUs: 122, durationUs: 22, vars: 'tick = 42', status: 'COMMITTED' },
  { id: 'op_12', worker: 0, tx: 'T1', type: 'write', name: 'UPDATE accounts SET balance = 900 WHERE id = 1', startUs: 130, durationUs: 40, vars: 'balance = 900', status: 'COMMITTED' },
  { id: 'op_13', worker: 1, tx: 'T2', type: 'conflict', name: 'UPDATE accounts SET balance = 900 WHERE id = 1', startUs: 145, durationUs: 45, vars: 'balance = 900 [OVERWRITE COLLISION]', status: 'P4_LOST_UPDATE (Triggered at 184μs)' },
  { id: 'op_14', worker: 2, tx: 'T3', type: 'write', name: 'COMMIT; -- Worker 2 audit finished', startUs: 160, durationUs: 18, vars: "status = 'OK'", status: 'COMMITTED' },
  { id: 'op_15', worker: 3, tx: 'T4', type: 'read', name: 'SELECT sum(balance) AS total FROM accounts', startUs: 175, durationUs: 25, vars: 'total = 1400 (Expected 1500)', status: 'INVARIANT_FAIL' },
  { id: 'op_16', worker: 0, tx: 'T1', type: 'write', name: 'COMMIT; -- Worker 0 committed', startUs: 185, durationUs: 16, vars: "status = 'COMMITTED'", status: 'COMMITTED' },
  { id: 'op_17', worker: 1, tx: 'T2', type: 'write', name: 'COMMIT; -- Worker 1 committed (Overwritten)', startUs: 192, durationUs: 18, vars: "status = 'COMMITTED'", status: 'COMMITTED' },
  { id: 'op_18', worker: 2, tx: 'T3', type: 'write', name: "INSERT INTO anomaly_log VALUES ('P4', 184)", startUs: 205, durationUs: 22, vars: 'logged = true', status: 'COMMITTED' },
  { id: 'op_19', worker: 3, tx: 'T4', type: 'write', name: 'ROLLBACK; -- Fuzzer teardown context', startUs: 220, durationUs: 15, vars: 'done = true', status: 'ROLLED_BACK' },
];

// 1-Minimal reproduction isolated by Andreas Zeller's ddmin algorithm
export const SHRUNK_TRACE_OPS: TraceOp[] = [
  { id: 'op_s0', worker: 0, tx: 'T1', type: 'read', name: 'SELECT balance FROM accounts WHERE id = 1 -> cur', startUs: 10, durationUs: 35, vars: 'cur = 1000', status: 'OK' },
  { id: 'op_s1', worker: 1, tx: 'T2', type: 'read', name: 'SELECT balance FROM accounts WHERE id = 1 -> cur', startUs: 25, durationUs: 35, vars: 'cur = 1000', status: 'OK' },
  { id: 'op_s2', worker: 0, tx: 'T1', type: 'write', name: 'UPDATE accounts SET balance = {cur - 100} WHERE id = 1', startUs: 85, durationUs: 45, vars: 'balance = 900', status: 'COMMITTED' },
  { id: 'op_s3', worker: 1, tx: 'T2', type: 'conflict', name: 'UPDATE accounts SET balance = {cur - 100} WHERE id = 1', startUs: 105, durationUs: 50, vars: 'balance = 900 [OVERWRITE]', status: '1-MINIMAL COLLISION' },
];

export interface PresetDef {
  id: string;
  namePt: string;
  nameEn: string;
  anomaly: string;
  descriptionPt: string;
  descriptionEn: string;
  yaml: string;
  nodes: AdyaNodeDef[];
  edges: AdyaEdgeDef[];
  cycleFormula: string;
  cycleExplanationPt: string;
  cycleExplanationEn: string;
  reducedOps: number;
}

export const PLAYGROUND_PRESETS: PresetDef[] = [
  {
    id: 'banking',
    namePt: 'Perda de Atualização Bancária (P4)',
    nameEn: 'Banking Lost Update (P4)',
    anomaly: 'P4_LOST_UPDATE',
    descriptionPt: 'Dois saques concorrentes leem o mesmo saldo e sobrescrevem uma alteração.',
    descriptionEn: 'Two concurrent withdrawals read the same balance and overwrite updates.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Saque $100 (Worker 0)', roleEn: 'Withdraw $100 (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Saque $100 (Worker 1)', roleEn: 'Withdraw $100 (Worker 1)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'rw', item: 'accounts.balance', label: 'rw (anti-dep)' },
      { from: 'T2', to: 'T1', type: 'ww', item: 'accounts.balance', label: 'ww (overwrite)' },
    ],
    cycleFormula: 'T₁ ──(rw)──► T₂ ──(ww)──► T₁',
    cycleExplanationPt: 'A transação T₁ lê o saldo original antes de T₂ escrever (rw). T₂ comita sua alteração e T₁ sobrescreve cegamente o saldo baseado no valor obsoleto (ww). Nenhuma ordem serial estrita (T1 < T2 ou T2 < T1) preserva a conservação contábil.',
    cycleExplanationEn: 'Transaction T₁ reads initial balance before T₂ writes (rw). T₂ commits its update, and T₁ blind-overwrites the balance using stale data (ww). No strict serial order preserves the balance invariant.',
    reducedOps: 4,
    yaml: `version: "1.0"
name: "banking_lost_update"
database:
  driver: "sqlite"
  schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);"
  seed: "INSERT INTO accounts VALUES (1, 1000);"
invariants:
  - name: "total_balance"
    query: "SELECT sum(balance) AS total FROM accounts;"
    assert: "total == 1000"
operations:
  - name: "withdraw_100"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1"
        capture: "cur_bal"
      - sql: "UPDATE accounts SET balance = {cur_bal - 100} WHERE id = 1"`,
  },
  {
    id: 'inventory',
    namePt: 'Oversell de Estoque (G2-item)',
    nameEn: 'Inventory Oversell (G2-item)',
    anomaly: 'G2_ANTI_DEPENDENCY',
    descriptionPt: 'Compras paralelas verificam estoque positivo simultaneamente, resultando em estoque negativo.',
    descriptionEn: 'Parallel checkouts simultaneously observe positive stock, causing negative inventory.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Compra Item (Worker 0)', roleEn: 'Buy Item (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Compra Item (Worker 1)', roleEn: 'Buy Item (Worker 1)' },
      { id: 'T3', label: 'T₃', rolePt: 'Compra Item (Worker 2)', roleEn: 'Buy Item (Worker 2)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'rw', item: 'inventory.stock', label: 'rw (stock check)' },
      { from: 'T2', to: 'T3', type: 'rw', item: 'inventory.stock', label: 'rw (stock check)' },
      { from: 'T3', to: 'T1', type: 'rw', item: 'inventory.stock', label: 'rw (cycle closure)' },
    ],
    cycleFormula: 'T₁ ──(rw)──► T₂ ──(rw)──► T₃ ──(rw)──► T₁',
    cycleExplanationPt: 'Ciclo anti-dependência fechado de 3 nós (Adya G2): T₁, T₂ e T₃ observam o mesmo estoque positivo e decremetam em paralelo. Nenhuma transação observa a escrita das outras antes de comitar, violando serializabilidade e gerando estoque negativo.',
    cycleExplanationEn: 'Closed 3-node anti-dependency cycle (Adya G2): T₁, T₂, and T₃ concurrently observe positive stock and decrement in parallel. No transaction observes the other updates before committing, producing negative inventory.',
    reducedOps: 6,
    yaml: `version: "1.0"
name: "inventory_oversell"
database:
  driver: "sqlite"
  schema: "CREATE TABLE inventory (item_id INT PRIMARY KEY, stock INT NOT NULL);"
  seed: "INSERT INTO inventory VALUES (1, 10);"
invariants:
  - name: "no_negative_stock"
    query: "SELECT stock FROM inventory WHERE item_id = 1;"
    assert: "stock >= 0"
operations:
  - name: "buy_item"
    steps:
      - sql: "SELECT stock FROM inventory WHERE item_id = 1"
        capture: "cur"
      - sql: "UPDATE inventory SET stock = {cur - 1} WHERE item_id = 1 AND {cur > 0}"`,
  },
  {
    id: 'hospital',
    namePt: 'Plantão Médico (Write Skew / G-skew)',
    nameEn: 'Doctor On-Call (Write Skew / G-skew)',
    anomaly: 'G_SKEW',
    descriptionPt: 'Dois médicos solicitam dispensa checando se há ao menos 2 ativos; ambos saem ao mesmo tempo.',
    descriptionEn: 'Two doctors sign off simultaneously verifying count >= 2, leaving hospital unmanned.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Dispensa Dr. Alice (Worker 0)', roleEn: 'Sign-off Dr. Alice (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Dispensa Dr. Bob (Worker 1)', roleEn: 'Sign-off Dr. Bob (Worker 1)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'rw', item: 'doctors(on_duty>=2)', label: 'rw (anti-dep)' },
      { from: 'T2', to: 'T1', type: 'rw', item: 'doctors(on_duty>=2)', label: 'rw (anti-dep)' },
    ],
    cycleFormula: 'T₁ ──(rw)──► T₂ ──(rw)──► T₁',
    cycleExplanationPt: 'Anti-dependência mútua simétrica (Write Skew): T₁ e T₂ leem o mesmo predicado compartilhado (médicos ativos >= 2) e alteram linhas disjuntas. Ambas as transações são válidas individualmente, mas a intercalação desativa ambos os médicos.',
    cycleExplanationEn: 'Symmetric mutual anti-dependency (Write Skew): T₁ and T₂ both evaluate shared predicate (active >= 2) and mutate disjoint rows. Both are valid in isolation, but interleaving leaves zero active doctors.',
    reducedOps: 4,
    yaml: `version: "1.0"
name: "hospital_write_skew"
database:
  driver: "sqlite"
  schema: "CREATE TABLE doctors (id INT PRIMARY KEY, name TEXT NOT NULL, on_duty BOOLEAN NOT NULL);"
  seed: "INSERT INTO doctors VALUES (1, 'Dr. Alice', 1), (2, 'Dr. Bob', 1);"
invariants:
  - name: "at_least_one_doctor"
    query: "SELECT count(*) AS active FROM doctors WHERE on_duty = 1;"
    assert: "active >= 1"
operations:
  - name: "sign_off_alice"
    steps:
      - sql: "SELECT count(*) AS active FROM doctors WHERE on_duty = 1"
        capture: "act"
      - sql: "UPDATE doctors SET on_duty = 0 WHERE id = 1 AND {act >= 2}"
  - name: "sign_off_bob"
    steps:
      - sql: "SELECT count(*) AS active FROM doctors WHERE on_duty = 1"
        capture: "act"
      - sql: "UPDATE doctors SET on_duty = 0 WHERE id = 2 AND {act >= 2}"`,
  },
  {
    id: 'financial',
    namePt: 'Auditoria de Leitura Inconsistente (A5A Read Skew)',
    nameEn: 'Financial Read Skew (A5A)',
    anomaly: 'A5A_READ_SKEW',
    descriptionPt: 'Transações de transferência modificam contas enquanto a auditoria lê saldos desalinhados.',
    descriptionEn: 'Transfer transactions mutate accounts while audit reads misaligned balances.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Auditoria Balanço (Worker 0)', roleEn: 'Balance Audit (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Transferência $100 (Worker 1)', roleEn: 'Transfer $100 (Worker 1)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'rw', item: 'accounts[id=1]', label: 'rw (pre-debit)' },
      { from: 'T2', to: 'T1', type: 'wr', item: 'accounts[id=2]', label: 'wr (post-credit)' },
    ],
    cycleFormula: 'T₁ ──(rw)──► T₂ ──(wr)──► T₁',
    cycleExplanationPt: 'Ciclo rw → wr (Read Skew / A5A): A transação de auditoria T₁ lê a Conta 1 antes do débito de T₂ (rw), mas lê a Conta 2 após o crédito de T₂ (wr). A soma calculada difere do total conservado de R$ 1.000.',
    cycleExplanationEn: 'Cycle rw → wr (Read Skew / A5A): The audit transaction T₁ reads Account 1 before T₂ debit (rw), but reads Account 2 after T₂ credit (wr). The calculated sum breaks the conservation invariant.',
    reducedOps: 3,
    yaml: `version: "1.0"
name: "read_skew_financial_audit"
database:
  driver: "sqlite"
  schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);"
  seed: "INSERT INTO accounts VALUES (1, 500), (2, 500);"
invariants:
  - name: "sum_matches"
    query: "SELECT sum(balance) AS s FROM accounts;"
    assert: "s == 1000"
operations:
  - name: "transfer_1_to_2"
    steps:
      - sql: "UPDATE accounts SET balance = balance - 100 WHERE id = 1"
      - sql: "UPDATE accounts SET balance = balance + 100 WHERE id = 2"`,
  },
  {
    id: 'auction',
    namePt: 'Leilão com Sobrescrita Suja (G0 Dirty Write)',
    nameEn: 'Auction Dirty Write (G0)',
    anomaly: 'G0_DIRTY_WRITE',
    descriptionPt: 'Dois lances simultâneos intercalam escrita de valor e arrematante gerando registro híbrido.',
    descriptionEn: 'Two bids interleave high bid and high bidder writes resulting in hybrid state.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Lance Alice $100 (Worker 0)', roleEn: 'Alice Bid $100 (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Lance Bob $200 (Worker 1)', roleEn: 'Bob Bid $200 (Worker 1)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'ww', item: 'auctions.high_bid', label: 'ww (bid amount)' },
      { from: 'T2', to: 'T1', type: 'ww', item: 'auctions.high_bidder', label: 'ww (bidder name)' },
    ],
    cycleFormula: 'T₁ ──(ww)──► T₂ ──(ww)──► T₁',
    cycleExplanationPt: 'Ciclo de Escrita Suja (Dirty Write / G0): T₁ grava o valor do lance, mas antes de gravar o titular, T₂ sobrescreve o valor com $200. T₁ então comita seu nome \'Alice\', gravando Alice como vencedora por $200 sem que essa oferta tenha existido.',
    cycleExplanationEn: 'Dirty Write Cycle (G0): T₁ writes bid amount, but before writing bidder name, T₂ overwrites bid with $200. T₁ then commits \'Alice\', generating an invalid hybrid winner record.',
    reducedOps: 4,
    yaml: `version: "1.0"
name: "dirty_write_auction"
database:
  driver: "sqlite"
  schema: "CREATE TABLE auctions (item_id INT PRIMARY KEY, high_bidder TEXT, high_bid INT);"
  seed: "INSERT INTO auctions VALUES (1, 'nobody', 0);"
invariants:
  - name: "bidder_consistent"
    query: "SELECT high_bid FROM auctions WHERE item_id = 1;"
    assert: "high_bid >= 0"
operations:
  - name: "bid_alice"
    steps:
      - sql: "UPDATE auctions SET high_bid = 100 WHERE item_id = 1"
      - sql: "UPDATE auctions SET high_bidder = 'Alice' WHERE item_id = 1"
  - name: "bid_bob"
    steps:
      - sql: "UPDATE auctions SET high_bid = 200 WHERE item_id = 1"
      - sql: "UPDATE auctions SET high_bidder = 'Bob' WHERE item_id = 1"`,
  },
  {
    id: 'crypto',
    namePt: 'Arbitragem Cripto Circular (Ciclo Adya G2)',
    nameEn: 'Circular Crypto Arbitrage (Adya G2)',
    anomaly: 'G2_CYCLE',
    descriptionPt: 'Transações cruzadas atualizam cotações BTC e ETH baseadas em taxas anteriores mutuamente.',
    descriptionEn: 'Cross transactions update BTC and ETH quotes based on mutual previous rates.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Arbitragem BTC (Worker 0)', roleEn: 'BTC Arbitrage (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Arbitragem ETH (Worker 1)', roleEn: 'ETH Arbitrage (Worker 1)' },
      { id: 'T3', label: 'T₃', rolePt: 'Rebalanceamento USD (Worker 2)', roleEn: 'USD Rebalance (Worker 2)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'rw', item: 'crypto_pairs[ETH]', label: 'rw (quote dependency)' },
      { from: 'T2', to: 'T3', type: 'rw', item: 'crypto_pairs[BTC]', label: 'rw (cross rate)' },
      { from: 'T3', to: 'T1', type: 'rw', item: 'crypto_pairs[USD]', label: 'rw (cyclic feedback)' },
    ],
    cycleFormula: 'T₁ ──(rw)──► T₂ ──(rw)──► T₃ ──(rw)──► T₁',
    cycleExplanationPt: 'Ciclo causal de arbitragem triangular: Transações concorrentes recalculam taxas cambiais cruzadas baseando-se em cotações obsoletas mútuas, desestabilizando a paridade de mercado.',
    cycleExplanationEn: 'Triangular arbitrage causal cycle: Concurrent transactions recompute cross-exchange rates based on mutually stale quotes, breaking pricing parity.',
    reducedOps: 6,
    yaml: `version: "1.0"
name: "circular_info_crypto"
database:
  driver: "sqlite"
  schema: "CREATE TABLE crypto_pairs (pair TEXT PRIMARY KEY, rate INT);"
  seed: "INSERT INTO crypto_pairs VALUES ('BTC', 60000), ('ETH', 3000);"
invariants:
  - name: "rate_positive"
    query: "SELECT count(*) AS c FROM crypto_pairs WHERE rate <= 0;"
    assert: "c == 0"
operations:
  - name: "trade_btc"
    steps:
      - sql: "SELECT rate FROM crypto_pairs WHERE pair = 'ETH'"
        capture: "eth"
      - sql: "UPDATE crypto_pairs SET rate = {eth * 20} WHERE pair = 'BTC'"
  - name: "trade_eth"
    steps:
      - sql: "SELECT rate FROM crypto_pairs WHERE pair = 'BTC'"
        capture: "btc"
      - sql: "UPDATE crypto_pairs SET rate = {btc / 20} WHERE pair = 'ETH'"`,
  },
  {
    id: 'flash_crash',
    namePt: 'Flash Crash (A1 Dirty Read)',
    nameEn: 'Orderbook Flash Crash (A1 Dirty Read)',
    anomaly: 'A1_DIRTY_READ',
    descriptionPt: 'Um dump não comitado no book de ofertas é lido por uma liquidação de margem.',
    descriptionEn: 'An uncommitted price dump is read by an automated liquidation bot.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Dump de Ofertas (Worker 0)', roleEn: 'Order Dump (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Bot Liquidação (Worker 1)', roleEn: 'Liquidation Bot (Worker 1)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'ww', item: 'orderbook.price', label: 'ww (uncommitted)' },
      { from: 'T2', to: 'T1', type: 'wr', item: 'orderbook.price', label: 'wr (dirty read)' },
    ],
    cycleFormula: 'T₁ ──(ww)──► T₂ ──(wr)──► T₁',
    cycleExplanationPt: 'Leitura Suja (Dirty Read / A1): O bot de liquidação T₂ lê a cotação colapsada escrita por T₁ antes de seu commit. Mesmo com o rollback de T₁, o bot executou ordens irreversíveis sob um preço falso.',
    cycleExplanationEn: 'Dirty Read (A1): Liquidation bot T₂ reads an uncommitted price crash written by T₁. Even when T₁ rolls back, the bot executed irreversible orders under false market depth.',
    reducedOps: 2,
    yaml: `version: "1.0"
name: "dirty_read_flash_crash"
database:
  driver: "sqlite"
  schema: "CREATE TABLE orderbook (id INT PRIMARY KEY, price INT, status TEXT);"
  seed: "INSERT INTO orderbook VALUES (1, 100, 'FILLED');"
invariants:
  - name: "no_uncommitted_prices"
    query: "SELECT price FROM orderbook WHERE id = 1;"
    assert: "price == 100"
operations:
  - name: "flash_crash_dump"
    steps:
      - sql: "UPDATE orderbook SET price = 10 WHERE id = 1"
  - name: "liquidate"
    steps:
      - sql: "SELECT price FROM orderbook WHERE id = 1"
        capture: "p"`,
  },
  {
    id: 'ticket',
    namePt: 'Reserva de Assentos Duplicada (Phantom / A3)',
    nameEn: 'Double Booking Seats (Phantom / A3)',
    anomaly: 'A3_PHANTOM_PREDICATE',
    descriptionPt: 'Dois passageiros encontram assentos vagos sob o mesmo predicado e reservam ambos.',
    descriptionEn: 'Two customers concurrently observe available seats and double-book them.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Compra Passageiro 1 (Worker 0)', roleEn: 'Book Passenger 1 (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Compra Passageiro 2 (Worker 1)', roleEn: 'Book Passenger 2 (Worker 1)' },
      { id: 'T3', label: 'T₃', rolePt: 'Auditoria Lotação (Worker 2)', roleEn: 'Seat Audit (Worker 2)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'rw', item: 'seats(booked=0)', label: 'rw (predicate check)' },
      { from: 'T2', to: 'T3', type: 'rw', item: 'seats(booked=0)', label: 'rw (predicate check)' },
      { from: 'T3', to: 'T1', type: 'rw', item: 'seats(booked=0)', label: 'rw (phantom closure)' },
    ],
    cycleFormula: 'T₁ ──(rw)──► T₂ ──(rw)──► T₃ ──(rw)──► T₁',
    cycleExplanationPt: 'Leitura Fantasma (Phantom / A3): Clientes buscam assentos livres via predicado \'booked = 0\'. Ambos observam vaga livre e comitam a reserva, duplicando o assento e violando o invariant do voo.',
    cycleExplanationEn: 'Phantom Predicate (A3): Buyers search for free seats with \'booked = 0\'. Both observe availability and commit reservations, double-booking seats.',
    reducedOps: 6,
    yaml: `version: "1.0"
name: "ticket_anti_dependency"
database:
  driver: "sqlite"
  schema: "CREATE TABLE seats (id INT PRIMARY KEY, booked BOOLEAN NOT NULL);"
  seed: "INSERT INTO seats VALUES (1, 0), (2, 0);"
invariants:
  - name: "not_double_booked"
    query: "SELECT count(*) AS c FROM seats WHERE booked = 1;"
    assert: "c <= 2"
operations:
  - name: "book_seat_1"
    steps:
      - sql: "SELECT count(*) AS free FROM seats WHERE booked = 0"
        capture: "f"
      - sql: "UPDATE seats SET booked = 1 WHERE id = 1 AND {f > 0}"
  - name: "book_seat_2"
    steps:
      - sql: "SELECT count(*) AS free FROM seats WHERE booked = 0"
        capture: "f"
      - sql: "UPDATE seats SET booked = 1 WHERE id = 2 AND {f > 0}"`,
  },
  {
    id: 'deadlock',
    namePt: 'Ciclo de Deadlock Bidirecional',
    nameEn: 'Bidirectional Deadlock Cycle',
    anomaly: 'CYCLE_DEADLOCK',
    descriptionPt: 'Transações concorrentes adquirem travas em ordem invertida gerando ciclo no grafo de espera.',
    descriptionEn: 'Concurrent transactions acquire locks in reversed order generating wait-for cycle.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Trava R1 -> Aguarda R2 (Worker 0)', roleEn: 'Lock R1 -> Wait R2 (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Trava R2 -> Aguarda R1 (Worker 1)', roleEn: 'Lock R2 -> Wait R1 (Worker 1)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'waits', item: 'locks[id=2]', label: 'waits_for (lock 2)' },
      { from: 'T2', to: 'T1', type: 'waits', item: 'locks[id=1]', label: 'waits_for (lock 1)' },
    ],
    cycleFormula: 'T₁ ──(waits)──► T₂ ──(waits)──► T₁',
    cycleExplanationPt: 'Ciclo no Grafo de Espera (Deadlock): T₁ detém a trava da linha 1 e solicita a linha 2. Concorrentemente, T₂ detém a linha 2 e solicita a linha 1. O deadlock é permanente até que o motor aborte uma das transações.',
    cycleExplanationEn: 'Wait-For Graph Deadlock Cycle: T₁ holds lock 1 and requests lock 2. Concurrently, T₂ holds lock 2 and requests lock 1. Both goroutines block permanently until engine aborts.',
    reducedOps: 4,
    yaml: `version: "1.0"
name: "deadlock_cycle"
database:
  driver: "sqlite"
  schema: "CREATE TABLE locks (id INT PRIMARY KEY, v INT);"
  seed: "INSERT INTO locks VALUES (1, 10), (2, 20);"
invariants:
  - name: "sum_check"
    query: "SELECT sum(v) AS s FROM locks;"
    assert: "s == 30"
operations:
  - name: "tx_1_2"
    steps:
      - sql: "UPDATE locks SET v = v + 1 WHERE id = 1"
      - sql: "UPDATE locks SET v = v - 1 WHERE id = 2"
  - name: "tx_2_1"
    steps:
      - sql: "UPDATE locks SET v = v + 1 WHERE id = 2"
      - sql: "UPDATE locks SET v = v - 1 WHERE id = 1"`,
  },
  {
    id: 'fk',
    namePt: 'Cascata de Chave Estrangeira (FK Deadlock)',
    nameEn: 'Foreign Key Cascade Deadlock',
    anomaly: 'FK_CASCADE_VIOLATION',
    descriptionPt: 'Exclusão do pai colide com inserção no filho através de trigger implícito de consistência.',
    descriptionEn: 'Parent order deletion collides with item insertion through cascade triggers.',
    nodes: [
      { id: 'T1', label: 'T₁', rolePt: 'Exclusão Pedido Pai (Worker 0)', roleEn: 'Delete Parent Order (Worker 0)' },
      { id: 'T2', label: 'T₂', rolePt: 'Novo Item Filho (Worker 1)', roleEn: 'Insert Child Item (Worker 1)' },
    ],
    edges: [
      { from: 'T1', to: 'T2', type: 'cascade', item: 'parent_orders', label: 'cascade (lock)' },
      { from: 'T2', to: 'T1', type: 'waits', item: 'child_items', label: 'waits (ref integrity)' },
    ],
    cycleFormula: 'T₁ ──(cascade)──► T₂ ──(waits)──► T₁',
    cycleExplanationPt: 'Colisão de Cascata de Chave Estrangeira: A exclusão do pedido mestre dispara lock em cascata sobre itens filhos enquanto uma transação simultânea tenta adicionar um novo item associado ao pedido.',
    cycleExplanationEn: 'Foreign Key Cascade Conflict: Parent order deletion invokes cascading locks on child items while a concurrent transaction inserts a child item referencing the doomed parent.',
    reducedOps: 4,
    yaml: `version: "1.0"
name: "fk_cascade_deadlock"
database:
  driver: "sqlite"
  schema: "CREATE TABLE parent_orders (id INT PRIMARY KEY, total INT); CREATE TABLE child_items (id INT PRIMARY KEY, order_id INT, price INT);"
  seed: "INSERT INTO parent_orders VALUES (1, 100); INSERT INTO child_items VALUES (1, 1, 50), (2, 1, 50);"
invariants:
  - name: "orders_exist"
    query: "SELECT count(*) AS c FROM parent_orders;"
    assert: "c >= 0"
operations:
  - name: "add_item"
    steps:
      - sql: "UPDATE parent_orders SET total = total + 25 WHERE id = 1"
      - sql: "INSERT INTO child_items VALUES (3, 1, 25)"
  - name: "delete_order"
    steps:
      - sql: "DELETE FROM child_items WHERE order_id = 1"
      - sql: "DELETE FROM parent_orders WHERE id = 1"`,
  },
];

/**
 * Generates realistic interleaved trace events and execution report for a scenario.
 */
export function buildExecutionReport(
  preset: PresetDef,
  workers: number,
  iterations: number,
  jitterMs: number,
  seed: number
): WasmExecutionReport {
  const totalOps = workers * iterations;
  const durationMs = Math.round(26 + jitterMs * 1.6 + totalOps * 0.12 + (seed % 12));

  // Generate trace ops mapped to workers
  const trace: WasmExecutionReport['trace'] = [];
  const baseOps = [
    { type: 'SELECT', sql: `SELECT * FROM state WHERE id = 1`, dur: 22000 },
    { type: 'SELECT', sql: `SELECT count(*) FROM state`, dur: 24000 },
    { type: 'UPDATE', sql: `UPDATE state SET v = v + 1 WHERE id = 1`, dur: 38000 },
    { type: 'UPDATE', sql: `UPDATE state SET v = v - 1 WHERE id = 1 [COLLISION]`, dur: 45000 },
    { type: 'INSERT', sql: `INSERT INTO log VALUES ('EVENT')`, dur: 30000 },
    { type: 'COMMIT', sql: `COMMIT`, dur: 18000 },
  ];

  for (let w = 0; w < workers; w++) {
    const opsCount = Math.min(iterations, 6);
    for (let i = 0; i < opsCount; i++) {
      const template = baseOps[(w + i + seed) % baseOps.length];
      const isCollision = (w === 1 && i === opsCount - 2) || (w === 0 && i === opsCount - 1);
      trace.push({
        worker_id: w,
        type: isCollision ? 'UPDATE' : template.type,
        sql: isCollision ? `${template.sql} [COLLISION / ${preset.anomaly}]` : template.sql,
        duration_ns: template.dur + (seed * 100) + (w * 1500),
      });
    }
  }

  const reducedTrace = trace.filter((t) => (t.sql || '').includes('COLLISION') || t.type === 'UPDATE').slice(0, preset.reducedOps);

  return {
    totalOps,
    reducedOps: preset.reducedOps,
    durationMs,
    anomalyType: preset.anomaly,
    adyaEdges: preset.edges,
    trace,
    reducedTrace: reducedTrace.length > 0 ? reducedTrace : trace.slice(0, preset.reducedOps),
  };
}

/**
 * ChaosSqlWasmBridge: Web Worker controller for ChaosSQL WebAssembly Engine
 */
export class ChaosSqlWasmBridge {
  private worker: Worker | null = null;
  private isReady = false;
  private onReadyCallbacks: Array<() => void> = [];
  private onReportCallback: ((report: WasmExecutionReport) => void) | null = null;
  private onProgressCallback: ((progress: any) => void) | null = null;
  private onValidationCallback: ((result: ValidationResult) => void) | null = null;
  private onErrorCallback: ((error: string) => void) | null = null;

  constructor() {
    this.init();
  }

  public init(): void {
    if (typeof window === 'undefined' || typeof Worker === 'undefined') {
      return;
    }
    if (this.worker) {
      return;
    }

    try {
      this.worker = new Worker('/wasm/wasm-worker.js');
      this.worker.onmessage = (e: MessageEvent) => {
        this.handleMessage(e.data);
      };
      this.worker.onerror = (err: ErrorEvent) => {
        console.warn('ChaosSQL Worker encountered error:', err.message);
        if (this.onErrorCallback) {
          this.onErrorCallback(err.message || 'Worker runtime error');
        }
      };

      // Request initialization of the Go WASM runtime
      this.worker.postMessage({
        action: 'INIT',
        wasmUrl: '/wasm/chaossql.wasm',
      });
    } catch (err) {
      console.warn('Could not spawn ChaosSQL Web Worker:', err);
    }
  }

  private handleMessage(data: any): void {
    if (!data) return;

    switch (data.type) {
      case 'READY':
        this.isReady = true;
        this.onReadyCallbacks.forEach((cb) => cb());
        this.onReadyCallbacks = [];
        break;

      case 'VALIDATION_RESULT':
        if (this.onValidationCallback) {
          this.onValidationCallback({
            valid: Boolean(data.valid),
            error: data.error,
            name: data.name,
            operations: data.operations,
            invariants: data.invariants,
          });
        }
        break;

      case 'CYCLE_DETECTED':
      case 'ANOMALY':
      case 'PROGRESS':
        if (this.onProgressCallback) {
          this.onProgressCallback(data);
        }
        break;

      case 'COMPLETE':
      case 'REPORT':
      case 'DONE': {
        let rep: WasmExecutionReport = data;
        if (data.report) {
          try {
            rep = typeof data.report === 'string' ? JSON.parse(data.report) : data.report;
          } catch (_) {
            rep = data.report;
          }
        }
        if (this.onReportCallback) {
          this.onReportCallback(rep);
        }
        break;
      }

      case 'ERROR':
        if (this.onErrorCallback) {
          this.onErrorCallback(data.error || 'Unknown WASM runtime error');
        }
        break;
    }
  }

  public ready(): Promise<boolean> {
    if (this.isReady) return Promise.resolve(true);
    return new Promise((resolve) => {
      this.onReadyCallbacks.push(() => resolve(true));
      setTimeout(() => resolve(this.isReady), 5000);
    });
  }

  public isEngineReady(): boolean {
    return this.isReady;
  }

  public validateYaml(
    yamlContent: string,
    onResult: (result: ValidationResult) => void
  ): void {
    this.onValidationCallback = onResult;
    if (!this.worker || !this.isReady) {
      const hasDatabase = yamlContent.includes('database:');
      const hasOps = yamlContent.includes('operations:');
      const valid = hasDatabase && hasOps;
      onResult({
        valid,
        error: valid ? undefined : 'YAML deve conter os blocos "database:" e "operations:".',
        name: 'heuristic-validation',
      });
      return;
    }

    this.worker.postMessage({
      action: 'VALIDATE',
      yamlContent,
    });
  }

  public runScenario(
    config: PlaygroundConfig,
    onProgress: (progress: any) => void,
    onReport: (report: WasmExecutionReport) => void,
    onError: (error: string) => void
  ): void {
    this.onProgressCallback = onProgress;
    this.onReportCallback = onReport;
    this.onErrorCallback = onError;

    if (!this.worker || !this.isReady) {
      onError('Motor WASM ainda inicializando ou indisponível.');
      return;
    }

    this.worker.postMessage({
      action: 'RUN',
      config: {
        yamlContent: config.yamlContent,
        workers: config.workers,
        iterations: config.iterations,
        jitterMs: config.jitterMs,
        seed: config.seed,
      },
    });
  }

  public cancel(): void {
    if (this.worker) {
      this.worker.postMessage({ action: 'CANCEL' });
    }
  }

  public terminate(): void {
    if (this.worker) {
      this.worker.terminate();
      this.worker = null;
      this.isReady = false;
    }
  }
}

let singletonInstance: ChaosSqlWasmBridge | null = null;
export function getWasmBridge(): ChaosSqlWasmBridge {
  if (!singletonInstance) {
    singletonInstance = new ChaosSqlWasmBridge();
  }
  return singletonInstance;
}
