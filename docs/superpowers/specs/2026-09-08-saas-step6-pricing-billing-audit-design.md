# ChaosSQL SaaS — Passo 6: Pricing, Billing Limits & Concurrency Audit Product Pack (Design Spec)

- **Status:** Approved (Commercial Impact Lens - Sections 20, 21, 22, 23 of 90-Days Plan)
- **Date:** 2026-09-08
- **Author:** Bregalda Engineering & Antigravity
- **Scope:** Fase 4 / Semanas 8-10 do Plano de 90 Dias (Tiers Comerciais, Limites de Planos Repository-Based e Produto de Auditoria de Concorrência)

---

## 1. Visão Geral & Objetivo Comercial

Conforme as **Seções 20 a 23 do Plano de 90 Dias**:
1. **Regra de Ouro da Precificação (Seção 21):**
   > *"Não cobrar por usuário inicialmente. Ferramentas de CI ficam mais fáceis de vender quando o preço não cresce a cada desenvolvedor. Preferência: Repository-based."*
2. **Tiers de Preços Definidos (Seção 20):**
   - **Community OSS**: Gratuito, local, sem limites em projetos de código aberto.
   - **Cloud Developer**: $0/mês (1 repositório privado, histórico de 7 dias).
   - **Cloud Team**: $39/mês (10 repositórios privados, histórico de 90 dias, detecção de regressões de concorrência e bot de PR).
   - **Cloud Pro**: $99/mês (30 repositórios privados, histórico de 1 ano, webhooks, suporte prioritário).
   - **Enterprise**: Personalizado (on-premise control plane, SSO, retenção customizada).
3. **Monetização Paralela de Alto Valor (Seções 22 e 23):**
   - **ChaosSQL Concurrency Audit ($1.490 taxa única):**
     - O produto de ativação mais rápida para gerar fluxo de caixa e validação empresarial com times reais de banco de dados.
     - Entrega: Concurrency Safety Report, teste de Hermitage nos fluxos críticos, reproduções em Go sintetizadas e recomendações cirúrgicas de mitigação de concorrência.

---

## 2. Arquitetura Técnica

```text
                                  PORTAL WEB
                                      │
                 ┌────────────────────┴────────────────────┐
                 │                                         │
                 ▼                                         ▼
         #/pricing (Tiers)                      #/pricing (Audit Pack)
   Developer ($0) · Team ($39/mo)              1-Week Deep Concurrency Audit
   Pro ($99/mo) · Enterprise                   $1,490 One-Time Booking
                 │                                         │
                 ▼                                         ▼
         POST /api/waitlist                     POST /api/waitlist
     (Tier checkout lead capture)          (VIP Concurrency Audit intent)
                 │                                         │
                 └────────────────────┬────────────────────┘
                                      │
                                      ▼
                           CONTROL PLANE (Backend)
                          internal/server/billing.go
                           - PlanConfig & Limits
                           - Repo limit enforcement
                           - GET /v1/organizations/{id}/subscription
```

---

## 3. Modelo de Dados & Limites no Servidor

### `internal/server/billing.go`:
```go
type PlanConfig struct {
    ID                    string `json:"id"`
    Name                  string `json:"name"`
    PriceMonthlyUSD       int    `json:"price_monthly_usd"`
    PriceAnnualUSD        int    `json:"price_annual_usd"`
    MaxRepositories       int    `json:"max_repositories"`
    RetentionDays         int    `json:"retention_days"`
    HasRegressionGating   bool   `json:"has_regression_gating"`
    HasPRComments         bool   `json:"has_pr_comments"`
    HasPrioritySupport    bool   `json:"has_priority_support"`
}
```

Endpoint REST:
- `GET /v1/organizations/{id}/subscription`: retorna o plano da organização, contagem de repositórios cadastrados e limites aplicáveis.
