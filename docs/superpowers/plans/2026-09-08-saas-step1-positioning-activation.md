# ChaosSQL SaaS — Passo 1: Posicionamento & Ativação Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transformar a apresentação do ChaosSQL em uma máquina de aquisição comercial SaaS, atualizando o README com demo imediata, adicionando showcase interativo de cenários, formulário de Early Access/Waitlist e endpoint serverless Cloudflare Pages.

**Architecture:** Abordagem "Conversão Direta & Dual-Funnel": O README.md foca na dor comercial imediata com bloco ASCII de Lost Update; o portal Vite/React em `site/` ganha os componentes `DemoShowcase` e `CloudWaitlistSection`; um endpoint Cloudflare Pages Function em `site/functions/api/waitlist.ts` processa inscrições com validação e despacho de webhook.

**Tech Stack:** Go (CLI & Docs), React 19, TypeScript 5.7, Vite 6.2, CSS Modules, Cloudflare Pages Functions, Lucide Icons.

**Spec:** `docs/superpowers/specs/2026-09-08-saas-step1-positioning-activation-design.md`

## Global Constraints

- Preserve todas as capacidades técnicas e rigor científico do motor ChaosSQL existente (Adya DSG, PCT, ddmin, pkg/chaostest).
- Suporte a bilíngue (PT e EN) na Landing Page e componentes novos, respeitando a prop `lang`.
- Seguir o Bregalda Design System (tema dark, paleta de roxos/dourados, tipografia de alta legibilidade, bordas suaves e transições rápidas).
- Sem dependências de bibliotecas de terceiros pesadas para formulários; utilizar formulário React nativo e Cloudflare Pages Functions.
- Validação estrita do TypeScript (`npm --prefix site run typecheck` deve passar com zero erros).

---

### Task 1: Reestruturação Comercial do README.md

**Files:**
- Modify: `/root/chaossql/README.md`

**Interfaces:**
- Consumes: Exemplos existentes em `examples/banking_lost_update/` e portal `https://chaossql.bregalda.com`.
- Produces: Novo topo de funil com posicionamento comercial, demo visual de Lost Update, quickstart de 30s e seção de Cloud & Audit.

- [ ] **Step 1: Inspecionar o topo atual do README.md e preparar a reestruturação**
Verificar as seções existentes para assegurar que nenhuma documentação técnica sobre Adya ou ddmin seja perdida ao mover para o rodapé.

- [ ] **Step 2: Reescrever a introdução e o bloco de demo imediata no README.md**
Substituir o cabeçalho técnico pelas novas taglines e pelo card de impacto visual:
```markdown
<div align="center">

<img src="site/assets/icone_bregalda.svg" width="96" height="96" alt="ChaosSQL Logo" />

# ChaosSQL

### Catch database concurrency bugs before production does.

[![Documentation Portal](https://img.shields.io/badge/Docs-chaossql.bregalda.com-4B2E83?style=for-the-badge&logo=cloudflare&logoColor=white)](https://chaossql.bregalda.com)
[![Release Version](https://img.shields.io/badge/Release-v1.4.0-F5C400?style=for-the-badge&logo=github&labelColor=2A2140)](https://github.com/bregaldahq/chaossql/releases/tag/v1.4.0)
[![CI Pipeline](https://img.shields.io/badge/CI-Passing-22C55E?style=for-the-badge&logo=githubactions&logoColor=white)](https://github.com/bregaldahq/chaossql/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)

<p align="center">
  <b>Deterministic Concurrency & Invariant Testing for PostgreSQL, MySQL, and SQLite.</b><br />
  Stop silent data corruption, write skew, and lost updates in CI/CD before they reach production.
</p>

<p align="center">
  <a href="https://chaossql.bregalda.com">🌐 Official Portal & Documentation</a> •
  <a href="https://chaossql.bregalda.com/#/playground">🧪 Interactive WASM Playground</a> •
  <a href="#-quickstart">🚀 Quickstart</a> •
  <a href="#-flagship-scenarios">🔬 3 Key Demos</a> •
  <a href="#️-chaossql-cloud--concurrency-audit">☁️ Cloud Early Access</a>
</p>

</div>

---

## ⚡ The Concurrency Blindspot

Traditional unit and integration tests run transactions in isolation. In production, simultaneous transactions trigger silent data corruption that logs cannot explain:

```text
❌ Lost Update Detected (P4 Anomaly)
────────────────────────────────────────────────────────
Expected Account Balance:   $2,000.00
Actual Account Balance:     $1,950.00 (Silent race condition)
Root Cause:                 T1 and T2 concurrently read balance=2000 under READ COMMITTED
Deterministic Seed:         184729
Minimal Reproduction:       4 operations synthesized in repro_test.go
────────────────────────────────────────────────────────
```

**ChaosSQL finds these bugs deterministically, isolates the minimal failing trace, and prevents them from returning to your main branch.**
```

- [ ] **Step 3: Adicionar seções de Quickstart, 3 Demos Principais, GitHub Action e Cloud Early Access**
Inserir:
1. Quickstart de 1 linha.
2. Tabela comparativa dos cenários (Banking, Hospital Write Skew, Deadlock Cycle) com links para o playground.
3. Configuração do GitHub Action em 6 linhas.
4. Seção comercial: "ChaosSQL Cloud — Concurrency Regression Detection in CI" e "ChaosSQL Concurrency Audit".
5. Preservar no rodapé a arquitetura formal (Adya DSG, PCT, ddmin causal, Go SDK, SARIF).

- [ ] **Step 4: Validar formatação markdown e links do README.md**
Comando: `git diff README.md | head -n 40`
Garantir que as tags e links estão corretos.

- [ ] **Step 5: Commitar a reestruturação do README.md**
```bash
git -C /root/chaossql add README.md
git -C /root/chaossql commit -m "docs: restructure README for commercial positioning and instant demo clarity"
```

---

### Task 2: Implementar Componente DemoShowcase na Landing Page

**Files:**
- Create: `/root/chaossql/site/src/components/ui/DemoShowcase.tsx`
- Create: `/root/chaossql/site/src/components/ui/DemoShowcase.module.css`

**Interfaces:**
- Consumes: Props `{ lang?: 'pt' | 'en' }`.
- Produces: Export `DemoShowcase` que renderiza seletor de 3 cenários com visualização de trace, copy explicativo, comando CLI e link para o Playground WASM.

- [ ] **Step 1: Criar estilos CSS Modules para DemoShowcase**
Criar `site/src/components/ui/DemoShowcase.module.css` com estilos para tabs selecionáveis, badge de anomalia (P4, A5B, Deadlock), bloco de comparação (Esperado vs Obtido), container de código e botões de ação com suporte ao tema Bregalda.

- [ ] **Step 2: Implementar o componente DemoShowcase.tsx**
Implementar o seletor com os 3 cenários:
1. `banking`: Lost Update (saldo $2000 -> $1950, P4).
2. `reservation`: Write Skew / Oversell (vagas hospitalares/ingressos, A5B).
3. `deadlock`: Deadlock Cycle (travamento circular de recursos).
Incluir botão para copiar comando CLI (`chaossql run examples/...`) e botão para navegar para `#/playground`.

- [ ] **Step 3: Validar tipagem TypeScript do componente**
Comando: `npm --prefix /root/chaossql/site run typecheck`
Esperado: Exit code 0 sem erros.

- [ ] **Step 4: Commitar o componente DemoShowcase**
```bash
git -C /root/chaossql add site/src/components/ui/DemoShowcase.tsx site/src/components/ui/DemoShowcase.module.css
git -C /root/chaossql commit -m "feat(site): add interactive DemoShowcase component with 3 flagship concurrency scenarios"
```

---

### Task 3: Implementar Componente CloudWaitlistSection

**Files:**
- Create: `/root/chaossql/site/src/components/ui/CloudWaitlistSection.tsx`
- Create: `/root/chaossql/site/src/components/ui/CloudWaitlistSection.module.css`

**Interfaces:**
- Consumes: Props `{ lang?: 'pt' | 'en' }`.
- Produces: Export `CloudWaitlistSection` com formulário de captação dual (Early Access Cloud + Concurrency Audit), validação no cliente e envio via `POST /api/waitlist` com fallback em `localStorage`.

- [ ] **Step 1: Criar estilos CSS Modules para CloudWaitlistSection**
Criar `site/src/components/ui/CloudWaitlistSection.module.css` contendo layout elegante, campos com foco suave, checkbox estilizado para a Concurrency Audit, estados de loading com spinner, alerta de sucesso com confirmação e alerta de erro.

- [ ] **Step 2: Implementar a lógica e renderização do formulário em CloudWaitlistSection.tsx**
Gerenciar estado:
- `formData`: `{ name: string, email: string, company: string, database: string, wantAudit: boolean, notes: string }`
- `status`: `'idle' | 'submitting' | 'success' | 'error'`
- `errorMessage`: `string`
Submissão:
- Envia payload via `fetch('/api/waitlist', { method: 'POST', body: JSON.stringify(formData) })`.
- Em caso de falha de rede/desenvolvimento local, salva em `localStorage.setItem('chaossql_waitlist_offline', ...)` e exibe tela de sucesso confirmando o interesse do lead.

- [ ] **Step 3: Validar tipagem TypeScript do componente**
Comando: `npm --prefix /root/chaossql/site run typecheck`
Esperado: Exit code 0 sem erros.

- [ ] **Step 4: Commitar o componente CloudWaitlistSection**
```bash
git -C /root/chaossql add site/src/components/ui/CloudWaitlistSection.tsx site/src/components/ui/CloudWaitlistSection.module.css
git -C /root/chaossql commit -m "feat(site): add CloudWaitlistSection component with dual-intent lead capture"
```

---

### Task 4: Implementar API Serverless Cloudflare Pages Function

**Files:**
- Create: `/root/chaossql/site/functions/api/waitlist.ts`

**Interfaces:**
- Consumes: Requisições HTTP `POST /api/waitlist` com JSON payload.
- Produces: Resposta JSON `{ success: boolean, message: string }`, com disparo opcional de webhook configurado em `WAITLIST_WEBHOOK_URL`.

- [ ] **Step 1: Implementar o handler em site/functions/api/waitlist.ts**
Criar a função PagesFunction:
- Validação do método: aceitar apenas `POST` (retornar 405 para outros).
- Parse do JSON do body.
- Validação: `email` obrigatório e válido (regex de e-mail), `name` obrigatório.
- Se a variável de ambiente `context.env.WAITLIST_WEBHOOK_URL` estiver presente, despacha um POST para o webhook com payload formatado contendo os dados do lead.
- Retorno HTTP 200 com cabeçalho `Content-Type: application/json`.

- [ ] **Step 2: Validar compilação do TypeScript**
Comando: `npm --prefix /root/chaossql/site run typecheck`
Esperado: Exit code 0 sem erros.

- [ ] **Step 3: Commitar a função Cloudflare Pages**
```bash
git -C /root/chaossql add site/functions/api/waitlist.ts
git -C /root/chaossql commit -m "feat(site): add Cloudflare Pages serverless function for waitlist lead ingestion"
```

---

### Task 5: Integrar Componentes na LandingPage.tsx e Atualizar Copy

**Files:**
- Modify: `/root/chaossql/site/src/pages/LandingPage.tsx`
- Modify: `/root/chaossql/site/src/pages/LandingPage.module.css`

**Interfaces:**
- Consumes: Componentes `DemoShowcase` e `CloudWaitlistSection`.
- Produces: Landing page completa e renovada, com novo Hero, CTA para Early Access, visualizador do Lost Update, seletor de demos, fluxo de workflow e formulário de conversão.

- [ ] **Step 1: Atualizar os estilos de LandingPage.module.css**
Adicionar classes para a visualização do card de Lost Update no Hero, alinhamentos da seção de workflow e espaçamento harmonioso com o design system da Bregalda.

- [ ] **Step 2: Atualizar LandingPage.tsx para integrar os novos componentes**
- Substituir a headline antiga pelas novas taglines comerciais em PT e EN.
- Adicionar no Hero o card do Lost Update e os botões `[ Testar no Playground WASM ]` e `[ Entrar no Cloud Early Access ]`.
- Incluir a seção `DemoShowcase` após o Hero.
- Incluir o diagrama visual do Workflow CI (Invariante -> Exploração de Escalas -> ddmin -> Bloqueio no CI).
- Substituir o `ContactSection` pelo novo `CloudWaitlistSection`.

- [ ] **Step 3: Validar tipagem TypeScript completa**
Comando: `npm --prefix /root/chaossql/site run typecheck`
Esperado: Exit code 0 sem erros.

- [ ] **Step 4: Executar build completo do portal**
Comando: `npm --prefix /root/chaossql/site run build`
Esperado: Compilação de produção gerada com sucesso em `site/dist` e na raiz de `site/`.

- [ ] **Step 5: Commitar a integração na LandingPage**
```bash
git -C /root/chaossql add site/src/pages/LandingPage.tsx site/src/pages/LandingPage.module.css site/dist site/index.html
git -C /root/chaossql commit -m "feat(site): integrate commercial hero, DemoShowcase, and CloudWaitlistSection into landing page"
```

---

### Task 6: Verificação End-to-End e Fechamento da Fase 1

**Files:**
- Test/Verify: `README.md`, `site/dist/index.html`, `site/src/`

- [ ] **Step 1: Executar verificação de build e sintaxe**
Comandos:
```bash
npm --prefix /root/chaossql/site run typecheck
npm --prefix /root/chaossql/site run build
```
Esperado: 0 erros, bundles estáticos prontos para deploy na Cloudflare Pages.

- [ ] **Step 2: Validar integridade dos exemplos e testes do core Go**
Comando: `go test -v ./pkg/...`
Esperado: PASS em todos os pacotes do ChaosSQL para certificar que nada do engine foi afetado.

- [ ] **Step 3: Commitar qualquer ajuste final de build**
```bash
git -C /root/chaossql status
git -C /root/chaossql commit -am "chore: finalize step 1 positioning and activation release assets" || true
```
