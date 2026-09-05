# Plano de Implementação: Portal Web & Centro de Documentação Robusta do ChaosSQL (v1.2.0)

> **Sub-habilidade Requerida:** Execução com `superpowers:executing-plans` ou `superpowers:subagent-driven-development`.
> **Objetivo:** Transformar o site atual do ChaosSQL em um portal duplo de alta fidelidade: uma **Landing Page cinematográfica e elegante** (focada em conversão, impacto visual e visão geral rápida) conectada a um **Centro de Documentação Técnica Robusto e Aprofundado** (com abas dedicadas, referência de DSL, CLI completo, visualizador de traces, guias de CI/CD, SDK Go, drivers e soluções para os 10 cenários).

---

## 1. Visão Geral da Arquitetura

O site do ChaosSQL é distribuído via Cloudflare Pages através do diretório `site/` (`index.html`, `app.js`, `assets/style.css`, ativos SVG/PNG), configurado com roteamento de SPA (`not_found_handling = "single-page-application"` no `wrangler.toml`).

Para atender ao pedido de **separar em abas/páginas detalhadas** sem depender de frameworks pesados e mantendo o carregamento ultrarrápido com zero dependências externas de runtime, adotaremos uma **arquitetura de portal multi-view com roteamento por Hash**:
* `#/` ou vazio: **Landing Page** (visual refinado, impacto, visão geral, terminal interativo, showcase do motor).
* `#/docs`: **Centro de Documentação Técnica** (portal com sidebar expansível e 8 capítulos aprofundados).
* `#/scenarios`: **Catálogo Interativo dos 10 Cenários** (com nova aba de *Fix / Mitigação no Banco*).
* `#/visualizer`: **Showcase do Trace Visualizer** (`chaossql ui`, Gantt swimlanes e grafos de conflito Adya).
* `#/matrix`: **Matriz Empírica de Isolamento Hermitage** (comparativo entre motores SQL).

---

## 2. Estrutura de Arquivos

```
/root/chaossql/site/
├── index.html                   # Estrutura do portal com views para Landing Page, Docs Hub, Scenarios e Visualizer
├── app.js                       # Roteador por hash, controller dos cenários, simulador terminal e visualizador interativo
├── docs-data.js                 # Base de dados estruturada dos 8 capítulos de documentação técnica
├── assets/
│   ├── style.css                # Design system Studio Bregalda: tokens, grid, docs layout, sidebar, código, animações
│   ├── icone_bregalda.svg       # Logomarca Bregalda
│   ├── bregalda_wordmark.svg    # Wordmark
│   └── ...                      # Demais assets de marca
└── wrangler.toml                # Configuração Cloudflare Pages
```

---

## 3. Fases de Execução

### Fase 1: Arquitetura de Roteamento, Layout Multi-View e Design System

- [x] **1.1 Configurar Roteador por Hash (`app.js`)**
  - Implementar roteamento por eventos `hashchange` e `DOMContentLoaded`.
  - Suportar navegação fluida entre `#/` (Home), `#/docs`, `#/docs/:section`, `#/scenarios`, `#/visualizer` e `#/matrix`.
  - Manter histórico do navegador e suporte a deep-linking para cada capítulo de documentação.

- [x] **1.2 Atualizar Navbar Global e Header (`index.html` & `style.css`)**
  - Adicionar menu com navegação ativa e indicador luminoso:
    - **Início** (`#/`)
    - **Documentação** (`#/docs`)
    - **Cenários & Soluções** (`#/scenarios`)
    - **Trace Visualizer** (`#/visualizer`)
    - **Matriz de Isolamento** (`#/matrix`)
    - Botão CTA com ícone e badge para o repositório GitHub.

- [x] **1.3 Estruturar Containers de Visualização (`index.html`)**
  - Separar as visões em containers independentes com transição suave (`fade-in`):
    - `<div id="view-landing" class="portal-view active">`
    - `<div id="view-docs" class="portal-view">`
    - `<div id="view-scenarios" class="portal-view">`
    - `<div id="view-visualizer" class="portal-view">`
    - `<div id="view-matrix" class="portal-view">`

---

### Fase 2: Redesign Cinematográfico da Landing Page

- [x] **2.1 Novo Hero Section com Impacto Visual**
  - Typography refinada (Inter + JetBrains Mono) com títulos hierárquicos e badge editorial `Studio Bregalda • Systems Engineering • Version 1.2.0`.
  - Headline afiado: *"Deterministic Concurrency & Isolation Fuzzer for SQL Databases"*.
  - Subtítulo equilibrando rigor acadêmico com engenharia de sistemas.
  - CTAs duplos: *"Explorar Documentação"* e *"Ver 10 Cenários"*, mais barra com comando `go install` com botão de cópia instantânea.

- [x] **2.2 Dual-Split Showcase no Hero: Terminal Replay + Grafo Adya SVG**
  - Lado Esquerdo: Simulador interativo de terminal reproduzindo execução caótica, detecção da anomalia e redução Zeller $ddmin$ (20 ops → 2 ops, 90% em 68ms).
  - Lado Direito: Widget SVG ilustrando a relação cíclica de dependência ($T_1 \xrightarrow{rw} T_2 \xrightarrow{ww} T_1$) com nós pulsantes e arestas rotuladas.

- [x] **2.3 Grid de Pilares & Capacidades de Engenharia (Cards Visuais)**
  - 6 cartões refinados com micro-interações:
    1. **Adya Cycle Classification**: Mapeamento formal de grafos de dependência ($wr, ww, rw$).
    2. **PCT-SQL Scheduling**: Garantia probabilística $\mathbb{P} \ge \frac{1}{n \cdot k^{d-1}}$ evitando heurísticas ingênuas.
    3. **Causal Delta-Debugging ($ddmin$)**: Redução de traces barulhentos para 1-minimalidade em $< 200\text{ms}$.
    4. **Interactive Trace Visualizer (`chaossql ui`)**: Servidor local embutido com Gantt swimlane e inspeção de queries.
    5. **OASIS SARIF 2.1.0 & CI/CD**: Gating de segurança nativo no GitHub Code Scanning com 9 regras formais.
    6. **Zero CGO & Pure Go**: Compilação estática com SQLite sem dependências de C runtime e 13.9M ops/s no benchmark.

- [x] **2.4 Seção Teaser do Visualizador de Traces**
  - Mockup visual editorial demonstrando a timeline de concorrência e a transição entre o trace ruidoso de 100 operações e o código isolado de 2 passos.

---

### Fase 3: Centro de Documentação Técnica Robusta (Docs Hub)

- [x] **3.1 Criar Módulo de Conteúdo dos Docs (`docs-data.js`)**
  - Estruturar os 8 capítulos completos em formato rico com metadados, títulos, tags de navegação e exemplos prontos para uso:
    1. **Guia de Início Rápido & Arquitetura**:
       - O que é o ChaosSQL e quais dores resolve.
       - Requisitos (Go 1.22+, sem CGO), instalação binária e compilação do código-fonte.
       - Execução do primeiro teste em 5 minutos com `chaossql init` e `chaossql run`.
    2. **Referência Completa da Especificação (`chaos.yaml` DSL)**:
       - Estrutura hierárquica do arquivo de configuração (`version`, `name`, `database`, `engine`, `operations`, `invariants`, `faults`).
       - Sintaxe de captura de variáveis (`-> var_name`), expressões dinâmicas (`{var_name - 100}`) e predicados (`AND {var > 0}`).
       - Injeção de falhas controladas: aborts intencionais, latência estocástica e desconexões simuladas.
    3. **Manual Completo da CLI & Todas as Flags**:
       - Tabela exaustiva e exemplos para todos os 9 subcomandos:
         - `run`: `--seed`, `--workers`, `--iterations`, `--export-sarif`, `--export-html`, `--export-otel`, `--export-junit`, `--export-summary`, `--export-mermaid`, `--export-repro`, `--ui`, `--json`.
         - `demo`: os 10 cenários pré-configurados.
         - `ui`: servidor visualizador interativo em `127.0.0.1:8090`.
         - `diff`: comparação diferencial entre motores e traces.
         - `replay`: reprodução idêntica com mesma semente PRNG.
         - `bench`: microbenchmarks de alta vazão (13.9M ops/s).
         - `validate`: linter estático e validação de schema de arquivos YAML.
         - `init`: scaffolding guiado com templates para SQLite e Postgres.
         - `matrix`: geração de matriz empírica Hermitage.
    4. **Visualizador Interativo de Traces (`chaossql ui`)**:
       - Arquitetura do visualizador web local e modo embutido (`--ui`).
       - Anatomia do Gantt Swimlane (marcações em microssegundos por goroutine).
       - Grafo de Forças Adya interativo (nós de transação e arestas de conflito).
       - Inspetor de Queries e parâmetros resolvidos.
    5. **Integração com CI/CD, SARIF 2.1.0 & GitHub Actions**:
       - Como plugar o ChaosSQL em Pull Requests usando o `action.yml` oficial.
       - Configuração de alertas de segurança com OASIS SARIF no GitHub Code Scanning.
       - Tabela com as 9 regras formais (`chaossql/P4-lost-update`, `chaossql/A5B-write-skew`, etc.).
       - Relatórios JUnit XML e GitHub Step Summaries automáticos.
       - Exportação de traces OpenTelemetry (OTLP JSON) para Jaeger e Datadog.
    6. **Suporte e Configuração de Motores & Drivers**:
       - SQLite em memória e em arquivo (WAL mode, pure Go `modernc.org/sqlite`, zero CGO).
       - PostgreSQL (`pgx`): strings de conexão DSN, comportamento sob `READ COMMITTED` vs `SERIALIZABLE` (SSI).
       - MySQL / MariaDB: InnoDB Repeatable Read, gap locks e deadlocks.
    7. **Go Developer Testing SDK (`pkg/chaostest`)**:
       - Referência da API fluente do builder `chaostest.New(t)`.
       - Métodos: `WithDriver`, `WithSchema`, `WithSeed`, `WithInvariant`, `WithJitter`, `AddOperation`, `AddOperationWithParams`.
       - Asserções: `AssertNoAnomalies` e `AssertAnomalyDetected`.
       - Como o Zeller $ddmin$ sintetiza testes `repro_test.go` automaticamente durante o `go test`.
    8. **Fundamentação Matemática & Teoria Formal**:
       - Condições de Conflito de Bernstein.
       - Teorema Fundamental da Serializabilidade por Conflito (CSR).
       - Teorema de Burckhardt-Musuvathi sobre probabilidade de agendamento PCT:
         $$\mathbb{P}(\text{Detecção}) \ge \frac{1}{n \cdot k^{d-1}}$$
       - Algoritmo de Delta-Debugging ($ddmin$) e prova de 1-minimalidade de Andreas Zeller.
       - Verificação de Linearizabilidade de Registradores estilo Elle/Jepsen.

- [x] **3.2 Implementar Interface do Docs Hub (`index.html` & `app.js`)**
  - Layout clássico de documentação moderna: Sidebar lateral retrátil e navegável com busca por tópicos + Área de Leitura central com índice (ToC) na direita.
  - Formatação com blocos de código com botão de cópia rápida, alerts estilizados (`[!NOTE]`, `[!TIP]`, `[!WARNING]`) e tabelas responsivas.

---

### Fase 4: Enriquecimento dos 10 Cenários com Soluções e Mitigações

- [x] **4.1 Adicionar 4ª Aba: "Fix & Mitigação" aos Cenários**
  - Atualmente os cenários exibem: *Schema & Seed*, *Chaos Workload* e *Invariant / Reduction*.
  - Adicionar a aba **Fix / Mitigação em Produção**, fornecendo o código SQL e arquitetural para sanar cada uma das 10 anomalias:
    1. **Banking ($P4$)**: Migração de Read-Modify-Write para Update Atômico ou Pessimistic Lock (`SELECT ... FOR UPDATE`).
    2. **Inventory ($A3$)**: Predicado de guarda com decremento condicional (`WHERE id = 1 AND stock >= 1`).
    3. **Hospital ($A5B$)**: Elevação para Serializable Snapshot Isolation (SSI) ou lock de tabela/materializado.
    4. **Financial ($A5A$)**: Execução de auditoria em transação com isolamento `REPEATABLE READ` ou `SNAPSHOT`.
    5. **Auction ($G0$)**: Agrupamento em transação atômica única com locks exclusivos.
    6. **Crypto ($G1c$)**: Inclusão de verificação de timestamps monotônicos e oráculos atômicos.
    7. **Flash Crash ($G1a$)**: Envolvimento de margens em transação com commit prévio validado antes de liquidar.
    8. **Ticket Booking ($G2$)**: Indexação por restrição de unicidade e locks de predicado serializáveis.
    9. **Deadlock ($G-DL$)**: Ordenação canônica e determinística de locks (sempre ordenar por ID crescente).
    10. **Foreign Key Cascade ($G-DL$)**: Padronização de bloqueios pai-filho e isolamento de triggers de deleção.

---

### Fase 5: Showcase Interativo do Trace Visualizer

- [x] **5.1 Criar Seção/Aba do Visualizador de Traces**
  - Permitir que o visitante explore interativamente como funciona o `chaossql ui`:
    - Visualizador de Gantt simulado com timeline de workers simultâneos.
    - Grafo SVG interativo de conflitos que pode ser clicado para inspecionar nós ($T_1, T_2$) e arestas ($rw, ww, wr$).
    - Comparador antes/depois: trace original barulhento de 20 passos vs reprodução 1-minimal de 2 passos.

---

### Fase 6: Responsividade, Performance & Verificação Final

- [x] **6.1 Auditoria CSS & Mobile Responsiveness**
  - Garantir adaptação perfeita em resoluções mobile (smartphones), tablets e telas ultrawide.
  - Sidebar do Docs Hub com gaveta móvel (*drawer*) em telas menores que 768px.
- [x] **6.2 Verificação de Servidor Local & Validação do Harness**
  - Executar `make serve-site` e testar navegação de todas as abas e links.
  - Validar compatibilidade estática com `wrangler.toml` para deploy no Cloudflare Pages.
  - Garantir zero erros de console no JavaScript (`app.js`, `docs-data.js`).
