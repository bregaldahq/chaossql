# ChaosSQL SaaS — Passo 1: Posicionamento & Ativação (Design Spec)

- **Status:** Approved (Commercial Impact Lens)
- **Date:** 2026-09-08
- **Author:** Bregalda Engineering & Antigravity
- **Scope:** Semana 1 do Plano de 90 Dias (GTM, README, Landing Page, Waitlist & Demos)

---

## 1. Visão Geral & Problema Comercial

### 1.1 Contexto
O **ChaosSQL** possui um core técnico e científico sofisticado (classificação formal de anomalias de Adya, agendamento determinístico PCT, algoritmo causal ddmin, gerador de reproduções repro_test.go e playground WASM).

Entretanto, o repositório e a página inicial foram apresentados originalmente com foco acadêmico/fuzzing, criando atrito na percepção de valor imediato por equipes de engenharia, tech leads e decisores técnicos.

### 1.2 Objetivo Deste Passo
Transformar o ChaosSQL em um **produto comercial atraente**, reduzindo o tempo de compreensão do valor para menos de 30 segundos, permitindo experimentação instantânea e abrindo o canal de captação de leads qualificados para:
1. **ChaosSQL Cloud Early Access** (detecção de regressões de concorrência no CI/CD).
2. **ChaosSQL Concurrency Audit** (serviço consultivo de alto valor para transações críticas de empresas, US$ 500 – US$ 2.000).

---

## 2. Decisões Estratégicas de Posicionamento

### 2.1 Taglines e Mensagens
- **Tagline Primária (EN):** *"Catch database concurrency bugs before production does."*
- **Tagline Secundária (EN):** *"Deterministic concurrency testing for PostgreSQL, MySQL, and SQLite. Built for CI/CD."*
- **Tagline em Português (PT):** *"Encontre race conditions e anomalias de banco antes que cheguem à produção."*

### 2.2 Ideal Customer Profile (ICP)
- **Perfil Técnico:** Backend Engineers (Go, Node.js, Python, Java), Tech Leads, Staff Engineers, Database / Platform Architects.
- **Setores Prioritários:** Fintechs, pagamentos, carteiras digitais, e-commerce (inventário/estoque), sistemas de reserva com vagas limitadas e plataformas com filas e concorrência pesada.
- **A Dor Real:** Race conditions intermitentes que passam pelos testes unitários convencionais e corrompem o saldo ou estado em produção, sem logs rastreáveis.

### 2.3 Proposta de Valor Visual
Em vez de focar no jargão matemático, o produto destaca a cadeia de valor tangível:
```text
ENCONTRAR      -> Identifica a anomalia (ex: P4 Lost Update)
REPRODUZIR     -> Gera seed determinístico e teste mínimo (repro_test.go)
EXPLICAR       -> Mostra o intercalamento exato das transações (T1 vs T2)
IMPEDIR NO CI  -> Bloqueia Pull Requests que reintroduzam a regressão
```

---

## 3. Reestruturação do README.md

O `README.md` será reorganizado de acordo com a ordem recomendada de conversão:

1. **Header Comercial:** Logo Bregalda, badges essenciais e links rápidos de navegação.
2. **Hero com Demonstração Visual:**
   ```text
   ❌ Lost Update Detected (P4 Anomaly)
   Expected Balance: $2,000.00
   Actual Balance:   $1,950.00 (Race condition between concurrent updates)
   Deterministic Seed: 184729
   Minimal Reproduction: 4 operations synthesized in repro_test.go
   ```
3. **Quickstart de 30 Segundos:**
   - Instalação via `go install github.com/bregaldahq/chaossql/cmd/chaossql@latest`
   - Comando de 1 linha: `chaossql run examples/banking_lost_update/chaos.yaml`
4. **3 Cenários Emblemáticos & Playground:**
   - Tabela comparativa com links para o Playground WASM interativo:
     - 🏦 *Banking Lost Update* (Inconsistência de saldo).
     - 🏥 *Hospital / Reservation Write Skew* (Violação de capacidade mínima).
     - 🔒 *Deadlock Cycle* (Travamento cruzado de transações).
5. **Integração com CI/CD (GitHub Action):**
   - Snippet conciso demonstrando como rodar o ChaosSQL em Pull Requests.
6. **ChaosSQL Cloud & Concurrency Audit (Nova Seção Comercial):**
   - Chamada para a lista de espera do Cloud (Regression Guard).
   - Oferta de Concurrency Audit para equipes com prazos de entrega críticos.
7. **Arquitetura & Teoria Profunda:**
   - Mantém as seções detalhadas de PCT, Adya DSG, ddmin causal, Go SDK e SARIF no rodapé para consolidar a credibilidade técnica.

---

## 4. Redesign da Landing Page (`site/src/pages/LandingPage.tsx`)

### 4.1 Componentes e Layout
A Landing Page será atualizada para incorporar:
1. **Hero Section Renovada:**
   - Taglines em PT e EN (com suporte ao seletor de idioma existente).
   - Dual Call-to-Action:
     - Primário: `[ Testar no Playground WASM ]` (navegação direta).
     - Secundário / Ativação: `[ Entrar no Cloud Early Access ]` (âncora suave para o formulário).
   - Card lateral com preview do *Lost Update Detected* ($2,000 -> $1,950).
2. **Componente `DemoShowcase`:**
   - Tabs interativas para alternar entre Banking, Reservation e Deadlock.
   - Demonstração do trace mínimo (Esperado vs. Obtido).
   - Botão para carregar o cenário no playground WASM.
   - Snippet de terminal com comando CLI copiado em 1 clique.
3. **Seção de Workflow CI/CD:**
   - 4 passos visuais destacando a transição do invariante local para o CI.
4. **Componente `CloudWaitlistSection`:**
   - Formulário nativo com o design system da Bregalda.
   - Campos:
     - `name`: Nome do contato.
     - `email`: E-mail de trabalho ou pessoal (obrigatório).
     - `company`: Empresa ou repositório GitHub.
     - `database`: PostgreSQL / MySQL / SQLite / Outro.
     - `wantAudit`: Checkbox para manifestar interesse na *ChaosSQL Concurrency Audit*.
     - `notes`: Observações opcionais sobre o desafio de concorrência.
   - Feedback de estado: Idle, Submitting, Success e Error.
   - Fallback offline/local via `localStorage` para garantir perda zero de leads.

---

## 5. API Serverless (`site/functions/api/waitlist.ts`)

### 5.1 Endpoint
- **URL:** `POST /api/waitlist`
- **Ambiente:** Cloudflare Pages Functions
- **Headers:** `Content-Type: application/json`

### 5.2 Validação e Fluxo
- Valida campos obrigatórios (`email` válido, `name`).
- Sanitiza dados de entrada para evitar injeções.
- Se a variável de ambiente `WAITLIST_WEBHOOK_URL` estiver configurada no Cloudflare Pages, despacha um webhook HTTP POST (formato compatível com Slack/Discord/Make/Zapier) notificando a equipe imediatamente sobre a nova inscrição.
- Retorna `200 OK` com `{ "success": true, "message": "Inscrição confirmada" }`.

---

## 6. Plano de Verificação e Testes

1. **Testes Estáticos de Front-end:**
   - Executar `npm --prefix site run typecheck` para garantir integridade estrita do TypeScript.
   - Executar `npm --prefix site run build` para validar geração dos bundles de produção (`dist/`).
2. **Validação da Experiência Visual e Responsividade:**
   - Testar o fluxo de seleção das tabs do `DemoShowcase`.
   - Testar preenchimento, validação e envio do formulário no `CloudWaitlistSection`.
   - Garantir visualização perfeita em resoluções mobile e desktop.
3. **Verificação do README.md:**
   - Checagem de links, formatação de blocos de código e coerência das mensagens.
