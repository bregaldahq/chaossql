# ChaosSQL SaaS — Passo 5: Concurrency Health Dashboard MVP Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar a página do Dashboard MVP no portal web (`site/src/pages/DashboardPage.tsx` e CSS module), com cálculo de Concurrency Health Score, métricas agregadas dos últimos 30 dias, tabela de execuções com filtros dinâmicos e o drawer/modal de detalhe de finding (Seção 16).

---

### Task 1: Componente do Dashboard (`site/src/pages/DashboardPage.tsx` e `DashboardPage.module.css`)
- Implementar cards de métricas (Health 98.7%, Runs, Schedules, Regressões abertas/resolvidas).
- Implementar tabela de execuções recentes com status badges e tags de anomalias.
- Implementar filtros: All, Regressions, PRs, Passed, e barra de busca.
- Implementar Finding Detail modal interativo com rastro causal, invariante falha e download do reprodutor Go.

### Task 2: Atualização de Roteamento e Navegação
- Atualizar `site/src/components/ui/SiteNav.tsx` adicionando item "Dashboard" com badge de status.
- Atualizar `site/src/App.tsx` para suporte à rota `#/dashboard`.

### Task 3: Validação de Build do Frontend
- Executar `npm run build` em `site/` garantindo zero erros de tipagem TypeScript ou CSS.

### Task 4: Merge e Deploy
- Merge `feat/saas-step5-cloud-dashboard-mvp` em `main` e push para origin.
