# ChaosSQL SaaS — Fase 3 / Passo 1: Onboarding de 60s & Concurrency CI Workflow Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o fluxo de onboarding rápido de 60 segundos para early adopters, adicionar o workflow oficial `.github/workflows/concurrency-ci.yml` para validação contínua em PRs e integrar o modal de conexão de repositórios no Dashboard MVP.

---

### Task 1: Workflow Oficial de Concorrência (`.github/workflows/concurrency-ci.yml`)
- Create: `/root/chaossql/.github/workflows/concurrency-ci.yml`
- Steps:
  - Checkout
  - Setup Go
  - Run ChaosSQL Action contra `examples/banking_lost_update/chaos.yaml`
  - Pass `--cloud-token` e `--github-token`

### Task 2: Experiência de Conexão de Repositório no Dashboard (`site/src/pages/DashboardPage.tsx`)
- Adicionar botão **"+ Conectar Repositório"** no topo da tabela de runs.
- Adicionar modal passo-a-passo com cópia rápida do arquivo YAML `.github/workflows/concurrency.yml`.
- Feedback visual com badges de passos e validação.

### Task 3: Atualizar README.md com o Snippet de 60s
- Adicionar guia de 60 segundos de instalação da Action no `README.md`.

### Task 4: Verificação Completa e Merge
- Executar suíte de testes Go (`go test ./...`).
- Executar build do frontend (`npm run build`).
- Merge da branch `feat/saas-early-adopters-step1-onboarding-ci` na `main` e push para `origin`.
