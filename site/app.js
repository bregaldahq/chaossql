// ChaosSQL — Studio Bregalda Interactive Controller (v1.2.0)
// High-fidelity Multi-View Portal, Hash Router, Docs Hub, Scenarios with Fixes, & Trace Visualizer

// ============================================================================
// 1. SCENARIOS DATABASE (10 Flagship Scenarios with Invariant Reduction & Fixes)
// ============================================================================
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
    },
    fix: {
      strategy: "Update Atômico com Predicado de Guarda ou Bloqueio Pessimista",
      sql: `-- Opção 1: Update Atômico com Aritmética no Motor (Recomendado)
UPDATE accounts 
SET balance = balance - 100 
WHERE id = 1 AND balance >= 100;

-- Opção 2: Bloqueio Pessimista dentro de Transação Explícita
BEGIN;
SELECT balance FROM accounts WHERE id = 1 FOR UPDATE;
-- Aplicação valida balance >= 100 antes de submeter escrita:
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
COMMIT;`,
      explanation: "A anomalia ocorre pelo padrão vulnerável Read-Modify-Write sob READ COMMITTED: a aplicação lê o valor em um SELECT e computa a subtração na memória do servidor web. Se duas goroutines lerem balance=1000 ao mesmo tempo, ambas calculam 900 e o último UPDATE sobrescreve o primeiro. A mitigação transfere o cálculo para o motor de banco em uma única instrução atômica (UPDATE accounts SET balance = balance - 100 WHERE id = 1 AND balance >= 100), serializando a alteração sob o bloqueio exclusivo de linha.",
      engines: ["PostgreSQL", "SQLite (WAL)", "MySQL (InnoDB)"]
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
    },
    fix: {
      strategy: "Guarded Decrement Atômico com Verificação de Linhas Afetadas",
      sql: `-- Decremento Atômico Protegido no Banco de Dados
UPDATE products 
SET stock = stock - 1 
WHERE id = 1 AND stock >= 1;

-- No código de aplicação (Go / Node):
-- Verificar se rows_affected == 1.
-- Se rows_affected == 0, retornar ErrProdutoEsgotado sem prosseguir para pagamento.`,
      explanation: "A condição de guarda {avail > 0} avaliada na memória da aplicação é obsoleta (stale) no momento em que a escrita chega ao disco. Ao mover a verificação diretamente para o WHERE da instrução UPDATE (AND stock >= 1), o motor de banco garante a avaliação atômica do predicado e o bloqueio de escrita sobre a linha. Se múltiplos clientes tentarem comprar a última unidade simultaneamente, exatamente um decrementará o estoque para 0 e os demais receberão 0 linhas afetadas.",
      engines: ["PostgreSQL", "SQLite", "MySQL (InnoDB)"]
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
    on_duty BOOLEAN NOT NULL
);

-- Seed
INSERT INTO doctors VALUES (1, 'Dr. Alice', true), (2, 'Dr. Bob', true);`,
    chaos: `version: "1.0"
name: "hospital_write_skew"
operations:
  - name: "sign_off_alice"
    steps:
      - "SELECT count(*) AS active FROM doctors WHERE on_duty = true -> act"
      - "UPDATE doctors SET on_duty = false WHERE id = 1 AND {act >= 2}"
  - name: "sign_off_bob"
    steps:
      - "SELECT count(*) AS active FROM doctors WHERE on_duty = true -> act"
      - "UPDATE doctors SET on_duty = false WHERE id = 2 AND {act >= 2}"
invariants:
  - name: "at_least_one_doctor_on_duty"
    query: "SELECT count(*) AS active FROM doctors WHERE on_duty = true;"
    assert: "active >= 1"`,
    reduction: {
      originalOps: 40,
      minimalOps: 2,
      reductionPct: "95.0%",
      elapsed: "82ms",
      cycle: "T1 ──(rw)──► T2 ──(rw)──► T1",
      explanation: "Under Snapshot Isolation, T1 and T2 read the same snapshot (2 active doctors). T1 updates Alice, T2 updates Bob. Because their write sets are disjoint, standard SI permits both commits, leaving 0 doctors on duty."
    },
    fix: {
      strategy: "Elevação para Serializable Snapshot Isolation (SSI) ou Bloqueio de Conflito",
      sql: `-- Solução A: Elevação do Nível de Isolamento para SERIALIZABLE (SSI)
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
BEGIN;
SELECT count(*) FROM doctors WHERE on_duty = true;
-- Se count >= 2:
UPDATE doctors SET on_duty = false WHERE id = 1;
COMMIT; -- O motor PostgreSQL detectará o ciclo rw e abortará T2 com código 40001!

-- Solução B: Bloqueio Pessimista Compartilhado em Linha Pai de Plantão
BEGIN;
SELECT shift_id FROM duty_shifts WHERE shift_id = 1 FOR UPDATE;
SELECT count(*) FROM doctors WHERE on_duty = true;
UPDATE doctors SET on_duty = false WHERE id = 1;
COMMIT;`,
      explanation: "Write Skew (A5B) é o exemplo clássico de falha sob Snapshot Isolation (SI): como T1 altera a linha 1 e T2 altera a linha 2, os conjuntos de escrita não colidem (sem conflito ww), mas geram um ciclo de anti-dependência de leitura-escrita (rw). A solução canônica é configurar SERIALIZABLE (onde motores modernos com SSI rastreiam dependências rw e abortam transações concorrentes com serialization_failure) ou serializar os acessos através de um lock FOR UPDATE em uma linha mestre de escala.",
      engines: ["PostgreSQL (SSI)", "CockroachDB", "MySQL (SERIALIZABLE)"]
    }
  },
  {
    id: "financial",
    name: "Financial Read Skew",
    code: "A5A",
    summary: "An audit query observes a partially applied transfer, detecting an artificial discrepancy in total wealth.",
    schema: `-- Schema
CREATE TABLE accounts (
    id INT PRIMARY KEY,
    balance INT NOT NULL
);

-- Seed
INSERT INTO accounts VALUES (1, 1000), (2, 1000);`,
    chaos: `version: "1.0"
name: "financial_read_skew"
operations:
  - name: "transfer_100"
    steps:
      - "UPDATE accounts SET balance = balance - 100 WHERE id = 1"
      - "UPDATE accounts SET balance = balance + 100 WHERE id = 2"
  - name: "audit_total"
    steps:
      - "SELECT balance FROM accounts WHERE id = 1 -> b1"
      - "SELECT balance FROM accounts WHERE id = 2 -> b2"
invariants:
  - name: "total_wealth_preserved"
    query: "SELECT sum(balance) AS total FROM accounts;"
    assert: "total == 2000"`,
    reduction: {
      originalOps: 25,
      minimalOps: 3,
      reductionPct: "88.0%",
      elapsed: "61ms",
      cycle: "T1 ──(rw)──► T2 ──(wr)──► T1",
      explanation: "Under READ COMMITTED, each SELECT inside the audit transaction takes a separate snapshot. Audit reads account 1 before transfer (-$100), but reads account 2 after transfer (+$100), seeing $1000 + $1100 = $2100."
    },
    fix: {
      strategy: "Isolamento REPEATABLE READ / SNAPSHOT para Transações de Auditoria",
      sql: `-- Transação de Auditoria em REPEATABLE READ
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;

BEGIN;
-- Todas as consultas subsequentes leem a mesma fotografia temporal do banco:
SELECT balance AS bal1 FROM accounts WHERE id = 1;
-- Mesmo que a transferência ocorra e comite aqui no meio...
SELECT balance AS bal2 FROM accounts WHERE id = 2;
-- bal1 + bal2 sempre somará exatamente 2000!
COMMIT;`,
      explanation: "Sob READ COMMITTED, cada instrução SELECT cria uma nova fotografia temporal (snapshot). Se uma transação de transferência comitar entre a leitura da conta 1 e da conta 2, o relatório financeiro observará o estado de contas em momentos díspares do tempo. Ao elevar a transação de leitura para REPEATABLE READ (ou SNAPSHOT ISOLATION), o banco congela um único Point-in-Time Snapshot imutável no primeiro SELECT, garantindo consistência estrita de leitura sem necessidade de bloqueios.",
      engines: ["PostgreSQL", "MySQL (InnoDB)", "SQLite (WAL)"]
    }
  },
  {
    id: "auction",
    name: "Auction Dirty Write",
    code: "G0",
    summary: "Two concurrent bidders overwrite separate columns of the same auction row without strict 2PL locking.",
    schema: `-- Schema
CREATE TABLE auctions (
    id INT PRIMARY KEY,
    item TEXT NOT NULL,
    highest_bid INT NOT NULL,
    winner TEXT NOT NULL
);

-- Seed
INSERT INTO auctions VALUES (1, 'Rare Stamp', 100, 'Original');`,
    chaos: `version: "1.0"
name: "auction_dirty_write"
operations:
  - name: "bid_alice"
    steps:
      - "UPDATE auctions SET highest_bid = 150 WHERE id = 1"
      - "UPDATE auctions SET winner = 'Alice' WHERE id = 1"
  - name: "bid_bob"
    steps:
      - "UPDATE auctions SET highest_bid = 200 WHERE id = 1"
      - "UPDATE auctions SET winner = 'Bob' WHERE id = 1"
invariants:
  - name: "bid_winner_consistency"
    query: "SELECT highest_bid, winner FROM auctions WHERE id = 1;"
    assert: "(highest_bid == 150 AND winner == 'Alice') OR (highest_bid == 200 AND winner == 'Bob')"`,
    reduction: {
      originalOps: 15,
      minimalOps: 2,
      reductionPct: "86.7%",
      elapsed: "49ms",
      cycle: "T1 ──(ww)──► T2 ──(ww)──► T1",
      explanation: "Without atomic row locking or single-statement updates, Alice writes bid 150, Bob writes bid 200, but Alice's second step writes winner='Alice', creating an inconsistent state: $200 bid won by Alice."
    },
    fix: {
      strategy: "Atualização Atômica Multi-Coluna com Predicado Monotônico",
      sql: `-- Update Atômico em Instrução Única com Cláusula Monotônica
BEGIN;
UPDATE auctions 
SET highest_bid = 200, 
    winner = 'Bob' 
WHERE id = 1 AND highest_bid < 200;

-- Se rows_affected == 0, o lance já foi superado concorrentemente;
-- Fazer ROLLBACK e notificar o usuário.
COMMIT;`,
      explanation: "Dirty Writes (G0) ocorrem quando escritas não comitadas de transações simultâneas se entrelaçam sobre a mesma linha de dados. A correção definitiva consiste em unificar as atualizações de valor e titular em uma única cláusula UPDATE atômica no banco, condicionada estritamente a um lance superior ao atual (WHERE id = 1 AND highest_bid < novo_lance). Isso assegura retenção de locks 2PL estritos e garante monotonicidade sem risco de estados corrompidos híbridos.",
      engines: ["PostgreSQL", "SQLite", "MySQL"]
    }
  },
  {
    id: "crypto",
    name: "Crypto Arbitrage Circular Info",
    code: "G1c",
    summary: "Cross-exchange arbitrage loop observes cyclic dependency edges, leading to phantom profit execution.",
    schema: `-- Schema
CREATE TABLE orderbooks (
    id INT PRIMARY KEY,
    exchange TEXT NOT NULL,
    pair TEXT NOT NULL,
    ask_price INT NOT NULL,
    version INT NOT NULL
);

-- Seed
INSERT INTO orderbooks VALUES (1, 'Binance', 'BTC/USDT', 64000, 1), (2, 'Coinbase', 'BTC/USDT', 64200, 1);`,
    chaos: `version: "1.0"
name: "crypto_arbitrage_circular"
operations:
  - name: "arb_bot_1"
    steps:
      - "SELECT ask_price FROM orderbooks WHERE id = 1 -> p1"
      - "UPDATE orderbooks SET ask_price = {p1 + 50}, version = version + 1 WHERE id = 2"
  - name: "arb_bot_2"
    steps:
      - "SELECT ask_price FROM orderbooks WHERE id = 2 -> p2"
      - "UPDATE orderbooks SET ask_price = {p2 - 50}, version = version + 1 WHERE id = 1"
invariants:
  - name: "monotonic_versioning"
    query: "SELECT sum(version) AS total_ver FROM orderbooks;"
    assert: "total_ver <= 100"`,
    reduction: {
      originalOps: 30,
      minimalOps: 2,
      reductionPct: "93.3%",
      elapsed: "71ms",
      cycle: "T1 ──(wr)──► T2 ──(wr)──► T1",
      explanation: "Bot 1 reads Binance price and modifies Coinbase. Concurrently, Bot 2 reads Coinbase price and modifies Binance. The circular information flow creates a G1c cycle violating causal order."
    },
    fix: {
      strategy: "Optimistic Concurrency Control (OCC) com Versionamento ou Oráculo Serializado",
      sql: `-- Controle Otimista de Concorrência (OCC) com Versionamento Monotônico
BEGIN;
SELECT id, ask_price, version FROM orderbooks WHERE id IN (1, 2) FOR UPDATE;

-- Validar spread de lucro e atualizar com incremento estrito:
UPDATE orderbooks 
SET ask_price = :novo_preco, version = version + 1 
WHERE id = :exchange_id AND version = :versao_esperada;

COMMIT;`,
      explanation: "A anomalia G1c demonstra fluxos cíclicos de leitura-escrita entre pares distribuídos. Para assegurar linearizabilidade global e evitar cálculos baseados em spreads defasados, emprega-se OCC com número de versão atômico ou bloqueio explícito FOR UPDATE em ordem canônica das chaves primárias dos livros antes da validação e execução da ordem de arbitragem.",
      engines: ["PostgreSQL", "MySQL (InnoDB)", "CockroachDB"]
    }
  },
  {
    id: "flashcrash",
    name: "Flash Crash Dirty Read",
    code: "G1a",
    summary: "Liquidation bot reads uncommitted margin drop from a transaction that subsequently rolls back.",
    schema: `-- Schema
CREATE TABLE margin_accounts (
    id INT PRIMARY KEY,
    trader TEXT NOT NULL,
    collateral INT NOT NULL,
    status TEXT NOT NULL
);

-- Seed
INSERT INTO margin_accounts VALUES (1, 'Whale_01', 50000, 'ACTIVE');`,
    chaos: `version: "1.0"
name: "flash_crash_dirty_read"
operations:
  - name: "market_order_rollback"
    steps:
      - "UPDATE margin_accounts SET collateral = 5000 WHERE id = 1"
      - "SELECT 1/0" # Forced division by zero fault to trigger ROLLBACK
  - name: "liquidation_bot"
    steps:
      - "SELECT collateral FROM margin_accounts WHERE id = 1 -> col"
      - "UPDATE margin_accounts SET status = 'LIQUIDATED' WHERE id = 1 AND {col < 10000}"
invariants:
  - name: "no_spurious_liquidation"
    query: "SELECT collateral, status FROM margin_accounts WHERE id = 1;"
    assert: "collateral == 50000 AND status == 'ACTIVE'"`,
    reduction: {
      originalOps: 10,
      minimalOps: 2,
      reductionPct: "80.0%",
      elapsed: "38ms",
      cycle: "w1(collateral=5000) ... r2(collateral) ... a1",
      explanation: "T1 updates collateral to $5000 and then aborts. Under READ UNCOMMITTED, T2 reads the uncommitted $5000 drop and triggers premature liquidation of a solvent trader."
    },
    fix: {
      strategy: "Isolamento Mínimo READ COMMITTED no Pool de Conexões",
      sql: `-- Assegurar Nível de Isolamento Mínimo READ COMMITTED no Banco
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- Rotina de Liquidação com Validação Transacional
BEGIN;
SELECT collateral, status FROM margin_accounts WHERE id = 1 FOR UPDATE;
-- Apenas dados comitados são lidos! Transações abortadas nunca vazam.
UPDATE margin_accounts SET status = 'LIQUIDATED' WHERE id = 1 AND collateral < 10000;
COMMIT;`,
      explanation: "Dirty Reads (G1a / Aborted Reads) ocorrem quando uma transação lê modificações feitas por outra transação que posteriormente sofre abort (ROLLBACK). Em finanças e criptoativos, isso acarreta liquidações catastróficas espúrias. A mitigação arquitetural exige que o pool de conexões configure como piso inviolável o isolamento READ COMMITTED, garantindo que leituras acessem apenas tuplas comitadas de forma duradoura no write-ahead log.",
      engines: ["PostgreSQL (Default)", "SQLite (WAL)", "MySQL (Default)"]
    }
  },
  {
    id: "ticket",
    name: "Ticket Anti-Dependency Cycle",
    code: "G2",
    summary: "Concert ticket reservations generate phantom conflicts across seat range predicates under Snapshot Isolation.",
    schema: `-- Schema
CREATE TABLE seats (
    id INT PRIMARY KEY,
    section TEXT NOT NULL,
    seat_no INT NOT NULL,
    reserved_by TEXT
);

-- Seed
INSERT INTO seats VALUES (1, 'VIP', 101, NULL), (2, 'VIP', 102, NULL);`,
    chaos: `version: "1.0"
name: "ticket_anti_dependency_g2"
operations:
  - name: "book_adjacent_left"
    steps:
      - "SELECT count(*) AS reserved FROM seats WHERE section = 'VIP' AND reserved_by IS NOT NULL -> c"
      - "UPDATE seats SET reserved_by = 'Fan_A' WHERE id = 1 AND {c == 0}"
  - name: "book_adjacent_right"
    steps:
      - "SELECT count(*) AS reserved FROM seats WHERE section = 'VIP' AND reserved_by IS NOT NULL -> c"
      - "UPDATE seats SET reserved_by = 'Fan_B' WHERE id = 2 AND {c == 0}"
invariants:
  - name: "max_one_seat_rule"
    query: "SELECT count(*) AS reserved FROM seats WHERE section = 'VIP' AND reserved_by IS NOT NULL;"
    assert: "reserved <= 1"`,
    reduction: {
      originalOps: 20,
      minimalOps: 2,
      reductionPct: "90.0%",
      elapsed: "65ms",
      cycle: "T1 ──(rw)──► T2 ──(rw)──► T1",
      explanation: "Predicate anti-dependency cycle: T1 checks if any VIP seat is booked (finds 0) and books seat 1. T2 simultaneously checks if any VIP seat is booked (finds 0) and books seat 2. Both commit, violating VIP exclusivity."
    },
    fix: {
      strategy: "Restrição de Unicidade Estrutural e Locks de Predicado Serializáveis",
      sql: `-- 1. Restrição de Chave Única Composta para Garantir Atomicidade de Índice
ALTER TABLE seats 
ADD CONSTRAINT uq_section_reservation UNIQUE (section, seat_no);

-- 2. Predicate Lock com SELECT FOR UPDATE ou SSI
BEGIN;
SELECT id FROM seats 
WHERE section = 'VIP' AND reserved_by IS NULL 
LIMIT 1 FOR UPDATE;

UPDATE seats SET reserved_by = 'Fan_A' WHERE id = :seat_id;
COMMIT;`,
      explanation: "Ciclos G2 ocorrem quando consultas de agregação ou intervalos checam a existência de registros através de predicados (COUNT(*) WHERE ...). Sob Snapshot Isolation, as transações não bloqueiam predicados ausentes. A adição de uma restrição UNIQUE a nível de schema converte a corrida silenciosa em um erro determinístico de integridade de índice, enquanto SELECT ... FOR UPDATE ou SSI serializam a reserva de forma infalível.",
      engines: ["PostgreSQL", "MySQL", "SQLite"]
    }
  },
  {
    id: "deadlock",
    name: "Deadlock Cycle & Recovery",
    code: "G-DL",
    summary: "Two transfer workers lock accounts in reverse order, inducing a cyclic lock-wait graph (WFG) deadlock.",
    schema: `-- Schema
CREATE TABLE accounts (
    id INT PRIMARY KEY,
    balance INT NOT NULL
);

-- Seed
INSERT INTO accounts VALUES (1, 500), (2, 500);`,
    chaos: `version: "1.0"
name: "deadlock_cycle_recovery"
operations:
  - name: "transfer_1_to_2"
    steps:
      - "SELECT balance FROM accounts WHERE id = 1 -> b1"
      - "UPDATE accounts SET balance = balance - 100 WHERE id = 1"
      - "UPDATE accounts SET balance = balance + 100 WHERE id = 2"
  - name: "transfer_2_to_1"
    steps:
      - "SELECT balance FROM accounts WHERE id = 2 -> b2"
      - "UPDATE accounts SET balance = balance - 100 WHERE id = 2"
      - "UPDATE accounts SET balance = balance + 100 WHERE id = 1"
invariants:
  - name: "total_balance_exact"
    query: "SELECT sum(balance) AS total FROM accounts;"
    assert: "total == 1000"`,
    reduction: {
      originalOps: 15,
      minimalOps: 2,
      reductionPct: "86.7%",
      elapsed: "54ms",
      cycle: "T1 ──(waits)──► T2 ──(waits)──► T1",
      explanation: "T1 locks Account 1 and requests lock on Account 2. Concurrently, T2 locks Account 2 and requests lock on Account 1. Mutual wait forms a cycle in the Wait-For Graph (WFG)."
    },
    fix: {
      strategy: "Ordenação Canônica Determinística de Bloqueios (Lock Ordering)",
      sql: `-- Ordenação Padronizada de Locks pelo Menor ID
-- Na aplicação:
-- first_id = MIN(conta_a, conta_b)
-- second_id = MAX(conta_a, conta_b)

BEGIN;
-- Bloquear sempre os recursos na mesma ordem canônica crescente:
SELECT id, balance FROM accounts WHERE id = :first_id FOR UPDATE;
SELECT id, balance FROM accounts WHERE id = :second_id FOR UPDATE;

UPDATE accounts SET balance = balance - 100 WHERE id = :origem;
UPDATE accounts SET balance = balance + 100 WHERE id = :destino;
COMMIT;`,
      explanation: "Deadlocks ocorrem quando diferentes transações disputam os mesmos registros em ordens inversas (T1: 1 depois 2; T2: 2 depois 1). A mitigação canônica consiste em estabelecer um protocolo estrito de ordenação hierárquica de recursos antes de adquirir qualquer bloqueio (ex: travar sempre pelo menor ID primeiro). Com ordenação canônica garantida na camada de repositório, ciclos no Wait-For Graph tornam-se matematicamente impossíveis de se formarem.",
      engines: ["PostgreSQL (Code 40P01)", "MySQL (Error 1213)", "SQLite"]
    }
  },
  {
    id: "fk_cascade",
    name: "Foreign Key Cascade Deadlock",
    code: "G-DL",
    summary: "Concurrent inserts into child items and cascaded parent order deletes invert the lock hierarchy.",
    schema: `-- Schema
CREATE TABLE parent_orders (
    id INT PRIMARY KEY,
    status TEXT NOT NULL,
    total_cents INT NOT NULL
);

CREATE TABLE child_items (
    id INT PRIMARY KEY,
    order_id INT NOT NULL REFERENCES parent_orders(id) ON DELETE CASCADE,
    sku TEXT NOT NULL,
    price_cents INT NOT NULL
);

-- Seed
INSERT INTO parent_orders VALUES (1, 'OPEN', 5000);
INSERT INTO child_items VALUES (1, 1, 'ITEM-A', 2500), (2, 1, 'ITEM-B', 2500);`,
    chaos: `version: "1.0"
name: "foreign_key_cascade_deadlock"
operations:
  - name: "add_order_item"
    steps:
      - "INSERT INTO child_items VALUES ($monotonic_counter(10, 1), 1, 'ITEM-NEW', 1500)"
      - "UPDATE parent_orders SET total_cents = total_cents + 1500 WHERE id = 1"
  - name: "cancel_order_cascade"
    steps:
      - "UPDATE parent_orders SET status = 'CANCELLED' WHERE id = 1"
      - "DELETE FROM parent_orders WHERE id = 1"
invariants:
  - name: "referential_integrity"
    query: "SELECT count(*) AS orphan_items FROM child_items LEFT JOIN parent_orders ON child_items.order_id = parent_orders.id WHERE parent_orders.id IS NULL;"
    assert: "orphan_items == 0"`,
    reduction: {
      originalOps: 20,
      minimalOps: 2,
      reductionPct: "90.0%",
      elapsed: "58ms",
      cycle: "T1: Child ──► Parent ◄──► T2: Parent ──► Child",
      explanation: "T1 locks child row and requests parent lock to update total; T2 locks parent row and requests cascaded child locks to delete. Lock hierarchy inversion creates deadlock (WFG cycle)."
    },
    fix: {
      strategy: "Índice na Chave Estrangeira e Padronização de Locks Pai-Filho",
      sql: `-- 1. Criar Índice Dedicado na Coluna de Chave Estrangeira (CRÍTICO!)
CREATE INDEX idx_child_items_order_id ON child_items(order_id);

-- 2. Padronizar a Sequência de Bloqueios Pai -> Filhos em Cancelamentos:
BEGIN;
-- Bloquear a linha pai explicitamente antes de disparar deleções:
SELECT id FROM parent_orders WHERE id = 1 FOR UPDATE;

-- Deletar os filhos de forma explícita e controlada:
DELETE FROM child_items WHERE order_id = 1;

-- Deletar o pedido pai:
DELETE FROM parent_orders WHERE id = 1;
COMMIT;`,
      explanation: "Deadlocks em deleções em cascata ocorrem porque: (1) a ausência de índice na coluna de FK força table scans na tabela filha, gerando bloqueios abrangentes desnecessários; e (2) transações de inserção travam primeiro a tabela filha e depois o pai, enquanto cascatas travam primeiro o pai e depois os filhos. A criação do índice na FK associada ao bloqueio prévio do registro pai (FOR UPDATE) equaliza a hierarquia e extingue o risco de deadlock.",
      engines: ["PostgreSQL", "MySQL (InnoDB)", "SQLite"]
    }
  }
];

// ============================================================================
// 2. HASH ROUTER & MULTI-VIEW CONTROLLER
// ============================================================================
const ROUTES = {
  landing: "view-landing",
  docs: "view-docs",
  scenarios: "view-scenarios",
  visualizer: "view-visualizer",
  matrix: "view-matrix"
};

function initRouter() {
  window.addEventListener("hashchange", handleRoute);
}

function handleRoute() {
  const hash = window.location.hash || "#/";
  let route = "landing";
  let param = null;

  if (hash.startsWith("#/docs")) {
    route = "docs";
    const parts = hash.split("/");
    if (parts.length >= 3 && parts[2]) {
      param = parts[2];
    }
  } else if (hash.startsWith("#/scenarios")) {
    route = "scenarios";
    const parts = hash.split("/");
    if (parts.length >= 3 && parts[2]) {
      param = parts[2];
    }
  } else if (hash.startsWith("#/visualizer")) {
    route = "visualizer";
  } else if (hash.startsWith("#/matrix")) {
    route = "matrix";
  } else {
    route = "landing";
  }

  // Update Portal Views
  document.querySelectorAll(".portal-view").forEach(view => {
    view.classList.remove("active");
  });

  const activeViewId = ROUTES[route] || "view-landing";
  const activeView = document.getElementById(activeViewId);
  if (activeView) {
    activeView.classList.add("active");
  }

  // Update Navigation Active State
  document.querySelectorAll(".nav-item").forEach(item => {
    if (item.getAttribute("data-route") === route) {
      item.classList.add("active");
    } else {
      item.classList.remove("active");
    }
  });

  // Update Mobile Drawer
  document.querySelectorAll(".mobile-nav-link").forEach(item => {
    if (item.getAttribute("data-route") === route) {
      item.classList.add("active");
    } else {
      item.classList.remove("active");
    }
  });

  const drawer = document.getElementById("mobileNavDrawer");
  if (drawer) drawer.classList.remove("open");

  // Route-Specific Setup
  if (route === "docs") {
    loadDocChapter(param || currentDocChapterId);
  } else if (route === "scenarios") {
    if (param) {
      const idx = SCENARIOS.findIndex(s => s.id === param);
      if (idx !== -1) {
        currentScenarioIndex = idx;
      }
    }
    renderScenarioNav();
    renderScenarioStage();
  } else if (route === "visualizer") {
    initVisualizer();
  }

  window.scrollTo({ top: 0, behavior: "instant" });
}

// ============================================================================
// 3. TERMINAL REPLAY SIMULATOR (Landing Page)
// ============================================================================
let terminalTimer = null;

function setupTerminalSimulator() {
  const runBtn = document.getElementById("termRunBtn");
  const jitterBtn = document.getElementById("termJitterBtn");
  const shrinkBtn = document.getElementById("termShrinkBtn");
  const resetBtn = document.getElementById("termResetBtn");

  if (runBtn) runBtn.addEventListener("click", () => runTerminalSimulation("full"));
  if (jitterBtn) jitterBtn.addEventListener("click", () => runTerminalSimulation("jitter"));
  if (shrinkBtn) shrinkBtn.addEventListener("click", () => runTerminalSimulation("shrink"));
  if (resetBtn) resetBtn.addEventListener("click", resetTerminal);
}

function runTerminalSimulation(mode) {
  const termEl = document.getElementById("terminalOutput");
  if (!termEl) return;

  clearTimeout(terminalTimer);
  termEl.innerHTML = "";

  document.querySelectorAll(".term-action-btn").forEach(b => b.classList.remove("active"));
  if (mode === "full") document.getElementById("termRunBtn")?.classList.add("active");
  if (mode === "jitter") document.getElementById("termJitterBtn")?.classList.add("active");
  if (mode === "shrink") document.getElementById("termShrinkBtn")?.classList.add("active");

  const steps = [];

  if (mode === "full" || mode === "jitter") {
    steps.push({ text: "# Initializing PCT-SQL concurrency fuzzer (4 workers, 20 iterations, seed=42)...", class: "term-dim", delay: 80 });
    steps.push({ text: "# Injecting micro-jitter [1ms, 5ms] on SQLite in-memory driver (Zero CGO)...", class: "term-dim", delay: 200 });
    steps.push({ text: "✘ ISOLATION ANOMALY DETECTED: P4_LOST_UPDATE", class: "term-err", delay: 450 });
    steps.push({ text: "  Cycle: T1 ──(rw)──► T2 ──(ww)──► T1", class: "term-dim", delay: 650 });
    steps.push({ text: "  Violated Invariant: total_balance == 1000 (Actual: 850)", class: "term-dim", delay: 850 });
  }

  if (mode === "full" || mode === "shrink") {
    steps.push({ text: "▶ Starting Causal Delta-Debugging (ddmin)...", class: "term-info", delay: 1050 });
    steps.push({ text: "  [Iteration 1] Testing subset of 10 operations ──► <span class=\"term-err\">FAIL (Anomaly Preserved)</span>", class: "", delay: 1300 });
    steps.push({ text: "  [Iteration 2] Testing subset of 4 operations  ──► <span class=\"term-err\">FAIL (Anomaly Preserved)</span>", class: "", delay: 1550 });
    steps.push({ text: "  [Iteration 3] Testing subset of 2 operations  ──► <span class=\"term-err\">FAIL (1-minimal achieved)</span>", class: "", delay: 1800 });
    steps.push({ text: "✔ Trace shrunk from 20 to 2 operations (90.0% reduction in 68ms)", class: "term-ok", delay: 2050 });
    steps.push({ text: "  Synthesized standalone repro: bin/repro_test.go", class: "term-dim", delay: 2250 });
  }

  let delaySum = 0;
  steps.forEach(step => {
    delaySum += step.delay;
    terminalTimer = setTimeout(() => {
      const line = document.createElement("div");
      if (step.class) line.className = step.class;
      line.innerHTML = step.text;
      termEl.appendChild(line);
      termEl.scrollTop = termEl.scrollHeight;
    }, delaySum);
  });
}

function resetTerminal() {
  clearTimeout(terminalTimer);
  const termEl = document.getElementById("terminalOutput");
  if (!termEl) return;

  document.querySelectorAll(".term-action-btn").forEach(b => b.classList.remove("active"));
  document.getElementById("termResetBtn")?.classList.add("active");

  termEl.innerHTML = `
    <div class="term-dim"># Initializing PCT-SQL concurrency fuzzer (4 workers, 20 iterations, seed=42)...</div>
    <div class="term-dim"># Injecting micro-jitter [1ms, 5ms] on SQLite in-memory driver (Zero CGO)...</div>
    <div class="term-err">✘ ISOLATION ANOMALY DETECTED: P4_LOST_UPDATE</div>
    <div class="term-dim">  Cycle: T1 ──(rw)──► T2 ──(ww)──► T1</div>
    <div class="term-dim">  Violated Invariant: total_balance == 1000 (Actual: 850)</div>
    <div class="term-info">▶ Starting Causal Delta-Debugging (ddmin)...</div>
    <div>  [Iteration 1] Testing subset of 10 operations ──► <span class="term-err">FAIL (Anomaly Preserved)</span></div>
    <div>  [Iteration 2] Testing subset of 4 operations  ──► <span class="term-err">FAIL (Anomaly Preserved)</span></div>
    <div>  [Iteration 3] Testing subset of 2 operations  ──► <span class="term-err">FAIL (1-minimal achieved)</span></div>
    <div class="term-ok">✔ Trace shrunk from 20 to 2 operations (90.0% reduction in 68ms)</div>
    <div class="term-dim">  Synthesized standalone repro: bin/repro_test.go</div>
  `;
}

// ============================================================================
// 4. DOCUMENTATION HUB CONTROLLER (View: #view-docs)
// ============================================================================
let currentDocChapterId = "getting-started";

function initDocsHub() {
  if (!window.DOCS_DATA || !Array.isArray(window.DOCS_DATA)) return;

  renderDocsSidebar();
  setupDocsSearch();
  setupDocsFooterNav();
}

function renderDocsSidebar(filterQuery = "") {
  const navEl = document.getElementById("docsSidebarNav");
  if (!navEl || !window.DOCS_DATA) return;

  const query = filterQuery.toLowerCase().trim();
  const categories = {};

  window.DOCS_DATA.forEach(doc => {
    const match = !query || 
      doc.title.toLowerCase().includes(query) || 
      doc.summary.toLowerCase().includes(query) || 
      doc.category.toLowerCase().includes(query);

    if (match) {
      if (!categories[doc.category]) categories[doc.category] = [];
      categories[doc.category].push(doc);
    }
  });

  if (Object.keys(categories).length === 0) {
    navEl.innerHTML = `<div style="padding: 12px; color: var(--text-muted); font-size: 0.85rem;">Nenhum capítulo encontrado.</div>`;
    return;
  }

  let html = "";
  for (const [catName, docs] of Object.entries(categories)) {
    html += `
      <div class="docs-nav-group">
        <div class="docs-nav-group-title">${catName}</div>
        ${docs.map(doc => `
          <button class="docs-nav-link ${doc.id === currentDocChapterId ? 'active' : ''}" data-doc-id="${doc.id}">
            <span>${doc.title}</span>
            <div class="docs-nav-indicator"></div>
          </button>
        `).join("")}
      </div>
    `;
  }

  navEl.innerHTML = html;

  navEl.querySelectorAll(".docs-nav-link").forEach(btn => {
    btn.addEventListener("click", () => {
      const docId = btn.getAttribute("data-doc-id");
      window.location.hash = `#/docs/${docId}`;
    });
  });
}

function loadDocChapter(chapterId) {
  if (!window.DOCS_DATA) return;
  const doc = window.DOCS_DATA.find(d => d.id === chapterId) || window.DOCS_DATA[0];
  if (!doc) return;

  currentDocChapterId = doc.id;

  // Breadcrumbs (Docs > Category > Chapter Title)
  const breadcrumbEl = document.getElementById("docsBreadcrumbs");
  if (breadcrumbEl) {
    breadcrumbEl.innerHTML = `
      <a href="#/docs">Docs</a> <span style="color: var(--text-muted); margin: 0 4px;">&gt;</span> 
      <span style="color: var(--text-secondary);">${escapeHtml(doc.category)}</span> <span style="color: var(--text-muted); margin: 0 4px;">&gt;</span> 
      <span style="color: var(--color-yellow);">${escapeHtml(doc.title)}</span>
    `;
  }

  // Header
  const badgeEl = document.getElementById("docsCategoryBadge");
  if (badgeEl) badgeEl.textContent = doc.category;

  const titleEl = document.getElementById("docsChapterTitle");
  if (titleEl) titleEl.textContent = doc.title;

  const summaryEl = document.getElementById("docsSummaryBox");
  if (summaryEl) summaryEl.textContent = doc.summary;

  // Content
  const bodyEl = document.getElementById("docsRenderedBody");
  if (bodyEl) {
    bodyEl.innerHTML = doc.content;

    // Attach copy button logic to any code block inside
    bodyEl.querySelectorAll(".code-container").forEach(container => {
      if (!container.querySelector(".copy-code-btn")) {
        const btn = document.createElement("button");
        btn.className = "copy-code-btn";
        btn.textContent = "Copiar";
        btn.onclick = () => {
          const code = container.querySelector("code")?.innerText || "";
          copyTextToClipboard(code);
          btn.textContent = "Copiado!";
          setTimeout(() => btn.textContent = "Copiar", 1500);
        };
        container.style.position = "relative";
        container.prepend(btn);
      }
    });
  }

  // Update Sidebar active item
  document.querySelectorAll(".docs-nav-link").forEach(link => {
    if (link.getAttribute("data-doc-id") === doc.id) {
      link.classList.add("active");
    } else {
      link.classList.remove("active");
    }
  });

  updateDocsFooterNav();
}

function setupDocsSearch() {
  const searchInput = document.getElementById("docsSearchInput");
  if (!searchInput) return;

  searchInput.addEventListener("input", (e) => {
    renderDocsSidebar(e.target.value);
  });
}

function setupDocsFooterNav() {
  const prevBtn = document.getElementById("docsPrevBtn");
  const nextBtn = document.getElementById("docsNextBtn");

  if (prevBtn) {
    prevBtn.addEventListener("click", () => {
      if (!window.DOCS_DATA) return;
      const idx = window.DOCS_DATA.findIndex(d => d.id === currentDocChapterId);
      if (idx > 0) {
        window.location.hash = `#/docs/${window.DOCS_DATA[idx - 1].id}`;
      }
    });
  }

  if (nextBtn) {
    nextBtn.addEventListener("click", () => {
      if (!window.DOCS_DATA) return;
      const idx = window.DOCS_DATA.findIndex(d => d.id === currentDocChapterId);
      if (idx < window.DOCS_DATA.length - 1) {
        window.location.hash = `#/docs/${window.DOCS_DATA[idx + 1].id}`;
      }
    });
  }
}

function updateDocsFooterNav() {
  if (!window.DOCS_DATA) return;
  const idx = window.DOCS_DATA.findIndex(d => d.id === currentDocChapterId);
  const prevBtn = document.getElementById("docsPrevBtn");
  const nextBtn = document.getElementById("docsNextBtn");

  if (prevBtn) {
    if (idx > 0) {
      prevBtn.style.visibility = "visible";
      prevBtn.innerHTML = `← ${window.DOCS_DATA[idx - 1].title}`;
    } else {
      prevBtn.style.visibility = "hidden";
    }
  }

  if (nextBtn) {
    if (idx < window.DOCS_DATA.length - 1) {
      nextBtn.style.visibility = "visible";
      nextBtn.innerHTML = `${window.DOCS_DATA[idx + 1].title} →`;
    } else {
      nextBtn.style.visibility = "hidden";
    }
  }
}

// ============================================================================
// 5. SCENARIOS EXPLORER CONTROLLER (View: #view-scenarios)
// ============================================================================
let currentScenarioIndex = 0;
let currentTab = "schema";

function renderScenarioNav() {
  const navContainer = document.getElementById("scenarioNavList");
  if (!navContainer) return;

  navContainer.innerHTML = SCENARIOS.map((sc, index) => `
    <button class="scenario-nav-item ${index === currentScenarioIndex ? 'active' : ''}" data-index="${index}" data-id="${sc.id}">
      <span class="scenario-nav-name">${sc.name}</span>
      <span class="scenario-nav-code">${sc.code}</span>
    </button>
  `).join('');

  navContainer.querySelectorAll(".scenario-nav-item").forEach(btn => {
    btn.addEventListener("click", () => {
      currentScenarioIndex = parseInt(btn.getAttribute("data-index"), 10);
      const sc = SCENARIOS[currentScenarioIndex];
      window.location.hash = `#/scenarios/${sc.id}`;
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
        <div style="font-family: var(--font-mono); font-size: 0.78rem; color: var(--color-yellow); margin-bottom: 6px; text-transform: uppercase;">Grafo de Conflito Formal (Adya)</div>
        <div style="font-family: var(--font-mono); font-size: 1rem; color: var(--color-cream); background: var(--bg-terminal); padding: 10px 14px; border-radius: var(--radius-sm); border: 1px solid var(--border-subtle);">${sc.reduction.cycle}</div>
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
  } else if (currentTab === "fix") {
    contentEl.innerHTML = `
      <div>
        <span class="fix-header-pill">PRODUÇÃO • RECOMENDAÇÃO ARQUITETURAL</span>
        <h4 class="fix-strategy-title">${sc.fix.strategy}</h4>
        <p class="fix-desc-text">${sc.fix.explanation}</p>
        
        <div class="code-container">
          <button class="copy-code-btn" onclick="copySnippet(this)">Copiar Fix SQL</button>
          <pre><code>${escapeHtml(sc.fix.sql)}</code></pre>
        </div>

        <div class="fix-badge-row">
          <span style="font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-muted); align-self: center;">Motores validados:</span>
          ${sc.fix.engines.map(eng => `<span class="fix-engine-badge">${eng}</span>`).join("")}
        </div>
      </div>
    `;
  }
}

function setupScenarioTabs() {
  document.querySelectorAll(".stage-tab-btn").forEach(tabBtn => {
    tabBtn.addEventListener("click", () => {
      document.querySelectorAll(".stage-tab-btn").forEach(b => b.classList.remove("active"));
      tabBtn.classList.add("active");
      currentTab = tabBtn.getAttribute("data-tab");
      renderScenarioStage();
    });
  });
}

// ============================================================================
// 6. TRACE VISUALIZER SHOWCASE CONTROLLER (View: #view-visualizer)
// ============================================================================
let vizMode = "raw"; // "raw" or "shrunk"
let vizWorkerFilter = "all";

// 20 realistic operations interleaved across 4 workers (W0, W1, W2, W3)
const RAW_TRACE_OPS = [
  { id: "op_0", worker: 0, tx: "T1", type: "read", name: "SELECT balance FROM accounts WHERE id = 1 -> cur", startUs: 5, durationUs: 25, vars: "cur = 1000", status: "OK" },
  { id: "op_1", worker: 1, tx: "T2", type: "read", name: "SELECT balance FROM accounts WHERE id = 1 -> cur", startUs: 12, durationUs: 28, vars: "cur = 1000", status: "OK" },
  { id: "op_2", worker: 2, tx: "T3", type: "read", name: "SELECT count(*) FROM accounts -> cnt", startUs: 20, durationUs: 22, vars: "cnt = 2", status: "OK" },
  { id: "op_3", worker: 3, tx: "T4", type: "read", name: "SELECT id, status FROM ledger WHERE id = 10", startUs: 32, durationUs: 24, vars: "status = 'ACTIVE'", status: "OK" },
  { id: "op_4", worker: 0, tx: "T1", type: "read", name: "SELECT balance FROM accounts WHERE id = 2", startUs: 45, durationUs: 26, vars: "balance = 500", status: "OK" },
  { id: "op_5", worker: 1, tx: "T2", type: "read", name: "SELECT min(balance) FROM accounts", startUs: 58, durationUs: 25, vars: "min = 500", status: "OK" },
  { id: "op_6", worker: 2, tx: "T3", type: "write", name: "UPDATE audit_heartbeat SET last_ping = 6500", startUs: 70, durationUs: 30, vars: "ping = 6500", status: "COMMITTED" },
  { id: "op_7", worker: 3, tx: "T4", type: "read", name: "SELECT max(id) FROM audit_heartbeat", startUs: 85, durationUs: 20, vars: "max = 1", status: "OK" },
  { id: "op_8", worker: 0, tx: "T1", type: "read", name: "BEGIN; -- Worker 0 critical section", startUs: 98, durationUs: 15, vars: "tx = T1", status: "OK" },
  { id: "op_9", worker: 1, tx: "T2", type: "read", name: "BEGIN; -- Worker 1 critical section", startUs: 105, durationUs: 15, vars: "tx = T2", status: "OK" },
  { id: "op_10", worker: 2, tx: "T3", type: "read", name: "SELECT balance FROM accounts WHERE id = 1", startUs: 115, durationUs: 20, vars: "cur = 1000", status: "OK" },
  { id: "op_11", worker: 3, tx: "T4", type: "write", name: "INSERT INTO trace_events VALUES ('SCHED_TICK')", startUs: 122, durationUs: 22, vars: "tick = 42", status: "COMMITTED" },
  { id: "op_12", worker: 0, tx: "T1", type: "write", name: "UPDATE accounts SET balance = 900 WHERE id = 1", startUs: 130, durationUs: 40, vars: "balance = 900", status: "COMMITTED" },
  { id: "op_13", worker: 1, tx: "T2", type: "conflict", name: "UPDATE accounts SET balance = 900 WHERE id = 1", startUs: 145, durationUs: 45, vars: "balance = 900 [OVERWRITE COLLISION]", status: "P4_LOST_UPDATE (Triggered at 184μs)" },
  { id: "op_14", worker: 2, tx: "T3", type: "write", name: "COMMIT; -- Worker 2 audit finished", startUs: 160, durationUs: 18, vars: "status = 'OK'", status: "COMMITTED" },
  { id: "op_15", worker: 3, tx: "T4", type: "read", name: "SELECT sum(balance) AS total FROM accounts", startUs: 175, durationUs: 25, vars: "total = 1400 (Expected 1500)", status: "INVARIANT_FAIL" },
  { id: "op_16", worker: 0, tx: "T1", type: "write", name: "COMMIT; -- Worker 0 committed", startUs: 185, durationUs: 16, vars: "status = 'COMMITTED'", status: "COMMITTED" },
  { id: "op_17", worker: 1, tx: "T2", type: "write", name: "COMMIT; -- Worker 1 committed (Overwritten)", startUs: 192, durationUs: 18, vars: "status = 'COMMITTED'", status: "COMMITTED" },
  { id: "op_18", worker: 2, tx: "T3", type: "write", name: "INSERT INTO anomaly_log VALUES ('P4', 184)", startUs: 205, durationUs: 22, vars: "logged = true", status: "COMMITTED" },
  { id: "op_19", worker: 3, tx: "T4", type: "write", name: "ROLLBACK; -- Fuzzer teardown context", startUs: 220, durationUs: 15, vars: "done = true", status: "ROLLED_BACK" }
];

// 1-Minimal reproduction isolated by Andreas Zeller's ddmin algorithm
const SHRUNK_TRACE_OPS = [
  { id: "op_s0", worker: 0, tx: "T1", type: "read", name: "SELECT balance FROM accounts WHERE id = 1 -> cur", startUs: 10, durationUs: 35, vars: "cur = 1000", status: "OK" },
  { id: "op_s1", worker: 1, tx: "T2", type: "read", name: "SELECT balance FROM accounts WHERE id = 1 -> cur", startUs: 25, durationUs: 35, vars: "cur = 1000", status: "OK" },
  { id: "op_s2", worker: 0, tx: "T1", type: "write", name: "UPDATE accounts SET balance = {cur - 100} WHERE id = 1", startUs: 85, durationUs: 45, vars: "balance = 900", status: "COMMITTED" },
  { id: "op_s3", worker: 1, tx: "T2", type: "conflict", name: "UPDATE accounts SET balance = {cur - 100} WHERE id = 1", startUs: 105, durationUs: 50, vars: "balance = 900 [OVERWRITE]", status: "1-MINIMAL COLLISION" }
];

function initVisualizer() {
  setupVizControls();
  renderGantt();
  renderVizAdyaDag();
  renderQueryInspector(vizMode === "raw" ? RAW_TRACE_OPS[13] : SHRUNK_TRACE_OPS[3]);
}

function setupVizControls() {
  const rawBtn = document.getElementById("vizModeRaw");
  const shrunkBtn = document.getElementById("vizModeShrunk");
  const animBtn = document.getElementById("vizAnimateBtn");

  if (rawBtn && shrunkBtn) {
    rawBtn.onclick = () => {
      vizMode = "raw";
      rawBtn.classList.add("active");
      shrunkBtn.classList.remove("active");
      renderGantt();
      renderQueryInspector(RAW_TRACE_OPS[13]);
    };

    shrunkBtn.onclick = () => {
      vizMode = "shrunk";
      shrunkBtn.classList.add("active");
      rawBtn.classList.remove("active");
      renderGantt();
      renderQueryInspector(SHRUNK_TRACE_OPS[3]);
    };
  }

  document.querySelectorAll(".viz-worker-filter").forEach(btn => {
    btn.onclick = () => {
      document.querySelectorAll(".viz-worker-filter").forEach(b => b.classList.remove("active"));
      btn.classList.add("active");
      vizWorkerFilter = btn.getAttribute("data-worker");
      renderGantt();
    };
  });

  if (animBtn) {
    animBtn.onclick = runGanttAnimation;
  }
}

function renderGantt() {
  const container = document.getElementById("ganttContainer");
  if (!container) return;

  const ops = vizMode === "raw" ? RAW_TRACE_OPS : SHRUNK_TRACE_OPS;
  const filteredOps = vizWorkerFilter === "all" ? ops : ops.filter(o => o.worker.toString() === vizWorkerFilter);

  const workers = [0, 1, 2, 3];
  const maxTime = 250; // μs

  let html = `
    <div class="gantt-axis">
      <div class="gantt-axis-tick">0μs</div>
      <div class="gantt-axis-tick">50μs</div>
      <div class="gantt-axis-tick">100μs</div>
      <div class="gantt-axis-tick">150μs</div>
      <div class="gantt-axis-tick">200μs</div>
      <div class="gantt-axis-tick">250μs</div>
    </div>
    <div class="gantt-lanes" style="position: relative;">
  `;

  // Collision Marker Line
  const collisionX = vizMode === "raw" ? (184 / maxTime) * 100 : (155 / maxTime) * 100;
  html += `
    <div class="gantt-collision-marker" style="left: calc(80px + (100% - 80px) * (${collisionX} / 100));">
      <div class="gantt-collision-label">P4 Collision (${vizMode === "raw" ? '184μs' : '155μs'})</div>
    </div>
  `;

  workers.forEach(w => {
    if (vizWorkerFilter !== "all" && vizWorkerFilter !== w.toString()) return;

    const workerOps = filteredOps.filter(o => o.worker === w);
    html += `
      <div class="gantt-lane">
        <div class="gantt-lane-label">Worker ${w}</div>
        <div class="gantt-track">
          ${workerOps.map(op => {
            const leftPct = (op.startUs / maxTime) * 100;
            const widthPct = (op.durationUs / maxTime) * 100;
            let opClass = "op-read";
            if (op.type === "write") opClass = "op-write";
            if (op.type === "conflict") opClass = "op-conflict";

            return `
              <div class="gantt-block ${opClass}" style="left: ${leftPct}%; width: ${Math.max(widthPct, 7)}%;" data-op-id="${op.id}" title="${op.name}">
                ${op.tx}: ${op.type.toUpperCase()} (${op.durationUs}μs)
              </div>
            `;
          }).join("")}
        </div>
      </div>
    `;
  });

  html += `</div>`;
  container.innerHTML = html;

  container.querySelectorAll(".gantt-block").forEach(block => {
    block.onclick = () => {
      const opId = block.getAttribute("data-op-id");
      const found = (vizMode === "raw" ? RAW_TRACE_OPS : SHRUNK_TRACE_OPS).find(o => o.id === opId);
      if (found) {
        container.querySelectorAll(".gantt-block").forEach(b => b.classList.remove("active"));
        block.classList.add("active");
        renderQueryInspector(found);
      }
    };
  });
}

function renderQueryInspector(op) {
  const inspector = document.getElementById("queryInspector");
  if (!inspector || !op) return;

  inspector.innerHTML = `
    <div class="viz-pane-header">
      <span class="viz-pane-title">Inspetor de Operações &amp; Queries</span>
      <span style="font-family: var(--font-mono); font-size: 0.72rem; color: ${op.type === 'conflict' ? 'var(--color-red)' : 'var(--color-green)'}; font-weight: 600;">${op.status}</span>
    </div>
    <div class="inspector-grid">
      <span class="inspector-lbl">Transação / Worker:</span>
      <span class="inspector-val"><strong>${op.tx}</strong> (Goroutine Worker ${op.worker})</span>

      <span class="inspector-lbl">Timestamp / Latência:</span>
      <span class="inspector-val">${op.startUs}μs (+${op.durationUs}μs de execução)</span>

      <span class="inspector-lbl">Parâmetros / Vars:</span>
      <span class="inspector-val" style="color: var(--color-yellow); font-family: var(--font-mono); font-size: 0.8rem;">${op.vars}</span>

      <span class="inspector-lbl">Grafo de Conflito:</span>
      <span class="inspector-val" style="font-family: var(--font-mono); font-size: 0.8rem; color: ${op.type === 'conflict' ? 'var(--color-red)' : 'var(--text-primary)'};">${op.type === "conflict" ? "T1 ──(rw)──► T2 ──(ww)──► T1 [CYCLE DETECTED]" : "Linha serializável sem ciclos"}</span>

      <div class="inspector-code">
        <code>${escapeHtml(op.name)}</code>
      </div>
    </div>
  `;
}

function renderVizAdyaDag() {
  const dagContainer = document.getElementById("vizAdyaDag");
  if (!dagContainer) return;

  dagContainer.innerHTML = `
    <svg viewBox="0 0 380 180" style="width: 100%; max-height: 180px; cursor: pointer;">
      <defs>
        <marker id="viz-arrow" viewBox="0 0 10 10" refX="6" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#F5C400"/>
        </marker>
        <marker id="viz-arrow-red" viewBox="0 0 10 10" refX="6" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M 0 1 L 10 5 L 0 9 z" fill="#EF4444"/>
        </marker>
      </defs>

      <!-- Path T1 -> T2 (rw anti-dependency) -->
      <g id="edgeRw">
        <path d="M 100 80 C 140 20, 240 20, 280 80" stroke="#F5C400" stroke-width="2.4" fill="none" marker-end="url(#viz-arrow)" stroke-dasharray="4, 2"/>
        <rect x="165" y="24" width="50" height="20" rx="4" fill="#0D0A17" stroke="#F5C400" stroke-width="1.2"/>
        <text x="190" y="38" fill="#F5C400" font-family="JetBrains Mono" font-size="10" font-weight="700" text-anchor="middle">rw</text>
      </g>

      <!-- Path T2 -> T1 (ww write-write conflict) -->
      <g id="edgeWw">
        <path d="M 280 100 C 240 160, 140 160, 100 100" stroke="#EF4444" stroke-width="2.4" fill="none" marker-end="url(#viz-arrow-red)" stroke-dasharray="4, 2"/>
        <rect x="165" y="136" width="50" height="20" rx="4" fill="#0D0A17" stroke="#EF4444" stroke-width="1.2"/>
        <text x="190" y="150" fill="#EF4444" font-family="JetBrains Mono" font-size="10" font-weight="700" text-anchor="middle">ww</text>
      </g>

      <!-- Node T1 -->
      <g id="nodeT1">
        <circle cx="80" cy="90" r="28" fill="#1F1934" stroke="#4B2E83" stroke-width="2.4"/>
        <text x="80" y="94" fill="#FCFBF8" font-family="Inter" font-size="14" font-weight="700" text-anchor="middle">T₁</text>
      </g>

      <!-- Node T2 -->
      <g id="nodeT2">
        <circle cx="300" cy="90" r="28" fill="#1F1934" stroke="#F5C400" stroke-width="2.4"/>
        <text x="300" y="94" fill="#FCFBF8" font-family="Inter" font-size="14" font-weight="700" text-anchor="middle">T₂</text>
      </g>
    </svg>
  `;

  // Make nodes/edges clickable in DAG
  const nodeT1 = dagContainer.querySelector("#nodeT1");
  const nodeT2 = dagContainer.querySelector("#nodeT2");
  const edgeRw = dagContainer.querySelector("#edgeRw");
  const edgeWw = dagContainer.querySelector("#edgeWw");

  if (nodeT1) nodeT1.onclick = () => renderQueryInspector(vizMode === "raw" ? RAW_TRACE_OPS[12] : SHRUNK_TRACE_OPS[2]);
  if (nodeT2) nodeT2.onclick = () => renderQueryInspector(vizMode === "raw" ? RAW_TRACE_OPS[13] : SHRUNK_TRACE_OPS[3]);
  if (edgeRw) edgeRw.onclick = () => renderQueryInspector(vizMode === "raw" ? RAW_TRACE_OPS[0] : SHRUNK_TRACE_OPS[0]);
  if (edgeWw) edgeWw.onclick = () => renderQueryInspector(vizMode === "raw" ? RAW_TRACE_OPS[13] : SHRUNK_TRACE_OPS[3]);
}

function runGanttAnimation() {
  const container = document.getElementById("ganttContainer");
  if (!container) return;

  const blocks = container.querySelectorAll(".gantt-block");
  blocks.forEach((b, i) => {
    b.style.opacity = "0.2";
    setTimeout(() => {
      b.style.opacity = "1";
      b.style.transform = "scale(1.08)";
      setTimeout(() => b.style.transform = "", 250);
    }, i * 140);
  });
}

// ============================================================================
// 7. GLOBAL UTILITIES & INITIALIZATION
// ============================================================================
function setupGlobalUI() {
  // Mobile nav toggle
  const toggleBtn = document.getElementById("navMobileToggle");
  const drawer = document.getElementById("mobileNavDrawer");
  if (toggleBtn && drawer) {
    toggleBtn.addEventListener("click", () => {
      drawer.classList.toggle("open");
    });
  }

  // Copy install command button
  const copyInstallBtn = document.getElementById("copyInstallBtn");
  if (copyInstallBtn) {
    copyInstallBtn.addEventListener("click", () => {
      copyTextToClipboard("go install github.com/bregaldahq/chaossql/cmd/chaossql@latest");
      copyInstallBtn.style.color = "var(--color-green)";
      setTimeout(() => copyInstallBtn.style.color = "", 1500);
    });
  }
}

function copyTextToClipboard(text) {
  if (navigator.clipboard && window.isSecureContext) {
    return navigator.clipboard.writeText(text);
  } else {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";
    textArea.style.left = "-999999px";
    textArea.style.top = "-999999px";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    try {
      document.execCommand("copy");
    } catch (err) {
      console.error("Failed to copy text: ", err);
    }
    textArea.remove();
    return Promise.resolve();
  }
}

function copySnippet(button) {
  const code = button.parentElement.querySelector("code")?.innerText;
  if (code) {
    copyTextToClipboard(code);
    const orig = button.textContent;
    button.textContent = "Copiado!";
    setTimeout(() => button.textContent = orig, 1500);
  }
}

function escapeHtml(str) {
  return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

// Robust Bootstrap
function bootstrap() {
  setupGlobalUI();
  setupTerminalSimulator();
  initDocsHub();
  setupScenarioTabs();
  initRouter();
  handleRoute();
}

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", bootstrap);
} else {
  bootstrap();
}
