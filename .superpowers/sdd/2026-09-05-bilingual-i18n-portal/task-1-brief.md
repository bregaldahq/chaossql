# Task 1: Switcher Visual [ PT | EN ] e Marcação Declarativa i18n

## Contexto & Localização
Repositório: `/root/chaossql`
Arquivos a modificar:
- `/root/chaossql/site/index.html`
- `/root/chaossql/site/assets/style.css`

## Objetivos Exatos
1. Inserir o componente `.lang-switch` na barra de navegação desktop de `site/index.html`:
   - Posicionar entre `nav.nav-links` (imediatamente antes do link GitHub `.nav-cta`).
   - Estrutura:
     ```html
     <div class="lang-switch" role="group" aria-label="Language Selector">
       <button class="lang-btn active" data-lang="pt" type="button" aria-pressed="true">PT</button>
       <button class="lang-btn" data-lang="en" type="button" aria-pressed="false">EN</button>
     </div>
     ```
2. Inserir o componente `.lang-switch` também no topo da gaveta de navegação móvel (`.mobile-nav-drawer`):
   - Estrutura idêntica para permitir alternância em telas pequenas.
3. Adicionar atributos declarativos `data-i18n="<chave>"` nos elementos estáticos do HTML para permitir que o motor de i18n substitua os textos dinamicamente:
   - Links da nav: `data-i18n="nav.home"`, `data-i18n="nav.docs"`, `data-i18n="nav.scenarios"`, `data-i18n="nav.visualizer"`, `data-i18n="nav.matrix"`.
   - Links do mobile drawer: mesmos atributos `data-i18n`.
   - Hero: `data-i18n="landing.heroMeta"`, `data-i18n="landing.heroTitle"`, `data-i18n="landing.heroSubtitle"`, `data-i18n="landing.ctaDocs"`, `data-i18n="landing.ctaScenarios"`.
   - Terminal interativo no hero: botões `data-i18n="landing.termRun"`, `data-i18n="landing.termJitter"`, `data-i18n="landing.termShrink"`, `data-i18n="landing.termReset"`.
   - 6 Feature cards na landing page: títulos e descrições com `data-i18n`.
   - Banners de conversão na landing page.
   - Docs view: placeholder de busca (`data-i18n-attr="placeholder:docs.searchPlaceholder"`), botões de capítulo anterior e próximo.
   - Scenarios view: cabeçalho e abas.
   - Visualizer view: cabeçalho, botão `data-i18n="visualizer.animateBtn"`, seletor Raw vs Shrunk, tabs.
   - Matrix view: cabeçalhos e legendas.
4. Adicionar estilização em `site/assets/style.css`:
   - Classes `.lang-switch`, `.lang-btn`, `.lang-btn:hover`, `.lang-btn.active`.
   - Seguir fielmente o design system do Studio Bregalda:
     - Fundo `.lang-switch`: `var(--color-surface)` (`#181328`), border `1px solid rgba(255, 255, 255, 0.1)`, `border-radius: 999px`, `padding: 2px`, `gap: 2px`.
     - Botão `.lang-btn`: `font-family: 'JetBrains Mono', monospace`, `font-size: 0.72rem`, `font-weight: 600`, `padding: 3px 10px`, `border-radius: 999px`, cor `var(--color-cream)`, `opacity: 0.65`, `transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1)`.
     - Ativo `.lang-btn.active`: `background: var(--color-yellow)` (`#F5C400`), `color: var(--color-ink)` (`#2A2140`), `opacity: 1`, `box-shadow: 0 0 10px rgba(245, 196, 0, 0.35)`.
     - Ajustes responsivos para mobile drawer.

## Regras
- Nenhuma dependência externa.
- Não quebrar o HTML nem a hierarquia existente das 5 views (`#view-landing`, `#view-docs`, `#view-scenarios`, `#view-visualizer`, `#view-matrix`).
- Ao concluir, commitar as alterações no git:
  `git add site/index.html site/assets/style.css`
  `git commit -m "feat(ui): add bilingual segmented pill switch and i18n markers"`
