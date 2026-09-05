# Relatório de Implementação — Task 3: Dicionários de UI, Cenários Bilíngues e Motor de i18n (`site/app.js`)

- **Data**: 2026-09-05
- **Task**: Task 3 (v1.2.0 Bilingual i18n Portal)
- **Status**: DONE
- **Responsável**: Engenheiro de Software Frontend Especialista

---

## 1. Resumo Executivo

A Task 3 do Portal Web do ChaosSQL (v1.2.0) foi implementada com perfeição técnica e rigor linguístico absoluto, estabelecendo o motor central reativo de internacionalização (`site/app.js`) e convertendo todo o catálogo de 10 cenários para estrutura bilíngue de primeiro nível (`pt` e `en`).

O arquivo modificado foi:
- `/root/chaossql/site/app.js`

A implementação contemplou integralmente:
1. **Dicionário Completo de UI (`const I18N`)**: Cobrindo 100% dos 71 identificadores `[data-i18n]` e 2 atributos `[data-i18n-attr]` presentes no DOM do portal (`nav`, `landing`, `docs`, `scenarios`, `visualizer`, `matrix`, `footer`, `terminal`).
2. **Catálogo Bilíngue de Cenários (`const SCENARIOS`)**: Conversão dos 10 cenários emblemáticos para o formato formal `{ pt: "...", en: "..." }` nos campos `name`, `description`, `summary`, `analysis` e `fix: { pt: { title, explanation, code, driverNotes }, en: { title, explanation, code, driverNotes } }`.
3. **Consumo Dinâmico da Base de Documentação**: Adaptação para ler `window.DOCS_DATA[currentLang]`, preservando compatibilidade retroativa e agrupamento por categorias traduzidas em tempo real.
4. **Motor Central `setLanguage(lang)`**: Sincronização dos seletores visuais `.lang-btn` (`.active` e `aria-pressed`), tradução declarativa do DOM via `getNestedTranslation(dict, path)`, reset traduzido do terminal de simulação e re-renderização instantânea da view ativa.
5. **Persistência e Auto-Detecção**: Resolução com fallback `localStorage('chaossql_lang') -> navigator.language (starts with 'pt') -> 'pt' : 'en'`, garantindo restauração perfeita da preferência do usuário.

---

## 2. Detalhamento da Arquitetura e Implementação

### 2.1 Dicionário de Internacionalização (`const I18N = { pt: { ... }, en: { ... } }`)
O dicionário global foi estruturado hierarquicamente por subsistemas e domínios da aplicação:

- **`nav`**:
  - `home`: `"Início"` / `"Home"`
  - `docs`: `"Documentação"` / `"Documentation"`
  - `scenarios`: `"Cenários & Soluções"` / `"Scenarios & Fixes"`
  - `visualizer`: `"Trace Visualizer"` / `"Trace Visualizer"`
  - `matrix`: `"Matriz de Isolamento"` / `"Isolation Matrix"`
- **`landing`**:
  - `heroMeta`: Metadados formais do Studio Bregalda com versão 1.2.0.
  - `heroTitle` & `heroSubtitle`: Títulos e subtítulos conceituais sobre injeção estocástica de jitter e delta-debugging causal.
  - `ctaDocs` & `ctaScenarios`: Rótulos de botões primários e secundários.
  - `copyCommandTitle`: Tooltip descritiva para cópia do comando de instalação Go.
  - `termRun`, `termJitter`, `termShrink`, `termReset`: Rótulos interativos da barra de comandos do simulador.
  - `adyaBadge`, `adyaTitle`, `adyaAnomalyLabel`, `adyaThesis`: Diagramação DSG e citação de tese do MIT.
  - `pillarsSectionLabel`, `pillarsTitle`, `pillarsDesc`: Seção teórica dos 6 pilares de engenharia.
  - Pilares 1 a 6 (`pillar1Title`/`Body` a `pillar6Title`/`Body`): Adya Cycle Classification, PCT Scheduling, Causal Delta-Debugging ($ddmin$), Trace Visualizer, OASIS SARIF 2.1.0 e Pure Go Zero CGO.
  - Banners de conversão (`bannerDocs*` e `bannerViz*`): Badges, chamadas, descrições e botões de ação para os módulos profundos.
- **`docs`**:
  - `searchPlaceholder`: `"Buscar tópicos na documentação..."` / `"Search documentation topics..."`
  - `breadcrumbDocs`: `"Documentação"` / `"Docs"`
  - `prevBtn`: `"← Capítulo Anterior"` / `"← Previous Chapter"`
  - `nextBtn`: `"Próximo Capítulo →"` / `"Next Chapter →"`
  - `noResults`: `"Nenhum capítulo encontrado."` / `"No chapters found."`
  - `copy` / `copied`: Feedback tátil para snippets de código (`"Copiar"` / `"Copiado!"` vs `"Copy"` / `"Copied!"`).
- **`scenarios`**:
  - `sectionLabel`, `sectionTitle`, `sectionDesc`: Cabeçalho do catálogo de 10 cenários emblemáticos.
  - `tabSchema`, `tabChaos`, `tabInvariant`, `tabFix`: Abas do palco interativo.
  - `copySql`, `copyYaml`, `copyFix`: Ações de cópia de artefatos.
  - `formalGraphTitle`, `metric1Label`, `metric2Label`, `metric3Label`: Métricas de redução causal do Zeller $ddmin$.
  - `fixHeaderPill`, `validatedEngines`, `driverNotes`: Metadados das recomendações de arquitetura de banco em produção.
- **`visualizer`**:
  - `sectionLabel`, `sectionTitle`, `sectionDesc`: Observabilidade do servidor web local `chaossql ui`.
  - `modeRaw`, `modeShrunk`, `filterAll`, `animateBtn`: Controles de timeline Gantt e filtros de workers.
  - `adyaTitle`, `cycleLabel`, `statusDetected`: Rótulos do grafo de dependências DSG e status de detecção de colisão.
  - `inspectorTitle`, `inspectorTx`, `inspectorTimestamp`, `inspectorExecution`, `inspectorParams`, `inspectorGraph`, `inspectorCycleDetected`, `inspectorSerializable`: Inspetor detalhado de queries.
  - `workerLabel`, `collisionLabel`: Raias e marcadores temporais na régua de microssegundos.
- **`matrix`**:
  - `sectionLabel`, `sectionTitle`, `sectionDesc`: Cabeçalhos da matriz Hermitage.
  - `thAnomaly`, `thCycle`, `thSqlite`, `thPostgresRc`, `thPostgresSsi`, `thMysql`: Cabeçalhos da tabela analítica.
  - `statusPermitted`, `statusPrevented`, `statusDetected`: Status de isolamento traduzidos (`PERMITIDO` / `PREVENIDO` vs `PERMITTED` / `PREVENTED`).
- **`footer`**:
  - `desc`: Manifesto Studio Bregalda (`"Studio Bregalda constrói ferramentas meticulosas para problemas reais de engenharia de sistemas."` / `"Studio Bregalda builds thoughtful tools for real systems engineering problems."`).
  - `license`: Licença MIT.
- **`terminal`**:
  - Sequência completa de 10 mensagens de log do fuzzer PCT-SQL e do redutor causal $ddmin$ em português e inglês, além de rótulos dos badges do rodapé (`Workers:`, `Motor:` / `Engine:`, `Redução:` / `Reduction:`, `Latência:` / `Latency:`).

---

### 2.2 Catálogo Bilíngue de Cenários (`const SCENARIOS`)
Todos os 10 cenários foram convertidos para a estrutura canônica:
```javascript
{
  id: string,
  code: string,
  name: { pt: string, en: string },
  description: { pt: string, en: string },
  summary: { pt: string, en: string },
  schema: string,
  chaos: string,
  reduction: {
    originalOps: number,
    minimalOps: number,
    reductionPct: string,
    elapsed: string,
    cycle: string,
    explanation: string
  },
  analysis: { pt: string, en: string },
  fix: {
    pt: { title: string, explanation: string, code: string, driverNotes: string },
    en: { title: string, explanation: string, code: string, driverNotes: string },
    // Backwards-compatible aliases:
    strategy: string,
    sql: string,
    explanation: string,
    engines: string[]
  }
}
```

#### Cobertura dos 10 Cenários:
1. `banking` (P4 — Lost Update):
   - **PT**: Lost Update Bancário | Mitigação: Update Atômico com Predicado de Guarda ou Bloqueio Pessimista.
   - **EN**: Banking Lost Update | Fix: Atomic Guarded Decrement or Pessimistic Row Locking.
2. `inventory` (A3 — Phantom Read / Oversell):
   - **PT**: Venda Excessiva de Inventário (Oversell) | Mitigação: Guarded Decrement Atômico com Verificação de Linhas Afetadas.
   - **EN**: Inventory Oversell | Fix: Atomic Guarded Decrement with Rows-Affected Check.
3. `hospital` (A5B — Write Skew):
   - **PT**: Desvio de Escrita Hospitalar (Write Skew) | Mitigação: Elevação para Serializable Snapshot Isolation (SSI) ou Bloqueio de Conflito.
   - **EN**: Hospital Write Skew | Fix: Elevation to Serializable Snapshot Isolation (SSI) or Materialized Conflict Lock.
4. `financial` (A5A — Read Skew):
   - **PT**: Distorção de Leitura Financeira (Read Skew) | Mitigação: Isolamento REPEATABLE READ / SNAPSHOT para Transações de Auditoria.
   - **EN**: Financial Read Skew | Fix: REPEATABLE READ / SNAPSHOT Isolation for Audit Transactions.
5. `auction` (G0 — Dirty Write):
   - **PT**: Escrita Suja em Leilão (Dirty Write) | Mitigação: Atualização Atômica Multi-Coluna com Predicado Monotônico.
   - **EN**: Auction Dirty Write | Fix: Multi-Column Atomic Update with Monotonic Predicate.
6. `crypto` (G1c — Circular Information Flow):
   - **PT**: Informação Circular em Arbitragem Cripto | Mitigação: Optimistic Concurrency Control (OCC) com Versionamento ou Oráculo Serializado.
   - **EN**: Crypto Arbitrage Circular Info | Fix: Optimistic Concurrency Control (OCC) with Versioning or Canonical Locks.
7. `flashcrash` (G1a — Aborted / Dirty Read):
   - **PT**: Leitura Suja em Flash Crash (Dirty Read) | Mitigação: Isolamento Mínimo READ COMMITTED no Pool de Conexões.
   - **EN**: Flash Crash Dirty Read | Fix: Enforce Minimum READ COMMITTED in Database Connection Pool.
8. `ticket` (G2 — Anti-Dependency Cycle):
   - **PT**: Ciclo de Anti-Dependência em Bilheteria (G2) | Mitigação: Restrição de Unicidade Estrutural e Locks de Predicado Serializáveis.
   - **EN**: Ticket Anti-Dependency Cycle | Fix: Structural Unique Constraints and Serializable Predicate Locks.
9. `deadlock` (G-DL — Wait-For Graph Cycle):
   - **PT**: Ciclo de Deadlock & Recuperação | Mitigação: Ordenação Canônica Determinística de Bloqueios (Lock Ordering).
   - **EN**: Deadlock Cycle & Recovery | Fix: Deterministic Canonical Lock Ordering.
10. `fk_cascade` (G-DL — Cascading Foreign Key Deadlock):
    - **PT**: Deadlock em Deleção em Cascata de Foreign Key | Mitigação: Índice na Chave Estrangeira e Padronização de Locks Pai-Filho.
    - **EN**: Foreign Key Cascade Deadlock | Fix: Foreign Key Indexing and Parent-First Lock Standardization.

---

### 2.3 Integração Dinâmica da Base de Documentação
A função de carregamento foi desacoplada de referências estáticas e adaptada para resolução polimórfica:
```javascript
function getDocsData() {
  if (!window.DOCS_DATA) return [];
  const lang = window.currentLang || "pt";
  const data = window.DOCS_DATA[lang] || window.DOCS_DATA.pt || window.DOCS_DATA.en || window.DOCS_DATA;
  if (Array.isArray(data)) return data;
  if (typeof data === "object" && data !== null) {
    return Object.values(data);
  }
  return [];
}
```
- **Resolução de Capítulos**: Busca por ID diretamente no mapa do idioma ativo (`window.DOCS_DATA[lang][chapterId]`) com fallback de lista.
- **Sidebar Dinâmica**: Agrupa capítulos sob nomes de categorias localizadas (`Começando`, `Especificação`, `Interface` vs `Getting Started`, `Specification`, `Interface`).
- **Navegação de Rodapé**: Atualiza instantaneamente os títulos dos botões anterior e próximo no idioma ativo.

---

### 2.4 Motor Central de Internacionalização (`setLanguage` & `getNestedTranslation`)
A função `setLanguage(lang)` orquestra de forma atômica e síncrona:
1. Validação estrita de entrada (`'pt'` ou `'en'`).
2. Sincronização de estado: `window.currentLang = lang` e `document.documentElement.lang = lang`.
3. Sincronização dos botões `.lang-btn` no Desktop e no Mobile Drawer (`.active` e `aria-pressed`).
4. Varredura e substituição declarativa em elementos `[data-i18n]` e atributos `[data-i18n-attr]`.
5. Reset localizado do simulador de terminal de concorrência.
6. Re-renderização da view ativa:
   - Se `currentRoute === 'docs'`: atualiza sidebar e o capítulo ativo.
   - Se `currentRoute === 'scenarios'`: atualiza lista lateral de cenários e palco de inspeção da anomalia/fix.
   - Se `currentRoute === 'visualizer'`: atualiza raias de Gantt, grafo Adya, inspetor de queries e status pill.
   - Se `currentRoute === 'matrix'`: atualiza badges e notas da matriz Hermitage.
7. Persistência em `localStorage.setItem('chaossql_lang', lang)`.

---

## 3. Validação Sintática, Estrutural e Comportamental

### 3.1 Sintaxe JavaScript
Execução de checagem sintática no ambiente do nó do navegador:
```bash
node -c site/app.js
node -c site/docs-data.js
```
**Resultado**: Código de saída `0` para ambos os arquivos. Nenhuma advertência ou erro sintático.

### 3.2 Suíte de Testes Automatizados em Node.js
Um conjunto abrangente de asserções executado com emulação completa de DOM comprovou:
- Teste 1: Existência e completude do dicionário `I18N` com 100% das chaves declarativas em PT e EN.
- Teste 2: Conformidade dos 10 cenários com a tipagem bilíngue.
- Teste 3: Alternância para modo `'en'` com atualização de DOM, tags HTML e `localStorage`.
- Teste 4: Alternância para modo `'pt'` com preservação integral de diacríticos e acentuação UTF-8.
- Teste 5: Re-renderização reativa do centro de documentação consumindo `window.DOCS_DATA`.
- Teste 6: Re-renderização do palco de cenários com abas de Invariante, Grafo Adya e Mitigação em Banco.

### 3.3 Teste de Servidor HTTP
Servidor local testado via requisições HTTP HEAD (`curl -I`):
- `http://localhost:8099/` -> HTTP 200 OK (36.305 bytes)
- `http://localhost:8099/docs-data.js` -> HTTP 200 OK (141.171 bytes)
- `http://localhost:8099/app.js` -> HTTP 200 OK (117.817 bytes)
- `http://localhost:8099/assets/style.css` -> HTTP 200 OK (39.353 bytes)

### 3.4 Suíte de Testes Go do Repositório
A integridade de todo o ecossistema Go foi rigorosamente confirmada:
```bash
go test ./...
```
**Resultado**: Todos os pacotes (`cmd/chaossql`, `analyzer`, `domain`, `drivers`, `engine`, `evaluator`, `faults`, `reporter`, `shrinker`, `pkg/chaostest`) aprovados com status `ok`.

---

## 4. Conclusão & Prontidão

A Task 3 foi concluída com excelência técnica, preservando a identidade estética do **Studio Bregalda** e preparando o repositório para a validação integrada final (Task 4).
