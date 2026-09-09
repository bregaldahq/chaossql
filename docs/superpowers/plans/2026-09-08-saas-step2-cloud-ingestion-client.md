# ChaosSQL SaaS — Passo 2: Cliente de Ingestão Cloud & Instrumentação da Action Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o cliente de ingestão Cloud no ChaosSQL Core (`internal/cloud`), adicionar as flags `--cloud-token` e `--cloud-url` no CLI (`cmd/chaossql`), atualizar `action.yml` e garantir detecção e publicação automática de resultados de CI com sanitização e tolerância a falhas.

**Architecture:** O pacote `internal/cloud` fornece structs de contrato v1 (`types.go`), extração automática de metadados de CI (`ci.go`), sanitização de traces (`sanitizer.go`) e um cliente HTTP resiliente com retries exponenciais (`client.go`). A CLI (`cmd/chaossql/main.go`) e `action.yml` orquestram o despacho pós-execução quando `--cloud-token` ou `CHAOSSQL_CLOUD_TOKEN` estiverem presentes.

**Tech Stack:** Go 1.22+, `net/http`, `net/http/httptest`, Cobra CLI, GitHub Actions Composite Action.

**Spec:** `docs/superpowers/specs/2026-09-08-saas-step2-cloud-ingestion-client-design.md`

## Global Constraints

- O banco e dados sensíveis do cliente NUNCA devem sair do runner local.
- Resiliência: falhas temporárias na nuvem nunca devem quebrar os testes no CI por padrão (`--cloud-fail-fast` default: `false`).
- Zero CGO: manter compatibilidade pura em Go (`CGO_ENABLED=0`).
- Manter 100% de cobertura de testes no pacote `internal/cloud` usando `httptest.Server`.

---

### Task 1: Criar Modelos de Dados e Tipos do Contrato v1

**Files:**
- Create: `/root/chaossql/internal/cloud/types.go`
- Test: `/root/chaossql/internal/cloud/types_test.go`

**Interfaces:**
- Produces: Structs `RunIngestRequest`, `RunIngestResponse`, `CIContext`, `ScenarioMetadata`, `ExecutionSummary`, `ReproductionData`, `SanitizedTraceEvent`, `BaselineComparison`.

- [ ] **Step 1: Escrever teste de serialização/deserialização JSON em types_test.go**
- [ ] **Step 2: Implementar tipos fortemente tipados em types.go**
- [ ] **Step 3: Executar testes para validar conformidade do contrato v1**
- [ ] **Step 4: Commitar Task 1**

---

### Task 2: Implementar Extrator de Contexto de CI

**Files:**
- Create: `/root/chaossql/internal/cloud/ci.go`
- Test: `/root/chaossql/internal/cloud/ci_test.go`

**Interfaces:**
- Produces: `DetectCIContext() *CIContext` com parsing de `GITHUB_REPOSITORY`, `GITHUB_SHA`, `GITHUB_REF_NAME`, PR Number e fallback local Git.

- [ ] **Step 1: Escrever testes unitários em ci_test.go mockando variáveis de ambiente**
- [ ] **Step 2: Implementar lógica de extração segura em ci.go**
- [ ] **Step 3: Executar testes de extração de CI**
- [ ] **Step 4: Commitar Task 2**

---

### Task 3: Implementar Sanitizador de Traces e Queries

**Files:**
- Create: `/root/chaossql/internal/cloud/sanitizer.go`
- Test: `/root/chaossql/internal/cloud/sanitizer_test.go`

**Interfaces:**
- Produces: `SanitizeTrace(trace []domain.TraceEvent) []SanitizedTraceEvent` que remove segredos e mascara literais sensíveis.

- [ ] **Step 1: Escrever testes unitários em sanitizer_test.go**
- [ ] **Step 2: Implementar algoritmo de sanitização em sanitizer.go**
- [ ] **Step 3: Executar testes unitários do sanitizador**
- [ ] **Step 4: Commitar Task 3**

---

### Task 4: Implementar Cliente HTTP Resiliente com Retries

**Files:**
- Create: `/root/chaossql/internal/cloud/client.go`
- Test: `/root/chaossql/internal/cloud/client_test.go`

**Interfaces:**
- Produces: `type Client struct` com método `PublishRun(ctx context.Context, req *RunIngestRequest) (*RunIngestResponse, error)`.

- [ ] **Step 1: Escrever testes unitários em client_test.go usando httptest.Server (casos: 200 OK, 401 Unauthorized, 503 com retry bem-sucedido, falha de conexão)**
- [ ] **Step 2: Implementar cliente HTTP com timeout de 10s e exponential backoff em client.go**
- [ ] **Step 3: Executar testes unitários do cliente**
- [ ] **Step 4: Commitar Task 4**

---

### Task 5: Integrar Flags e Publicação Cloud no CLI

**Files:**
- Modify: `/root/chaossql/cmd/chaossql/main.go`

**Interfaces:**
- Consumes: Pacote `internal/cloud`.
- Produces: Flags `--cloud-token`, `--cloud-url`, `--cloud-fail-fast` no comando `chaossql run` e `chaossql demo`.

- [ ] **Step 1: Declarar novas flags e vincular a variáveis de ambiente em main.go**
- [ ] **Step 2: Implementar chamada de publicação cloud pós-execução do chaos test**
- [ ] **Step 3: Testar execução com flag --cloud-token contra httptest.Server mock local**
- [ ] **Step 4: Commitar Task 5**

---

### Task 6: Atualizar a GitHub Action

**Files:**
- Modify: `/root/chaossql/action.yml`

**Interfaces:**
- Produces: Inputs `cloud-token`, `cloud-url` e outputs `cloud-run-id`, `cloud-run-url`, `is-regression`.

- [ ] **Step 1: Adicionar inputs e outputs em action.yml**
- [ ] **Step 2: Repassar flags --cloud-token e --cloud-url na execução do binário na action**
- [ ] **Step 3: Commitar Task 6**

---

### Task 7: Verificação Integrada End-to-End

**Files:**
- Test/Verify: `internal/cloud/...`, `cmd/chaossql/...`, `action.yml`

- [ ] **Step 1: Executar todos os testes do repositório Go (`go test -v ./...`)**
- [ ] **Step 2: Executar chaossql run com teste local completo**
- [ ] **Step 3: Finalizar commits da branch**
