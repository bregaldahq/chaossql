# ChaosSQL SaaS — Passo 4: PR Experience & GitHub Action Comment Bot Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o bot gerador de relatórios de Pull Request e resumo de CI (`$GITHUB_STEP_SUMMARY` e GitHub Issue/PR Comments API), garantindo que desenvolvedores recebam evidências instantâneas de concorrência direto no GitHub.

**Architecture:** O pacote `internal/cloud` recebe `pr_reporter.go` com formatação markdown rica, suporte ao `$GITHUB_STEP_SUMMARY` e integração com a API do GitHub via HTTP client resiliente. O CLI `cmd/chaossql/main.go` e a `action.yml` orquestram a publicação automática.

---

### Task 1: Gerador de Markdown e Publicador de PR (`internal/cloud/pr_reporter.go`)
- Create: `/root/chaossql/internal/cloud/pr_reporter.go`
- Test: `/root/chaossql/internal/cloud/pr_reporter_test.go`
- Interfaces:
  - `FormatPRMarkdown(req *RunIngestRequest, resp *RunIngestResponse) string`
  - `WriteStepSummary(summaryPath, markdown string) error`
  - `PostPRComment(ctx context.Context, client *http.Client, ghToken, repo string, prNumber int, body string) (string, error)`

### Task 2: Integração no CLI (`cmd/chaossql/main.go`)
- Modify: `/root/chaossql/cmd/chaossql/main.go`
- Test: `/root/chaossql/cmd/chaossql/cloud_integration_test.go`
- Add flags `--github-token`, `--pr-comment`.
- Post comment and step summary after run.

### Task 3: Atualização da GitHub Action (`action.yml`)
- Modify: `/root/chaossql/action.yml`
- Add input `github-token` with default `${{ github.token }}` and `post-pr-comment`.

### Task 4: Verificação Completa e Merge
- Run all tests.
- Merge `feat/saas-step4-pr-bot-and-reporter` to `main` and push.
