# ChaosSQL — Implementação do Bregalda Design System & Arquitetura Frontend React

- **Data**: 2026-09-06
- **Status**: Em Revisão Formal (Pós-Brainstorming com Usuário)
- **Autor**: Antigravity & Ricardo Bregalda (Studio Bregalda)
- **Contexto**: Modernização Visual e Arquitetural do Portal e Documentação Técnica do ChaosSQL (`chaossql.bregalda.com`)

---

## 1. Contexto e Motivação

O **ChaosSQL** é um fuzzer determinístico de concorrência e integridade transacional SQL desenvolvido pelo Studio Bregalda. O portal web atual em `/root/chaossql/site` foi implementado inicialmente com uma estética puramente dark (`#120E1F`) em HTML5/CSS/JS estático.

Paralelamente, o Studio Bregalda estabeleceu o seu ecossistema canônico de marca e design através da skill `bregalda-design-system` e da implementação de referência em `/root/me/bregalda-site` (branch de produção `feat/portfolio`).

Este design system une dois pilares inegociáveis:
1. **Calma Editorial Elegante**: Canvas predominante em Warm Cream (`#FCFBF8`), tipografia Inter equilibrada com tracking negativo afiado (`-0.06em`), entrelinha generosa e divisores sutis em hairline.
2. **Autenticidade Técnica Profunda**: Metadados em JetBrains Mono com tracking alargado (`0.13em`), grafos de dependência serial de Adya, tabelas de intercalação de concorrência T1 vs T2, evidências verificáveis e ausência estrita de dados sintéticos sem rotulagem honesta.

O objetivo deste projeto é reestruturar integralmente o portal web e o centro de documentação do ChaosSQL, migrando sua base para uma arquitetura moderna em **Vite + React 19 + TypeScript**, adotando 100% das diretrizes visuais e componentes canônicos do **Bregalda Design System**.

---

## 2. Objetivos e Não-Objetivos

### Objetivos
- **Alinhamento Estrito à Identidade de Marca Bregalda**:
  - Implantação da paleta canônica: Warm Cream `#FCFBF8` (68%), Deep Ink `#2A2140` (18%), Bregalda Purple `#4B2E83` (9%), Signal Yellow `#F5C400` (3%) e Ship Green `#22C55E` (2%).
  - Arquitetura Dual-Surface acessível (`[data-surface="light"]` no canvas e `[data-surface="dark"]` nos artefatos técnicos e código), com conformidade estrita de contraste WCAG AA (anel de foco roxo no creme; anel de foco amarelo no escuro).
  - Geometria arquitetônica: controles e botões com raio estrito de `3px` (proibição de pill buttons), cartões de artefato com raio de `9px` e ritmo espacial base 8px (`--space-1` a `--space-6`).
  - Incorporação dos ativos vetoriais canônicos em SVG (`bregalda_wordmark.svg`, `bregalda_monogram.svg`, `icone_bregalda.svg`, `bregalda_primary_lockup.svg`, `bregalda_secondary_lockup.svg`).
- **Arquitetura Frontend Moderna (Vite + React 19 + TypeScript)**:
  - Base de código modular, tipada e com hot reload instantâneo no WSL.
  - Componentização dos padrões canônicos: `SiteNav`, `ProjectCycle`, `ArtifactCard`, `DocsLayout`, `MatrixTable` e `ContactSection`.
  - Preservação da infraestrutura existente de deploy no Cloudflare Pages gerando pasta estática `dist/` via `wrangler.toml`.
- **Experiência de Leitura Editorial na Documentação (`#/docs`)**:
  - Canvas creme com leitura confortável (medida máxima de `42rem` / ~62ch), busca instantânea indexada, categorias claras e blocos de código com destaque de sintaxe em containers escuros isolados.
- **Interatividade Completa Preservada & Aprimorada**:
  - **Playground WebAssembly (`#/playground`)**: Execução do binário Go `chaossql.wasm` (8.1MB) em Web Worker isolado, com sliders de controle, editor YAML de cargas de trabalho, geração de grafo DSG Adya interativo e raias Gantt.
  - **Trace Visualizer (`#/visualizer`)**: Alternador entre rastro bruto (20 ops) e rastro encolhido 1-minimal (2 ops) via algoritmo causal $ddmin$.
  - **Matriz de Isolamento Hermitage (`#/matrix`)**: Tabela comparativa dinâmica entre SQLite, PostgreSQL e MySQL.
- **Suporte Bilíngue Contínuo (PT/EN)**:
  - Preservação e refinamento do dicionário de internacionalização instantâneo sem recarregamento.

### Não-Objetivos
- Não alterar a API interna do CLI Go do ChaosSQL nem a especificação dos arquivos `chaos.yaml`.
- Não criar backend dinâmico ou servidor Node.js em produção (a aplicação permanece 100% estática e client-side, servida pela Cloudflare CDN).
- Não inventar elementos visuais ou logotipos fora dos arquivos vetoriais oficiais contidos em `bregalda_brand/`.

---

## 3. Arquitetura de Design & Identidade Visual

### 3.1. Tokens Canônicos de Cor
Conforme extraído de `bregalda-design-system/tokens.css` e `/root/me/bregalda-site`:

```css
:root {
  /* Paleta Canônica */
  --cream: #fcfbf8;
  --ink: #2a2140;
  --purple: #4b2e83;
  --yellow: #f5c400;
  --green: #22c55e;

  /* Cores de Profundidade e Atmosfera */
  --cosmos-deep: #05030c;
  --cosmos-sky: color-mix(in srgb, var(--ink) 55%, #05030c);
  --cosmos-center: #150f26;

  /* Tipografia */
  --font-inter: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-jetbrains-mono: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace;

  /* Escala de Tipos */
  --type-display: clamp(2.9rem, 4.65vw, 6.5rem);
  --type-h2: clamp(2.5rem, 5.25vw, 6rem);
  --type-h3: clamp(1.8rem, 3.2vw, 3rem);
  --type-body: 1rem;       /* 16px */
  --type-body-lg: 1.125rem;/* 18px */
  --type-action: 0.875rem; /* 14px */
  --type-meta: 0.75rem;    /* 12px */

  /* Ritmo Espacial (Base 8px) */
  --space-1: 0.5rem;   /* 8px */
  --space-2: 1rem;     /* 16px */
  --space-3: 1.5rem;   /* 24px */
  --space-4: 2rem;     /* 32px */
  --space-5: 3rem;     /* 48px */
  --space-6: 4rem;     /* 64px */

  /* Geometria e Dimensões */
  --radius-control: 3px;
  --radius-badge: 4px;
  --radius-card: 9px;
  --radius-modal: 16px;
  --page-gutter: clamp(1.1rem, 4.5vw, 6rem);
  --content-max: 90rem;   /* 1440px */
  --text-max: 42rem;      /* ~62ch */
  --nav-height: 76px;
}
```

### 3.2. Regras de Superfície (Dual-Surface)
- **Canvas Geral (`data-surface="light"`)**:
  - `--background: var(--cream)`
  - `--foreground: var(--ink)`
  - `--text-secondary: color-mix(in srgb, var(--ink) 72%, var(--cream))`
  - `--focus-color: var(--purple)` *(Garante contraste WCAG AA 3:1+ sobre o creme; amarelo é proibido sobre creme)*.
- **Painéis Técnicos e Artefatos (`data-surface="dark"`)**:
  - `--background: var(--cosmos-sky)`
  - `--foreground: var(--cream)`
  - `--text-secondary: color-mix(in srgb, var(--cream) 72%, transparent)`
  - `--focus-color: var(--yellow)` *(Alto contraste sobre fundos cósmicos)*.

---

## 4. Arquitetura de Componentes e Telas

### 4.1. Cabeçalho Fixo Canônico (`SiteNav`)
- **Fixação e Backdrop**: Altura de 76px (desktop) / 64px (mobile), `position: sticky; top: 0; z-index: 40`, fundo `color-mix(in srgb, var(--cream) 94%, transparent)` com `backdrop-filter: blur(16px)` e borda hairline inferior.
- **Marca**: `bregalda_wordmark.svg` oficial (138x36px) no lado esquerdo com descritor `ChaosSQL` em JetBrains Mono.
- **Links de Navegação**: `Início`, `Documentação`, `Cenários`, `Visualizador`, `Matriz`, `Playground`.
- **Seletor de Idioma**: Switch `[ PT | EN ]` com raio de `3px`.
- **CTA Primário**: Botão com fundo Roxo Bregalda (`#4B2E83`), texto Warm Cream, indicador Signal Yellow (`↗`), raio de `3px` apontando para o GitHub.
- **Mobile**: Gaveta lateral acessível com fallback `<noscript>`.

### 4.2. Landing Page & Grid de Demonstração Técnica (`ProjectCycle`)
- **Hero Section**:
  - Ponto focal com `bregalda_monogram.svg` centralizado em badge circular com sombra suave.
  - H1 Display: *"Turn chaos into a test."* (Inter 500, tracking `-0.06em`).
  - Lead Editorial: max-width 42rem, entrelinha 1.8.
  - Caixa de comando de instalação rápida com cópia em um clique (`go install ...`).
- **Grid de Demonstração Técnica (`ProjectCycle`)**:
  - Layout assimétrico 2 colunas (`1fr 1.1fr` desktop; narrativa → artefato → evidências no mobile).
  - **Coluna Esquerda**: Narrativa editorial com rótulo técnico `01 / DETERMINISTIC FUZZER`, título H2, descrição causal dos 3 pilares, lista `<dl>` de evidências em JetBrains Mono e tags de tecnologias.
  - **Coluna Direita**: Artefato sticky canônico (`ChaosSqlArtifact`):
    - Toolbar de 64px com ícone `Database`, título `chaossql`, badge `"Illustrative example"` em `#F5C400` e tag `banking_lost_update`.
    - Tabela de passos concorrentes T1 vs T2.
    - Card de violação de invariante (`actual_balance == expected_balance`).
    - Sequência visual de redução: `Trace ➔ Shrink ➔ repro_test.go`.
    - Disclaimer sintético obrigatório em JetBrains Mono.
- **Divisor de Seção**: Inclusão de `bregalda_primary_lockup.svg` (`BREGALDA · BUILD · LEARN · SHIP`).
- **Fechamento Studio / Contact**: Componente `ContactSection` com fundo Roxo Bregalda (`#4B2E83`), tipografia monumental e CTA de conexão com o estúdio.

### 4.3. Portal de Documentação Técnica (`DocsLayout`)
- **Estrutura**: Barra lateral com busca rápida por atalho (`/` ou `Ctrl+K`), agrupamento em 7 categorias canônicas e marcadores de estado ativo.
- **Tipografia e Leitura**: Breadcrumb técnico mono (`CHAOSSQL / DOCS / CATEGORY / 01 QUICKSTART`), H1 refinado, caixa de resumo executivo e largura de leitura limitada a 42rem.
- **Blocos de Código**: Encapsulados em containers escuros (`[data-surface="dark"]`) com sintaxe realçada, etiqueta de linguagem e botão de cópia com feedback visual.
- **Navegação**: Cartões bidirecionais de próximo/anterior capítulo no rodapé.

### 4.4. Módulos Interativos Avançados
1. **Playground WASM (`PlaygroundPage`)**:
   - Casca de artefato (`ArtifactCard`) com toolbar de 64px.
   - Sliders em JetBrains Mono (Workers, Iterations, Jitter, Seed).
   - Editor YAML com sintaxe destacada.
   - Visualização integrada de grafo Adya DSG, raias Gantt e console de logs.
2. **Trace Visualizer (`VisualizerPage`)**:
   - Alternador dinâmico de modo: `Raw Trace (20 ops)` vs `1-Minimal Shrunk (2 ops)`.
   - Timeline interativa com inspeção de queries SQL individuais.
3. **Matriz de Isolamento (`MatrixPage`)**:
   - Tabela comparativa Hermitage com status `Ship Green` (seguro) e `Signal Yellow` (anomalia permitida).
   - Clique em célula para exibição do diagnóstico formal de Adya e comando de reprodução CLI.

---

## 5. Arquitetura Frontend & Estrutura de Diretórios

O novo código-fonte da interface será hospedado em `/root/chaossql/site`, mantendo os arquivos antigos preservados em `/root/chaossql/site_legacy`:

```text
/root/chaossql/site/
├── index.html                    # HTML canônico com fontes Inter, JetBrains Mono e meta tags SEO
├── package.json                  # React 19, Vite 8, TypeScript, Lucide React, Wouter
├── vite.config.ts                # Bundler Vite configurado para assets WASM e Workers
├── tsconfig.json                 # TypeScript strict mode
├── wrangler.toml                 # Cloudflare Pages: pages_build_output_dir = "dist"
├── public/
│   ├── brand/                    # SVGs canônicos (wordmark, monogram, icone, lockups)
│   ├── wasm/
│   │   ├── chaossql.wasm         # Binário oficial compilado em Go (8.1MB)
│   │   ├── wasm_exec.js          # Runtime Go WebAssembly
│   │   └── wasm-worker.js        # Execução isolada em Web Worker
│   ├── _headers                  # Configurações de cache e security headers Cloudflare
│   └── _redirects                # Roteamento SPA no Cloudflare (/* /index.html 200)
└── src/
    ├── styles/
    │   ├── tokens.css            # Tokens canônicos do Bregalda Design System
    │   └── globals.css           # Reset, layout e utilitários (.technical-label, etc.)
    ├── components/
    │   ├── ui/
    │   │   ├── SiteNav.tsx       # Header sticky com wordmark e CTA Roxo 3px
    │   │   ├── SiteFooter.tsx    # Rodapé institucional e links do ecossistema
    │   │   └── ContactSection.tsx# Fechamento institucional Bregalda
    │   ├── artifacts/
    │   │   ├── ArtifactCard.tsx  # Casca canônica de demonstração (toolbar 64px)
    │   │   ├── ChaosSqlArtifact.tsx # Demonstração T1 vs T2, violação de invariante e ddmin
    │   │   └── MatrixTable.tsx   # Tabela Hermitage estilizada com tokens Bregalda
    │   └── docs/
    │       ├── DocsSidebar.tsx   # Categorias e campo de busca rápida
    │       ├── DocsContent.tsx   # Renderizador editorial (medida max 42rem, callouts)
    │       └── CodeBlock.tsx     # Bloco de código escuro com botão de cópia
    ├── pages/
    │   ├── LandingPage.tsx       # Hero + ProjectCycle + Grid de Evidências
    │   ├── DocsPage.tsx          # Portal editorial de documentação técnica
    │   ├── ScenariosPage.tsx     # Catálogo interativo dos 9 cenários de concorrência
    │   ├── VisualizerPage.tsx    # Trace Visualizer (Raw vs 1-Minimal Shrunk)
    │   ├── MatrixPage.tsx        # Matriz Hermitage comparativa
    │   └── PlaygroundPage.tsx    # Workbench WASM completo com controles
    ├── data/
    │   ├── docs-content.ts       # Capítulos bilíngues da documentação técnica
    │   └── scenarios-data.ts     # Definição e queries dos 9 cenários de concorrência
    ├── lib/
    │   ├── wasm-bridge.ts        # Ponte tipada para comunicação com o Web Worker
    │   └── i18n.ts               # Dicionário de internacionalização (PT / EN)
    ├── App.tsx                   # Roteamento SPA e gerenciador de estado de tema/idioma
    └── main.tsx                  # Ponto de entrada React
```

---

## 6. Plano de Verificação e Critérios de Aceite

1. **Fidelidade Visual e de Tokens (Bregalda Design System)**:
   - Verificação de que o canvas padrão é `#FCFBF8` (Warm Cream).
   - Verificação de ausência de pill buttons em CTAs principais (raio obrigatório de `3px`).
   - Verificação de anéis de foco (`focus-visible` roxo no creme, amarelo no escuro, conforme WCAG AA).
   - Renderização correta dos SVGs canônicos sem distorção ou recoloração indevida.
2. **Integridade da Documentação Técnica**:
   - Todos os capítulos da documentação são renderizados com formatação editorial de alta legibilidade.
   - A busca rápida filtra capítulos instantaneamente por título, categoria e termos técnicos.
3. **Execução Segura do WebAssembly**:
   - O binário `chaossql.wasm` carrega via Web Worker sem travar a interface gráfica.
   - O playground executa os presets (ex: Banking Lost Update) e gera as métricas e grafos Adya esperados.
4. **Compilação e Deploy Estático**:
   - `npm run build` compila com zero erros de TypeScript e gera os assets na pasta `dist/`.
   - `wrangler pages dev` ou validação do `wrangler.toml` confirma prontidão para deploy no Cloudflare Pages.
