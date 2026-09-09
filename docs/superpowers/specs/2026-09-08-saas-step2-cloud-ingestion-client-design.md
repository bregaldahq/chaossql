# ChaosSQL SaaS — Passo 2: Cliente de Ingestão Cloud & Instrumentação da Action (Design Spec)

- **Status:** Approved (Commercial Impact Lens)
- **Date:** 2026-09-08
- **Author:** Bregalda Engineering & Antigravity
- **Scope:** Frente 2A da Fase 2 (Contrato de Ingestão v1, Pacote internal/cloud, Flags CLI e GitHub Action)

---

## 1. Visão Geral & Objetivo

### 1.1 Contexto
Para habilitar o valor comercial do **ChaosSQL Cloud** (histórico centralizado, baselines de branch principal, detecção de regressão em Pull Requests e comentários automatizados), o motor ChaosSQL precisa ser capaz de publicar de forma segura e determinística os resultados de execução diretamente para a API do Cloud.

### 1.2 Princípios Centrais
1. **Segurança e Privacidade Absolutas:** O banco de dados do cliente e os dados reais NUNCA saem do runner de CI. Apenas metadados de execução, anomalias classificadas e o rastro causal mínimo (sem credenciais ou dados pessoais) são enviados.
2. **Resiliência e Tolerância a Falhas:** Uma falha de rede temporária na nuvem não deve quebrar a suíte de testes do cliente no CI, a menos que a flag `--cloud-fail-fast` seja explicitamente solicitada.
3. **Zero Configuração no CI:** Quando executado dentro do GitHub Actions, o ChaosSQL detecta automaticamente o repositório, commit SHA, branch e número do PR sem necessidade de flags manuais.

---

## 2. Contrato de Ingestão da API v1 (`POST /v1/runs`)

### 2.1 Endpoint
- **URL:** `POST {cloud-url}/v1/runs`
- **Headers:**
  - `Authorization: Bearer <cloud-token>`
  - `Content-Type: application/json`
  - `User-Agent: ChaosSQL-CLI/<version> (<os>/<arch>)`

### 2.2 Payload de Requisição (`RunIngestRequest`)
```json
{
  "version": "1.0",
  "ci": {
    "provider": "github-actions",
    "repository": "bregaldahq/payments-service",
    "commit_sha": "a8129c1f73b8...",
    "branch": "feature/wallet-concurrency",
    "base_branch": "main",
    "pull_request_number": 382,
    "run_id": "1283918239",
    "actor": "octocat"
  },
  "scenario": {
    "name": "wallet_transfer",
    "fingerprint": "sha256:8f2a...",
    "driver": "postgres",
    "driver_version": "16.2",
    "workers": 4,
    "iterations": 100,
    "seed": 184729
  },
  "result": {
    "status": "failed",
    "success": false,
    "violation_detected": true,
    "anomaly_type": "P4",
    "duration_ms": 482,
    "total_schedules": 100,
    "failed_schedules": 31,
    "failing_invariant": {
      "name": "total_wealth_conserved",
      "query": "SELECT sum(balance) AS total FROM accounts;",
      "assertion": "total == 2000",
      "actual": "1950"
    }
  },
  "reproduction": {
    "minimal_operations_count": 4,
    "shrink_duration_ms": 120,
    "repro_go_code": "package repro_test...",
    "mermaid_diagram": "sequenceDiagram...",
    "sanitized_minimal_trace": [
      { "worker": "T1", "op_type": "read", "table": "accounts", "sql": "SELECT balance FROM accounts WHERE id = 1" },
      { "worker": "T2", "op_type": "read", "table": "accounts", "sql": "SELECT balance FROM accounts WHERE id = 1" },
      { "worker": "T1", "op_type": "write", "table": "accounts", "sql": "UPDATE accounts SET balance = 1950 WHERE id = 1" },
      { "worker": "T2", "op_type": "write", "table": "accounts", "sql": "UPDATE accounts SET balance = 1900 WHERE id = 1" }
    ]
  }
}
```

### 2.3 Resposta do Servidor (`RunIngestResponse`)
```json
{
  "success": true,
  "run_id": "run_01j7k9...",
  "url": "https://chaossql.bregalda.com/runs/run_01j7k9...",
  "is_regression": true,
  "baseline": {
    "run_id": "run_01j6a4...",
    "status": "passed",
    "commit_sha": "f4b18c...",
    "branch": "main"
  },
  "pr_comment": {
    "posted": true,
    "comment_id": "1928491"
  }
}
```

---

## 3. Arquitetura do Pacote `internal/cloud`

### 3.1 Componentes e Responsabilidades
1. **`internal/cloud/types.go`**: Modelos de dados fortemente tipados com tags JSON estritas.
2. **`internal/cloud/client.go`**:
   - Cliente HTTP configurado com timeouts controlados (`10s`).
   - Algoritmo de Retry com exponential backoff (até 3 tentativas) para erros transitórios (HTTP 502, 503, 504, conexões truncadas).
   - Autenticação Bearer Token.
3. **`internal/cloud/ci.go`**:
   - Extrator de contexto de CI: lê variáveis de ambiente (`GITHUB_REPOSITORY`, `GITHUB_SHA`, `GITHUB_REF_NAME`, `GITHUB_BASE_REF`, etc.).
   - Parser seguro para número de Pull Request via `GITHUB_REF` (`refs/pull/<N>/merge`) ou `GITHUB_EVENT_PATH`.
4. **`internal/cloud/sanitizer.go`**:
   - Sanitizador de statements SQL: remove credenciais conhecidas, tokens de API e mascara literals sensíveis antes do despacho.

---

## 4. Integração no CLI (`cmd/chaossql`)

### 4.1 Novas Flags no `chaossql run`
- `--cloud-token`: Token de API da organização/repositório (suporta env `CHAOSSQL_CLOUD_TOKEN`).
- `--cloud-url`: URL base da API Cloud (default: `https://api.chaossql.bregalda.com`, suporta env `CHAOSSQL_CLOUD_URL`).
- `--cloud-fail-fast`: Se habilitado, uma falha na nuvem aborta o processo com exit code 1 (default: `false`).

### 4.2 Feedback no Terminal
```text
[☁️ ChaosSQL Cloud] Publishing execution metadata to https://api.chaossql.bregalda.com...
[✓] Run recorded: https://chaossql.bregalda.com/runs/run_01j7k9a...
[🚨 REGRESSION DETECTED]
  Scenario: wallet_transfer
  Baseline (main): PASS
  Current PR: FAIL (P4 Lost Update)
```

---

## 5. Atualização da GitHub Action (`action.yml`)

### 5.1 Novos Inputs
- `cloud-token`: Token da organização ChaosSQL Cloud (`secrets.CHAOSSQL_CLOUD_TOKEN`).
- `cloud-url`: URL opcional para self-hosted ou staging.

### 5.2 Novos Outputs
- `cloud-run-id`: ID único da execução no Cloud.
- `cloud-run-url`: Link direto para visualização do trace no Cloud.
- `is-regression`: Booleano (`true` / `false`) indicando se uma regressão foi detectada.

---

## 6. Plano de Testes & Verificação

1. **Testes Unitários de Sanitização e CI Extraction (`internal/cloud`):**
   - Testar extração de metadados em mocks de GitHub Actions.
   - Testar sanitização de traces com queries contendo dados mascaráveis.
2. **Testes de Integração com Servidor Mock (`httptest.Server`):**
   - Testar fluxo de envio com sucesso (HTTP 200).
   - Testar comportamento com erro de autenticação (HTTP 401).
   - Testar retries com falhas 503 transitórias e recuperação no 3º retry.
   - Testar tolerância a falhas quando o servidor está offline.
3. **Verificação do CLI:**
   - Executar `chaossql run examples/banking_lost_update/chaos.yaml --cloud-token=test_token` contra servidor de teste mock.
4. **Verificação da Action:**
   - Validar sintaxe do `action.yml`.
