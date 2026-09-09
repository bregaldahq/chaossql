# ChaosSQL SaaS — Passo 6: Pricing, Billing Limits & Concurrency Audit Product Pack Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o catálogo de precificação e monetização do SaaS no portal web (`site/src/pages/PricingPage.tsx`) com foco em cobrança repository-based, pacote de Concurrency Safety Audit ($1.490), e suporte no backend (`internal/server/billing.go`) para consulta e validação de limites de planos.

---

### Task 1: Módulo de Planos e Assinaturas no Servidor (`internal/server/billing.go` e `billing_test.go`)
- Create: `/root/chaossql/internal/server/billing.go`
- Test: `/root/chaossql/internal/server/billing_test.go`
- Interfaces:
  - `type PlanConfig struct`
  - `GetPlan(id string) PlanConfig`
  - `(s *Store) GetOrgSubscription(orgID string) (*OrgSubscription, error)`
  - Endpoint `GET /v1/organizations/{id}/subscription` em `handlers.go`

### Task 2: Página de Preços & Pacote de Auditoria (`site/src/pages/PricingPage.tsx` e `.module.css`)
- Implementar toggle Mensal / Anual (-20%).
- Implementar os 4 cards de planos: Community ($0), Developer ($0), Team ($39/mo - Popular), Pro ($99/mo).
- Implementar a seção em destaque: **ChaosSQL Concurrency Audit ($1.490)**.
- Implementar modal interativo de contratação com lead capture dual.

### Task 3: Integração na Navegação (`SiteNav.tsx` e `App.tsx`)
- Adicionar rota `#/pricing` no roteamento hash.
- Adicionar link "Preços" no menu principal de navegação.

### Task 4: Verificação Completa e Merge
- Validar suíte Go (`go test ./...`).
- Validar build Vite (`npm run build`).
- Merge `feat/saas-step6-pricing-billing-audit` na branch `main` e push para `origin`.
