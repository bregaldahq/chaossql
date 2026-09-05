# Task 2: Expansão Bilíngue da Documentação Técnica (`site/docs-data.js`)

## Contexto & Localização
Repositório: `/root/chaossql`
Arquivo a modificar:
- `/root/chaossql/site/docs-data.js`

## Objetivos Exatos
1. Reestruturar `window.DOCS_DATA` para suportar idiomas múltiplos como chave primária:
   ```javascript
   window.DOCS_DATA = {
     pt: {
       'getting-started': { ... },
       'dsl-spec': { ... },
       'cli-reference': { ... },
       'trace-visualizer': { ... },
       'cicd-sarif': { ... },
       'drivers': { ... },
       'go-sdk': { ... },
       'academic-theory': { ... }
     },
     en: {
       'getting-started': { ... },
       'dsl-spec': { ... },
       'cli-reference': { ... },
       'trace-visualizer': { ... },
       'cicd-sarif': { ... },
       'drivers': { ... },
       'go-sdk': { ... },
       'academic-theory': { ... }
     }
   };
   ```
2. Manter 100% da integridade dos 8 capítulos atuais em português dentro de `window.DOCS_DATA.pt`.
3. Escrever a versão integral, técnica e idiomática em Inglês (`en`) para todos os 8 capítulos dentro de `window.DOCS_DATA.en`:
   - `getting-started`:
     - title: 'Quickstart Guide & Core Architecture'
     - category: 'Getting Started'
     - summary: 'Complete guide to installing, configuring, and executing your first deterministic database race condition stress test with zero CGO dependencies.'
     - content: Instalação Go (`go install`), binário pré-compilado, container Docker, pipeline de 4 estágios (Fuzzer, Detector, Shrinker, Reporter), primeiro arquivo `chaos.yaml`, execução local com `--workers 4 --duration 10s`.
   - `dsl-spec`:
     - title: 'Declarative Language Specification'
     - category: 'Specification'
     - summary: 'Exhaustive syntax reference for chaos.yaml: transaction definitions, dynamic expressions, variable bindings, and invariant assertions.'
     - content: Estrutura do documento, blocos `setup`, `transactions`, `workload`, `invariants`, injeção de faltas (`fault_injection`), captura de variáveis com `-> var`, expressões dinâmicas `{var - 100}`, validação de predicados e asserções SQL (`==`, `!=`, `<`, `<=`, `>`, `>=`).
   - `cli-reference`:
     - title: 'CLI Manual, Subcommands & All 12 Flags'
     - category: 'Interface'
     - summary: 'Comprehensive manual for the chaossql command-line interface, detailed options, exit codes, and operational flags.'
     - content: Todos os 9 subcomandos (`run`, `demo`, `ui`, `diff`, `replay`, `bench`, `validate`, `init`, `matrix`), tabela completa das 12 flags (`--driver`, `--dsn`, `--config`, `--workers`, `--duration`, `--jitter-min`, `--jitter-max`, `--seed`, `--pct-depth`, `--ddmin`, `--output`, `--port`), exemplos práticos de CLI e códigos de saída (0=Clean, 1=Anomaly detected, 2=Config error).
   - `trace-visualizer`:
     - title: 'Interactive Trace Visualizer (chaossql ui)'
     - category: 'Visualization'
     - summary: 'Embedded HTTP visualizer with microsecond-resolution Gantt swimlanes, interactive Adya dependency graphs, and statement inspector.'
     - content: Inicialização do servidor web embutido (`chaossql ui --port 8090`), swimlanes de goroutines por worker, marcação de conflitos críticos ($rw, ww, wr$), visualização do grafo de serialização direta (DSG), e comparação entre trace bruto (20 ops) vs rastro reduzido (2 ops).
   - `cicd-sarif`:
     - title: 'CI/CD, SARIF 2.1.0 & GitHub Actions'
     - category: 'CI/CD & Security'
     - summary: 'Automated pipeline integration with GitHub Code Scanning, OASIS SARIF 2.1.0 output, JUnit XML, and OpenTelemetry OTLP tracing.'
     - content: Configuração de GitHub Actions (`action.yml`), exportação SARIF 2.1.0 para a aba Security/Code Scanning do GitHub, tabela das 9 regras formais (`CHAOS001` a `CHAOS009`), relatório JUnit XML e spans OTLP.
   - `drivers`:
     - title: 'Supported Database Engines & Drivers'
     - category: 'Engines'
     - summary: 'Deep dive into pure-Go SQLite, PostgreSQL pgx connection pooling, and MySQL/MariaDB isolation quirks and gap locks.'
     - content: SQLite em Go puro (`modernc.org/sqlite`, CGO=0, 13.9M ops/s), PostgreSQL (`pgx/v5`, Read Committed vs Serializable Snapshot Isolation), MySQL/MariaDB (InnoDB Repeatable Read, gap locks e phantom protection), strings de conexão DSN suportadas.
   - `go-sdk`:
     - title: 'Go Developer Testing SDK (chaostest)'
     - category: 'SDK'
     - summary: 'Programmatic API in pkg/chaostest for embedding deterministic isolation fuzzing directly into standard go test suites.'
     - content: Instalação do pacote `pkg/chaostest`, exemplo completo de teste unitário Go com `chaostest.NewHarness(t)`, API fluente de transações e asserções, e geração automática de testes de regressão independentes `repro_test.go` a partir de traços mínimos do $ddmin$.
   - `academic-theory`:
     - title: 'Formal Theory & Mathematical Foundation'
     - category: 'Theory'
     - summary: 'Mathematical foundations of concurrency fuzzing: Bernstein conditions, Adya CSR theorems, Burckhardt PCT probability bounds, and Zeller delta-debugging.'
     - content: Condições de Bernstein para concorrência de acessos, Teorema de Serializabilidade por Grafo de Conflito (CSR de Papadimitriou & Adya), Prova e Teorema de Agendamento por Prioridade Probabilística (PCT de Burckhardt et al. ASPLOS 2010: $\mathbb{P} \ge \frac{1}{n \cdot k^{d-1}}$), algoritmo $ddmin$ de Zeller para minimização de contraexemplos, e comparação com linearizabilidade de Elle/Jepsen.

## Regras
- Código SQL, fórmulas matemáticas KaTeX e nomes de parâmetros/comandos não devem ser alterados.
- Todo o texto em inglês deve ser fluido, idiomático, técnico e rigoroso.
- Validar sintaxe: `node -c site/docs-data.js`.
- Fazer commit das alterações: `git commit -m "docs: add comprehensive english documentation suite (v1.2.0)"`.
