const y=[{id:"op_0",worker:0,tx:"T1",type:"read",name:"SELECT balance FROM accounts WHERE id = 1 -> cur",startUs:5,durationUs:25,vars:"cur = 1000",status:"OK"},{id:"op_1",worker:1,tx:"T2",type:"read",name:"SELECT balance FROM accounts WHERE id = 1 -> cur",startUs:12,durationUs:28,vars:"cur = 1000",status:"OK"},{id:"op_2",worker:2,tx:"T3",type:"read",name:"SELECT count(*) FROM accounts -> cnt",startUs:20,durationUs:22,vars:"cnt = 2",status:"OK"},{id:"op_3",worker:3,tx:"T4",type:"read",name:"SELECT id, status FROM ledger WHERE id = 10",startUs:32,durationUs:24,vars:"status = 'ACTIVE'",status:"OK"},{id:"op_4",worker:0,tx:"T1",type:"read",name:"SELECT balance FROM accounts WHERE id = 2",startUs:45,durationUs:26,vars:"balance = 500",status:"OK"},{id:"op_5",worker:1,tx:"T2",type:"read",name:"SELECT min(balance) FROM accounts",startUs:58,durationUs:25,vars:"min = 500",status:"OK"},{id:"op_6",worker:2,tx:"T3",type:"write",name:"UPDATE audit_heartbeat SET last_ping = 6500",startUs:70,durationUs:30,vars:"ping = 6500",status:"COMMITTED"},{id:"op_7",worker:3,tx:"T4",type:"read",name:"SELECT max(id) FROM audit_heartbeat",startUs:85,durationUs:20,vars:"max = 1",status:"OK"},{id:"op_8",worker:0,tx:"T1",type:"read",name:"BEGIN; -- Worker 0 critical section",startUs:98,durationUs:15,vars:"tx = T1",status:"OK"},{id:"op_9",worker:1,tx:"T2",type:"read",name:"BEGIN; -- Worker 1 critical section",startUs:105,durationUs:15,vars:"tx = T2",status:"OK"},{id:"op_10",worker:2,tx:"T3",type:"read",name:"SELECT balance FROM accounts WHERE id = 1",startUs:115,durationUs:20,vars:"cur = 1000",status:"OK"},{id:"op_11",worker:3,tx:"T4",type:"write",name:"INSERT INTO trace_events VALUES ('SCHED_TICK')",startUs:122,durationUs:22,vars:"tick = 42",status:"COMMITTED"},{id:"op_12",worker:0,tx:"T1",type:"write",name:"UPDATE accounts SET balance = 900 WHERE id = 1",startUs:130,durationUs:40,vars:"balance = 900",status:"COMMITTED"},{id:"op_13",worker:1,tx:"T2",type:"conflict",name:"UPDATE accounts SET balance = 900 WHERE id = 1",startUs:145,durationUs:45,vars:"balance = 900 [OVERWRITE COLLISION]",status:"P4_LOST_UPDATE (Triggered at 184μs)"},{id:"op_14",worker:2,tx:"T3",type:"write",name:"COMMIT; -- Worker 2 audit finished",startUs:160,durationUs:18,vars:"status = 'OK'",status:"COMMITTED"},{id:"op_15",worker:3,tx:"T4",type:"read",name:"SELECT sum(balance) AS total FROM accounts",startUs:175,durationUs:25,vars:"total = 1400 (Expected 1500)",status:"INVARIANT_FAIL"},{id:"op_16",worker:0,tx:"T1",type:"write",name:"COMMIT; -- Worker 0 committed",startUs:185,durationUs:16,vars:"status = 'COMMITTED'",status:"COMMITTED"},{id:"op_17",worker:1,tx:"T2",type:"write",name:"COMMIT; -- Worker 1 committed (Overwritten)",startUs:192,durationUs:18,vars:"status = 'COMMITTED'",status:"COMMITTED"},{id:"op_18",worker:2,tx:"T3",type:"write",name:"INSERT INTO anomaly_log VALUES ('P4', 184)",startUs:205,durationUs:22,vars:"logged = true",status:"COMMITTED"},{id:"op_19",worker:3,tx:"T4",type:"write",name:"ROLLBACK; -- Fuzzer teardown context",startUs:220,durationUs:15,vars:"done = true",status:"ROLLED_BACK"}],_=[{id:"op_s0",worker:0,tx:"T1",type:"read",name:"SELECT balance FROM accounts WHERE id = 1 -> cur",startUs:10,durationUs:35,vars:"cur = 1000",status:"OK"},{id:"op_s1",worker:1,tx:"T2",type:"read",name:"SELECT balance FROM accounts WHERE id = 1 -> cur",startUs:25,durationUs:35,vars:"cur = 1000",status:"OK"},{id:"op_s2",worker:0,tx:"T1",type:"write",name:"UPDATE accounts SET balance = {cur - 100} WHERE id = 1",startUs:85,durationUs:45,vars:"balance = 900",status:"COMMITTED"},{id:"op_s3",worker:1,tx:"T2",type:"conflict",name:"UPDATE accounts SET balance = {cur - 100} WHERE id = 1",startUs:105,durationUs:50,vars:"balance = 900 [OVERWRITE]",status:"1-MINIMAL COLLISION"}],A=[{id:"banking",namePt:"Banking Lost Update (P4)",nameEn:"Banking Lost Update (P4)",anomaly:"P4_LOST_UPDATE",descriptionPt:"Dois saques concorrentes leem o mesmo saldo e sobrescrevem uma alteração.",descriptionEn:"Two concurrent withdrawals read the same balance and overwrite updates.",reducedOps:4,edges:[{from:"T1",to:"T2",type:"RW",item:"accounts:1"},{from:"T2",to:"T1",type:"WW",item:"accounts:1"},{from:"T1",to:"T4",type:"RW",item:"accounts:1"},{from:"T4",to:"T3",type:"WW",item:"accounts:1"},{from:"T3",to:"T2",type:"RW",item:"accounts:1"},{from:"T2",to:"T4",type:"WW",item:"accounts:1"}],yaml:`version: "1.0"
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
      - sql: "UPDATE accounts SET balance = {cur_bal - 100} WHERE id = 1"`},{id:"inventory",namePt:"Inventory Oversell (A3)",nameEn:"Inventory Oversell (A3)",anomaly:"G2_ANTI_DEPENDENCY",descriptionPt:"Compras paralelas verificam estoque positivo simultaneamente, resultando em estoque negativo.",descriptionEn:"Parallel checkouts simultaneously observe positive stock, causing negative inventory.",reducedOps:6,edges:[{from:"T1",to:"T2",type:"RW",item:"inventory:stock"},{from:"T2",to:"T3",type:"RW",item:"inventory:stock"},{from:"T3",to:"T4",type:"RW",item:"inventory:stock"},{from:"T4",to:"T1",type:"RW",item:"inventory:stock"},{from:"T1",to:"T3",type:"RW",item:"inventory:stock"},{from:"T2",to:"T4",type:"RW",item:"inventory:stock"}],yaml:`version: "1.0"
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
      - sql: "UPDATE inventory SET stock = {cur - 1} WHERE item_id = 1 AND {cur > 0}"`},{id:"hospital",namePt:"Hospital Write Skew (A5B)",nameEn:"Hospital Write Skew (A5B)",anomaly:"A5A_READ_SKEW",descriptionPt:"Dois médicos solicitam dispensa checando se há ao menos 2 ativos; ambos saem ao mesmo tempo.",descriptionEn:"Two doctors sign off simultaneously verifying count >= 2, leaving hospital unmanned.",reducedOps:4,edges:[{from:"T1",to:"T2",type:"RW",item:"doctors:2"},{from:"T2",to:"T1",type:"RW",item:"doctors:2"},{from:"T1",to:"T4",type:"RW",item:"doctors:2"},{from:"T2",to:"T3",type:"RW",item:"doctors:2"},{from:"T4",to:"T3",type:"RW",item:"doctors:1"},{from:"T3",to:"T4",type:"RW",item:"doctors:1"},{from:"T1",to:"T3",type:"RW",item:"doctors:2"},{from:"T4",to:"T2",type:"RW",item:"doctors:1"}],yaml:`version: "1.0"
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
      - sql: "UPDATE doctors SET on_duty = 0 WHERE id = 2 AND {act >= 2}"`},{id:"financial",namePt:"Financial Audit Read Skew (A5A)",nameEn:"Financial Audit Read Skew (A5A)",anomaly:"A5A_READ_SKEW",descriptionPt:"Transações de transferência modificam contas enquanto a auditoria lê saldos desalinhados.",descriptionEn:"Transfer transactions mutate accounts while audit reads misaligned balances.",reducedOps:3,edges:[{from:"T1",to:"T2",type:"RW",item:"accounts:1"},{from:"T2",to:"T1",type:"WR",item:"accounts:2"},{from:"T3",to:"T4",type:"RW",item:"accounts:1"},{from:"T4",to:"T3",type:"WR",item:"accounts:2"},{from:"T1",to:"T3",type:"WR",item:"accounts:2"},{from:"T4",to:"T2",type:"RW",item:"accounts:1"}],yaml:`version: "1.0"
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
      - sql: "UPDATE accounts SET balance = balance + 100 WHERE id = 2"`},{id:"auction",namePt:"Auction Bidding Dirty Write (G0)",nameEn:"Auction Bidding Dirty Write (G0)",anomaly:"G0_DIRTY_WRITE",descriptionPt:"Dois lances simultâneos intercalam escrita de valor e arrematante gerando registro híbrido.",descriptionEn:"Two bids interleave high bid and high bidder writes resulting in hybrid state.",reducedOps:4,edges:[{from:"T1",to:"T2",type:"WW",item:"auctions:bid"},{from:"T2",to:"T1",type:"WW",item:"auctions:bidder"},{from:"T3",to:"T4",type:"WW",item:"auctions:bid"},{from:"T4",to:"T3",type:"WW",item:"auctions:bidder"},{from:"T1",to:"T4",type:"WW",item:"auctions:bid"},{from:"T2",to:"T3",type:"WW",item:"auctions:bidder"}],yaml:`version: "1.0"
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
      - sql: "UPDATE auctions SET high_bidder = 'Bob' WHERE item_id = 1"`},{id:"crypto",namePt:"Crypto Arbitrage Circular Info (G1c)",nameEn:"Crypto Arbitrage Circular Info (G1c)",anomaly:"G2_CYCLE",descriptionPt:"Transações cruzadas atualizam cotações BTC e ETH baseadas em taxas anteriores mutuamente.",descriptionEn:"Cross transactions update BTC and ETH quotes based on mutual previous rates.",reducedOps:6,edges:[{from:"T1",to:"T2",type:"RW",item:"crypto:ETH"},{from:"T2",to:"T3",type:"RW",item:"crypto:BTC"},{from:"T3",to:"T4",type:"RW",item:"crypto:USD"},{from:"T4",to:"T1",type:"RW",item:"crypto:BTC"},{from:"T1",to:"T3",type:"RW",item:"crypto:ETH"},{from:"T4",to:"T2",type:"RW",item:"crypto:USD"}],yaml:`version: "1.0"
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
      - sql: "UPDATE crypto_pairs SET rate = {btc / 20} WHERE pair = 'ETH'"`},{id:"flash_crash",namePt:"Flash Crash Dirty Read (G1a)",nameEn:"Flash Crash Dirty Read (G1a)",anomaly:"A1_DIRTY_READ",descriptionPt:"Um dump não comitado no book de ofertas é lido por uma liquidação de margem.",descriptionEn:"An uncommitted price dump is read by an automated liquidation bot.",reducedOps:2,edges:[{from:"T1",to:"T2",type:"WW",item:"orderbook:price"},{from:"T2",to:"T1",type:"WR",item:"orderbook:price"},{from:"T3",to:"T4",type:"WW",item:"orderbook:price"},{from:"T4",to:"T3",type:"WR",item:"orderbook:price"},{from:"T1",to:"T4",type:"WR",item:"orderbook:price"},{from:"T3",to:"T2",type:"WW",item:"orderbook:price"}],yaml:`version: "1.0"
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
        capture: "p"`},{id:"ticket",namePt:"Ticket Seat Anti-Dependency (G2)",nameEn:"Ticket Seat Anti-Dependency (G2)",anomaly:"A3_PHANTOM_PREDICATE",descriptionPt:"Dois passageiros encontram assentos vagos sob o mesmo predicado e reservam ambos.",descriptionEn:"Two customers concurrently observe available seats and double-book them.",reducedOps:6,edges:[{from:"T1",to:"T2",type:"RW",item:"seats:free"},{from:"T2",to:"T3",type:"RW",item:"seats:free"},{from:"T3",to:"T4",type:"RW",item:"seats:free"},{from:"T4",to:"T1",type:"RW",item:"seats:free"},{from:"T1",to:"T3",type:"RW",item:"seats:free"},{from:"T2",to:"T4",type:"RW",item:"seats:free"}],yaml:`version: "1.0"
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
      - sql: "UPDATE seats SET booked = 1 WHERE id = 2 AND {f > 0}"`},{id:"deadlock",namePt:"Deadlock Cycle Diagnostics (G-DL)",nameEn:"Deadlock Cycle Diagnostics (G-DL)",anomaly:"CYCLE_DEADLOCK",descriptionPt:"Transações concorrentes adquirem travas em ordem invertida gerando ciclo no grafo de espera.",descriptionEn:"Concurrent transactions acquire locks in reversed order generating wait-for cycle.",reducedOps:4,edges:[{from:"T1",to:"T2",type:"WR",item:"locks:wait_2"},{from:"T2",to:"T1",type:"WR",item:"locks:wait_1"},{from:"T3",to:"T4",type:"WR",item:"locks:wait_2"},{from:"T4",to:"T3",type:"WR",item:"locks:wait_1"},{from:"T1",to:"T4",type:"WR",item:"locks:wait_1"},{from:"T3",to:"T2",type:"WR",item:"locks:wait_2"}],yaml:`version: "1.0"
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
      - sql: "UPDATE locks SET v = v - 1 WHERE id = 1"`},{id:"fk",namePt:"Foreign Key Cascade Deadlock (FK)",nameEn:"Foreign Key Cascade Deadlock (FK)",anomaly:"FK_CASCADE_VIOLATION",descriptionPt:"Exclusão do pai colide com inserção no filho através de trigger implícito de consistência.",descriptionEn:"Parent order deletion collides with item insertion through cascade triggers.",reducedOps:4,edges:[{from:"T1",to:"T2",type:"WW",item:"orders:parent"},{from:"T2",to:"T1",type:"WR",item:"items:child"},{from:"T3",to:"T4",type:"WW",item:"orders:parent"},{from:"T4",to:"T3",type:"WR",item:"items:child"},{from:"T1",to:"T3",type:"WW",item:"orders:parent"},{from:"T4",to:"T2",type:"WR",item:"items:child"}],yaml:`version: "1.0"
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
      - sql: "DELETE FROM parent_orders WHERE id = 1"`}];function k(r,e,t,o,s){const i=e*t,u=Math.round(25+o*1.6+i*.14+s%10),c=[],m=[{type:"SELECT",sql:"SELECT * FROM data WHERE id = 1",dur:22e3},{type:"SELECT",sql:"SELECT count(*) FROM data",dur:25e3},{type:"UPDATE",sql:"UPDATE data SET v = v + 1 WHERE id = 1",dur:38e3},{type:"UPDATE",sql:"UPDATE data SET v = v - 1 WHERE id = 1 [COLLISION]",dur:45e3},{type:"INSERT",sql:"INSERT INTO log VALUES ('EVENT')",dur:3e4},{type:"COMMIT",sql:"COMMIT",dur:18e3}];for(let a=0;a<e;a++){const T=Math.min(t,6);for(let n=0;n<T;n++){const d=m[(a+n+s)%m.length],p=a===1&&n===T-2||a===0&&n===T-1;c.push({worker_id:a,type:p?"UPDATE":d.type,sql:p?`${d.sql} [COLLISION / ${r.anomaly}]`:d.sql,duration_ns:d.dur+s*100+a*1500})}}const l=c.filter(a=>(a.sql||"").includes("COLLISION")||a.type==="UPDATE").slice(0,r.reducedOps);return{totalOps:i,reducedOps:r.reducedOps,durationMs:u,anomalyType:r.anomaly,adyaEdges:r.edges,trace:c,reducedTrace:l.length>0?l:c.slice(0,r.reducedOps)}}class R{worker=null;isReady=!1;onReadyCallbacks=[];onReportCallback=null;onProgressCallback=null;onValidationCallback=null;onErrorCallback=null;constructor(){this.init()}init(){if(!(typeof window>"u"||typeof Worker>"u")&&!this.worker)try{this.worker=new Worker("/wasm/wasm-worker.js"),this.worker.onmessage=e=>{this.handleMessage(e.data)},this.worker.onerror=e=>{console.warn("ChaosSQL Worker encountered error:",e.message),this.onErrorCallback&&this.onErrorCallback(e.message||"Worker runtime error")},this.worker.postMessage({action:"INIT",wasmUrl:"/wasm/chaossql.wasm"})}catch(e){console.warn("Could not spawn ChaosSQL Web Worker:",e)}}handleMessage(e){if(e)switch(e.type){case"READY":this.isReady=!0,this.onReadyCallbacks.forEach(t=>t()),this.onReadyCallbacks=[];break;case"VALIDATION_RESULT":this.onValidationCallback&&this.onValidationCallback({valid:!!e.valid,error:e.error,name:e.name,operations:e.operations,invariants:e.invariants});break;case"CYCLE_DETECTED":case"ANOMALY":case"PROGRESS":this.onProgressCallback&&this.onProgressCallback(e);break;case"COMPLETE":case"REPORT":case"DONE":{let t=e;if(e.report)try{t=typeof e.report=="string"?JSON.parse(e.report):e.report}catch{t=e.report}this.onReportCallback&&this.onReportCallback(t);break}case"ERROR":this.onErrorCallback&&this.onErrorCallback(e.error||"Unknown WASM runtime error");break}}ready(){return this.isReady?Promise.resolve(!0):new Promise(e=>{this.onReadyCallbacks.push(()=>e(!0)),setTimeout(()=>e(this.isReady),5e3)})}isEngineReady(){return this.isReady}validateYaml(e,t){if(this.onValidationCallback=t,!this.worker||!this.isReady){const o=e.includes("database:"),s=e.includes("operations:"),i=o&&s;t({valid:i,error:i?void 0:'YAML deve conter os blocos "database:" e "operations:".',name:"heuristic-validation"});return}this.worker.postMessage({action:"VALIDATE",yamlContent:e})}runScenario(e,t,o,s){if(this.onProgressCallback=t,this.onReportCallback=o,this.onErrorCallback=s,!this.worker||!this.isReady){s("Motor WASM ainda inicializando ou indisponível.");return}this.worker.postMessage({action:"RUN",config:{yamlContent:e.yamlContent,workers:e.workers,iterations:e.iterations,jitterMs:e.jitterMs,seed:e.seed}})}cancel(){this.worker&&this.worker.postMessage({action:"CANCEL"})}terminate(){this.worker&&(this.worker.terminate(),this.worker=null,this.isReady=!1)}}let E=null;function W(){return E||(E=new R),E}export{A as P,y as R,_ as S,k as b,W as g};
//# sourceMappingURL=wasm-bridge-B2O4-O7V.js.map
