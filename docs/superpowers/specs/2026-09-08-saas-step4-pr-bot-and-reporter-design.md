# ChaosSQL SaaS — Passo 4: PR Experience & GitHub Action Comment Bot (Design Spec)

- **Status:** Approved (Commercial Impact Lens - Section 7 of 90-Days Plan)
- **Date:** 2026-09-08
- **Author:** Bregalda Engineering & Antigravity
- **Scope:** Fase 3 / Semanas 5-7 do Plano de 90 Dias (Pull Request Experience, Automated PR Comment, GitHub Step Summary)

---

## 1. Visão Geral & Objetivo Comercial

De acordo com o **Plano de 90 Dias (Seção 7)**:
> *"O principal ponto de interação do usuário não deveria ser o dashboard. Deve ser o GitHub. O dashboard existe para análise histórica. O Pull Request deve entregar valor instantaneamente."*

O PR bot do ChaosSQL é a principal alavanca de **ativação, retenção e expansão viral** da plataforma:
1. Todo desenvolvedor e revisor de código que abre um Pull Request é impactado diretamente pela análise de concorrência.
2. Quando uma regressão de concorrência ocorre (ex: P4 Lost Update ou Deadlock), o bot entrega a evidência incontestável:
   - Status da branch principal vs status do PR.
   - Rastro causal mínimo com passos exatos das threads (`T1`, `T2`).
   - Violação de invariante (`Expected: 2000`, `Actual: 1950`).
   - Link direto para o Cloud e comando de reprodução local.
3. Se não houver regressão e o teste passar, o bot emite um selo de integridade de concorrência limpo e conciso.

---

## 2. Arquitetura do Componente

```text
               ChaosSQL CLI / Runner
                         │
        ┌────────────────┴────────────────┐
        │                                 │
        ▼                                 ▼
   Cloud Ingest API              PR Reporter Generator
 (POST /v1/runs)             (internal/cloud/pr_reporter.go)
        │                                 │
        ▼                                 ├────────────────────────┐
   Run Response                           ▼                        ▼
 (IsRegression, ID)             $GITHUB_STEP_SUMMARY      GitHub PR Comment
                                 (Actions UI Summary)   (POST /repos/.../comments)
```

---

## 3. Formato do Comentário no Pull Request

### Cenário A: Regressão Detectada (P0 Alerta Vermelho)
```markdown
## ChaosSQL ❌ Concurrency Regression Detected

> **Pull Request introduces a concurrency violation that broke the default branch baseline.**

| Metric | Value |
| :--- | :--- |
| **Scenario** | `wallet_transfer` |
| **Anomaly** | **P4 (Lost Update)** |
| **Baseline (`main`)** | `PASS` |
| **Current PR** | `FAIL` (31/100 schedules failed) |
| **Database** | PostgreSQL 16 (READ COMMITTED) |

### Failing Invariant
- **Assertion:** `balance_sum: total == 2000`
- **Actual:** `1950`

### Minimal Causal Trace
```sql
T1: SELECT balance FROM accounts WHERE id = 1;
T2: SELECT balance FROM accounts WHERE id = 1;
T1: UPDATE accounts SET balance = balance - 50 WHERE id = 1;
T2: UPDATE accounts SET balance = balance - 50 WHERE id = 1;
```

[🔍 View Detailed Trace in ChaosSQL Cloud](https://app.chaossql.bregalda.com/runs/run_123) • [⚡ Run Reproducer Locally](https://app.chaossql.bregalda.com/runs/run_123#repro)
```

### Cenário B: Execução com Sucesso (PASS)
```markdown
## ChaosSQL ✅ Concurrency Verification Passed

All **100** interleaved schedules passed without concurrency anomalies.

- **Scenario:** `wallet_transfer`
- **Isolation:** `READ COMMITTED`
- **Baseline Status:** Up to date
```

---

## 4. Estrutura de Código

1. `internal/cloud/pr_reporter.go`:
   - `FormatPRMarkdown(req *RunIngestRequest, resp *RunIngestResponse) string`
   - `WriteStepSummary(markdown string) error`
   - `PostPRComment(ctx context.Context, ghToken, repo string, prNumber int, body string) (string, error)`
2. `internal/cloud/pr_reporter_test.go`:
   - Teste de formatação para regressão, falha simples e sucesso.
   - Teste de envio de comentário via mock HTTP da GitHub API (`api.github.com`).
3. `cmd/chaossql/main.go`:
   - Flags `--github-token` (env `GITHUB_TOKEN`), `--pr-comment` (padrão true se token existir).
   - Execução automática de `WriteStepSummary` e `PostPRComment`.
4. `action.yml`:
   - Configuração do input `github-token` com default `${{ github.token }}`.
