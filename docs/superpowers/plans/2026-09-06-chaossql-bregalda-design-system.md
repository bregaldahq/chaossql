# ChaosSQL — Bregalda Design System & React Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reestruturar integralmente o portal e documentação técnica do ChaosSQL (`chaossql.bregalda.com`), migrando para Vite + React 19 + TypeScript e aplicando 100% dos tokens, identidade visual e componentes canônicos do Bregalda Design System.

**Architecture:** A aplicação será uma Single Page Application (SPA) modular em React 19 compilada via Vite, com estilização em CSS Modules e tokens canônicos (`tokens.css`), renderizando o canvas editorial em Warm Cream (`#FCFBF8`) com artefatos técnicos em Dark Surface (`#05030c` / `#2A2140`). O binário Go `chaossql.wasm` (8.1MB) será isolado em um Web Worker dedicado sem bloquear a interface gráfica, e o deploy continuará estático no Cloudflare Pages.

**Tech Stack:** Vite 8, React 19, TypeScript 5.9, Lucide React, Wouter (ou roteamento hash declarativo), CSS Modules, Go WebAssembly (`chaossql.wasm`), Cloudflare Pages (`wrangler.toml`).

**Spec:** [`docs/superpowers/specs/2026-09-06-chaossql-bregalda-design-system-design.md`](file://wsl.localhost/Ubuntu-22.04/root/chaossql/docs/superpowers/specs/2026-09-06-chaossql-bregalda-design-system-design.md)

## Global Constraints

- **Canvas Principal**: `--cream: #fcfbf8` como superfície padrão de leitura (68% de peso visual).
- **Contraste & Código**: `--ink: #2a2140` e `--cosmos-sky: color-mix(in srgb, var(--ink) 55%, #05030c)` (18%).
- **Identidade Bregalda**: `--purple: #4b2e83` (9%), `--yellow: #f5c400` (3%), `--green: #22c55e` (2%).
- **Geometria Rígida**: Controles e botões com raio de `3px` (proibido pill buttons em CTAs), cards de artefato com `9px`, modais com `16px`.
- **Acessibilidade WCAG AA**: Anel de foco roxo sobre creme; anel de foco amarelo sobre escuro (amarelo sobre creme é estritamente proibido).
- **Ativos Vetoriais Canônicos**: `bregalda_wordmark.svg`, `bregalda_monogram.svg`, `icone_bregalda.svg`, `bregalda_primary_lockup.svg`, `bregalda_secondary_lockup.svg` — nunca redesenhar ou sobrescrever cores internas.
- **Deploy**: Geração estática na pasta `dist/` com suporte a Cloudflare Pages via `wrangler.toml`.

---

### Task 1: Scaffolding do Projeto Vite + React 19 + TypeScript e Preservação de Legado

**Files:**
- Create: `site/package.json`
- Create: `site/vite.config.ts`
- Create: `site/tsconfig.json`
- Create: `site/index.html`
- Move: `site/*` atual (exceto novo app) para `site_legacy/`
- Copy: `bregalda_brand/*.svg` para `site/public/brand/`
- Copy: `site/assets/chaossql.wasm`, `wasm_exec.js`, `wasm-worker.js` para `site/public/wasm/`

**Interfaces:**
- Produces: Base de compilação Vite/React funcional com TypeScript strict mode e assets estáticos mapeados.

- [ ] **Step 1: Fazer backup seguro dos arquivos existentes para `site_legacy`**
  Executar comando em WSL criando `/root/chaossql/site_legacy` e copiando todo o conteúdo atual de `site/` para lá.

- [ ] **Step 2: Configurar `site/package.json`**
  Definir scripts (`dev`, `build`, `preview`, `typecheck`), dependências de produção (`react`, `react-dom`, `lucide-react`, `wouter`) e de desenvolvimento (`vite`, `@vitejs/plugin-react`, `typescript`, `@types/react`, `@types/react-dom`).

- [ ] **Step 3: Configurar `site/vite.config.ts` e `site/tsconfig.json`**
  Configurar resolução de caminhos, otimização de assets grandes e suporte a Web Workers com target ES2022.

- [ ] **Step 4: Criar `site/index.html` com metadados e fontes Inter e JetBrains Mono**
  Configurar o arquivo com preconnect para Google Fonts (`Inter:wght@400;500;600;700` e `JetBrains Mono:wght@400;500`), favicon apontando para `public/brand/icone_bregalda.svg` e ponto de entrada `<div id="root"></div>`.

- [ ] **Step 5: Copiar os ativos canônicos da marca e os arquivos Go WASM para `site/public`**
  Copiar os SVGs de `/root/chaossql/bregalda_brand` para `/root/chaossql/site/public/brand` e os binários WASM para `/root/chaossql/site/public/wasm`.

- [ ] **Step 6: Instalar dependências e validar compilação inicial**
  Executar `pnpm install` ou `npm install` em `site/` e validar que `npm run build` cria a pasta `dist/` com sucesso.

- [ ] **Step 7: Commit**
  `git add site site_legacy && git commit -m "chore: scaffold Vite + React 19 + TypeScript app and preserve legacy site"`

---

### Task 2: Tokens Canônicos e Sistema Dual-Surface de Estilos

**Files:**
- Create: `site/src/styles/tokens.css`
- Create: `site/src/styles/globals.css`

**Interfaces:**
- Consumes: Arquivos canônicos de tokens do `bregalda-design-system/tokens.css`
- Produces: Variáveis CSS globais (`--cream`, `--ink`, `--purple`, `--yellow`, `--green`, `--space-1` a `--space-6`, classes `.technical-label`, `[data-surface="light"]`, `[data-surface="dark"]`).

- [ ] **Step 1: Criar `site/src/styles/tokens.css`**
  Portar os tokens exatos do Bregalda Design System, incluindo paleta canônica, escalas espaciais de 8px, escala tipográfica com `clamp()` e regras de foco WCAG AA.

- [ ] **Step 2: Criar `site/src/styles/globals.css`**
  Implementar o reset acessível, tipografia padrão com Inter, estilização de seleção em amarelo `#F5C400`, classes utilitárias `.technical-label` e regras de `@media (prefers-reduced-motion: reduce)`.

- [ ] **Step 3: Testar e validar renderização dos tokens**
  Criar um componente simples de teste que renderiza as caixas de cor dos tokens e valida que `data-surface="light"` e `data-surface="dark"` alternam as variáveis de contraste perfeitamente.

- [ ] **Step 4: Commit**
  `git add site/src/styles && git commit -m "feat(ui): implement Bregalda canonical design tokens and dual-surface styles"`

---

### Task 3: Componente de Cabeçalho Canônico (`SiteNav`) e Identidade Visual

**Files:**
- Create: `site/src/components/ui/SiteNav.tsx`
- Create: `site/src/components/ui/SiteNav.module.css`

**Interfaces:**
- Produces: Header sticky com `bregalda_wordmark.svg`, links de seções (`Work`, `Docs`, `Scenarios`, `Visualizer`, `Matrix`, `Playground`), seletor bilíngue `PT/EN` e botão CTA Roxo com raio de `3px`.

- [ ] **Step 1: Implementar `site/src/components/ui/SiteNav.module.css`**
  Estilizar a barra fixa de 76px de altura, `backdrop-filter: blur(16px)`, fundo translúcido em creme, borda hairline inferior e botão CTA com indicador amarelo `span { color: var(--yellow) }`.

- [ ] **Step 2: Implementar `site/src/components/ui/SiteNav.tsx`**
  Renderizar o logotipo vetorial oficial (`/brand/bregalda_wordmark.svg`), links com suporte à rota ativa, controle do switch de idioma e drawer mobile acessível.

- [ ] **Step 3: Testar acessibilidade e responsividade**
  Verificar contraste das cores no Lighthouse/navegador e testar o colapso responsivo para dispositivos móveis (<768px).

- [ ] **Step 4: Commit**
  `git add site/src/components/ui/SiteNav* && git commit -m "feat(ui): implement canonical SiteNav component with Bregalda wordmark"`

---

### Task 4: Casca de Artefato Técnico (`ArtifactCard`) e Artefato de Concorrência (`ChaosSqlArtifact`)

**Files:**
- Create: `site/src/components/artifacts/ArtifactCard.tsx`
- Create: `site/src/components/artifacts/ArtifactCard.module.css`
- Create: `site/src/components/artifacts/ChaosSqlArtifact.tsx`
- Create: `site/src/components/artifacts/ChaosSqlArtifact.module.css`

**Interfaces:**
- Produces:
  - `<ArtifactCard title="..." icon={...} tag="..." disclaimer="...">`: container reutilizável para demonstrações técnicas com barra de 64px.
  - `<ChaosSqlArtifact />`: demonstração interativa de concorrência com tabela T1 vs T2, quebra de invariante e sequência de redução.

- [ ] **Step 1: Implementar `ArtifactCard.tsx` e CSS Module**
  Criar a casca visual de elevação profunda com raio de 9px, fundo escuro cósmico, toolbar de 64px, tag mono, etiqueta `"Illustrative example"` em `#F5C400` e disclaimer no rodapé.

- [ ] **Step 2: Implementar `ChaosSqlArtifact.tsx` e CSS Module**
  Transpor o componente canônico de `/root/me/bregalda-site/app/portfolio/projects/artifacts/chaossql-artifact.tsx`, adicionando interatividade de alternância de passos concorrentes.

- [ ] **Step 3: Validar visualmente o componente**
  Confirmar que a tabela T1 vs T2, o card de violação `actual_balance == expected_balance` e a sequência `Trace ➔ Shrink ➔ repro_test.go` obedecem com precisão milimétrica à estética de `/root/me`.

- [ ] **Step 4: Commit**
  `git add site/src/components/artifacts && git commit -m "feat(ui): implement ArtifactCard and ChaosSqlArtifact components"`

---

### Task 5: Grid de Demonstração Técnica (`ProjectCycle`) e Landing Page

**Files:**
- Create: `site/src/components/ui/ProjectCycle.tsx`
- Create: `site/src/components/ui/ProjectCycle.module.css`
- Create: `site/src/components/ui/ContactSection.tsx`
- Create: `site/src/components/ui/ContactSection.module.css`
- Create: `site/src/components/ui/SiteFooter.tsx`
- Create: `site/src/pages/LandingPage.tsx`

**Interfaces:**
- Produces: Landing page completa integrando o Hero com `bregalda_monogram.svg`, o grid `ProjectCycle` com a narrativa editorial e o artefato sticky, divisor de seção e o fechamento `ContactSection`.

- [ ] **Step 1: Implementar `ProjectCycle.tsx` e CSS Module**
  Estruturar o grid de 2 colunas assimétricas (`1fr 1.1fr`), com sticky positioning no artefato e ordenação responsiva no mobile (Narrativa → Artefato → Evidências).

- [ ] **Step 2: Implementar `ContactSection.tsx` e `SiteFooter.tsx`**
  Criar a seção institucional de encerramento em Roxo Bregalda `#4B2E83` com display typography monumental e botão de CTA de alta conversão.

- [ ] **Step 3: Construir `LandingPage.tsx`**
  Montar o Hero com badge circular de monograma, H1 Display *"Turn chaos into a test."*, caixa de instalação rápida `go install ...` em JetBrains Mono e o grid `ProjectCycle`.

- [ ] **Step 4: Validar build e visual da Landing Page**
  Verificar fluidez de scroll, tipografia balanceada e ausência de overflow horizontal.

- [ ] **Step 5: Commit**
  `git add site/src/components/ui site/src/pages/LandingPage.tsx && git commit -m "feat(ui): implement ProjectCycle showcase and Bregalda Landing Page"`

---

### Task 6: Portal de Documentação Técnica (`DocsLayout`, `DocsSidebar`, `DocsContent`)

**Files:**
- Create: `site/src/data/docs-content.ts`
- Create: `site/src/components/docs/DocsSidebar.tsx`
- Create: `site/src/components/docs/DocsContent.tsx`
- Create: `site/src/components/docs/CodeBlock.tsx`
- Create: `site/src/pages/DocsPage.tsx`

**Interfaces:**
- Produces: Portal de documentação editorial com Warm Cream canvas, sidebar de busca com atalho de teclado, categorização por tópicos, breadcrumbs mono e blocos de código com cópia.

- [ ] **Step 1: Migrar e tipar dados da documentação em `site/src/data/docs-content.ts`**
  Portar os 8 capítulos existentes com títulos, resumos executivos, categorias e conteúdo bilíngue (PT/EN).

- [ ] **Step 2: Implementar `CodeBlock.tsx`**
  Criar bloco de código com superfície escura (`[data-surface="dark"]`), sintaxe destacada, tag da linguagem e botão discreto de cópia.

- [ ] **Step 3: Implementar `DocsSidebar.tsx` e `DocsContent.tsx`**
  Criar a barra lateral com busca rápida por atalho (`/` ou `Ctrl+K`), e a área de conteúdo com largura máxima de 42rem, títulos com tracking `-0.06em` e caixas de resumo com hairline sutil.

- [ ] **Step 4: Montar `DocsPage.tsx` com navegação anterior/próximo**
  Integrar sidebar e conteúdo com sincronização de rota hash (`#/docs?chapter=...`) e botões de rodapé.

- [ ] **Step 5: Testar leitura e pesquisa**
  Verificar pesquisa por palavras-chave e navegação entre todos os capítulos da documentação.

- [ ] **Step 6: Commit**
  `git add site/src/data/docs-content.ts site/src/components/docs site/src/pages/DocsPage.tsx && git commit -m "feat(docs): implement editorial DocsLayout with search, sidebar and code blocks"`

---

### Task 7: Catálogo de Cenários (`ScenariosPage`) e Matriz Hermitage (`MatrixPage`)

**Files:**
- Create: `site/src/data/scenarios-data.ts`
- Create: `site/src/components/artifacts/MatrixTable.tsx`
- Create: `site/src/pages/ScenariosPage.tsx`
- Create: `site/src/pages/MatrixPage.tsx`

**Interfaces:**
- Produces:
  - Catálogo interativo dos 9 cenários acadêmicos com abas para Schema, Seed, Chaos YAML e Análise Adya.
  - Tabela comparativa Hermitage entre SQLite, PostgreSQL e MySQL com inspeção detalhada de anomalias.

- [ ] **Step 1: Estruturar dados em `scenarios-data.ts`**
  Definir os 9 cenários canônicos com metadados acadêmicos (Adya P4, A3, A5B, G0, G1c, G1a, G2, G-DL).

- [ ] **Step 2: Construir `ScenariosPage.tsx`**
  Implementar o catálogo de cenários utilizando o padrão de abas arquitetônicas com raio de 3px e visualização de SQL/YAML.

- [ ] **Step 3: Construir `MatrixTable.tsx` e `MatrixPage.tsx`**
  Criar a tabela Hermitage com cabeçalhos em JetBrains Mono, status `Ship Green` e `Signal Yellow`, com modal/painel lateral de inspeção formal ao clicar em cada célula.

- [ ] **Step 4: Validar interatividade**
  Testar alternância de cenários e abertura de detalhes da matriz de isolamento.

- [ ] **Step 5: Commit**
  `git add site/src/data/scenarios-data.ts site/src/components/artifacts/MatrixTable.tsx site/src/pages/ScenariosPage.tsx site/src/pages/MatrixPage.tsx && git commit -m "feat(scenarios): implement Scenarios catalog and Hermitage isolation matrix"`

---

### Task 8: Trace Visualizer (`VisualizerPage`) e Integração WebAssembly (`PlaygroundPage` + Web Worker)

**Files:**
- Create: `site/src/lib/wasm-bridge.ts`
- Create: `site/src/pages/VisualizerPage.tsx`
- Create: `site/src/pages/PlaygroundPage.tsx`

**Interfaces:**
- Produces:
  - Trace Visualizer simulando `chaossql ui` com alternador Raw (20 ops) vs Shrunk (2 ops).
  - Workbench WebAssembly completo conectado a `chaossql.wasm` via Web Worker, com renderização de grafo Adya DSG, raias Gantt e console de logs.

- [ ] **Step 1: Implementar `site/src/lib/wasm-bridge.ts`**
  Gerenciar o ciclo de vida do Web Worker apontando para `/wasm/wasm-worker.js` e `chaossql.wasm`, com tipagem TypeScript estrita para mensagens de entrada/saída.

- [ ] **Step 2: Implementar `VisualizerPage.tsx`**
  Criar o visualizador de traços com alternador dinâmico de modo, timeline Gantt com raias de transações $T1$ e $T2$ e inspetor de queries.

- [ ] **Step 3: Implementar `PlaygroundPage.tsx`**
  Montar o playground completo encapsulado em `ArtifactCard`: controles de sliders para workers/jitter/iterations, editor de carga YAML, visualizador de grafo Adya DSG via SVG, raias Gantt e régua de métricas.

- [ ] **Step 4: Testar execução WASM no navegador**
  Executar o fuzzer em WebAssembly no browser e garantir que a interface gráfica responda sem engasgos durante a simulação.

- [ ] **Step 5: Commit**
  `git add site/src/lib/wasm-bridge.ts site/src/pages/VisualizerPage.tsx site/src/pages/PlaygroundPage.tsx && git commit -m "feat(wasm): implement Trace Visualizer and in-browser WASM Playground"`

---

### Task 9: Integração Global, Roteamento SPA, Internacionalização e Verificação de Build

**Files:**
- Create: `site/src/App.tsx`
- Create: `site/src/main.tsx`
- Create: `site/src/lib/i18n.ts`
- Modify: `site/wrangler.toml`

**Interfaces:**
- Produces: Aplicação completa integrada com roteamento por hash, dicionário bilíngue PT/EN persistido no `localStorage`, deploy estático verificado para Cloudflare Pages.

- [ ] **Step 1: Implementar `i18n.ts`**
  Dicionário centralizado com todos os textos da UI e documentação, com hook `useI18n()` reativo.

- [ ] **Step 2: Implementar `App.tsx` e `main.tsx`**
  Roteador SPA mapeando rotas (`#/`, `#/docs`, `#/scenarios`, `#/visualizer`, `#/matrix`, `#/playground`), integrando `SiteNav`, as páginas e o rodapé institucional.

- [ ] **Step 3: Configurar `wrangler.toml`**
  Garantir que a diretiva `pages_build_output_dir = "dist"` aponte para o diretório de compilação do Vite.

- [ ] **Step 4: Executar suíte completa de verificação**
  - Executar `npm run typecheck` para verificar zero erros de TypeScript.
  - Executar `npm run build` para garantir que o bundle estático é gerado sem falhas.
  - Verificar tamanho dos assets e integridade dos links dos arquivos vetoriais.

- [ ] **Step 5: Commit final de integração**
  `git add site && git commit -m "feat: complete Bregalda Design System migration for ChaosSQL website and docs"`
