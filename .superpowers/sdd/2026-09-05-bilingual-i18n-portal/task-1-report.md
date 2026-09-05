# Relatório de Implementação — Task 1: Switcher Visual [ PT | EN ] e Marcação Declarativa i18n

- **Data**: 2026-09-05
- **Task**: Task 1 (v1.2.0 Bilingual i18n Portal)
- **Status**: DONE
- **Responsável**: Implementador Especialista Frontend

---

## 1. Resumo Executivo

A Task 1 do Portal Web do ChaosSQL (v1.2.0) foi implementada com sucesso e rigor estético absoluto, alinhada à especificação formal e às diretrizes visuais do **Studio Bregalda**.

Foram modificados com precisão cirúrgica:
- `/root/chaossql/site/index.html`
- `/root/chaossql/site/assets/style.css`

---

## 2. Detalhamento das Alterações

### 2.1 Componente Segmented Pill Switch (`.lang-switch`)
Inserido em duas posições estratégicas no `site/index.html`:
1. **Barra de Navegação Desktop (`.nav-links`)**: Imediatamente antes do botão GitHub (`.nav-cta`).
2. **Gaveta Móvel (`#mobileNavDrawer`)**: No topo do container, garantindo acesso imediato em telas pequenas e dispositivos móveis.

Estrutura HTML semântica e acessível implementada:
```html
<div class="lang-switch" role="group" aria-label="Language Selector">
  <button class="lang-btn active" data-lang="pt" type="button" aria-pressed="true">PT</button>
  <button class="lang-btn" data-lang="en" type="button" aria-pressed="false">EN</button>
</div>
```

### 2.2 Estilização e Design Tokens Studio Bregalda (`site/assets/style.css`)
- **Container (`.lang-switch`)**:
  - Background em dark surface: `var(--color-surface, var(--bg-surface))` (`#181328`).
  - Borda sutil: `1px solid rgba(255, 255, 255, 0.1)`.
  - Geometria em pílula: `border-radius: 999px`, `padding: 2px`, `gap: 2px`.
- **Botões (`.lang-btn`)**:
  - Tipografia: `font-family: 'JetBrains Mono', monospace`.
  - Tamanho: `font-size: 0.72rem`, `font-weight: 600`.
  - Geometria e espaçamento: `border-radius: 999px`, `padding: 3px 10px`.
  - Cor normal: `var(--color-cream)` com `opacity: 0.65`.
  - Transição cinemática: `all 0.2s cubic-bezier(0.16, 1, 0.3, 1)`.
- **Hover (`.lang-btn:hover`)**:
  - `opacity: 1`, `color: var(--color-yellow)`.
- **Ativo (`.lang-btn.active`)**:
  - Background em amarelo ouro canônico: `var(--color-yellow)` (`#F5C400`).
  - Cor do texto em contraste elegante: `var(--color-ink)` (`#2A2140`).
  - Opacidade total `1` e brilho difuso: `box-shadow: 0 0 10px rgba(245, 196, 0, 0.35)`.
- **Mobile Drawer Adaptations**:
  - Alinhamento superior compacto via `.mobile-nav-drawer .lang-switch { align-self: flex-start; margin-bottom: 8px; }`.

### 2.3 Mapeamento Declarativo de Chaves i18n (`data-i18n` e `data-i18n-attr`)
Foram adicionados atributos declarativos cobrindo 100% dos elementos estáticos do DOM para consumo direto pelo motor de internacionalização (`site/app.js` na Task 3):

| Seção / Componente | Elemento | Atributo Declarativo |
|---|---|---|
| **Navegação Desktop & Mobile** | Início | `data-i18n="nav.home"` |
| | Documentação / Docs | `data-i18n="nav.docs"` |
| | Cenários & Soluções | `data-i18n="nav.scenarios"` |
| | Trace Visualizer | `data-i18n="nav.visualizer"` |
| | Matriz de Isolamento | `data-i18n="nav.matrix"` |
| **Hero Section** | Metadados de versão | `data-i18n="landing.heroMeta"` |
| | Título principal | `data-i18n="landing.heroTitle"` |
| | Subtítulo conceitual | `data-i18n="landing.heroSubtitle"` |
| | Botão CTA Documentação | `data-i18n="landing.ctaDocs"` |
| | Botão CTA Cenários | `data-i18n="landing.ctaScenarios"` |
| | Botão copiar comando | `data-i18n-attr="title:landing.copyCommandTitle"` |
| **Terminal Interativo** | Botão Run Fuzzer | `data-i18n="landing.termRun"` |
| | Botão Injetar Jitter | `data-i18n="landing.termJitter"` |
| | Botão ddmin Shrink | `data-i18n="landing.termShrink"` |
| | Botão Reset | `data-i18n="landing.termReset"` |
| **Grafo Adya (Hero)** | Badge DSG | `data-i18n="landing.adyaBadge"` |
| | Título do Grafo | `data-i18n="landing.adyaTitle"` |
| | Rótulo da anomalia | `data-i18n="landing.adyaAnomalyLabel"` |
| | Citação da Tese | `data-i18n="landing.adyaThesis"` |
| **6 Pilares de Engenharia** | Rótulo da seção | `data-i18n="landing.pillarsSectionLabel"` |
| | Título da seção | `data-i18n="landing.pillarsTitle"` |
| | Descrição da seção | `data-i18n="landing.pillarsDesc"` |
| | Pillar 1: Adya Classification | `data-i18n="landing.pillar1Title"`, `data-i18n="landing.pillar1Body"` |
| | Pillar 2: PCT Scheduling | `data-i18n="landing.pillar2Title"`, `data-i18n="landing.pillar2Body"` |
| | Pillar 3: Delta-Debugging | `data-i18n="landing.pillar3Title"`, `data-i18n="landing.pillar3Body"` |
| | Pillar 4: Trace Visualizer | `data-i18n="landing.pillar4Title"`, `data-i18n="landing.pillar4Body"` |
| | Pillar 5: SARIF & Security | `data-i18n="landing.pillar5Title"`, `data-i18n="landing.pillar5Body"` |
| | Pillar 6: Pure Go & Zero CGO | `data-i18n="landing.pillar6Title"`, `data-i18n="landing.pillar6Body"` |
| **Banners de Conversão** | Banner Docs (badge, título, desc, CTA) | `landing.bannerDocsBadge`, `landing.bannerDocsTitle`, `landing.bannerDocsDesc`, `landing.bannerDocsCta` |
| | Banner Viz (badge, título, desc, CTA) | `landing.bannerVizBadge`, `landing.bannerVizTitle`, `landing.bannerVizDesc`, `landing.bannerVizCta` |
| **Docs Hub (#view-docs)** | Input de busca | `data-i18n-attr="placeholder:docs.searchPlaceholder"` |
| | Breadcrumb raiz | `data-i18n="docs.breadcrumbDocs"` |
| | Botão anterior | `data-i18n="docs.prevBtn"` |
| | Botão próximo | `data-i18n="docs.nextBtn"` |
| **Scenarios (#view-scenarios)** | Rótulo, título e descrição | `scenarios.sectionLabel`, `scenarios.sectionTitle`, `scenarios.sectionDesc` |
| | Abas do cenário | `scenarios.tabSchema`, `scenarios.tabChaos`, `scenarios.tabInvariant`, `scenarios.tabFix` |
| **Visualizer (#view-visualizer)** | Rótulo, título e descrição | `visualizer.sectionLabel`, `visualizer.sectionTitle`, `visualizer.sectionDesc` |
| | Modos Raw vs Shrunk | `visualizer.modeRaw`, `visualizer.modeShrunk` |
| | Filtro de workers | `visualizer.filterAll` |
| | Botão de animação | `visualizer.animateBtn` |
| | Cabeçalho e label do grafo Adya | `visualizer.adyaTitle`, `visualizer.cycleLabel` |
| **Matrix (#view-matrix)** | Rótulo, título e descrição | `matrix.sectionLabel`, `matrix.sectionTitle`, `matrix.sectionDesc` |
| | Cabeçalhos da tabela | `matrix.thAnomaly`, `matrix.thCycle`, `matrix.thSqlite`, `matrix.thPostgresRc`, `matrix.thPostgresSsi`, `matrix.thMysql` |
| **Footer Global** | Descrição Studio Bregalda | `data-i18n="footer.desc"` |
| | Licença MIT | `data-i18n="footer.license"` |

---

## 3. Validação e Integridade

1. **Integridade de Estrutura do HTML**:
   - Analisador sintático HTML do Python (`html.parser.HTMLParser`) validou o arquivo `site/index.html` sem nenhum erro ou tag malformada.
   - Todas as 5 views (`#view-landing`, `#view-docs`, `#view-scenarios`, `#view-visualizer`, `#view-matrix`) mantiveram seus identificadores originais e ordem no fluxo DOM.
2. **Sintaxe dos Arquivos JS**:
   - `node -c site/app.js` e `node -c site/docs-data.js` executados com código de saída 0.
3. **Servidor HTTP & Resolução de Assets**:
   - Teste local via HTTP com retorno HTTP 200 verificado para `/`, `/assets/style.css`, `/app.js` e `/docs-data.js`.
4. **Isolamento de Dependências**:
   - Zero dependências externas introduzidas.
