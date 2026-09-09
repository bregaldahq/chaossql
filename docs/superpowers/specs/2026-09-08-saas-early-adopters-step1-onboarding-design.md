# ChaosSQL SaaS — Fase 3 / Passo 1: Onboarding de 60s & Concurrency CI Workflow (Design Spec)

- **Status:** Approved (Commercial Impact Lens - Section 28 & 60 of 90-Days Plan)
- **Date:** 2026-09-08
- **Author:** Bregalda Engineering & Antigravity
- **Scope:** Fase 3 / Early Adopters (Semanas 5-6 do Plano de 90 Dias - Redução de Fricção, Onboarding de 60s e Validação Contínua de PRs no GitHub)

---

## 1. Visão Geral & Objetivo Comercial

De acordo com a **Seção 28 (Definition of Done do MVP)** e a **Seção 60 (Próximo passo imediato)**:
> *"O MVP está pronto quando:
> 1. usuário entra com GitHub;
> 2. conecta repositório;
> 3. adiciona ChaosSQL no CI;
> 4. abre PR;
> 5. ChaosSQL executa;
> 6. resultado aparece no Cloud;
> 7. PR recebe comentário;
> 8. uma regressão pode ser comparada contra main."*

Para atingir a meta da Semana 5 (**10 instalações reais de early adopters**), a barreira de entrada para conectar um repositório e rodar no GitHub Actions deve ser **menor que 60 segundos**.

---

## 2. Entregas Técnicas

1. **Workflow de Validação de Concorrência (`.github/workflows/concurrency-ci.yml`):**
   - Workflow oficial executando em `push` para `main` e em `pull_request`.
   - Permissões explícitas: `contents: read`, `pull-requests: write`.
   - Executa a action com `github-token: ${{ secrets.GITHUB_TOKEN }}` para habilitar comentários e `GITHUB_STEP_SUMMARY`.

2. **Fluxo de Onboarding de 60s no Dashboard Web (`site/src/pages/DashboardPage.tsx`):**
   - Botão **"+ Conectar Repositório"** com modal interativo de 4 passos:
     - **Passo 1:** Gerar/Copiar API Token do ChaosSQL Cloud.
     - **Passo 2:** Adicionar Secret no GitHub (`CHAOSSQL_CLOUD_TOKEN`).
     - **Passo 3:** Copiar snippet pronto de 6 linhas para `.github/workflows/concurrency.yml`.
     - **Passo 4:** Abrir Pull Request de teste e verificar o selo/alerta de concorrência.

3. **Validação & Documentação no README:**
   - Adicionar o bloco de 60 segundos com instruções de instalação para GitHub Actions.
