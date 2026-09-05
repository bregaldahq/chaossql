# ChaosSQL Portal Web — Plano de Implementação de Internacionalização Bilíngue (PT-BR / EN)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar um sistema de internacionalização (i18n) bilíngue completo (PT-BR e EN) com alternador *Segmented Pill Switch* `[ PT | EN ]`, cobrindo 100% da interface, dos 8 capítulos técnicos da documentação e dos 10 cenários com mitigações em produção sem recarregar a página.

**Architecture:** Vanilla JavaScript puro no cliente com estrutura de dados `DOCS_DATA = { pt, en }` em `docs-data.js`, dicionário `I18N = { pt, en }` em `app.js`, atributos declarativos `data-i18n` no `index.html`, e motor de sincronização de estado `setLanguage(lang)` que preserva rotas hash, contexto ativo e preferência no `localStorage`.

**Tech Stack:** HTML5 semântico, CSS3 Moderno (Custom Properties, Flexbox, Grid), Vanilla JavaScript ES6+, Node.js (validação de sintaxe).

**Spec:** [`docs/superpowers/specs/2026-09-05-bilingual-i18n-portal-design.md`](file:///wsl.localhost/Ubuntu-22.04/root/chaossql/docs/superpowers/specs/2026-09-05-bilingual-i18n-portal-design.md)

## Global Constraints
- Nenhuma dependência externa (zero CDNs, zero NPM packages em runtime).
- 100% de compatibilidade com hospedagem estática no Cloudflare Pages.
- Preservar palavras-chave canônicas de banco de dados (SQL: `SELECT`, `UPDATE`, `BEGIN`, `COMMIT`, etc.) e código Go.
- Não alterar as rotas hash (`#/`, `#/docs`, `#/scenarios`, `#/visualizer`, `#/matrix`).
- Transição instantânea em tempo real (zero reload de página).

---

### Task 1: Switcher Visual [ PT | EN ] e Marcação Declarativa i18n
**Files:**
- Modify: `site/index.html`
- Modify: `site/assets/style.css`

**Interfaces:**
- Produz: Elementos `.lang-switch` com botões `.lang-btn[data-lang="pt"]` e `.lang-btn[data-lang="en"]` na navbar e na gaveta móvel (`#mobileNavDrawer`).
- Produz: Atributos `data-i18n="caminho.chave"` nos títulos, links, badges e seções estáticas do HTML.
- Produz: Classes CSS `.lang-switch`, `.lang-btn`, `.lang-btn.active` seguindo o design system do Studio Bregalda.

- [ ] **Step 1: Adicionar o componente `.lang-switch` na navbar desktop e no `#mobileNavDrawer` de `site/index.html`**
  Inserir entre `nav-links` e o botão `nav-cta` (GitHub):
  ```html
  <div class="lang-switch" role="group" aria-label="Language Selector">
    <button class="lang-btn active" data-lang="pt" type="button" aria-pressed="true">PT</button>
    <button class="lang-btn" data-lang="en" type="button" aria-pressed="false">EN</button>
  </div>
  ```

- [ ] **Step 2: Adicionar atributos `data-i18n` e `data-i18n-attr` nos nós de texto de `site/index.html`**
  - Links de navegação: `nav.home`, `nav.docs`, `nav.scenarios`, `nav.visualizer`, `nav.matrix`.
  - Hero section: `landing.heroMeta`, `landing.heroTitle`, `landing.heroSubtitle`, `landing.ctaDocs`, `landing.ctaScenarios`.
  - Terminal interativo: `landing.termRun`, `landing.termJitter`, `landing.termShrink`, `landing.termReset`.
  - 6 Cards de Engenharia: títulos e descrições.
  - Seções Docs, Scenarios, Visualizer e Matrix: cabeçalhos e rótulos fixos.

- [ ] **Step 3: Adicionar estilização para `.lang-switch` e `.lang-btn` em `site/assets/style.css`**
  ```css
  .lang-switch {
    display: inline-flex;
    align-items: center;
    background: var(--color-surface);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 999px;
    padding: 2px;
    gap: 2px;
  }
  .lang-btn {
    background: transparent;
    border: none;
    color: var(--color-cream);
    opacity: 0.65;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    font-weight: 600;
    padding: 3px 10px;
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }
  .lang-btn:hover {
    opacity: 1;
    color: var(--color-yellow);
  }
  .lang-btn.active {
    background: var(--color-yellow);
    color: var(--color-ink);
    opacity: 1;
    box-shadow: 0 0 10px rgba(245, 196, 0, 0.35);
  }
  ```

- [ ] **Step 4: Validar sintaxe e visualização preliminar via curl**
  Executar: `wsl -d Ubuntu-22.04 -u root bash -c "grep -n 'lang-switch' /root/chaossql/site/index.html"`

- [ ] **Step 5: Commit das alterações da Task 1**
  ```bash
  git add site/index.html site/assets/style.css
  git commit -m "feat(ui): add bilingual segmented pill switch and i18n markers"
  ```

---

### Task 2: Expansão Bilíngue da Documentação Técnica (`site/docs-data.js`)
**Files:**
- Modify: `site/docs-data.js`

**Interfaces:**
- Consome: Os 8 capítulos em português já existentes.
- Produz: `window.DOCS_DATA = { pt: { ...8 capítulos... }, en: { ...8 capítulos... } }`.

- [ ] **Step 1: Preparar a estrutura dual `pt` e `en` em `site/docs-data.js`**
  Mover os 8 capítulos atuais para dentro de `window.DOCS_DATA.pt`.

- [ ] **Step 2: Escrever a versão integral em Inglês (`en`) dos 8 capítulos**
  1. `getting-started`: Quickstart Guide, Architecture Pipeline (Fuzzer -> Detector -> Shrinker -> Reporter), Installation (Binary, Pure Go Zero CGO, Docker).
  2. `dsl-spec`: Complete `chaos.yaml` DSL reference (variables `-> var`, dynamic expressions `{var - 100}`, invariants, fault injection).
  3. `cli-reference`: All 9 subcommands (`run`, `demo`, `ui`, `diff`, `replay`, `bench`, `validate`, `init`, `matrix`) and all 12 flags with CLI examples.
  4. `trace-visualizer`: Embedded web server (`chaossql ui`), microsecond Gantt swimlanes, SVG Adya conflict graph, statement inspector.
  5. `cicd-sarif`: CI/CD, `action.yml`, GitHub Code Scanning SARIF 2.1.0 (9 formal rules), JUnit XML, OTLP OpenTelemetry.
  6. `drivers`: SQLite (pure Go `modernc.org/sqlite`, CGO=0), PostgreSQL (`pgx/v5`, RC vs SSI), MySQL/MariaDB (InnoDB RR, gap locks).
  7. `go-sdk`: `pkg/chaostest` builder API, assertion functions, and automatic `repro_test.go` synthesis via $ddmin$.
  8. `academic-theory`: Formal foundation (Bernstein conditions, CSR theorem, Burckhardt PCT probability bound, Zeller $ddmin$, Elle linearizability).

- [ ] **Step 3: Validar a sintaxe JavaScript**
  Executar: `wsl -d Ubuntu-22.04 -u root bash -c "node -c /root/chaossql/site/docs-data.js"`
  Esperado: Código 0, sem erros de sintaxe.

- [ ] **Step 4: Commit das alterações da Task 2**
  ```bash
  git add site/docs-data.js
  git commit -m "docs: add comprehensive english documentation suite (v1.2.0)"
  ```

---

### Task 3: Dicionários de UI, Cenários Bilíngues e Motor de i18n (`site/app.js`)
**Files:**
- Modify: `site/app.js`

**Interfaces:**
- Consome: `window.DOCS_DATA[lang]`.
- Produz: `const I18N = { pt: { ... }, en: { ... } }`.
- Produz: `SCENARIOS` com `name: { pt, en }`, `description: { pt, en }`, `analysis: { pt, en }`, `fix: { pt, en }`.
- Produz: Função global `setLanguage(lang)` e inicializador de idioma com fallback `navigator.language`.

- [ ] **Step 1: Criar o dicionário completo `I18N` em `site/app.js`**
  Incluir textos em `pt` e `en` para:
  - `nav`: links do menu e gaveta mobile.
  - `landing`: Hero, badges, subtítulo, botões CTA, comando de cópia, labels do terminal interativo, logs do replay fuzzer, cards de engenharia e banners.
  - `docs`: Placeholder da busca, títulos das categorias, breadcrumb inicial, botões de anterior e próximo.
  - `scenarios`: Tabs (`SQL Code`, `Conflict DAG`, `Invariant Rule`, `Fix & Mitigation`), rótulos de severidade e anomalia.
  - `visualizer`: Títulos, botões de controle (`▶ Animate`, `Raw Trace`, `1-Minimal Shrunk`), filtros por worker, colunas da timeline e painel inspetor.
  - `matrix`: Cabeçalho, colunas e descrições dos fenômenos.

- [ ] **Step 2: Converter os 10 cenários de `SCENARIOS` para formato bilíngue**
  Atualizar os 10 cenários:
  1. Banking: Lost Update ($P4$)
  2. Inventory: Oversell Under Read Committed ($A3$)
  3. Hospital Shift: Write Skew Under Snapshot ($A5B$)
  4. Financial Balances: Read Skew ($A5A$)
  5. Auction Bidding: Dirty Write ($G0$)
  6. Crypto Exchange: Circular Information Flow ($G1c$)
  7. Flash Crash: Dirty Read ($G1a$)
  8. Ticket Booking: Anti-Dependency Cycle ($G2$)
  9. Transactional Deadlock: Lock Inversion ($G-DL$)
  10. Foreign Key Cascade Deadlock ($G-DL$)
  Com `name`, `description`, `analysis` e `fix` (title, explanation, code com comentários no idioma correto, driverNotes) tanto em `pt` quanto em `en`.

- [ ] **Step 3: Implementar a função `setLanguage(lang)` e os listeners de clique**
  - Atualizar `currentLang = lang`.
  - Atualizar `document.documentElement.lang = lang`.
  - Sincronizar classes `.active` e `aria-pressed` em todos os `.lang-btn[data-lang]`.
  - Iterar nós `[data-i18n]` e traduzir `textContent`.
  - Iterar nós `[data-i18n-attr]` e traduzir atributos como `placeholder`.
  - Re-renderizar dinamicamente a view ativa:
    - Se rota atual for `docs`: `renderDocsSidebar()` e `renderDocChapter(currentChapter)`.
    - Se rota atual for `scenarios`: `renderScenarioList()` e `renderScenarioDetail(activeScenarioIndex)`.
    - Se rota atual for `visualizer`: `updateVisualizerTexts()`.
    - Se rota atual for `matrix`: `renderMatrix()`.
  - Salvar no `localStorage.setItem('chaossql_lang', lang)`.

- [ ] **Step 4: Implementar a inicialização do idioma**
  - Checar `localStorage.getItem('chaossql_lang')`.
  - Se ausente, detectar via `(navigator.language || '').toLowerCase().startsWith('pt') ? 'pt' : 'en'`.
  - Chamar `setLanguage(initialLang)` durante o `initApp()`.
  - Conectar listeners `click` em todos os botões `.lang-btn`.

- [ ] **Step 5: Validar sintaxe de `site/app.js`**
  Executar: `wsl -d Ubuntu-22.04 -u root bash -c "node -c /root/chaossql/site/app.js"`
  Esperado: Código 0, sem erros.

- [ ] **Step 6: Commit das alterações da Task 3**
  ```bash
  git add site/app.js
  git commit -m "feat(i18n): implement dynamic bilingual engine and scenario translations"
  ```

---

### Task 4: Validação Integrada, Auditoria de Qualidade e Deploy Test
**Files:**
- Inspect & Test: `site/`

- [ ] **Step 1: Validar sintaxe de todos os arquivos JS**
  `node -c site/docs-data.js && node -c site/app.js`

- [ ] **Step 2: Executar testes automatizados do repositório Go**
  `go test ./...`

- [ ] **Step 3: Executar teste de requisição HTTP local**
  Subir servidor de teste temporário em porta livre e validar com `curl -I`:
  - `GET /` (HTTP 200)
  - `GET /docs-data.js` (HTTP 200)
  - `GET /app.js` (HTTP 200)
  - `GET /assets/style.css` (HTTP 200)

- [ ] **Step 4: Auditoria de coerência linguística**
  Garantir que não haja frases em inglês no modo `pt` e nenhuma frase em português no modo `en`.

- [ ] **Step 5: Commit final e relatório de walkthrough**
