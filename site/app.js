// ChaosSQL — Studio Bregalda Interactive Controller (v1.2.0)
// High-fidelity Multi-View Portal, Hash Router, Docs Hub, Scenarios with Fixes, & Trace Visualizer
// Dynamic Bilingual i18n Engine: [ PT | EN ]

// ============================================================================
// 1. INTERNATIONALIZATION DICTIONARIES (PT & EN)
// ============================================================================
const I18N = {
  "pt": {
    "nav": {
      "home": "Início",
      "docs": "Documentação",
      "scenarios": "Cenários & Soluções",
      "visualizer": "Trace Visualizer",
      "matrix": "Matriz de Isolamento"
    },
    "landing": {
      "heroMeta": "Studio Bregalda — Engenharia de Sistemas — Versão 1.2.0",
      "heroTitle": "Fuzzer Determinístico de Concorrência & Isolamento SQL",
      "heroSubtitle": "O ChaosSQL injeta micro-jitter estocástico para disparar condições de corrida raras em bancos de dados, classifica anomalias de isolamento via grafos de dependência Adya e reduz rastros de execução de 100 operações para reproduções 1-minimais em milissegundos.",
      "ctaDocs": "Explorar Documentação",
      "ctaScenarios": "Ver 10 Cenários",
      "copyCommandTitle": "Copiar comando de instalação",
      "termRun": "▶ Run Fuzzer",
      "termJitter": "⚡ Injetar Jitter",
      "termShrink": "🔍 Reduzir ddmin",
      "termReset": "↺ Reiniciar",
      "adyaBadge": "GRAFO DE SERIALIZAÇÃO DIRETA (DSG)",
      "adyaTitle": "Classificação de Ciclos de Adya",
      "adyaAnomalyLabel": "Anomalia:",
      "adyaThesis": "Atul Adya, Tese de Doutorado (MIT 1999)",
      "pillarsSectionLabel": "Fundações Acadêmicas & Sistemas Rigorosos",
      "pillarsTitle": "Construído sobre Pesquisa Formal de Concorrência",
      "pillarsDesc": "O ChaosSQL substitui adivinhações empíricas por teoria rigorosa de isolamento, limites probabilísticos prováveis e minimização algorítmica de casos de teste.",
      "pillar1Title": "Classificação de Ciclos Adya",
      "pillar1Body": "Constrói Grafos de Serialização Direta dinâmicos DSG(H) = (V, E) rastreando arestas de escrita-leitura (wr), escrita-escrita (ww) e anti-dependências de leitura-escrita (rw) para categorizar anomalias (P4, A5B, A5A, G0, G2).",
      "pillar2Title": "Escalonamento com Prioridade PCT-SQL",
      "pillar2Body": "Adapta o Teste de Concorrência Probabilístico para motores SQL. Garante matematicamente encontrar bugs de isolamento de profundidade d com limite inferior P ≥ 1/(n · k^(d-1)), evitando deadlocks e inanição de threads.",
      "pillar3Title": "Delta-Debugging Causal (ddmin)",
      "pillar3Body": "Particiona históricos ruidosos de 100+ operações concorrentes para isolar o subconjunto exato 1-minimal de 2 passos necessário para reproduzir a anomalia em <200ms.",
      "pillar4Title": "Visualizador de Traces Interativo",
      "pillar4Body": "Motor HTTP embutido na porta 8090 com raias Gantt em microssegundos por goroutine, grafos de conflito Adya interativos e inspetor profundo de queries com parâmetros.",
      "pillar5Title": "OASIS SARIF 2.1.0 & GitHub",
      "pillar5Body": "Converte falhas de isolamento do banco diretamente em alertas de segurança no GitHub Code Scanning com 9 regras formais, resultados JUnit XML e resumos de steps de PR.",
      "pillar6Title": "Zero CGO & Go Puro",
      "pillar6Body": "Compilado com CGO_ENABLED=0 usando modernc.org/sqlite. Binário totalmente estático com zero bibliotecas C externas e vazão em micro-benchmark de 13.9M operações/segundo.",
      "bannerDocsBadge": "Documentação Técnica Completa",
      "bannerDocsTitle": "Centro de Engenharia & Referência DSL",
      "bannerDocsDesc": "Consulte os 8 capítulos completos: Guia de Início Rápido, DSL do chaos.yaml, manual com 12 flags da CLI, arquitetura de drivers (Postgres, SQLite, MySQL) e Go Testing SDK.",
      "bannerDocsCta": "Acessar Documentação →",
      "bannerVizBadge": "Showcase do Visualizador de Traces",
      "bannerVizTitle": "Concorrência em Nível de Microssegundos",
      "bannerVizDesc": "Explore interativamente o chaossql ui: compare o trace bruto com a redução 1-minimal de 2 passos, inspecione parâmetros de queries e navegue pelo grafo de conflitos Adya.",
      "bannerVizCta": "Abrir Trace Visualizer →"
    },
    "docs": {
      "searchPlaceholder": "Buscar tópicos na documentação...",
      "breadcrumbDocs": "Documentação",
      "prevBtn": "← Capítulo Anterior",
      "nextBtn": "Próximo Capítulo →",
      "noResults": "Nenhum capítulo encontrado.",
      "copy": "Copiar",
      "copied": "Copiado!"
    },
    "scenarios": {
      "sectionLabel": "Catálogo Interativo & Produção",
      "sectionTitle": "10 Cenários de Demonstração & Mitigações",
      "sectionDesc": "Explore anomalias de isolamento reais em livros bancários, inventário, plantões médicos e corretoras cripto — com redução causal e código de mitigação em produção.",
      "tabSchema": "Código SQL",
      "tabChaos": "Carga de Caos",
      "tabInvariant": "Regra de Invariante",
      "tabFix": "Fix & Mitigação no Banco",
      "copySql": "Copiar SQL",
      "copyYaml": "Copiar YAML",
      "copyFix": "Copiar Fix SQL",
      "formalGraphTitle": "Grafo de Conflito Formal (Adya)",
      "metric1Label": "1-Minimal Ops Shrunk",
      "metric2Label": "Ruído Causal Removido",
      "metric3Label": "Tempo de Convergência",
      "fixHeaderPill": "PRODUÇÃO • RECOMENDAÇÃO ARQUITETURAL",
      "validatedEngines": "Motores validados:",
      "driverNotes": "Notas do Driver:"
    },
    "visualizer": {
      "sectionLabel": "chaossql ui • Live Concurrency Explorer • 127.0.0.1:8090",
      "sectionTitle": "Visualizador de Traces Interativo",
      "sectionDesc": "Servidor web de observabilidade de alta resolução. Analise o entrelaçamento de goroutines em escala de microssegundos, o Grafo de Forças Adya e compare o trace original de 20 passos com a síntese minimal de 2 passos.",
      "modeRaw": "Raw Trace (20 ops)",
      "modeShrunk": "1-Minimal Shrunk (2 ops)",
      "filterAll": "Todos os Workers",
      "animateBtn": "▶ Iniciar Simulação",
      "adyaTitle": "Grafo de Dependências Adya (DSG)",
      "cycleLabel": "Ciclo: rw ∘ ww",
      "statusDetected": "P4_LOST_UPDATE detectado em t=184μs",
      "inspectorTitle": "Inspetor de Operações & Queries",
      "inspectorTx": "Transação / Worker:",
      "inspectorTimestamp": "Timestamp / Latência:",
      "inspectorExecution": "de execução",
      "inspectorParams": "Parâmetros / Variáveis:",
      "inspectorGraph": "Grafo de Conflito:",
      "inspectorCycleDetected": "T1 ──(rw)──► T2 ──(ww)──► T1 [CICLO DETECTADO]",
      "inspectorSerializable": "Linha serializável sem ciclos",
      "workerLabel": "Worker",
      "collisionLabel": "Colisão P4"
    },
    "matrix": {
      "sectionLabel": "Verificação Empírica",
      "sectionTitle": "Matriz de Anomalias de Isolamento Hermitage",
      "sectionDesc": "Comparação empírica dos limites de isolamento entre motores SQL através de baterias automáticas de testes concorrentes fuzzed com ChaosSQL (chaossql matrix).",
      "thAnomaly": "Anomalia / Fenômeno",
      "thCycle": "Ciclo Formal",
      "thSqlite": "SQLite (WAL)",
      "thPostgresRc": "PostgreSQL (RC)",
      "thPostgresSsi": "PostgreSQL (SSI)",
      "thMysql": "MySQL (InnoDB RR)",
      "statusPermitted": "PERMITIDO",
      "statusPrevented": "PREVENIDO",
      "statusDetected": "DETECTADO"
    },
    "footer": {
      "desc": "Studio Bregalda constrói ferramentas meticulosas para problemas reais de engenharia de sistemas.",
      "license": "Licença MIT"
    },
    "terminal": {
      "initFuzzer": "# Inicializando fuzzer de concorrência PCT-SQL (4 workers, 20 iterações, seed=42)...",
      "injectJitter": "# Injetando micro-jitter [1ms, 5ms] no driver em memória SQLite (Zero CGO)...",
      "anomalyDetected": "✘ ANOMALIA DE ISOLAMENTO DETECTADA: P4_LOST_UPDATE",
      "cycle": "  Ciclo: T1 ──(rw)──► T2 ──(ww)──► T1",
      "violatedInvariant": "  Invariante Violada: total_balance == 1000 (Valor real: 850)",
      "startDdmin": "▶ Iniciando Delta-Debugging Causal (ddmin)...",
      "iteration1": "  [Iteração 1] Testando subconjunto de 10 operações ──► <span class=\"term-err\">FALHA (Anomalia Preservada)</span>",
      "iteration2": "  [Iteração 2] Testando subconjunto de 4 operações  ──► <span class=\"term-err\">FALHA (Anomalia Preservada)</span>",
      "iteration3": "  [Iteração 3] Testando subconjunto de 2 operações  ──► <span class=\"term-err\">FALHA (1-minimal alcançado)</span>",
      "traceShrunk": "✔ Rastro reduzido de 20 para 2 operações (90.0% de redução em 68ms)",
      "synthesizedRepro": "  Repro autônomo sintetizado: bin/repro_test.go",
      "pillWorkers": "Workers:",
      "pillEngine": "Motor:",
      "pillReduction": "Redução:",
      "pillLatency": "Latência:"
    }
  },
  "en": {
    "nav": {
      "home": "Home",
      "docs": "Documentation",
      "scenarios": "Scenarios & Fixes",
      "visualizer": "Trace Visualizer",
      "matrix": "Isolation Matrix"
    },
    "landing": {
      "heroMeta": "Studio Bregalda — Systems Engineering — Version 1.2.0",
      "heroTitle": "Deterministic Concurrency & Isolation Fuzzer",
      "heroSubtitle": "ChaosSQL injects stochastic micro-jitter to trigger rare database race conditions, classifies isolation anomalies through Adya dependency graphs, and shrinks 100-operation execution traces to 1-minimal reproductions in milliseconds.",
      "ctaDocs": "Explore Documentation",
      "ctaScenarios": "View 10 Scenarios",
      "copyCommandTitle": "Copy installation command",
      "termRun": "▶ Run Fuzzer",
      "termJitter": "⚡ Inject Jitter",
      "termShrink": "🔍 ddmin Shrink",
      "termReset": "↺ Reset",
      "adyaBadge": "DIRECT SERIALIZATION GRAPH (DSG)",
      "adyaTitle": "Adya Cycle Classification",
      "adyaAnomalyLabel": "Anomaly:",
      "adyaThesis": "Atul Adya, Ph.D. Thesis (MIT 1999)",
      "pillarsSectionLabel": "Academic Foundations & Rigorous Systems",
      "pillarsTitle": "Built on Formal Concurrency Research",
      "pillarsDesc": "ChaosSQL replaces heuristic guesswork with rigorous isolation theory, provable scheduling bounds, and algorithmic test case minimization.",
      "pillar1Title": "Adya Cycle Classification",
      "pillar1Body": "Constructs dynamic Direct Serialization Graphs DSG(H) = (V, E) tracking write-read (wr), write-write (ww), and read-write anti-dependency (rw) edges to categorize anomalies (P4, A5B, A5A, G0, G2).",
      "pillar2Title": "PCT-SQL Priority Scheduling",
      "pillar2Body": "Adapts Probabilistic Concurrency Testing to SQL engines. Mathematically guarantees finding isolation bugs of bug-depth d with lower bound P ≥ 1/(n · k^(d-1)), avoiding deadlocks and thread starvation.",
      "pillar3Title": "Causal Delta-Debugging (ddmin)",
      "pillar3Body": "Partitions noisy execution histories of 100+ concurrent operations to isolate the exact 1-minimal subset of 2 steps required to reproduce the anomaly in <200ms.",
      "pillar4Title": "Interactive Trace Visualizer",
      "pillar4Body": "Embedded HTTP engine on port 8090 with microsecond Gantt swimlanes per goroutine, interactive Adya conflict graphs, and deep parameter query inspector.",
      "pillar5Title": "OASIS SARIF 2.1.0 & GitHub",
      "pillar5Body": "Converts database isolation flaws directly into security findings in GitHub Code Scanning with 9 formal rules, JUnit XML test results, and PR step summaries.",
      "pillar6Title": "Zero CGO & Pure Go",
      "pillar6Body": "Compiled with CGO_ENABLED=0 using modernc.org/sqlite. Completely static binary with zero external C libraries and micro-benchmark throughput of 13.9M operations/sec.",
      "bannerDocsBadge": "Complete Technical Documentation",
      "bannerDocsTitle": "Engineering Hub & DSL Reference",
      "bannerDocsDesc": "Explore all 8 chapters: Quickstart Guide, chaos.yaml DSL, manual with 12 CLI flags, driver architecture (Postgres, SQLite, MySQL), and Go Testing SDK.",
      "bannerDocsCta": "Explore Documentation →",
      "bannerVizBadge": "Trace Visualizer Showcase",
      "bannerVizTitle": "Microsecond-Level Concurrency",
      "bannerVizDesc": "Interactively explore chaossql ui: compare the raw trace with the 1-minimal 2-step reduction, inspect query parameters, and navigate the Adya conflict graph.",
      "bannerVizCta": "Open Trace Visualizer →"
    },
    "docs": {
      "searchPlaceholder": "Search documentation topics...",
      "breadcrumbDocs": "Docs",
      "prevBtn": "← Previous Chapter",
      "nextBtn": "Next Chapter →",
      "noResults": "No chapters found.",
      "copy": "Copy",
      "copied": "Copied!"
    },
    "scenarios": {
      "sectionLabel": "Interactive Catalog & Production",
      "sectionTitle": "10 Flagship Demonstration Scenarios & Fixes",
      "sectionDesc": "Explore real isolation anomalies in banking ledgers, inventory, hospital on-call rosters, and crypto exchanges — with causal reduction and production mitigation code.",
      "tabSchema": "SQL Code",
      "tabChaos": "Chaos Workload",
      "tabInvariant": "Invariant Rule",
      "tabFix": "Production Fix & Mitigation",
      "copySql": "Copy SQL",
      "copyYaml": "Copy YAML",
      "copyFix": "Copy Fix SQL",
      "formalGraphTitle": "Formal Conflict Graph (Adya)",
      "metric1Label": "1-Minimal Ops Shrunk",
      "metric2Label": "Causal Noise Removed",
      "metric3Label": "Convergence Time",
      "fixHeaderPill": "PRODUCTION — ARCHITECTURAL RECOMMENDATION",
      "validatedEngines": "Validated engines:",
      "driverNotes": "Driver Notes:"
    },
    "visualizer": {
      "sectionLabel": "chaossql ui — Live Concurrency Explorer — 127.0.0.1:8090",
      "sectionTitle": "Interactive Trace Visualizer",
      "sectionDesc": "High-resolution observability web engine. Analyze goroutine interleaving at microsecond scale, the Adya Conflict Graph, and compare the 20-step raw trace with the 2-step 1-minimal synthesis.",
      "modeRaw": "Raw Trace (20 ops)",
      "modeShrunk": "1-Minimal Shrunk (2 ops)",
      "filterAll": "All Workers",
      "animateBtn": "▶ Animate Execution",
      "adyaTitle": "Adya Dependency Graph (DSG)",
      "cycleLabel": "Cycle: rw → ww",
      "statusDetected": "P4_LOST_UPDATE detected at t=184μs",
      "inspectorTitle": "Operation & Query Inspector",
      "inspectorTx": "Transaction / Worker:",
      "inspectorTimestamp": "Timestamp / Latency:",
      "inspectorExecution": "execution",
      "inspectorParams": "Parameters / Variables:",
      "inspectorGraph": "Conflict Graph:",
      "inspectorCycleDetected": "T1 ──(rw)──► T2 ──(ww)──► T1 [CYCLE DETECTED]",
      "inspectorSerializable": "Serializable schedule with no cycles",
      "workerLabel": "Worker",
      "collisionLabel": "P4 Collision"
    },
    "matrix": {
      "sectionLabel": "Empirical Verification",
      "sectionTitle": "Hermitage Isolation Anomaly Matrix",
      "sectionDesc": "Empirical comparison of isolation boundaries across SQL engines through automated suites of concurrent fuzzed test cases with ChaosSQL (chaossql matrix).",
      "thAnomaly": "Anomaly / Phenomenon",
      "thCycle": "Formal Cycle",
      "thSqlite": "SQLite (WAL)",
      "thPostgresRc": "PostgreSQL (RC)",
      "thPostgresSsi": "PostgreSQL (SSI)",
      "thMysql": "MySQL (InnoDB RR)",
      "statusPermitted": "PERMITTED",
      "statusPrevented": "PREVENTED",
      "statusDetected": "DETECTED"
    },
    "footer": {
      "desc": "Studio Bregalda builds thoughtful tools for real systems engineering problems.",
      "license": "MIT License"
    },
    "terminal": {
      "initFuzzer": "# Initializing PCT-SQL concurrency fuzzer (4 workers, 20 iterations, seed=42)...",
      "injectJitter": "# Injecting micro-jitter [1ms, 5ms] on SQLite in-memory driver (Zero CGO)...",
      "anomalyDetected": "✘ ISOLATION ANOMALY DETECTED: P4_LOST_UPDATE",
      "cycle": "  Cycle: T1 ──(rw)──► T2 ──(ww)──► T1",
      "violatedInvariant": "  Violated Invariant: total_balance == 1000 (Actual: 850)",
      "startDdmin": "▶ Starting Causal Delta-Debugging (ddmin)...",
      "iteration1": "  [Iteration 1] Testing subset of 10 operations ──► <span class=\"term-err\">FAIL (Anomaly Preserved)</span>",
      "iteration2": "  [Iteration 2] Testing subset of 4 operations  ──► <span class=\"term-err\">FAIL (Anomaly Preserved)</span>",
      "iteration3": "  [Iteration 3] Testing subset of 2 operations  ──► <span class=\"term-err\">FAIL (1-minimal achieved)</span>",
      "traceShrunk": "✔ Trace shrunk from 20 to 2 operations (90.0% reduction in 68ms)",
      "synthesizedRepro": "  Synthesized standalone repro: bin/repro_test.go",
      "pillWorkers": "Workers:",
      "pillEngine": "Engine:",
      "pillReduction": "Reduction:",
      "pillLatency": "Latency:"
    }
  }
};

// ============================================================================
// 2. SCENARIOS DATABASE (10 Flagship Scenarios - Bilingual Format)
// ============================================================================
const SCENARIOS = [
  {
    "id": "banking",
    "name": {
      "pt": "Lost Update Bancário",
      "en": "Banking Lost Update"
    },
    "code": "P4",
    "description": {
      "pt": "Saques concorrentes sob READ COMMITTED sobrescrevem saldos sem serialização transacional, perdendo quantias silenciosamente.",
      "en": "Concurrent withdrawals under READ COMMITTED overwrite balances without transactional serialization, silently losing funds."
    },
    "summary": {
      "pt": "Saques concorrentes sob READ COMMITTED sobrescrevem saldos sem serialização transacional, perdendo quantias silenciosamente.",
      "en": "Concurrent withdrawals under READ COMMITTED overwrite balances without transactional serialization, silently losing funds."
    },
    "schema": "-- Schema\nCREATE TABLE accounts (\n    id INT PRIMARY KEY,\n    owner TEXT NOT NULL,\n    balance INT NOT NULL\n);\n\n-- Seed\nINSERT INTO accounts VALUES (1, 'Alice', 1000);",
    "chaos": "version: \"1.0\"\nname: \"banking_lost_update\"\ndatabase:\n  driver: \"sqlite\"\noperations:\n  - name: \"withdraw_100\"\n    steps:\n      - \"SELECT balance FROM accounts WHERE id = 1 -> cur\"\n      - \"UPDATE accounts SET balance = {cur - 100} WHERE id = 1\"\ninvariants:\n  - name: \"total_balance_check\"\n    query: \"SELECT balance FROM accounts WHERE id = 1;\"\n    assert: \"balance == 1000 - (total_completed * 100)\"",
    "reduction": {
      "originalOps": 20,
      "minimalOps": 2,
      "reductionPct": "90.0%",
      "elapsed": "68ms",
      "cycle": "T1 ──(rw)──► T2 ──(ww)──► T1",
      "explanation": "Two concurrent read-modify-write transactions read balance 1000 simultaneously. T1 writes 900, but T2 overwrites with 900 based on its stale read, silently losing $100."
    },
    "analysis": {
      "pt": "Duas transações concorrentes de leitura-modificação-escrita leem o saldo de 1000 simultaneamente. T1 escreve 900, mas T2 sobrescreve com 900 com base em sua leitura defasada, perdendo silenciosamente $100.",
      "en": "Two concurrent read-modify-write transactions read balance 1000 simultaneously. T1 writes 900, but T2 overwrites with 900 based on its stale read, silently losing $100."
    },
    "fix": {
      "pt": {
        "title": "Update Atômico com Predicado de Guarda ou Bloqueio Pessimista",
        "explanation": "A anomalia ocorre pelo padrão vulnerável Read-Modify-Write sob READ COMMITTED: a aplicação lê o valor em um SELECT e computa a subtração na memória do servidor web. Se duas goroutines lerem balance=1000 ao mesmo tempo, ambas calculam 900 e o último UPDATE sobrescreve o primeiro. A mitigação transfere o cálculo para o motor de banco em uma única instrução atômica (UPDATE accounts SET balance = balance - 100 WHERE id = 1 AND balance >= 100), serializando a alteração sob o bloqueio exclusivo de linha.",
        "code": "-- Opção 1: Update Atômico com Aritmética no Motor (Recomendado)\nUPDATE accounts \nSET balance = balance - 100 \nWHERE id = 1 AND balance >= 100;\n\n-- Opção 2: Bloqueio Pessimista dentro de Transação Explícita\nBEGIN;\nSELECT balance FROM accounts WHERE id = 1 FOR UPDATE;\n-- Aplicação valida balance >= 100 antes de submeter escrita:\nUPDATE accounts SET balance = balance - 100 WHERE id = 1;\nCOMMIT;",
        "driverNotes": "PostgreSQL, SQLite (WAL), MySQL (InnoDB)"
      },
      "en": {
        "title": "Atomic Guarded Decrement or Pessimistic Row Locking",
        "explanation": "The anomaly is caused by the vulnerable Read-Modify-Write anti-pattern under READ COMMITTED: the application reads the balance in a SELECT statement and calculates subtraction in web server memory. If two goroutines read balance=1000 concurrently, both compute 900 and the last UPDATE overwrites the first. The mitigation delegates arithmetic directly to the database engine in a single atomic statement (UPDATE accounts SET balance = balance - 100 WHERE id = 1 AND balance >= 100), serializing execution under row exclusive locks.",
        "code": "-- Option 1: Atomic Database-Side Arithmetic (Recommended)\nUPDATE accounts \nSET balance = balance - 100 \nWHERE id = 1 AND balance >= 100;\n\n-- Option 2: Explicit Transaction with Pessimistic Locking\nBEGIN;\nSELECT balance FROM accounts WHERE id = 1 FOR UPDATE;\n-- Application verifies balance >= 100 before issuing update:\nUPDATE accounts SET balance = balance - 100 WHERE id = 1;\nCOMMIT;",
        "driverNotes": "PostgreSQL, SQLite (WAL), MySQL (InnoDB)"
      },
      "strategy": "Update Atômico com Predicado de Guarda ou Bloqueio Pessimista",
      "sql": "-- Opção 1: Update Atômico com Aritmética no Motor (Recomendado)\nUPDATE accounts \nSET balance = balance - 100 \nWHERE id = 1 AND balance >= 100;\n\n-- Opção 2: Bloqueio Pessimista dentro de Transação Explícita\nBEGIN;\nSELECT balance FROM accounts WHERE id = 1 FOR UPDATE;\n-- Aplicação valida balance >= 100 antes de submeter escrita:\nUPDATE accounts SET balance = balance - 100 WHERE id = 1;\nCOMMIT;",
      "explanation": "A anomalia ocorre pelo padrão vulnerável Read-Modify-Write sob READ COMMITTED: a aplicação lê o valor em um SELECT e computa a subtração na memória do servidor web. Se duas goroutines lerem balance=1000 ao mesmo tempo, ambas calculam 900 e o último UPDATE sobrescreve o primeiro. A mitigação transfere o cálculo para o motor de banco em uma única instrução atômica (UPDATE accounts SET balance = balance - 100 WHERE id = 1 AND balance >= 100), serializando a alteração sob o bloqueio exclusivo de linha.",
      "engines": [
        "PostgreSQL",
        "SQLite (WAL)",
        "MySQL (InnoDB)"
      ]
    }
  },
  {
    "id": "inventory",
    "name": {
      "pt": "Venda Excessiva de Inventário (Oversell)",
      "en": "Inventory Oversell"
    },
    "code": "A3",
    "description": {
      "pt": "Checkouts simultâneos reservam o último item disponível devido a anti-dependências de predicado (leituras fantasmas).",
      "en": "Concurrent checkouts reserve the final available item due to predicate anti-dependencies (phantom reads)."
    },
    "summary": {
      "pt": "Checkouts simultâneos reservam o último item disponível devido a anti-dependências de predicado (leituras fantasmas).",
      "en": "Concurrent checkouts reserve the final available item due to predicate anti-dependencies (phantom reads)."
    },
    "schema": "-- Schema\nCREATE TABLE products (\n    id INT PRIMARY KEY,\n    name TEXT NOT NULL,\n    stock INT NOT NULL\n);\n\n-- Seed\nINSERT INTO products VALUES (1, 'RTX 5090 GPU', 1);",
    "chaos": "version: \"1.0\"\nname: \"inventory_oversell\"\noperations:\n  - name: \"checkout\"\n    steps:\n      - \"SELECT stock FROM products WHERE id = 1 -> avail\"\n      - \"UPDATE products SET stock = {avail - 1} WHERE id = 1 AND {avail > 0}\"\ninvariants:\n  - name: \"stock_non_negative\"\n    query: \"SELECT stock FROM products WHERE id = 1;\"\n    assert: \"stock >= 0\"",
    "reduction": {
      "originalOps": 30,
      "minimalOps": 2,
      "reductionPct": "93.3%",
      "elapsed": "74ms",
      "cycle": "T1 ──(rw)──► T2 ──(ww)──► T1",
      "explanation": "Both workers evaluate {avail > 0} as true before either commit updates the physical stock row, driving inventory to -1."
    },
    "analysis": {
      "pt": "Ambos os workers avaliam {avail > 0} como verdadeiro antes que qualquer commit atualize o estoque físico em disco, levando o inventário a -1.",
      "en": "Both workers evaluate {avail > 0} as true before either commit updates the physical stock row, driving inventory to -1."
    },
    "fix": {
      "pt": {
        "title": "Guarded Decrement Atômico com Verificação de Linhas Afetadas",
        "explanation": "A condição de guarda {avail > 0} avaliada na memória da aplicação é obsoleta (stale) no momento em que a escrita chega ao disco. Ao mover a verificação diretamente para o WHERE da instrução UPDATE (AND stock >= 1), o motor de banco garante a avaliação atômica do predicado e o bloqueio de escrita sobre a linha. Se múltiplos clientes tentarem comprar a última unidade simultaneamente, exatamente um decrementará o estoque para 0 e os demais receberão 0 linhas afetadas.",
        "code": "-- Decremento Atômico Protegido no Banco de Dados\nUPDATE products \nSET stock = stock - 1 \nWHERE id = 1 AND stock >= 1;\n\n-- No código de aplicação (Go / Node):\n-- Verificar se rows_affected == 1.\n-- Se rows_affected == 0, retornar ErrProdutoEsgotado sem prosseguir para pagamento.",
        "driverNotes": "PostgreSQL, SQLite, MySQL (InnoDB)"
      },
      "en": {
        "title": "Atomic Guarded Decrement with Rows-Affected Check",
        "explanation": "The guard condition {avail > 0} evaluated in application memory becomes stale by the time writes hit disk. By pushing the check into the WHERE clause of the UPDATE statement (AND stock >= 1), the database engine guarantees atomic predicate verification and exclusive row locks. When multiple clients purchase the final item concurrently, exactly one decrements stock to 0 while others receive 0 rows affected.",
        "code": "-- Atomic Guarded Decrement in Database Engine\nUPDATE products \nSET stock = stock - 1 \nWHERE id = 1 AND stock >= 1;\n\n-- In application code (Go / Node):\n-- Verify rows_affected == 1.\n-- If rows_affected == 0, return ErrOutOfStock without charging payment.",
        "driverNotes": "PostgreSQL, SQLite, MySQL (InnoDB)"
      },
      "strategy": "Guarded Decrement Atômico com Verificação de Linhas Afetadas",
      "sql": "-- Decremento Atômico Protegido no Banco de Dados\nUPDATE products \nSET stock = stock - 1 \nWHERE id = 1 AND stock >= 1;\n\n-- No código de aplicação (Go / Node):\n-- Verificar se rows_affected == 1.\n-- Se rows_affected == 0, retornar ErrProdutoEsgotado sem prosseguir para pagamento.",
      "explanation": "A condição de guarda {avail > 0} avaliada na memória da aplicação é obsoleta (stale) no momento em que a escrita chega ao disco. Ao mover a verificação diretamente para o WHERE da instrução UPDATE (AND stock >= 1), o motor de banco garante a avaliação atômica do predicado e o bloqueio de escrita sobre a linha. Se múltiplos clientes tentarem comprar a última unidade simultaneamente, exatamente um decrementará o estoque para 0 e os demais receberão 0 linhas afetadas.",
      "engines": [
        "PostgreSQL",
        "SQLite",
        "MySQL (InnoDB)"
      ]
    }
  },
  {
    "id": "hospital",
    "name": {
      "pt": "Desvio de Escrita Hospitalar (Write Skew)",
      "en": "Hospital Write Skew"
    },
    "code": "A5B",
    "description": {
      "pt": "Dois médicos cancelam plantão simultaneamente sob Snapshot Isolation porque seus conjuntos de escrita não colidem.",
      "en": "Two doctors independently sign off duty under Snapshot Isolation because their write sets do not overlap."
    },
    "summary": {
      "pt": "Dois médicos cancelam plantão simultaneamente sob Snapshot Isolation porque seus conjuntos de escrita não colidem.",
      "en": "Two doctors independently sign off duty under Snapshot Isolation because their write sets do not overlap."
    },
    "schema": "-- Schema\nCREATE TABLE doctors (\n    id INT PRIMARY KEY,\n    name TEXT NOT NULL,\n    on_duty BOOLEAN NOT NULL\n);\n\n-- Seed\nINSERT INTO doctors VALUES (1, 'Dr. Alice', true), (2, 'Dr. Bob', true);",
    "chaos": "version: \"1.0\"\nname: \"hospital_write_skew\"\noperations:\n  - name: \"sign_off_alice\"\n    steps:\n      - \"SELECT count(*) AS active FROM doctors WHERE on_duty = true -> act\"\n      - \"UPDATE doctors SET on_duty = false WHERE id = 1 AND {act >= 2}\"\n  - name: \"sign_off_bob\"\n    steps:\n      - \"SELECT count(*) AS active FROM doctors WHERE on_duty = true -> act\"\n      - \"UPDATE doctors SET on_duty = false WHERE id = 2 AND {act >= 2}\"\ninvariants:\n  - name: \"at_least_one_doctor_on_duty\"\n    query: \"SELECT count(*) AS active FROM doctors WHERE on_duty = true;\"\n    assert: \"active >= 1\"",
    "reduction": {
      "originalOps": 40,
      "minimalOps": 2,
      "reductionPct": "95.0%",
      "elapsed": "82ms",
      "cycle": "T1 ──(rw)──► T2 ──(rw)──► T1",
      "explanation": "Under Snapshot Isolation, T1 and T2 read the same snapshot (2 active doctors). T1 updates Alice, T2 updates Bob. Because their write sets are disjoint, standard SI permits both commits, leaving 0 doctors on duty."
    },
    "analysis": {
      "pt": "Sob Snapshot Isolation, T1 e T2 leem o mesmo snapshot (2 médicos ativos). T1 altera Alice, T2 altera Bob. Como seus conjuntos de escrita são disjuntos, o SI padrão permite ambos os commits, deixando o hospital sem médicos.",
      "en": "Under Snapshot Isolation, T1 and T2 read the same snapshot (2 active doctors). T1 updates Alice, T2 updates Bob. Because their write sets are disjoint, standard SI permits both commits, leaving 0 doctors on duty."
    },
    "fix": {
      "pt": {
        "title": "Elevação para Serializable Snapshot Isolation (SSI) ou Bloqueio de Conflito",
        "explanation": "Write Skew (A5B) é o exemplo clássico de falha sob Snapshot Isolation (SI): como T1 altera a linha 1 e T2 altera a linha 2, os conjuntos de escrita não colidem (sem conflito ww), mas geram um ciclo de anti-dependência de leitura-escrita (rw). A solução canônica é configurar SERIALIZABLE (onde motores modernos com SSI rastreiam dependências rw e abortam transações concorrentes com serialization_failure) ou serializar os acessos através de um lock FOR UPDATE em uma linha mestre de escala.",
        "code": "-- Solução A: Elevação do Nível de Isolamento para SERIALIZABLE (SSI)\nSET TRANSACTION ISOLATION LEVEL SERIALIZABLE;\nBEGIN;\nSELECT count(*) FROM doctors WHERE on_duty = true;\n-- Se count >= 2:\nUPDATE doctors SET on_duty = false WHERE id = 1;\nCOMMIT; -- O motor PostgreSQL detectará o ciclo rw e abortará T2 com código 40001!\n\n-- Solução B: Bloqueio Pessimista Compartilhado em Linha Pai de Plantão\nBEGIN;\nSELECT shift_id FROM duty_shifts WHERE shift_id = 1 FOR UPDATE;\nSELECT count(*) FROM doctors WHERE on_duty = true;\nUPDATE doctors SET on_duty = false WHERE id = 1;\nCOMMIT;",
        "driverNotes": "PostgreSQL (SSI), CockroachDB, MySQL (SERIALIZABLE)"
      },
      "en": {
        "title": "Elevation to Serializable Snapshot Isolation (SSI) or Materialized Conflict Lock",
        "explanation": "Write Skew (A5B) is the canonical failure mode of Snapshot Isolation (SI): since T1 writes row 1 and T2 writes row 2, write sets are disjoint (no ww conflict), but generate a cyclic read-write anti-dependency (rw) loop. The standard solution is SERIALIZABLE isolation (where modern SSI engines track rw dependencies and abort concurrent transactions with SQLSTATE 40001 serialization_failure) or synchronizing transactions via a parent roster lock with FOR UPDATE.",
        "code": "-- Solution A: Elevate Isolation Level to SERIALIZABLE (SSI)\nSET TRANSACTION ISOLATION LEVEL SERIALIZABLE;\nBEGIN;\nSELECT count(*) FROM doctors WHERE on_duty = true;\n-- If count >= 2:\nUPDATE doctors SET on_duty = false WHERE id = 1;\nCOMMIT; -- Engine detects rw conflict cycle and aborts T2 with 40001!\n\n-- Solution B: Shared Pessimistic Lock on Parent Shift Row\nBEGIN;\nSELECT shift_id FROM duty_shifts WHERE shift_id = 1 FOR UPDATE;\nSELECT count(*) FROM doctors WHERE on_duty = true;\nUPDATE doctors SET on_duty = false WHERE id = 1;\nCOMMIT;",
        "driverNotes": "PostgreSQL (SSI), CockroachDB, MySQL (SERIALIZABLE)"
      },
      "strategy": "Elevação para Serializable Snapshot Isolation (SSI) ou Bloqueio de Conflito",
      "sql": "-- Solução A: Elevação do Nível de Isolamento para SERIALIZABLE (SSI)\nSET TRANSACTION ISOLATION LEVEL SERIALIZABLE;\nBEGIN;\nSELECT count(*) FROM doctors WHERE on_duty = true;\n-- Se count >= 2:\nUPDATE doctors SET on_duty = false WHERE id = 1;\nCOMMIT; -- O motor PostgreSQL detectará o ciclo rw e abortará T2 com código 40001!\n\n-- Solução B: Bloqueio Pessimista Compartilhado em Linha Pai de Plantão\nBEGIN;\nSELECT shift_id FROM duty_shifts WHERE shift_id = 1 FOR UPDATE;\nSELECT count(*) FROM doctors WHERE on_duty = true;\nUPDATE doctors SET on_duty = false WHERE id = 1;\nCOMMIT;",
      "explanation": "Write Skew (A5B) é o exemplo clássico de falha sob Snapshot Isolation (SI): como T1 altera a linha 1 e T2 altera a linha 2, os conjuntos de escrita não colidem (sem conflito ww), mas geram um ciclo de anti-dependência de leitura-escrita (rw). A solução canônica é configurar SERIALIZABLE (onde motores modernos com SSI rastreiam dependências rw e abortam transações concorrentes com serialization_failure) ou serializar os acessos através de um lock FOR UPDATE em uma linha mestre de escala.",
      "engines": [
        "PostgreSQL (SSI)",
        "CockroachDB",
        "MySQL (SERIALIZABLE)"
      ]
    }
  },
  {
    "id": "financial",
    "name": {
      "pt": "Distorção de Leitura Financeira (Read Skew)",
      "en": "Financial Read Skew"
    },
    "code": "A5A",
    "description": {
      "pt": "Uma consulta de auditoria observa uma transferência parcialmente aplicada, detectando uma discrepância espúria no patrimônio total.",
      "en": "An audit query observes a partially applied transfer, detecting an artificial discrepancy in total wealth."
    },
    "summary": {
      "pt": "Uma consulta de auditoria observa uma transferência parcialmente aplicada, detectando uma discrepância espúria no patrimônio total.",
      "en": "An audit query observes a partially applied transfer, detecting an artificial discrepancy in total wealth."
    },
    "schema": "-- Schema\nCREATE TABLE accounts (\n    id INT PRIMARY KEY,\n    balance INT NOT NULL\n);\n\n-- Seed\nINSERT INTO accounts VALUES (1, 1000), (2, 1000);",
    "chaos": "version: \"1.0\"\nname: \"financial_read_skew\"\noperations:\n  - name: \"transfer_100\"\n    steps:\n      - \"UPDATE accounts SET balance = balance - 100 WHERE id = 1\"\n      - \"UPDATE accounts SET balance = balance + 100 WHERE id = 2\"\n  - name: \"audit_total\"\n    steps:\n      - \"SELECT balance FROM accounts WHERE id = 1 -> b1\"\n      - \"SELECT balance FROM accounts WHERE id = 2 -> b2\"\ninvariants:\n  - name: \"total_wealth_preserved\"\n    query: \"SELECT sum(balance) AS total FROM accounts;\"\n    assert: \"total == 2000\"",
    "reduction": {
      "originalOps": 25,
      "minimalOps": 3,
      "reductionPct": "88.0%",
      "elapsed": "61ms",
      "cycle": "T1 ──(rw)──► T2 ──(wr)──► T1",
      "explanation": "Under READ COMMITTED, each SELECT inside the audit transaction takes a separate snapshot. Audit reads account 1 before transfer (-$100), but reads account 2 after transfer (+$100), seeing $1000 + $1100 = $2100."
    },
    "analysis": {
      "pt": "Sob READ COMMITTED, cada SELECT na transação de auditoria tira um snapshot isolado. A auditoria lê a conta 1 antes da transferência (-$100), mas lê a conta 2 após a transferência (+$100), observando $1000 + $1100 = $2100.",
      "en": "Under READ COMMITTED, each SELECT inside the audit transaction takes a separate snapshot. Audit reads account 1 before transfer (-$100), but reads account 2 after transfer (+$100), seeing $1000 + $1100 = $2100."
    },
    "fix": {
      "pt": {
        "title": "Isolamento REPEATABLE READ / SNAPSHOT para Transações de Auditoria",
        "explanation": "Sob READ COMMITTED, cada instrução SELECT cria uma nova fotografia temporal (snapshot). Se uma transação de transferência comitar entre a leitura da conta 1 e da conta 2, o relatório financeiro observará o estado de contas em momentos díspares do tempo. Ao elevar a transação de leitura para REPEATABLE READ (ou SNAPSHOT ISOLATION), o banco congela um único Point-in-Time Snapshot imutável no primeiro SELECT, garantindo consistência estrita de leitura sem necessidade de bloqueios.",
        "code": "-- Transação de Auditoria em REPEATABLE READ\nSET TRANSACTION ISOLATION LEVEL REPEATABLE READ;\n\nBEGIN;\n-- Todas as consultas subsequentes leem a mesma fotografia temporal do banco:\nSELECT balance AS bal1 FROM accounts WHERE id = 1;\n-- Mesmo que a transferência ocorra e comite aqui no meio...\nSELECT balance AS bal2 FROM accounts WHERE id = 2;\n-- bal1 + bal2 sempre somará exatamente 2000!\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL (InnoDB), SQLite (WAL)"
      },
      "en": {
        "title": "REPEATABLE READ / SNAPSHOT Isolation for Audit Transactions",
        "explanation": "Under READ COMMITTED, each SELECT statement constructs a new point-in-time snapshot. If an independent fund transfer commits between reading account 1 and account 2, the audit query sees balances at different historical moments. Elevating the audit transaction to REPEATABLE READ (or Snapshot Isolation) pins an immutable Point-in-Time Snapshot across all queries, guaranteeing consistent multi-row reads without requiring locks.",
        "code": "-- Audit Transaction in REPEATABLE READ\nSET TRANSACTION ISOLATION LEVEL REPEATABLE READ;\n\nBEGIN;\n-- All subsequent queries read from the exact same point-in-time snapshot:\nSELECT balance AS bal1 FROM accounts WHERE id = 1;\n-- Even if an external transfer commits right here...\nSELECT balance AS bal2 FROM accounts WHERE id = 2;\n-- bal1 + bal2 will always sum to exactly 2000!\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL (InnoDB), SQLite (WAL)"
      },
      "strategy": "Isolamento REPEATABLE READ / SNAPSHOT para Transações de Auditoria",
      "sql": "-- Transação de Auditoria em REPEATABLE READ\nSET TRANSACTION ISOLATION LEVEL REPEATABLE READ;\n\nBEGIN;\n-- Todas as consultas subsequentes leem a mesma fotografia temporal do banco:\nSELECT balance AS bal1 FROM accounts WHERE id = 1;\n-- Mesmo que a transferência ocorra e comite aqui no meio...\nSELECT balance AS bal2 FROM accounts WHERE id = 2;\n-- bal1 + bal2 sempre somará exatamente 2000!\nCOMMIT;",
      "explanation": "Sob READ COMMITTED, cada instrução SELECT cria uma nova fotografia temporal (snapshot). Se uma transação de transferência comitar entre a leitura da conta 1 e da conta 2, o relatório financeiro observará o estado de contas em momentos díspares do tempo. Ao elevar a transação de leitura para REPEATABLE READ (ou SNAPSHOT ISOLATION), o banco congela um único Point-in-Time Snapshot imutável no primeiro SELECT, garantindo consistência estrita de leitura sem necessidade de bloqueios.",
      "engines": [
        "PostgreSQL",
        "MySQL (InnoDB)",
        "SQLite (WAL)"
      ]
    }
  },
  {
    "id": "auction",
    "name": {
      "pt": "Escrita Suja em Leilão (Dirty Write)",
      "en": "Auction Dirty Write"
    },
    "code": "G0",
    "description": {
      "pt": "Dois licitantes concorrentes sobrescrevem colunas distintas da mesma linha de leilão sem bloqueios 2PL estritos.",
      "en": "Two concurrent bidders overwrite separate columns of the same auction row without strict 2PL locking."
    },
    "summary": {
      "pt": "Dois licitantes concorrentes sobrescrevem colunas distintas da mesma linha de leilão sem bloqueios 2PL estritos.",
      "en": "Two concurrent bidders overwrite separate columns of the same auction row without strict 2PL locking."
    },
    "schema": "-- Schema\nCREATE TABLE auctions (\n    id INT PRIMARY KEY,\n    item TEXT NOT NULL,\n    highest_bid INT NOT NULL,\n    winner TEXT NOT NULL\n);\n\n-- Seed\nINSERT INTO auctions VALUES (1, 'Rare Stamp', 100, 'Original');",
    "chaos": "version: \"1.0\"\nname: \"auction_dirty_write\"\noperations:\n  - name: \"bid_alice\"\n    steps:\n      - \"UPDATE auctions SET highest_bid = 150 WHERE id = 1\"\n      - \"UPDATE auctions SET winner = 'Alice' WHERE id = 1\"\n  - name: \"bid_bob\"\n    steps:\n      - \"UPDATE auctions SET highest_bid = 200 WHERE id = 1\"\n      - \"UPDATE auctions SET winner = 'Bob' WHERE id = 1\"\ninvariants:\n  - name: \"bid_winner_consistency\"\n    query: \"SELECT highest_bid, winner FROM auctions WHERE id = 1;\"\n    assert: \"(highest_bid == 150 AND winner == 'Alice') OR (highest_bid == 200 AND winner == 'Bob')\"",
    "reduction": {
      "originalOps": 15,
      "minimalOps": 2,
      "reductionPct": "86.7%",
      "elapsed": "49ms",
      "cycle": "T1 ──(ww)──► T2 ──(ww)──► T1",
      "explanation": "Without atomic row locking or single-statement updates, Alice writes bid 150, Bob writes bid 200, but Alice's second step writes winner='Alice', creating an inconsistent state: $200 bid won by Alice."
    },
    "analysis": {
      "pt": "Sem travas atômicas ou updates em instrução única, Alice escreve lance 150, Bob escreve lance 200, mas o segundo passo de Alice grava vencedor='Alice', criando um estado inconsistente: lance de $200 vencido por Alice.",
      "en": "Without atomic row locking or single-statement updates, Alice writes bid 150, Bob writes bid 200, but Alice's second step writes winner='Alice', creating an inconsistent state: $200 bid won by Alice."
    },
    "fix": {
      "pt": {
        "title": "Atualização Atômica Multi-Coluna com Predicado Monotônico",
        "explanation": "Dirty Writes (G0) ocorrem quando escritas não comitadas de transações simultâneas se entrelaçam sobre a mesma linha de dados. A correção definitiva consiste em unificar as atualizações de valor e titular em uma única cláusula UPDATE atômica no banco, condicionada estritamente a um lance superior ao atual (WHERE id = 1 AND highest_bid < novo_lance). Isso assegura retenção de locks 2PL estritos e garante monotonicidade sem risco de estados corrompidos híbridos.",
        "code": "-- Update Atômico em Instrução Única com Cláusula Monotônica\nBEGIN;\nUPDATE auctions \nSET highest_bid = 200, \n    winner = 'Bob' \nWHERE id = 1 AND highest_bid < 200;\n\n-- Se rows_affected == 0, o lance já foi superado concorrentemente;\n-- Fazer ROLLBACK e notificar o usuário.\nCOMMIT;",
        "driverNotes": "PostgreSQL, SQLite, MySQL"
      },
      "en": {
        "title": "Multi-Column Atomic Update with Monotonic Predicate",
        "explanation": "Dirty Writes (G0) occur when uncommitted writes from concurrent transactions interleave on the same physical row. The definitive fix combines price and winner updates into a single atomic UPDATE statement guarded by a monotonic predicate (WHERE id = 1 AND highest_bid < new_bid). This enforces strict 2PL lock retention and guarantees monotonic bid progression without hybrid corruption.",
        "code": "-- Atomic Single-Statement Update with Monotonic Guard\nBEGIN;\nUPDATE auctions \nSET highest_bid = 200, \n    winner = 'Bob' \nWHERE id = 1 AND highest_bid < 200;\n\n-- If rows_affected == 0, the bid has already been surpassed concurrently;\n-- Issue ROLLBACK and notify bidder.\nCOMMIT;",
        "driverNotes": "PostgreSQL, SQLite, MySQL"
      },
      "strategy": "Atualização Atômica Multi-Coluna com Predicado Monotônico",
      "sql": "-- Update Atômico em Instrução Única com Cláusula Monotônica\nBEGIN;\nUPDATE auctions \nSET highest_bid = 200, \n    winner = 'Bob' \nWHERE id = 1 AND highest_bid < 200;\n\n-- Se rows_affected == 0, o lance já foi superado concorrentemente;\n-- Fazer ROLLBACK e notificar o usuário.\nCOMMIT;",
      "explanation": "Dirty Writes (G0) ocorrem quando escritas não comitadas de transações simultâneas se entrelaçam sobre a mesma linha de dados. A correção definitiva consiste em unificar as atualizações de valor e titular em uma única cláusula UPDATE atômica no banco, condicionada estritamente a um lance superior ao atual (WHERE id = 1 AND highest_bid < novo_lance). Isso assegura retenção de locks 2PL estritos e garante monotonicidade sem risco de estados corrompidos híbridos.",
      "engines": [
        "PostgreSQL",
        "SQLite",
        "MySQL"
      ]
    }
  },
  {
    "id": "crypto",
    "name": {
      "pt": "Informação Circular em Arbitragem Cripto",
      "en": "Crypto Arbitrage Circular Info"
    },
    "code": "G1c",
    "description": {
      "pt": "Loop de arbitragem cross-exchange observa arestas de dependência cíclicas, levando à execução de lucros fantasmas.",
      "en": "Cross-exchange arbitrage loop observes cyclic dependency edges, leading to phantom profit execution."
    },
    "summary": {
      "pt": "Loop de arbitragem cross-exchange observa arestas de dependência cíclicas, levando à execução de lucros fantasmas.",
      "en": "Cross-exchange arbitrage loop observes cyclic dependency edges, leading to phantom profit execution."
    },
    "schema": "-- Schema\nCREATE TABLE orderbooks (\n    id INT PRIMARY KEY,\n    exchange TEXT NOT NULL,\n    pair TEXT NOT NULL,\n    ask_price INT NOT NULL,\n    version INT NOT NULL\n);\n\n-- Seed\nINSERT INTO orderbooks VALUES (1, 'Binance', 'BTC/USDT', 64000, 1), (2, 'Coinbase', 'BTC/USDT', 64200, 1);",
    "chaos": "version: \"1.0\"\nname: \"crypto_arbitrage_circular\"\noperations:\n  - name: \"arb_bot_1\"\n    steps:\n      - \"SELECT ask_price FROM orderbooks WHERE id = 1 -> p1\"\n      - \"UPDATE orderbooks SET ask_price = {p1 + 50}, version = version + 1 WHERE id = 2\"\n  - name: \"arb_bot_2\"\n    steps:\n      - \"SELECT ask_price FROM orderbooks WHERE id = 2 -> p2\"\n      - \"UPDATE orderbooks SET ask_price = {p2 - 50}, version = version + 1 WHERE id = 1\"\ninvariants:\n  - name: \"monotonic_versioning\"\n    query: \"SELECT sum(version) AS total_ver FROM orderbooks;\"\n    assert: \"total_ver <= 100\"",
    "reduction": {
      "originalOps": 30,
      "minimalOps": 2,
      "reductionPct": "93.3%",
      "elapsed": "71ms",
      "cycle": "T1 ──(wr)──► T2 ──(wr)──► T1",
      "explanation": "Bot 1 reads Binance price and modifies Coinbase. Concurrently, Bot 2 reads Coinbase price and modifies Binance. The circular information flow creates a G1c cycle violating causal order."
    },
    "analysis": {
      "pt": "O Bot 1 lê o preço da Binance e altera a Coinbase. Concorrentemente, o Bot 2 lê o preço da Coinbase e altera a Binance. O fluxo circular de informação cria um ciclo G1c violando a ordem causal.",
      "en": "Bot 1 reads Binance price and modifies Coinbase. Concurrently, Bot 2 reads Coinbase price and modifies Binance. The circular information flow creates a G1c cycle violating causal order."
    },
    "fix": {
      "pt": {
        "title": "Optimistic Concurrency Control (OCC) com Versionamento ou Oráculo Serializado",
        "explanation": "A anomalia G1c demonstra fluxos cíclicos de leitura-escrita entre pares distribuídos. Para assegurar linearizabilidade global e evitar cálculos baseados em spreads defasados, emprega-se OCC com número de versão atômico ou bloqueio explícito FOR UPDATE em ordem canônica das chaves primárias dos livros antes da validação e execução da ordem de arbitragem.",
        "code": "-- Controle Otimista de Concorrência (OCC) com Versionamento Monotônico\nBEGIN;\nSELECT id, ask_price, version FROM orderbooks WHERE id IN (1, 2) FOR UPDATE;\n\n-- Validar spread de lucro e atualizar com incremento estrito:\nUPDATE orderbooks \nSET ask_price = :novo_preco, version = version + 1 \nWHERE id = :exchange_id AND version = :versao_esperada;\n\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL (InnoDB), CockroachDB"
      },
      "en": {
        "title": "Optimistic Concurrency Control (OCC) with Versioning or Canonical Locks",
        "explanation": "Anomaly G1c demonstrates cyclic read-write dependencies across distributed pairs. To preserve global linearizability and prevent execution on stale price spreads, applications should implement OCC with monotonic version numbers or acquire explicit FOR UPDATE locks in canonical primary key order across order books before evaluating arbitrage margins.",
        "code": "-- Optimistic Concurrency Control (OCC) with Monotonic Version Check\nBEGIN;\nSELECT id, ask_price, version FROM orderbooks WHERE id IN (1, 2) FOR UPDATE;\n\n-- Validate profit spread and execute conditional update with version check:\nUPDATE orderbooks \nSET ask_price = :new_price, version = version + 1 \nWHERE id = :exchange_id AND version = :expected_version;\n\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL (InnoDB), CockroachDB"
      },
      "strategy": "Optimistic Concurrency Control (OCC) com Versionamento ou Oráculo Serializado",
      "sql": "-- Controle Otimista de Concorrência (OCC) com Versionamento Monotônico\nBEGIN;\nSELECT id, ask_price, version FROM orderbooks WHERE id IN (1, 2) FOR UPDATE;\n\n-- Validar spread de lucro e atualizar com incremento estrito:\nUPDATE orderbooks \nSET ask_price = :novo_preco, version = version + 1 \nWHERE id = :exchange_id AND version = :versao_esperada;\n\nCOMMIT;",
      "explanation": "A anomalia G1c demonstra fluxos cíclicos de leitura-escrita entre pares distribuídos. Para assegurar linearizabilidade global e evitar cálculos baseados em spreads defasados, emprega-se OCC com número de versão atômico ou bloqueio explícito FOR UPDATE em ordem canônica das chaves primárias dos livros antes da validação e execução da ordem de arbitragem.",
      "engines": [
        "PostgreSQL",
        "MySQL (InnoDB), CockroachDB"
      ]
    }
  },
  {
    "id": "flashcrash",
    "name": {
      "pt": "Leitura Suja em Flash Crash (Dirty Read)",
      "en": "Flash Crash Dirty Read"
    },
    "code": "G1a",
    "description": {
      "pt": "Bot de liquidação lê queda de margem não comitada de uma transação que subsequentemente sofre rollback.",
      "en": "Liquidation bot reads uncommitted margin drop from a transaction that subsequently rolls back."
    },
    "summary": {
      "pt": "Bot de liquidação lê queda de margem não comitada de uma transação que subsequentemente sofre rollback.",
      "en": "Liquidation bot reads uncommitted margin drop from a transaction that subsequently rolls back."
    },
    "schema": "-- Schema\nCREATE TABLE margin_accounts (\n    id INT PRIMARY KEY,\n    trader TEXT NOT NULL,\n    collateral INT NOT NULL,\n    status TEXT NOT NULL\n);\n\n-- Seed\nINSERT INTO margin_accounts VALUES (1, 'Whale_01', 50000, 'ACTIVE');",
    "chaos": "version: \"1.0\"\nname: \"flash_crash_dirty_read\"\noperations:\n  - name: \"market_order_rollback\"\n    steps:\n      - \"UPDATE margin_accounts SET collateral = 5000 WHERE id = 1\"\n      - \"SELECT 1/0\" # Forced division by zero fault to trigger ROLLBACK\n  - name: \"liquidation_bot\"\n    steps:\n      - \"SELECT collateral FROM margin_accounts WHERE id = 1 -> col\"\n      - \"UPDATE margin_accounts SET status = 'LIQUIDATED' WHERE id = 1 AND {col < 10000}\"\ninvariants:\n  - name: \"no_spurious_liquidation\"\n    query: \"SELECT collateral, status FROM margin_accounts WHERE id = 1;\"\n    assert: \"collateral == 50000 AND status == 'ACTIVE'\"",
    "reduction": {
      "originalOps": 10,
      "minimalOps": 2,
      "reductionPct": "80.0%",
      "elapsed": "38ms",
      "cycle": "w1(collateral=5000) ... r2(collateral) ... a1",
      "explanation": "T1 updates collateral to $5000 and then aborts. Under READ UNCOMMITTED, T2 reads the uncommitted $5000 drop and triggers premature liquidation of a solvent trader."
    },
    "analysis": {
      "pt": "T1 atualiza o colateral para $5000 e aborta em seguida. Sob READ UNCOMMITTED, T2 lê o colateral de $5000 não comitado e dispara a liquidação indevida de um trader solvente.",
      "en": "T1 updates collateral to $5000 and then aborts. Under READ UNCOMMITTED, T2 reads the uncommitted $5000 drop and triggers premature liquidation of a solvent trader."
    },
    "fix": {
      "pt": {
        "title": "Isolamento Mínimo READ COMMITTED no Pool de Conexões",
        "explanation": "Dirty Reads (G1a / Aborted Reads) ocorrem quando uma transação lê modificações feitas por outra transação que posteriormente sofre abort (ROLLBACK). Em finanças e criptoativos, isso acarreta liquidações catastróficas espúrias. A mitigação arquitetural exige que o pool de conexões configure como piso inviolável o isolamento READ COMMITTED, garantindo que leituras acessem apenas tuplas comitadas de forma duradoura no write-ahead log.",
        "code": "-- Assegurar Nível de Isolamento Mínimo READ COMMITTED no Banco\nSET TRANSACTION ISOLATION LEVEL READ COMMITTED;\n\n-- Rotina de Liquidação com Validação Transacional\nBEGIN;\nSELECT collateral, status FROM margin_accounts WHERE id = 1 FOR UPDATE;\n-- Apenas dados comitados são lidos! Transações abortadas nunca vazam.\nUPDATE margin_accounts SET status = 'LIQUIDATED' WHERE id = 1 AND collateral < 10000;\nCOMMIT;",
        "driverNotes": "PostgreSQL (Default), SQLite (WAL), MySQL (Default)"
      },
      "en": {
        "title": "Enforce Minimum READ COMMITTED in Database Connection Pool",
        "explanation": "Dirty Reads (G1a / Aborted Reads) occur when a transaction reads uncommitted mutations from another transaction that subsequently aborts (ROLLBACK). In financial and DeFi systems, this triggers catastrophic phantom liquidations. The architectural fix enforces READ COMMITTED as the strict floor in the connection pool, ensuring queries only access tuples durably committed in the WAL.",
        "code": "-- Enforce Minimum READ COMMITTED Isolation Level\nSET TRANSACTION ISOLATION LEVEL READ COMMITTED;\n\n-- Liquidation Routine with Row Guard Lock\nBEGIN;\nSELECT collateral, status FROM margin_accounts WHERE id = 1 FOR UPDATE;\n-- Only durably committed state is visible; aborted transactions never leak!\nUPDATE margin_accounts SET status = 'LIQUIDATED' WHERE id = 1 AND collateral < 10000;\nCOMMIT;",
        "driverNotes": "PostgreSQL (Default), SQLite (WAL), MySQL (Default)"
      },
      "strategy": "Isolamento Mínimo READ COMMITTED no Pool de Conexões",
      "sql": "-- Assegurar Nível de Isolamento Mínimo READ COMMITTED no Banco\nSET TRANSACTION ISOLATION LEVEL READ COMMITTED;\n\n-- Rotina de Liquidação com Validação Transacional\nBEGIN;\nSELECT collateral, status FROM margin_accounts WHERE id = 1 FOR UPDATE;\n-- Apenas dados comitados são lidos! Transações abortadas nunca vazam.\nUPDATE margin_accounts SET status = 'LIQUIDATED' WHERE id = 1 AND collateral < 10000;\nCOMMIT;",
      "explanation": "Dirty Reads (G1a / Aborted Reads) ocorrem quando uma transação lê modificações feitas por outra transação que posteriormente sofre abort (ROLLBACK). Em finanças e criptoativos, isso acarreta liquidações catastróficas espúrias. A mitigação arquitetural exige que o pool de conexões configure como piso inviolável o isolamento READ COMMITTED, garantindo que leituras acessem apenas tuplas comitadas de forma duradoura no write-ahead log.",
      "engines": [
        "PostgreSQL (Default)",
        "SQLite (WAL)",
        "MySQL (Default)"
      ]
    }
  },
  {
    "id": "ticket",
    "name": {
      "pt": "Ciclo de Anti-Dependência em Bilheteria (G2)",
      "en": "Ticket Anti-Dependency Cycle"
    },
    "code": "G2",
    "description": {
      "pt": "Reservas de ingressos geram conflitos fantasmas entre predicados de assentos sob Snapshot Isolation.",
      "en": "Concert ticket reservations generate phantom conflicts across seat range predicates under Snapshot Isolation."
    },
    "summary": {
      "pt": "Reservas de ingressos geram conflitos fantasmas entre predicados de assentos sob Snapshot Isolation.",
      "en": "Concert ticket reservations generate phantom conflicts across seat range predicates under Snapshot Isolation."
    },
    "schema": "-- Schema\nCREATE TABLE seats (\n    id INT PRIMARY KEY,\n    section TEXT NOT NULL,\n    seat_no INT NOT NULL,\n    reserved_by TEXT\n);\n\n-- Seed\nINSERT INTO seats VALUES (1, 'VIP', 101, NULL), (2, 'VIP', 102, NULL);",
    "chaos": "version: \"1.0\"\nname: \"ticket_anti_dependency_g2\"\noperations:\n  - name: \"book_adjacent_left\"\n    steps:\n      - \"SELECT count(*) AS reserved FROM seats WHERE section = 'VIP' AND reserved_by IS NOT NULL -> c\"\n      - \"UPDATE seats SET reserved_by = 'Fan_A' WHERE id = 1 AND {c == 0}\"\n  - name: \"book_adjacent_right\"\n    steps:\n      - \"SELECT count(*) AS reserved FROM seats WHERE section = 'VIP' AND reserved_by IS NOT NULL -> c\"\n      - \"UPDATE seats SET reserved_by = 'Fan_B' WHERE id = 2 AND {c == 0}\"\ninvariants:\n  - name: \"max_one_seat_rule\"\n    query: \"SELECT count(*) AS reserved FROM seats WHERE section = 'VIP' AND reserved_by IS NOT NULL;\"\n    assert: \"reserved <= 1\"",
    "reduction": {
      "originalOps": 20,
      "minimalOps": 2,
      "reductionPct": "90.0%",
      "elapsed": "65ms",
      "cycle": "T1 ──(rw)──► T2 ──(rw)──► T1",
      "explanation": "Predicate anti-dependency cycle: T1 checks if any VIP seat is booked (finds 0) and books seat 1. T2 simultaneously checks if any VIP seat is booked (finds 0) and books seat 2. Both commit, violating VIP exclusivity."
    },
    "analysis": {
      "pt": "Ciclo de anti-dependência de predicado: T1 verifica se algum assento VIP está reservado (encontra 0) e reserva o assento 1. T2 simultaneamente verifica assentos VIP (encontra 0) e reserva o assento 2. Ambos comitam, violando a exclusividade do setor.",
      "en": "Predicate anti-dependency cycle: T1 checks if any VIP seat is booked (finds 0) and books seat 1. T2 simultaneously checks if any VIP seat is booked (finds 0) and books seat 2. Both commit, violating VIP exclusivity."
    },
    "fix": {
      "pt": {
        "title": "Restrição de Unicidade Estrutural e Locks de Predicado Serializáveis",
        "explanation": "Ciclos G2 ocorrem quando consultas de agregação ou intervalos checam a existência de registros através de predicados (COUNT(*) WHERE ...). Sob Snapshot Isolation, as transações não bloqueiam predicados ausentes. A adição de uma restrição UNIQUE a nível de schema converte a corrida silenciosa em um erro determinístico de integridade de índice, enquanto SELECT ... FOR UPDATE ou SSI serializam a reserva de forma infalível.",
        "code": "-- 1. Restrição de Chave Única Composta para Garantir Atomicidade de Índice\nALTER TABLE seats \nADD CONSTRAINT uq_section_reservation UNIQUE (section, seat_no);\n\n-- 2. Predicate Lock com SELECT FOR UPDATE ou SSI\nBEGIN;\nSELECT id FROM seats \nWHERE section = 'VIP' AND reserved_by IS NULL \nLIMIT 1 FOR UPDATE;\n\nUPDATE seats SET reserved_by = 'Fan_A' WHERE id = :seat_id;\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL, SQLite"
      },
      "en": {
        "title": "Structural Unique Constraints and Serializable Predicate Locks",
        "explanation": "G2 cycles manifest when aggregate queries or range checks inspect row existence via predicates (COUNT(*) WHERE ...). Under Snapshot Isolation, queries do not lock non-existent keys. Introducing a schema-level UNIQUE constraint converts silent race conditions into deterministic index integrity violations, while SELECT ... FOR UPDATE or SSI serializes reservations safely.",
        "code": "-- 1. Composite Unique Constraint for Index-Level Atomicity\nALTER TABLE seats \nADD CONSTRAINT uq_section_reservation UNIQUE (section, seat_no);\n\n-- 2. Predicate Lock via SELECT FOR UPDATE or SSI\nBEGIN;\nSELECT id FROM seats \nWHERE section = 'VIP' AND reserved_by IS NULL \nLIMIT 1 FOR UPDATE;\n\nUPDATE seats SET reserved_by = 'Fan_A' WHERE id = :seat_id;\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL, SQLite"
      },
      "strategy": "Restrição de Unicidade Estrutural e Locks de Predicado Serializáveis",
      "sql": "-- 1. Restrição de Chave Única Composta para Garantir Atomicidade de Índice\nALTER TABLE seats \nADD CONSTRAINT uq_section_reservation UNIQUE (section, seat_no);\n\n-- 2. Predicate Lock com SELECT FOR UPDATE ou SSI\nBEGIN;\nSELECT id FROM seats \nWHERE section = 'VIP' AND reserved_by IS NULL \nLIMIT 1 FOR UPDATE;\n\nUPDATE seats SET reserved_by = 'Fan_A' WHERE id = :seat_id;\nCOMMIT;",
      "explanation": "Ciclos G2 ocorrem quando consultas de agregação ou intervalos checam a existência de registros através de predicados (COUNT(*) WHERE ...). Sob Snapshot Isolation, as transações não bloqueiam predicados ausentes. A adição de uma restrição UNIQUE a nível de schema converte a corrida silenciosa em um erro determinístico de integridade de índice, enquanto SELECT ... FOR UPDATE ou SSI serializam a reserva de forma infalível.",
      "engines": [
        "PostgreSQL",
        "MySQL",
        "SQLite"
      ]
    }
  },
  {
    "id": "deadlock",
    "name": {
      "pt": "Ciclo de Deadlock & Recuperação",
      "en": "Deadlock Cycle & Recovery"
    },
    "code": "G-DL",
    "description": {
      "pt": "Dois workers de transferência bloqueiam contas em ordem inversa, induzindo um deadlock no grafo de espera (WFG).",
      "en": "Two transfer workers lock accounts in reverse order, inducing a cyclic lock-wait graph (WFG) deadlock."
    },
    "summary": {
      "pt": "Dois workers de transferência bloqueiam contas em ordem inversa, induzindo um deadlock no grafo de espera (WFG).",
      "en": "Two transfer workers lock accounts in reverse order, inducing a cyclic lock-wait graph (WFG) deadlock."
    },
    "schema": "-- Schema\nCREATE TABLE accounts (\n    id INT PRIMARY KEY,\n    balance INT NOT NULL\n);\n\n-- Seed\nINSERT INTO accounts VALUES (1, 500), (2, 500);",
    "chaos": "version: \"1.0\"\nname: \"deadlock_cycle_recovery\"\noperations:\n  - name: \"transfer_1_to_2\"\n    steps:\n      - \"SELECT balance FROM accounts WHERE id = 1 -> b1\"\n      - \"UPDATE accounts SET balance = balance - 100 WHERE id = 1\"\n      - \"UPDATE accounts SET balance = balance + 100 WHERE id = 2\"\n  - name: \"transfer_2_to_1\"\n    steps:\n      - \"SELECT balance FROM accounts WHERE id = 2 -> b2\"\n      - \"UPDATE accounts SET balance = balance - 100 WHERE id = 2\"\n      - \"UPDATE accounts SET balance = balance + 100 WHERE id = 1\"\ninvariants:\n  - name: \"total_balance_exact\"\n    query: \"SELECT sum(balance) AS total FROM accounts;\"\n    assert: \"total == 1000\"",
    "reduction": {
      "originalOps": 15,
      "minimalOps": 2,
      "reductionPct": "86.7%",
      "elapsed": "54ms",
      "cycle": "T1 ──(waits)──► T2 ──(waits)──► T1",
      "explanation": "T1 locks Account 1 and requests lock on Account 2. Concurrently, T2 locks Account 2 and requests lock on Account 1. Mutual wait forms a cycle in the Wait-For Graph (WFG)."
    },
    "analysis": {
      "pt": "T1 trava a Conta 1 e requisita trava na Conta 2. Concorrentemente, T2 trava a Conta 2 e requisita trava na Conta 1. A espera mútua induz um ciclo no Wait-For Graph (WFG).",
      "en": "T1 locks Account 1 and requests lock on Account 2. Concurrently, T2 locks Account 2 and requests lock on Account 1. Mutual wait forms a cycle in the Wait-For Graph (WFG)."
    },
    "fix": {
      "pt": {
        "title": "Ordenação Canônica Determinística de Bloqueios (Lock Ordering)",
        "explanation": "Deadlocks ocorrem quando diferentes transações disputam os mesmos registros em ordens inversas (T1: 1 depois 2; T2: 2 depois 1). A mitigação canônica consiste em estabelecer um protocolo estrito de ordenação hierárquica de recursos antes de adquirir qualquer bloqueio (ex: travar sempre pelo menor ID primeiro). Com ordenação canônica garantida na camada de repositório, ciclos no Wait-For Graph tornam-se matematicamente impossíveis de se formarem.",
        "code": "-- Ordenação Padronizada de Locks pelo Menor ID\n-- Na aplicação:\n-- first_id = MIN(conta_a, conta_b)\n-- second_id = MAX(conta_a, conta_b)\n\nBEGIN;\n-- Bloquear sempre os recursos na mesma ordem canônica crescente:\nSELECT id, balance FROM accounts WHERE id = :first_id FOR UPDATE;\nSELECT id, balance FROM accounts WHERE id = :second_id FOR UPDATE;\n\nUPDATE accounts SET balance = balance - 100 WHERE id = :origem;\nUPDATE accounts SET balance = balance + 100 WHERE id = :destino;\nCOMMIT;",
        "driverNotes": "PostgreSQL (Code 40P01), MySQL (Error 1213), SQLite"
      },
      "en": {
        "title": "Deterministic Canonical Lock Ordering",
        "explanation": "Deadlocks emerge when concurrent transactions acquire locks on the same resources in opposing sequences (T1 locks 1 then 2; T2 locks 2 then 1). The canonical solution enforces a strict resource hierarchy before acquiring any lock (e.g. always acquiring locks in ascending ID order). With canonical lock ordering at the repository layer, cycles in the Wait-For Graph become mathematically impossible.",
        "code": "-- Standardized Lock Ordering by Ascending Primary Key\n-- In application layer:\n-- first_id = MIN(account_a, account_b)\n-- second_id = MAX(account_a, account_b)\n\nBEGIN;\n-- Always acquire exclusive locks in ascending canonical order:\nSELECT id, balance FROM accounts WHERE id = :first_id FOR UPDATE;\nSELECT id, balance FROM accounts WHERE id = :second_id FOR UPDATE;\n\nUPDATE accounts SET balance = balance - 100 WHERE id = :source_id;\nUPDATE accounts SET balance = balance + 100 WHERE id = :target_id;\nCOMMIT;",
        "driverNotes": "PostgreSQL (Code 40P01), MySQL (Error 1213), SQLite"
      },
      "strategy": "Ordenação Canônica Determinística de Bloqueios (Lock Ordering)",
      "sql": "-- Ordenação Padronizada de Locks pelo Menor ID\n-- Na aplicação:\n-- first_id = MIN(conta_a, conta_b)\n-- second_id = MAX(conta_a, conta_b)\n\nBEGIN;\n-- Bloquear sempre os recursos na mesma ordem canônica crescente:\nSELECT id, balance FROM accounts WHERE id = :first_id FOR UPDATE;\nSELECT id, balance FROM accounts WHERE id = :second_id FOR UPDATE;\n\nUPDATE accounts SET balance = balance - 100 WHERE id = :origem;\nUPDATE accounts SET balance = balance + 100 WHERE id = :destino;\nCOMMIT;",
      "explanation": "Deadlocks ocorrem quando diferentes transações disputam os mesmos registros em ordens inversas (T1: 1 depois 2; T2: 2 depois 1). A mitigação canônica consiste em estabelecer um protocolo estrito de ordenação hierárquica de recursos antes de adquirir qualquer bloqueio (ex: travar sempre pelo menor ID primeiro). Com ordenação canônica garantida na camada de repositório, ciclos no Wait-For Graph tornam-se matematicamente impossíveis de se formarem.",
      "engines": [
        "PostgreSQL (Code 40P01)",
        "MySQL (Error 1213), SQLite"
      ]
    }
  },
  {
    "id": "fk_cascade",
    "name": {
      "pt": "Deadlock em Deleção em Cascata de Foreign Key",
      "en": "Foreign Key Cascade Deadlock"
    },
    "code": "G-DL",
    "description": {
      "pt": "Inserções simultâneas em itens filhos e deleção em cascata do pedido pai invertem a hierarquia de bloqueios.",
      "en": "Concurrent inserts into child items and cascaded parent order deletes invert the lock hierarchy."
    },
    "summary": {
      "pt": "Inserções simultâneas em itens filhos e deleção em cascata do pedido pai invertem a hierarquia de bloqueios.",
      "en": "Concurrent inserts into child items and cascaded parent order deletes invert the lock hierarchy."
    },
    "schema": "-- Schema\nCREATE TABLE parent_orders (\n    id INT PRIMARY KEY,\n    status TEXT NOT NULL,\n    total_cents INT NOT NULL\n);\n\nCREATE TABLE child_items (\n    id INT PRIMARY KEY,\n    order_id INT NOT NULL REFERENCES parent_orders(id) ON DELETE CASCADE,\n    sku TEXT NOT NULL,\n    price_cents INT NOT NULL\n);\n\n-- Seed\nINSERT INTO parent_orders VALUES (1, 'OPEN', 5000);\nINSERT INTO child_items VALUES (1, 1, 'ITEM-A', 2500), (2, 1, 'ITEM-B', 2500);",
    "chaos": "version: \"1.0\"\nname: \"foreign_key_cascade_deadlock\"\noperations:\n  - name: \"add_order_item\"\n    steps:\n      - \"INSERT INTO child_items VALUES ($monotonic_counter(10, 1), 1, 'ITEM-NEW', 1500)\"\n      - \"UPDATE parent_orders SET total_cents = total_cents + 1500 WHERE id = 1\"\n  - name: \"cancel_order_cascade\"\n    steps:\n      - \"UPDATE parent_orders SET status = 'CANCELLED' WHERE id = 1\"\n      - \"DELETE FROM parent_orders WHERE id = 1\"\ninvariants:\n  - name: \"referential_integrity\"\n    query: \"SELECT count(*) AS orphan_items FROM child_items LEFT JOIN parent_orders ON child_items.order_id = parent_orders.id WHERE parent_orders.id IS NULL;\"\n    assert: \"orphan_items == 0\"",
    "reduction": {
      "originalOps": 20,
      "minimalOps": 2,
      "reductionPct": "90.0%",
      "elapsed": "58ms",
      "cycle": "T1: Child ──► Parent ◄──► T2: Parent ──► Child",
      "explanation": "T1 locks child row and requests parent lock to update total; T2 locks parent row and requests cascaded child locks to delete. Lock hierarchy inversion creates deadlock (WFG cycle)."
    },
    "analysis": {
      "pt": "T1 bloqueia a linha filha e requisita lock na linha pai para atualizar o total; T2 bloqueia a linha pai e requisita lock em cascata nos filhos para deletar. A inversão de hierarquia induz um deadlock no Wait-For Graph.",
      "en": "T1 locks child row and requests parent lock to update total; T2 locks parent row and requests cascaded child locks to delete. Lock hierarchy inversion creates deadlock (WFG cycle)."
    },
    "fix": {
      "pt": {
        "title": "Índice na Chave Estrangeira e Padronização de Locks Pai-Filho",
        "explanation": "Deadlocks em deleções em cascata ocorrem porque: (1) a ausência de índice na coluna de FK força table scans na tabela filha, gerando bloqueios abrangentes desnecessários; e (2) transações de inserção travam primeiro a tabela filha e depois o pai, enquanto cascatas travam primeiro o pai e depois os filhos. A criação do índice na FK associada ao bloqueio prévio do registro pai (FOR UPDATE) equaliza a hierarquia e extingue o risco de deadlock.",
        "code": "-- 1. Criar Índice Dedicado na Coluna de Chave Estrangeira (CRÍTICO!)\nCREATE INDEX idx_child_items_order_id ON child_items(order_id);\n\n-- 2. Padronizar a Sequência de Bloqueios Pai -> Filhos em Cancelamentos:\nBEGIN;\n-- Bloquear a linha pai explicitamente antes de disparar deleções:\nSELECT id FROM parent_orders WHERE id = 1 FOR UPDATE;\n\n-- Deletar os filhos de forma explícita e controlada:\nDELETE FROM child_items WHERE order_id = 1;\n\n-- Deletar o pedido pai:\nDELETE FROM parent_orders WHERE id = 1;\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL (InnoDB), SQLite"
      },
      "en": {
        "title": "Foreign Key Indexing and Parent-First Lock Standardization",
        "explanation": "Foreign key cascade deadlocks occur because: (1) missing indexes on FK columns force table scans on child tables, acquiring broad table-level locks; and (2) insert transactions lock child first then parent, while cascades lock parent then child. Adding a dedicated B-tree index on the child FK and explicitly locking the parent record first (FOR UPDATE) normalizes the locking hierarchy and eliminates deadlock cycles.",
        "code": "-- 1. Add Dedicated Index on Child Foreign Key Column (CRITICAL!)\nCREATE INDEX idx_child_items_order_id ON child_items(order_id);\n\n-- 2. Standardize Parent -> Child Lock Sequence during Order Cancellation:\nBEGIN;\n-- Explicitly lock parent row before triggering deletions:\nSELECT id FROM parent_orders WHERE id = 1 FOR UPDATE;\n\n-- Explicitly delete child items:\nDELETE FROM child_items WHERE order_id = 1;\n\n-- Delete parent order:\nDELETE FROM parent_orders WHERE id = 1;\nCOMMIT;",
        "driverNotes": "PostgreSQL, MySQL (InnoDB), SQLite"
      },
      "strategy": "Índice na Chave Estrangeira e Padronização de Locks Pai-Filho",
      "sql": "-- 1. Criar Índice Dedicado na Coluna de Chave Estrangeira (CRÍTICO!)\nCREATE INDEX idx_child_items_order_id ON child_items(order_id);\n\n-- 2. Padronizar a Sequência de Bloqueios Pai -> Filhos em Cancelamentos:\nBEGIN;\n-- Bloquear a linha pai explicitamente antes de disparar deleções:\nSELECT id FROM parent_orders WHERE id = 1 FOR UPDATE;\n\n-- Deletar os filhos de forma explícita e controlada:\nDELETE FROM child_items WHERE order_id = 1;\n\n-- Deletar o pedido pai:\nDELETE FROM parent_orders WHERE id = 1;\nCOMMIT;",
      "explanation": "Deadlocks em deleções em cascata ocorrem porque: (1) a ausência de índice na coluna de FK força table scans na tabela filha, gerando bloqueios abrangentes desnecessários; e (2) transações de inserção travam primeiro a tabela filha e depois o pai, enquanto cascatas travam primeiro o pai e depois os filhos. A criação do índice na FK associada ao bloqueio prévio do registro pai (FOR UPDATE) equaliza a hierarquia e extingue o risco de deadlock.",
      "engines": [
        "PostgreSQL",
        "MySQL (InnoDB), SQLite"
      ]
    }
  }
];

// ============================================================================
// 3. HASH ROUTER & MULTI-VIEW CONTROLLER
// ============================================================================
const ROUTES = {
  landing: "view-landing",
  docs: "view-docs",
  scenarios: "view-scenarios",
  visualizer: "view-visualizer",
  matrix: "view-matrix"
};

let currentRoute = "landing";

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

  currentRoute = route;

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
  } else if (route === "matrix") {
    renderMatrixView();
  }

  window.scrollTo({ top: 0, behavior: "instant" });
}

// ============================================================================
// 4. TERMINAL REPLAY SIMULATOR (Landing Page)
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

function getTerminalSteps(mode, lang) {
  const t = (I18N[lang] && I18N[lang].terminal) ? I18N[lang].terminal : I18N.pt.terminal;
  const steps = [];

  if (mode === "full" || mode === "jitter") {
    steps.push({ text: t.initFuzzer, class: "term-dim", delay: 80 });
    steps.push({ text: t.injectJitter, class: "term-dim", delay: 200 });
    steps.push({ text: t.anomalyDetected, class: "term-err", delay: 450 });
    steps.push({ text: t.cycle, class: "term-dim", delay: 650 });
    steps.push({ text: t.violatedInvariant, class: "term-dim", delay: 850 });
  }

  if (mode === "full" || mode === "shrink") {
    steps.push({ text: t.startDdmin, class: "term-info", delay: 1050 });
    steps.push({ text: t.iteration1, class: "", delay: 1300 });
    steps.push({ text: t.iteration2, class: "", delay: 1550 });
    steps.push({ text: t.iteration3, class: "", delay: 1800 });
    steps.push({ text: t.traceShrunk, class: "term-ok", delay: 2050 });
    steps.push({ text: t.synthesizedRepro, class: "term-dim", delay: 2250 });
  }

  return steps;
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

  const lang = window.currentLang || "pt";
  const steps = getTerminalSteps(mode, lang);

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

  const lang = window.currentLang || "pt";
  const t = (I18N[lang] && I18N[lang].terminal) ? I18N[lang].terminal : I18N.pt.terminal;

  termEl.innerHTML = `
    <div class="term-dim">${t.initFuzzer}</div>
    <div class="term-dim">${t.injectJitter}</div>
    <div class="term-err">${t.anomalyDetected}</div>
    <div class="term-dim">${t.cycle}</div>
    <div class="term-dim">${t.violatedInvariant}</div>
    <div class="term-info">${t.startDdmin}</div>
    <div>${t.iteration1}</div>
    <div>${t.iteration2}</div>
    <div>${t.iteration3}</div>
    <div class="term-ok">${t.traceShrunk}</div>
    <div class="term-dim">${t.synthesizedRepro}</div>
  `;

  const footerPills = document.querySelector(".terminal-footer-bar .terminal-pills");
  if (footerPills) {
    footerPills.innerHTML = `
      <span class="term-pill">${t.pillWorkers} <strong class="term-pill-val">4</strong></span>
      <span class="term-pill">${t.pillEngine} <strong class="term-pill-val">SQLite Memory</strong></span>
      <span class="term-pill">${t.pillReduction} <strong class="term-pill-val" style="color: var(--color-green);">90.0% (ddmin)</strong></span>
    `;
  }
  const latencyDiv = document.querySelector(".terminal-footer-bar > div:last-child");
  if (latencyDiv) {
    latencyDiv.innerHTML = `${t.pillLatency} <strong style="color: var(--color-cream);">68ms</strong>`;
  }
}

// ============================================================================
// 5. DOCUMENTATION HUB CONTROLLER (View: #view-docs)
// ============================================================================
let currentDocChapterId = "getting-started";

function getDocsData() {
  if (!window.DOCS_DATA) return [];
  const lang = window.currentLang || "pt";
  const data = window.DOCS_DATA[lang] || window.DOCS_DATA.pt || window.DOCS_DATA.en || window.DOCS_DATA;
  if (Array.isArray(data)) return data;
  if (typeof data === "object" && data !== null) {
    return Object.values(data);
  }
  return [];
}

function getDocById(chapterId) {
  if (!window.DOCS_DATA) return null;
  const lang = window.currentLang || "pt";
  const langData = window.DOCS_DATA[lang] || window.DOCS_DATA.pt || window.DOCS_DATA.en;
  if (langData && typeof langData === "object" && !Array.isArray(langData) && langData[chapterId]) {
    return langData[chapterId];
  }
  const list = getDocsData();
  return list.find(d => d.id === chapterId) || list[0] || null;
}

function initDocsHub() {
  if (!window.DOCS_DATA) return;

  renderDocsSidebar();
  setupDocsSearch();
  setupDocsFooterNav();
}

function renderDocsSidebar(filterQuery = "") {
  const navEl = document.getElementById("docsSidebarNav");
  if (!navEl || !window.DOCS_DATA) return;

  const lang = window.currentLang || "pt";
  const docsList = getDocsData();
  if (!docsList || docsList.length === 0) return;

  const query = filterQuery.toLowerCase().trim();
  const categories = {};

  docsList.forEach(doc => {
    const match = !query || 
      (doc.title && doc.title.toLowerCase().includes(query)) || 
      (doc.summary && doc.summary.toLowerCase().includes(query)) || 
      (doc.category && doc.category.toLowerCase().includes(query));

    if (match) {
      if (!categories[doc.category]) categories[doc.category] = [];
      categories[doc.category].push(doc);
    }
  });

  if (Object.keys(categories).length === 0) {
    const noResults = (I18N[lang] && I18N[lang].docs && I18N[lang].docs.noResults) ? I18N[lang].docs.noResults : "Nenhum capítulo encontrado.";
    navEl.innerHTML = `<div style="padding: 12px; color: var(--text-muted); font-size: 0.85rem;">${noResults}</div>`;
    return;
  }

  let html = "";
  for (const [catName, docs] of Object.entries(categories)) {
    html += `
      <div class="docs-nav-group">
        <div class="docs-nav-group-title">${escapeHtml(catName)}</div>
        ${docs.map(doc => `
          <button class="docs-nav-link ${doc.id === currentDocChapterId ? 'active' : ''}" data-doc-id="${doc.id}">
            <span>${escapeHtml(doc.title)}</span>
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
  const lang = window.currentLang || "pt";
  const doc = getDocById(chapterId);
  if (!doc) return;

  currentDocChapterId = doc.id;

  // Breadcrumbs (ChaosSQL > Docs > Category > Chapter Title)
  const breadcrumbEl = document.getElementById("docsBreadcrumbs");
  const docsLabel = (I18N[lang] && I18N[lang].docs && I18N[lang].docs.breadcrumbDocs) ? I18N[lang].docs.breadcrumbDocs : "Docs";
  if (breadcrumbEl) {
    breadcrumbEl.innerHTML = `
      <a href="#/">ChaosSQL</a> <span>/</span> <a href="#/docs">${escapeHtml(docsLabel)}</a> <span>/</span> 
      <span style="color: var(--text-secondary);">${escapeHtml(doc.category)}</span> <span style="color: var(--text-muted); margin: 0 4px;">/</span> 
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

    const copyText = (I18N[lang] && I18N[lang].docs && I18N[lang].docs.copy) ? I18N[lang].docs.copy : "Copiar";
    const copiedText = (I18N[lang] && I18N[lang].docs && I18N[lang].docs.copied) ? I18N[lang].docs.copied : "Copiado!";

    // Attach copy button logic to any code block inside
    bodyEl.querySelectorAll(".code-container").forEach(container => {
      if (!container.querySelector(".copy-code-btn")) {
        const btn = document.createElement("button");
        btn.className = "copy-code-btn";
        btn.textContent = copyText;
        btn.onclick = () => {
          const code = container.querySelector("code")?.innerText || "";
          copyTextToClipboard(code);
          btn.textContent = copiedText;
          setTimeout(() => btn.textContent = copyText, 1500);
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

const renderDocChapter = loadDocChapter;

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
      const docsList = getDocsData();
      if (!docsList || docsList.length === 0) return;
      const idx = docsList.findIndex(d => d.id === currentDocChapterId);
      if (idx > 0) {
        window.location.hash = `#/docs/${docsList[idx - 1].id}`;
      }
    });
  }

  if (nextBtn) {
    nextBtn.addEventListener("click", () => {
      const docsList = getDocsData();
      if (!docsList || docsList.length === 0) return;
      const idx = docsList.findIndex(d => d.id === currentDocChapterId);
      if (idx < docsList.length - 1) {
        window.location.hash = `#/docs/${docsList[idx + 1].id}`;
      }
    });
  }
}

function updateDocsFooterNav() {
  const docsList = getDocsData();
  if (!docsList || docsList.length === 0) return;
  const idx = docsList.findIndex(d => d.id === currentDocChapterId);
  const prevBtn = document.getElementById("docsPrevBtn");
  const nextBtn = document.getElementById("docsNextBtn");

  if (prevBtn) {
    if (idx > 0) {
      prevBtn.style.visibility = "visible";
      prevBtn.innerHTML = `← ${escapeHtml(docsList[idx - 1].title)}`;
    } else {
      prevBtn.style.visibility = "hidden";
    }
  }

  if (nextBtn) {
    if (idx < docsList.length - 1) {
      nextBtn.style.visibility = "visible";
      nextBtn.innerHTML = `${escapeHtml(docsList[idx + 1].title)} →`;
    } else {
      nextBtn.style.visibility = "hidden";
    }
  }
}

// ============================================================================
// 6. SCENARIOS EXPLORER CONTROLLER (View: #view-scenarios)
// ============================================================================
let currentScenarioIndex = 0;
let currentTab = "schema";

function renderScenarioNav() {
  const navContainer = document.getElementById("scenarioNavList");
  if (!navContainer) return;

  const lang = window.currentLang || "pt";

  navContainer.innerHTML = SCENARIOS.map((sc, index) => {
    const scName = (sc.name && typeof sc.name === "object") ? (sc.name[lang] || sc.name.en || sc.name.pt) : sc.name;
    return `
      <button class="scenario-nav-item ${index === currentScenarioIndex ? 'active' : ''}" data-index="${index}" data-id="${sc.id}">
        <span class="scenario-nav-name">${escapeHtml(scName)}</span>
        <span class="scenario-nav-code">${sc.code}</span>
      </button>
    `;
  }).join('');

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

const renderScenarioList = renderScenarioNav;

function renderScenarioStage() {
  const sc = SCENARIOS[currentScenarioIndex];
  if (!sc) return;

  const lang = window.currentLang || "pt";
  const i18n = I18N[lang] || I18N.pt;

  const scName = (sc.name && typeof sc.name === "object") ? (sc.name[lang] || sc.name.en || sc.name.pt) : sc.name;
  const scSummary = (sc.description && typeof sc.description === "object") ? (sc.description[lang] || sc.description.en || sc.description.pt) : (typeof sc.summary === "object" ? (sc.summary[lang] || sc.summary.en) : (sc.description || sc.summary));
  const scAnalysis = (sc.analysis && typeof sc.analysis === "object") ? (sc.analysis[lang] || sc.analysis.en || sc.analysis.pt) : (sc.reduction?.explanation || "");

  const titleEl = document.getElementById("stageTitle");
  const summaryEl = document.getElementById("stageSummary");
  const contentEl = document.getElementById("stageTabContent");

  if (titleEl) titleEl.textContent = `${scName} (${sc.code})`;
  if (summaryEl) summaryEl.textContent = scSummary;

  if (!contentEl) return;

  if (currentTab === "schema") {
    contentEl.innerHTML = `
      <div class="code-container">
        <button class="copy-code-btn" onclick="copySnippet(this)">${i18n.scenarios.copySql || "Copy SQL"}</button>
        <pre><code>${escapeHtml(sc.schema)}</code></pre>
      </div>
    `;
  } else if (currentTab === "chaos") {
    contentEl.innerHTML = `
      <div class="code-container">
        <button class="copy-code-btn" onclick="copySnippet(this)">${i18n.scenarios.copyYaml || "Copy YAML"}</button>
        <pre><code>${escapeHtml(sc.chaos)}</code></pre>
      </div>
    `;
  } else if (currentTab === "invariant") {
    contentEl.innerHTML = `
      <div style="margin-bottom: 16px;">
        <div style="font-family: var(--font-mono); font-size: 0.78rem; color: var(--color-yellow); margin-bottom: 6px; text-transform: uppercase;">${i18n.scenarios.formalGraphTitle || "Grafo de Conflito Formal (Adya)"}</div>
        <div style="font-family: var(--font-mono); font-size: 1rem; color: var(--color-cream); background: var(--bg-terminal); padding: 10px 14px; border-radius: var(--radius-sm); border: 1px solid var(--border-subtle);">${sc.reduction.cycle}</div>
      </div>
      <p style="font-size: 0.95rem; color: var(--text-secondary); line-height: 1.6; margin-bottom: 20px;">${escapeHtml(scAnalysis)}</p>
      <div class="metrics-row">
        <div class="metric-box">
          <div class="metric-val">${sc.reduction.originalOps} → ${sc.reduction.minimalOps}</div>
          <div class="metric-lbl">${i18n.scenarios.metric1Label || "1-Minimal Ops Shrunk"}</div>
        </div>
        <div class="metric-box">
          <div class="metric-val">${sc.reduction.reductionPct}</div>
          <div class="metric-lbl">${i18n.scenarios.metric2Label || "Causal Noise Removed"}</div>
        </div>
        <div class="metric-box">
          <div class="metric-val">${sc.reduction.elapsed}</div>
          <div class="metric-lbl">${i18n.scenarios.metric3Label || "Convergence Time"}</div>
        </div>
      </div>
    `;
  } else if (currentTab === "fix") {
    const fixData = (sc.fix && (sc.fix[lang] || sc.fix.en || sc.fix.pt)) || sc.fix || {};
    const fixTitle = fixData.title || fixData.strategy || "";
    const fixExplanation = fixData.explanation || "";
    const fixCode = fixData.code || fixData.sql || "";
    const fixDriverNotes = fixData.driverNotes || (Array.isArray(fixData.engines) ? fixData.engines.join(", ") : "");
    const enginesList = Array.isArray(fixData.engines) ? fixData.engines : (fixDriverNotes ? fixDriverNotes.split(",").map(s => s.trim()) : []);

    contentEl.innerHTML = `
      <div>
        <span class="fix-header-pill">${i18n.scenarios.fixHeaderPill || "PRODUÇÃO • RECOMENDAÇÃO ARQUITETURAL"}</span>
        <h4 class="fix-strategy-title">${escapeHtml(fixTitle)}</h4>
        <p class="fix-desc-text">${escapeHtml(fixExplanation)}</p>
        
        <div class="code-container">
          <button class="copy-code-btn" onclick="copySnippet(this)">${i18n.scenarios.copyFix || "Copiar Fix SQL"}</button>
          <pre><code>${escapeHtml(fixCode)}</code></pre>
        </div>

        <div class="fix-badge-row">
          <span style="font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-muted); align-self: center;">${i18n.scenarios.validatedEngines || "Motores validados:"}</span>
          ${enginesList.map(eng => `<span class="fix-engine-badge">${escapeHtml(eng)}</span>`).join("")}
        </div>
      </div>
    `;
  }
}

const renderScenarioDetail = renderScenarioStage;

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
// 7. TRACE VISUALIZER SHOWCASE CONTROLLER (View: #view-visualizer)
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
  
  const statusPill = document.getElementById("vizStatusPill");
  if (statusPill) {
    const lang = window.currentLang || "pt";
    const statusText = (I18N[lang] && I18N[lang].visualizer && I18N[lang].visualizer.statusDetected) ? I18N[lang].visualizer.statusDetected : "P4_LOST_UPDATE detected at t=184μs";
    statusPill.innerHTML = `<span>⚡</span> ${statusText}`;
  }
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

  const lang = window.currentLang || "pt";
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
  const collisionLabelPrefix = (I18N[lang] && I18N[lang].visualizer && I18N[lang].visualizer.collisionLabel) ? I18N[lang].visualizer.collisionLabel : "P4 Collision";
  html += `
    <div class="gantt-collision-marker" style="left: calc(80px + (100% - 80px) * (${collisionX} / 100));">
      <div class="gantt-collision-label">${collisionLabelPrefix} (${vizMode === "raw" ? '184μs' : '155μs'})</div>
    </div>
  `;

  const workerWord = (I18N[lang] && I18N[lang].visualizer && I18N[lang].visualizer.workerLabel) ? I18N[lang].visualizer.workerLabel : "Worker";

  workers.forEach(w => {
    if (vizWorkerFilter !== "all" && vizWorkerFilter !== w.toString()) return;

    const workerOps = filteredOps.filter(o => o.worker === w);
    html += `
      <div class="gantt-lane">
        <div class="gantt-lane-label">${workerWord} ${w}</div>
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

  const lang = window.currentLang || "pt";
  const vI18n = (I18N[lang] && I18N[lang].visualizer) ? I18N[lang].visualizer : {};

  const title = vI18n.inspectorTitle || "Inspetor de Operações & Queries";
  const txLbl = vI18n.inspectorTx || "Transação / Worker:";
  const tsLbl = vI18n.inspectorTimestamp || "Timestamp / Latência:";
  const execSuffix = vI18n.inspectorExecution || "de execução";
  const paramsLbl = vI18n.inspectorParams || "Parâmetros / Variáveis:";
  const graphLbl = vI18n.inspectorGraph || "Grafo de Conflito:";
  const cycleText = vI18n.inspectorCycleDetected || "T1 ──(rw)──► T2 ──(ww)──► T1 [CYCLE DETECTED]";
  const serialText = vI18n.inspectorSerializable || "Linha serializável sem ciclos";

  inspector.innerHTML = `
    <div class="viz-pane-header">
      <span class="viz-pane-title">${title}</span>
      <span style="font-family: var(--font-mono); font-size: 0.72rem; color: ${op.type === 'conflict' ? 'var(--color-red)' : 'var(--color-green)'}; font-weight: 600;">${op.status}</span>
    </div>
    <div class="inspector-grid">
      <span class="inspector-lbl">${txLbl}</span>
      <span class="inspector-val"><strong>${op.tx}</strong> (Goroutine Worker ${op.worker})</span>

      <span class="inspector-lbl">${tsLbl}</span>
      <span class="inspector-val">${op.startUs}μs (+${op.durationUs}μs ${execSuffix})</span>

      <span class="inspector-lbl">${paramsLbl}</span>
      <span class="inspector-val" style="color: var(--color-yellow); font-family: var(--font-mono); font-size: 0.8rem;">${op.vars}</span>

      <span class="inspector-lbl">${graphLbl}</span>
      <span class="inspector-val" style="font-family: var(--font-mono); font-size: 0.8rem; color: ${op.type === 'conflict' ? 'var(--color-red)' : 'var(--text-primary)'};">${op.type === "conflict" ? cycleText : serialText}</span>

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
// 8. ISOLATION MATRIX CONTROLLER (View: #view-matrix)
// ============================================================================
function renderMatrixView() {
  const lang = window.currentLang || "pt";
  const permitted = (I18N[lang] && I18N[lang].matrix && I18N[lang].matrix.statusPermitted) ? I18N[lang].matrix.statusPermitted : (lang === "pt" ? "PERMITIDO" : "PERMITTED");
  const prevented = (I18N[lang] && I18N[lang].matrix && I18N[lang].matrix.statusPrevented) ? I18N[lang].matrix.statusPrevented : (lang === "pt" ? "PREVENIDO" : "PREVENTED");

  document.querySelectorAll(".matrix-table .badge-permitted").forEach(badge => {
    badge.textContent = permitted;
  });
  document.querySelectorAll(".matrix-table .badge-prevented").forEach(badge => {
    badge.textContent = prevented;
  });
  document.querySelectorAll(".matrix-table .badge-cycle").forEach(badge => {
    if (lang === "pt") {
      badge.textContent = badge.textContent.replace("DETECTED", "DETECTADO");
    } else {
      badge.textContent = badge.textContent.replace("DETECTADO", "DETECTED");
    }
  });
}

// ============================================================================
// 9. DYNAMIC BILINGUAL I18N ENGINE (PT & EN)
// ============================================================================
function getNestedTranslation(obj, path) {
  if (!obj || !path) return "";
  const parts = path.split(".");
  let current = obj;
  for (const part of parts) {
    if (current && typeof current === "object" && part in current) {
      current = current[part];
    } else {
      return null;
    }
  }
  return current;
}

function setLanguage(lang) {
  if (lang !== "pt" && lang !== "en") {
    lang = "pt";
  }

  window.currentLang = lang;
  if (document.documentElement) {
    document.documentElement.lang = lang;
  }

  // Synchronize .active class and aria-pressed attributes
  document.querySelectorAll(".lang-btn[data-lang]").forEach(btn => {
    const btnLang = btn.getAttribute("data-lang");
    if (btnLang === lang) {
      btn.classList.add("active");
      btn.setAttribute("aria-pressed", "true");
    } else {
      btn.classList.remove("active");
      btn.setAttribute("aria-pressed", "false");
    }
  });

  // Translate all [data-i18n] elements
  const dict = I18N[lang] || I18N.pt;
  document.querySelectorAll("[data-i18n]").forEach(el => {
    const key = el.getAttribute("data-i18n");
    const val = getNestedTranslation(dict, key);
    if (val !== null && val !== undefined) {
      el.textContent = val;
    }
  });

  // Translate all [data-i18n-attr] elements
  document.querySelectorAll("[data-i18n-attr]").forEach(el => {
    const attrSpec = el.getAttribute("data-i18n-attr");
    if (!attrSpec) return;
    attrSpec.split(";").forEach(entry => {
      const colonIdx = entry.indexOf(":");
      if (colonIdx === -1) return;
      const attr = entry.slice(0, colonIdx).trim();
      const key = entry.slice(colonIdx + 1).trim();
      if (attr && key) {
        const val = getNestedTranslation(dict, key);
        if (val !== null && val !== undefined) {
          el.setAttribute(attr, val);
        }
      }
    });
  });

  // Reset terminal simulator text to localized initial state
  resetTerminal();

  // Instant view re-rendering
  if (currentRoute === "docs") {
    renderDocsSidebar();
    loadDocChapter(currentDocChapterId);
  } else if (currentRoute === "scenarios") {
    renderScenarioNav();
    renderScenarioStage();
  } else if (currentRoute === "visualizer") {
    initVisualizer();
  } else if (currentRoute === "matrix") {
    renderMatrixView();
  }

  // Persist preference to localStorage
  try {
    localStorage.setItem("chaossql_lang", lang);
  } catch (e) {
    console.warn("Could not save language to localStorage:", e);
  }
}

function initLanguage() {
  let savedLang = null;
  try {
    savedLang = localStorage.getItem("chaossql_lang");
  } catch (e) {
    console.warn("localStorage not accessible:", e);
  }

  if (!savedLang || (savedLang !== "pt" && savedLang !== "en")) {
    const navLang = (navigator.language || "").toLowerCase();
    savedLang = navLang.startsWith("pt") ? "pt" : "en";
  }

  // Register click listeners on all language switcher buttons
  document.querySelectorAll(".lang-btn[data-lang]").forEach(btn => {
    btn.addEventListener("click", (e) => {
      e.preventDefault();
      const targetLang = btn.getAttribute("data-lang");
      if (targetLang && (targetLang === "pt" || targetLang === "en")) {
        setLanguage(targetLang);
      }
    });
  });

  setLanguage(savedLang);
}

// ============================================================================
// 10. GLOBAL UTILITIES & BOOTSTRAP
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
    const lang = window.currentLang || "pt";
    const copiedLabel = (I18N[lang] && I18N[lang].docs && I18N[lang].docs.copied) ? I18N[lang].docs.copied : "Copiado!";
    const orig = button.textContent;
    button.textContent = copiedLabel;
    setTimeout(() => button.textContent = orig, 1500);
  }
}

function escapeHtml(str) {
  if (typeof str !== "string") return str;
  return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

// Robust Bootstrap
function bootstrap() {
  initLanguage();
  setupGlobalUI();
  setupTerminalSimulator();
  initDocsHub();
  setupScenarioTabs();
  initRouter();
  handleRoute();
}

if (typeof document !== "undefined") {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", bootstrap);
  } else {
    bootstrap();
  }
}

// Global Exports
if (typeof window !== "undefined") {
  window.setLanguage = setLanguage;
  window.getNestedTranslation = getNestedTranslation;
  window.I18N = I18N;
  window.SCENARIOS = SCENARIOS;
  window.renderDocsSidebar = renderDocsSidebar;
  window.renderDocChapter = renderDocChapter;
  window.loadDocChapter = loadDocChapter;
  window.renderScenarioList = renderScenarioList;
  window.renderScenarioNav = renderScenarioNav;
  window.renderScenarioDetail = renderScenarioDetail;
  window.renderScenarioStage = renderScenarioStage;
  window.renderMatrixView = renderMatrixView;
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = { I18N, SCENARIOS };
}
