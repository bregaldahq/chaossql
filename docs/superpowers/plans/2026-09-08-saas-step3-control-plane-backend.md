# ChaosSQL SaaS — Passo 3: Backend Control Plane (`chaossql-server`) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o servidor backend `chaossql-server` em Go com persistência (SQLite/PostgreSQL), autenticação Bearer, ingestão de execuções (`POST /v1/runs`), armazenamento de findings e o motor de detecção de regressão contra o baseline da branch principal.

**Architecture:** O pacote `internal/server` encapsula o armazenamento SQL (`store.go`), o motor de regressão (`regression.go`) e os handlers HTTP REST (`handlers.go`). O entrypoint `cmd/chaossql-server/main.go` fornece um binário CLI independente configurável via flags e variáveis de ambiente.

**Tech Stack:** Go 1.22+, `modernc.org/sqlite` (Pure Go SQLite driver já presente no projeto), `net/http`, Cobra CLI.

**Spec:** `docs/superpowers/specs/2026-09-08-saas-step3-control-plane-backend-design.md`

## Global Constraints

- Reutilizar os tipos do contrato `internal/cloud` (`RunIngestRequest`, `RunIngestResponse`, etc.) diretamente para garantir consistência perfeita.
- Zero CGO: utilizar o driver SQLite puro em Go (`modernc.org/sqlite`) para que o servidor rode sem dependências externas de compilação.
- Manter 100% de cobertura de testes nos handlers e motor de regressão.

---

### Task 1: Store de Dados e Migrações de Schema

**Files:**
- Create: `/root/chaossql/internal/server/store.go`
- Test: `/root/chaossql/internal/server/store_test.go`

**Interfaces:**
- Produces: `type Store struct` com `AutoMigrate()`, `CreateOrGetRepo()`, `SaveRun()`, `GetRun()`, `GetBaseline()`, `SetBaseline()`, `ValidateToken()`.

- [ ] **Step 1: Implementar testes unitários em store_test.go**
- [ ] **Step 2: Implementar store.go com schema SQLite/PostgreSQL e métodos CRUD**
- [ ] **Step 3: Executar testes unitários do Store**
- [ ] **Step 4: Commitar Task 1**

---

### Task 2: Motor de Detecção de Regressão

**Files:**
- Create: `/root/chaossql/internal/server/regression.go`
- Test: `/root/chaossql/internal/server/regression_test.go`

**Interfaces:**
- Produces: `type RegressionEngine struct` com `Evaluate(run *RunRecord) (*BaselineComparison, bool, error)`.

- [ ] **Step 1: Implementar testes unitários em regression_test.go cobrindo cenários: baseline PASS -> PR FAIL (Regressão!), baseline PASS -> PR PASS, e atualização de baseline na main**
- [ ] **Step 2: Implementar lógica de avaliação de regressão em regression.go**
- [ ] **Step 3: Executar testes unitários do motor de regressão**
- [ ] **Step 4: Commitar Task 2**

---

### Task 3: Handlers HTTP REST e Middleware de Autenticação

**Files:**
- Create: `/root/chaossql/internal/server/handlers.go`
- Test: `/root/chaossql/internal/server/handlers_test.go`

**Interfaces:**
- Produces: `NewRouter(store *Store, engine *RegressionEngine) http.Handler` com rotas `POST /v1/runs`, `GET /v1/runs/{id}`, `GET /v1/health`.

- [ ] **Step 1: Implementar testes unitários em handlers_test.go (autenticação Bearer, ingestão, 401 para token inválido)**
- [ ] **Step 2: Implementar handlers HTTP em handlers.go**
- [ ] **Step 3: Executar testes dos handlers HTTP**
- [ ] **Step 4: Commitar Task 3**

---

### Task 4: Entrypoint do Servidor (`cmd/chaossql-server/main.go`)

**Files:**
- Create: `/root/chaossql/cmd/chaossql-server/main.go`

**Interfaces:**
- Produces: Binário `chaossql-server start --port 8080 --db ./data.db`.

- [ ] **Step 1: Implementar CLI Cobra com flags --port, --db, --token e graceful shutdown**
- [ ] **Step 2: Validar compilação do binário**
- [ ] **Step 3: Commitar Task 4**

---

### Task 5: Teste Integrado End-to-End e Merge

**Files:**
- Create: `/root/chaossql/cmd/chaossql-server/e2e_test.go`

- [ ] **Step 1: Implementar e2e_test.go iniciando o servidor e executando o cliente chaossql run**
- [ ] **Step 2: Validar regressão detectada de ponta a ponta**
- [ ] **Step 3: Fazer merge da branch feat/saas-step3-cloud-control-plane na main e push para origin**
