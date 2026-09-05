# Task 3: Dicionários de UI, Cenários Bilíngues e Motor de i18n (`site/app.js`)

## Contexto & Localização
Repositório: `/root/chaossql`
Arquivo a modificar:
- `/root/chaossql/site/app.js`

## Objetivos Exatos
1. Criar o dicionário completo `const I18N = { pt: { ... }, en: { ... } }`:
   - `nav`: textos dos links da navbar e gaveta móvel (`Início / Home`, `Documentação / Documentation`, `Cenários & Soluções / Scenarios & Fixes`, `Trace Visualizer`, `Matriz de Isolamento / Isolation Matrix`).
   - `landing`: hero meta (`Studio Bregalda • Systems Engineering • Version 1.2.0`), hero title, hero subtitle, cta buttons (`Explorar Documentação / Explore Documentation`, `Ver 10 Cenários / View 10 Scenarios`), terminal simulator buttons (`[▶ Run Fuzzer]`, `[⚡ Injetar Jitter / Inject Jitter]`, `[🔍 Reduzir ddmin / ddmin Shrink]`, `[↺ Reiniciar / Reset]`), terminal log messages em pt e en, 6 feature cards (títulos e descrições), banners de conversão.
   - `docs`: placeholder da busca (`Buscar tópicos na documentação... / Search documentation topics...`), títulos das 8 categorias, botão anterior e próximo, link inicial do breadcrumbs.
   - `scenarios`: tabs (`Código SQL / SQL Code`, `Grafo de Conflito DAG / Conflict DAG`, `Regra de Invariante / Invariant Rule`, `Fix & Mitigação no Banco / Production Fix & Mitigation`), badges de anomalia e severidade.
   - `visualizer`: títulos, descrições, botões de ação (`▶ Iniciar Simulação / ▶ Animate Execution`, `Raw Trace (20 ops)`, `1-Minimal Shrunk (2 ops)`), filtros por worker (`Todos os Workers / All Workers`), labels de latência, cabeçalhos do inspetor de queries.
   - `matrix`: títulos da tabela, notas explicativas de ANSI vs Adya, legendas de status.
2. Atualizar todos os 10 cenários da constante `SCENARIOS`:
   - `name: { pt: "...", en: "..." }`
   - `description: { pt: "...", en: "..." }`
   - `analysis: { pt: "...", en: "..." }`
   - `fix: { pt: { title: "...", explanation: "...", code: "...", driverNotes: "..." }, en: { title: "...", explanation: "...", code: "...", driverNotes: "..." } }`
   - SQL, graph e invariant permanecem inalterados.
3. Implementar a função central `setLanguage(lang)`:
   - Valida `lang === 'pt' || lang === 'en'`.
   - Atualiza `window.currentLang = lang` e `document.documentElement.lang = lang`.
   - Sincroniza classes `.active` e atributo `aria-pressed` nos botões `.lang-btn[data-lang]`.
   - Varre todos os elementos com `[data-i18n]` e substitui `textContent` usando `getNestedTranslation(I18N[lang], key)`.
   - Varre todos os elementos com `[data-i18n-attr]` (ex: `placeholder:docs.searchPlaceholder`) e atualiza o atributo.
   - Re-renderiza dinamicamente a view ativa:
     - Se `currentRoute === 'docs'`: recarrega a sidebar (`renderDocsSidebar()`) e re-renderiza o capítulo ativo (`renderDocChapter(currentChapter)`).
     - Se `currentRoute === 'scenarios'`: recarrega os botões de cenários (`renderScenarioList()`) e o detalhe do cenário ativo (`renderScenarioDetail(activeScenarioIndex)`).
     - Se `currentRoute === 'visualizer'`: atualiza textos do visualizador e do inspetor.
     - Se `currentRoute === 'matrix'`: atualiza a matriz.
   - Salva preferência no `localStorage.setItem('chaossql_lang', lang)`.
4. Inicialização de Idioma:
   - Ler `localStorage.getItem('chaossql_lang')`.
   - Se nulo, detectar via `(navigator.language || '').toLowerCase().startsWith('pt') ? 'pt' : 'en'`.
   - Chamar `setLanguage(initialLang)` na inicialização da aplicação.
   - Registrar event listeners `click` em todos os botões `.lang-btn` da navbar e mobile drawer.

## Regras
- Validar sintaxe: `node -c site/app.js`.
- Preservar integridade do simulador de terminal, visualizador de traces e hash router.
- Fazer commit: `git commit -m "feat(i18n): implement dynamic bilingual engine and scenario translations"`.
