# ChaosSQL SaaS — Passo 5: Concurrency Health Dashboard MVP & Finding Detail (Design Spec)

- **Status:** Approved (Commercial Impact Lens - Sections 15 & 16 of 90-Days Plan)
- **Date:** 2026-09-08
- **Author:** Bregalda Engineering & Antigravity
- **Scope:** Fase 3 / Semanas 6-8 do Plano de 90 Dias (Observabilidade de Concorrência, Concurrency Health Score, Histórico de Runs e Inspector de Findings)

---

## 1. Visão Geral & Objetivo

Conforme o **Plano de 90 Dias (Seção 15)**:
> *"Evitar dashboard enorme. O dashboard existe para análise histórica e saúde de concorrência contínua."*

O **Concurrency Health Dashboard MVP** é a interface central onde engenheiros, tech leads e gerentes de engenharia:
1. Monitoram o índice de saúde de concorrência da organização (**Concurrency Health %**).
2. Rastreiam todas as execuções de CI disparadas pelos repositórios.
3. Identificam regressões em tempo real introduzidas em Pull Requests antes de chegarem à produção.
4. Inspecionam o **Finding Detail (Seção 16)**: isolamento, violação de invariantes, rastro causal minimalista de `T1/T2`, download de reprodutores Go e abertura instantânea no Playground WASM.

---

## 2. Estrutura da Interface

```text
┌──────────────────────────────────────────────────────────────────────────┐
│ SiteNav: [Início] [Docs] [Cenários] [Visualizer] [Matrix] [Playground] [Dashboard ●]
├──────────────────────────────────────────────────────────────────────────┤
│ HEADER: ChaosSQL Cloud Dashboard                                          │
│ Org: Acme Corp · Concurrency Health: 98.7% (Healthy)                      │
│                                                                          │
│ [98.7% Health]  [184 Runs]  [284,128 Schedules]  [3 Regressões] [1 Aberta]│
├──────────────────────────────────────────────────────────────────────────┤
│ FILTROS: [Todos os Runs] [Regressões Apenas 🚨] [Pull Requests] [Passed]  │
│ Busca: [Pesquisar repositório ou cenário...                           ]  │
├──────────────────────────────────────────────────────────────────────────┤
│ RECENT RUNS TABLE                                                        │
│ Repositório    Branch/PR       Status   Finding      Driver       Ação   │
│ acme/payments  #382 (feat/...) ❌ FAIL   P4 Lost Upd  Postgres 16  [Ver]  │
│ acme/payments  main            ✅ PASS   Clean        Postgres 16  [Ver]  │
│ acme/checkout  #104 (fix/...)  ❌ FAIL   Deadlock     MySQL 8.0    [Ver]  │
├──────────────────────────────────────────────────────────────────────────┤
│ FINDING DETAIL MODAL (SEÇÃO 16)                                          │
│ • Anomaly: P4 — Lost Update                                              │
│ • Invariant: sum(balance) == 2000 (Actual: 1950)                         │
│ • Minimal Trace: T1: SELECT -> T2: SELECT -> T1: UPDATE -> T2: UPDATE    │
│ • [Download Repro Go Test] [Copy CLI Reproduce] [Open in WASM Playground]│
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Design Tokens & Bregalda Design System
- Fundo canvas: `--cream: #FCFBF8`
- Painéis escuros de alta densidade técnica: `--ink: #181226` / `#12151c`
- Indicadores: `--green: #22C55E`, `--yellow: #F5C400`, `--red: #EF4444`, `--purple: #4B2E83`
- Tipografia: Inter para títulos e métricas, JetBrains Mono para código, hashes e expressões.
- Precisão: 3px para botões/filtros, 9px para cartões e modais.
