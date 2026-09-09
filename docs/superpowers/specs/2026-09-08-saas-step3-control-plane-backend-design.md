# ChaosSQL SaaS — Passo 3: Backend Control Plane (`chaossql-server`) (Design Spec)

- **Status:** Approved (Commercial Impact Lens)
- **Date:** 2026-09-08
- **Author:** Bregalda Engineering & Antigravity
- **Scope:** Fase 2 / Semana 3 e 4 do Plano de 90 Dias (Control Plane, Persistência, Ingestão de Runs e Detecção de Regressão)

---

## 1. Visão Geral & Objetivo

### 1.1 Contexto
Nos Passos 1 e 2, estabelecemos o posicionamento público do produto e instrumentamos o motor ChaosSQL (CLI e GitHub Action) com o cliente de ingestão e sanitização de dados.

Agora, implementamos o **ChaosSQL Control Plane (`chaossql-server`)**, o serviço backend em Go responsável por:
1. Autenticar runners de CI via Bearer Token.
2. Ingerir execuções (`POST /v1/runs`).
3. Persistir repositórios, execuções e findings.
4. Manter o baseline da branch principal (`main`).
5. **Detectar Regressões de Concorrência em tempo real:** Comparar Pull Requests com o baseline da branch principal e sinalizar instantaneamente se uma anomalia (ex: P4 Lost Update) foi introduzida.

---

## 2. Arquitetura do Sistema

```text
GitHub Actions Runner / CLI
         │
         │ POST /v1/runs (Bearer Token)
         ▼
┌──────────────────────────────────────────────────────────┐
│                   chaossql-server                        │
│                                                          │
│  ┌────────────────────┐      ┌────────────────────────┐  │
│  │ Auth Middleware    │ ───► │ Ingest Handler         │  │
│  │ Bearer Token Check │      │ POST /v1/runs          │  │
│  └────────────────────┘      └────────────────────────┘  │
│                                          │               │
│                                          ▼               │
│                              ┌────────────────────────┐  │
│                              │ Regression Engine      │  │
│                              │ Baseline vs PR Compare │  │
│                              └────────────────────────┘  │
│                                          │               │
│                                          ▼               │
│                              ┌────────────────────────┐  │
│                              │ Store Interface        │  │
│                              │ SQLite (Dev) / PG(Prod)│  │
│                              └────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
```

---

## 3. Modelo de Dados

### 3.1 Entidades
1. **`organizations`**: `id`, `name`, `created_at`
2. **`api_tokens`**: `id`, `org_id`, `token_hash`, `name`, `created_at`
3. **`repositories`**: `id`, `org_id`, `full_name` (ex: `bregaldahq/payments`), `default_branch`, `created_at`
4. **`scenarios`**: `id`, `repo_id`, `name`, `driver`, `created_at`
5. **`runs`**: `id`, `repo_id`, `scenario_id`, `commit_sha`, `branch`, `pr_number`, `status` ("passed"|"failed"), `anomaly_type`, `seed`, `duration_ms`, `created_at`
6. **`findings`**: `id`, `run_id`, `anomaly_type`, `assertion`, `minimal_ops_count`, `repro_code`, `trace_json`, `created_at`
7. **`baselines`**: `id`, `repo_id`, `scenario_id`, `branch`, `last_successful_run_id`, `updated_at`

---

## 4. Endpoints REST da API v1

- `POST /v1/runs`: Ingestão de execução (Autenticado).
- `GET /v1/runs/{run_id}`: Consulta detalhada de uma execução e seu finding.
- `GET /v1/repositories/{owner}/{name}/runs`: Histórico de execuções paginado.
- `GET /v1/health`: Healthcheck do servidor (`{"status": "ok"}`).

---

## 5. Algoritmo de Detecção de Regressão

1. Quando um run é recebido:
   - Se `branch == repo.default_branch` (ex: `main`) e `status == "passed"`:
     - O run é registrado como o novo **baseline** do cenário.
   - Se `pr_number > 0` ou `branch != repo.default_branch`:
     - O sistema consulta o último baseline para o mesmo cenário na branch `main`.
     - Se o baseline era `passed` e o run atual é `failed`:
       - `is_regression = true`
       - Retorna os dados do baseline para gerar o alerta de regressão.

---

## 6. Verificação e Testes

1. Testes unitários do `Store` em memória / SQLite.
2. Testes unitários do `RegressionEngine`.
3. Testes de integração HTTP cobrindo autenticação, ingestão e detecção de regressão.
4. Teste ponta a ponta disparando `chaossql run` contra `chaossql-server`.
