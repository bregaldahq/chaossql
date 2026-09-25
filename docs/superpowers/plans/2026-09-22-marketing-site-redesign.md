# Redesign do site de divulgação do ChaosSQL: plano detalhado

**Data:** 2026-09-22
**Escopo:** `site/` (landing `/`, `/pricing` e a camada visual compartilhada: nav, footer, tokens). As ferramentas de produto (`/playground`, `/visualizer`, `/matrix`, `/dashboard`, `/docs`) só recebem o novo design system, sem mudanças de conteúdo.
**Objetivo de negócio:** mais conversão em três funis: adoção open source (instalar a CLI, estrelas no GitHub), leads do Cloud (waitlist) e vendas da Concurrency Audit ($1,490).

**Leitura de design:** landing de devtool B2B para engenheiros de backend e tech leads, com linguagem de "laboratório forense" (evidência, precisão, prova), puxando para CSS nativo + Motion, com animação usada para *explicar* o produto e não para enfeitar.

Dials: `DESIGN_VARIANCE 7` / `MOTION_INTENSITY 6` / `VISUAL_DENSITY 4`.

---

## 1. Diagnóstico: por que o site parece feito por IA

Tudo isto vem do código atual, não é impressão genérica.

### 1.1 Estrutura e layout
| Sintoma | Onde | Por que prejudica |
|---|---|---|
| Hero centralizado com **7 elementos empilhados**: monograma com glow roxo, eyebrow "Studio Bregalda · Database Concurrency Systems", H1, parágrafo de ~35 palavras, 2 CTAs, caixa de instalação e um card de anomalia | `LandingPage.tsx:150-238` | É o hero padrão de IA. Nada tem prioridade e o olho não sabe onde pousar. |
| **Grades de cards iguais**: 4 cards de workflow, 3 cards de "pilares", 3 itens de evidência | `LandingPage.tsx:258-270`, `:287-315`, `ProjectCycle.tsx` | É o padrão "três cards iguais", o sinal mais reconhecível de template. |
| **A mesma mensagem repetida 4 vezes** (detectar → reduzir com ddmin → entregar/CI): no hero, no workflow, no ProjectCycle e nos pilares | landing inteira | O leitor sente que o texto está enchendo linguiça, e a página fica longa sem acrescentar nada. |
| **Eyebrow técnico em toda seção** ("Fluxo do Desenvolvedor", "Fundamentos de Engenharia", "Evidência Determinística em Ação"…) + numeração `01 /`, `02 /` | todas as seções | É o ritmo repetitivo de página gerada. |
| Lockup "Build · Learn · Ship" solto no meio da página | `LandingPage.tsx:318-324` | Interrompe a narrativa e não comunica nada sobre o produto. |
| Card de anomalia do hero feito de `div`s estáticas | `LandingPage.tsx:216-236` | Parece um screenshot falso. |

### 1.2 Linguagem e copy
- **O jargão vem antes da dor.** "Directed Serialization Graphs", "1-minimal", "jitter micro-temporal", "arestas wr/ww/rw" aparecem antes de o visitante entender o problema de negócio (dinheiro sumindo, overbooking, estoque negativo).
- **Emojis na interface:** 🚨 🏦 🏥 🔒 🛡️ (12 ocorrências em `src/`).
- **Rótulos vagos de vendas:** "VIP Advisory", "VIP ENGINEERING ENGAGEMENT", "O padrão para equipes de produto".
- **Muitos CTAs disputando a mesma intenção:** "Participar do Cloud Early Access", "Solicitar Acesso", "Solicitar Cloud Team", "Solicitar Cloud Pro", "Solicitar Auditoria". São dois formulários de lead diferentes (`CloudWaitlistSection` e o modal do `PricingPage`).
- **Números sem fonte visível:** "< 200ms", "em milissegundos". Os evals (`evals/01_shrinking_ratio.md`) garantem **≥ 85% de redução e convergência < 2s**. A página precisa usar números reais e medidos.
- **Credibilidade técnica:** o cenário de deadlock usa o badge "G-DL ANOMALY", que não é uma classe de Adya. Para um público especialista, um erro desses derruba a confiança. Revisar todos os badges contra `docs/SCENARIO_ACADEMIC_AUDIT.md`.

### 1.3 i18n e consistência
- Textos fixos em PT dentro da versão EN: "Fundamentos de Engenharia" (`LandingPage.tsx:290`), "Copiar/Copiado!" (`:205-212`), "/permanente" (`PricingPage.tsx:156`), `aria-label` em PT no `ProjectCycle` e no `SiteNav`.
- `og:locale=pt_BR` com título e descrição OG em inglês. `softwareVersion: 1.4.0` no JSON-LD, mas a versão lançada é a v1.6.0.

### 1.4 Visual
- Inter + purple glow (`box-shadow` roxo no monograma e no CTA): é a combinação padrão de "site de IA".
- Raios misturados (3px, 4px, 9px, 16px, 2px no botão de copiar) sem uma regra clara.
- Movimento quase inexistente: só `translateY(-1px)` no hover. Nada ajuda a *entender* a concorrência, que é justamente o ponto em que a animação mais serviria.
- Imagem OG é o lockup da marca (2172×724) e não mostra o produto. Os compartilhamentos no LinkedIn/X não dizem o que a ferramenta faz.
- Nav com 8 itens, incluindo ferramentas internas (Visualizer, Hermitage Matrix), e um "●" decorativo em "Cloud Dashboard ●".

### 1.5 Marketing e medição
- **Nenhum analytics.** Hoje não dá para saber a taxa de conversão de nada, então qualquer redesign seria no escuro.
- Nenhuma prova social: sem contador de estrelas do GitHub, sem "usado por", sem citações, sem link para issues ou releases.
- O formulário da waitlist tem 6 campos (com textarea) e fica no fim da página, com atrito alto.

---

## 2. Estratégia comercial (antes do visual)

### 2.1 Públicos e o que cada um precisa ver
| Público | Dor | Prova que convence | CTA certo |
|---|---|---|---|
| **Engenheiro backend** (usuário) | "Tenho um bug intermitente de saldo/estoque que não reproduzo" | Ver o bug acontecer e a reprodução mínima gerada | Rodar no navegador / instalar a CLI |
| **Tech lead / Staff** (champion) | "Como garanto que isso não volta?" | Integração com CI, GitHub Action, `repro_test.go` versionado | Cloud early access |
| **CTO / Head de Eng** (comprador) | "Risco financeiro/regulatório antes de um lançamento" | Escopo, entregáveis e preço claros da auditoria, e um caso real | Agendar auditoria |

### 2.2 Funil e hierarquia de CTAs (um rótulo por intenção, em todo o site)
1. **Primário:** `Ver o bug acontecer` → leva à demo interativa na própria landing (e dali ao `/playground`).
2. **Secundário:** `Instalar` → copia `go install …` e registra o evento.
3. **Cloud:** `Entrar na lista do Cloud` (formulário curto: só o e-mail; o resto é pedido depois, no e-mail de onboarding).
4. **Auditoria:** `Agendar auditoria` (formulário qualificado, com banco, volume e prazo).

Todos os "Solicitar X" do pricing viram um destes quatro.

### 2.3 Mensagem central
**Decidido:** direção "seus testes passam, seus saldos não fecham". O mercado é global, então o EN é a versão canônica e o PT é uma tradução revisada.
- **H1 EN (canônico):** "Your tests pass. Your balances don't add up."
- **H1 PT:** "Seus testes passam. Seus saldos não fecham."
- **Subtext EN (≤ 20 palavras):** "ChaosSQL forces the races between transactions in PostgreSQL, MySQL and SQLite, then hands you the smallest test that reproduces the bug."
- **Subtexto PT:** "O ChaosSQL força as corridas entre transações no PostgreSQL, MySQL e SQLite e entrega o menor teste que reproduz o bug."

Regra de copy: **dor em linguagem de negócio → mecanismo em linguagem de engenheiro → teoria (Adya, DSG, ddmin) só como camada de aprofundamento**, em "Como funciona" e nos docs.

---

## 3. Nova arquitetura da landing

Hoje: hero → 3 cenários → workflow (4 cards) → ProjectCycle → pilares (3 cards) → lockup → waitlist.
Proposta: 8 seções, cada uma com um trabalho único e um layout diferente.

| # | Seção | Trabalho | Layout | Substitui |
|---|---|---|---|---|
| 1 | **Hero** | Promessa + CTA + ver o problema em 5s | Split assimétrico: texto à esquerda, **trace animado real** à direita | Hero atual |
| 2 | **Faixa de prova** | Credibilidade imediata | Linha única: logos PostgreSQL/MySQL/SQLite/Go/GitHub Actions (Simple Icons), estrelas do GitHub, licença MIT | (novo) |
| 3 | **"O bug em 30 segundos"** | Explicar lost update para qualquer pessoa | Scroll-story com painel fixo: 4 passos, e a animação do painel muda conforme o scroll | Workflow + pilares |
| 4 | **Galeria de cenários** | "Isso acontece no *meu* domínio" | Tabs com os 10 exemplos (banco, hospital, ingressos, estoque, leilão…), cada um com a invariante quebrada em uma frase | `DemoShowcase` (3 cenários) |
| 5 | **Do caos ao teste mínimo** | Diferencial técnico (ddmin) | Animação full-width: ~50 ops colapsando em 3–4, com os números reais do eval | Coluna "Redução" do ProjectCycle |
| 6 | **No seu CI** | Retenção / Cloud | Split: YAML da GitHub Action + mockup **real** (screenshot) de comentário de PR bloqueando o merge | Passo 04 do workflow |
| 7 | **Open source vs Cloud vs Auditoria** | Converter para o caminho certo | Três colunas *diferentes* (não cards iguais), com a auditoria destacada | Pilares + waitlist |
| 8 | **CTA final + FAQ** | Remover objeções | FAQ em acordeão (6 perguntas) + um CTA | Lockup + waitlist longa |

Detalhes por seção:

**1. Hero**
- No máximo 4 elementos de texto: H1, subtexto, CTA primário, CTA secundário (`Instalar` com cópia inline). Remover monograma com glow, eyebrow e card estático.
- Visual: duas swimlanes (T1/T2) executando o lost update **a partir de um trace real** exportado pela CLI com seed fixa (ex.: `examples/banking_lost_update`, `--seed 184729`). O saldo esperado (1900) e o real (1950) aparecem em contagem e a invariante "acende". Loop lento, pausável, e estático com `prefers-reduced-motion`.
- Isso respeita o princípio de determinismo do projeto: o hero **é** evidência, não ilustração.

**3. "O bug em 30 segundos"** (a seção que mais melhora as explicações)
1. *Duas pessoas leem o mesmo saldo.* (as swimlanes aparecem)
2. *Cada uma debita um valor válido.* (setas de escrita)
3. *A segunda escrita apaga a primeira. R$ 50 somem sem erro no log.* (valor pisca, invariante `total == 2000` falha)
4. *O ChaosSQL encontra essa ordem, prova a falha e gera `repro_test.go`.* (o arquivo aparece digitado)

Um glossário inline (tooltip/popover) para "isolamento", "invariante" e "seed" atende o leitor não especialista sem poluir o texto.

**4. Galeria de cenários:** usar os dados de `src/data/scenarios.json` (já existem 9–10). Cada tab mostra o domínio, a frase de dor, a invariante SQL, o trace mínimo e o botão "Abrir no Playground" com o cenário pré-carregado (hoje todos apontam só para `/playground`, sem parâmetro).

**5. ddmin:** contador real ("54 operações → 3 · 94,4% de redução · 0,41s") vindo de um run gravado no build, com o link "como medimos" para `evals/01_shrinking_ratio.md`.

**7. Planos:** a landing mostra o resumo e `/pricing` mantém o detalhe. A auditoria ganha uma lista concreta de entregáveis (ex.: "relatório com N cenários do seu schema", "PR com testes de regressão", "sessão de 1h com o time"), prazo (1 semana) e preço.

### 3.1 Navegação
- Desktop: `Como funciona` · `Cenários` · `Docs` · `Preços` · [GitHub ★ N] · [PT/EN] · botão `Ver o bug acontecer`.
- `Visualizer`, `Hermitage Matrix` e `Playground` vão para um menu "Ferramentas" ou para o footer. `Cloud Dashboard` sai da nav pública (entra como "Entrar" quando o Cloud abrir).
- As URLs não mudam (sem risco de SEO).

---

## 4. Direção visual

### 4.1 Preservar
Paleta Bregalda (`#4B2E83` roxo, `#F5C400` amarelo, `#22C55E` verde, `#FCFBF8` cream, `#2A2140` ink), monograma e JetBrains Mono. É um redesign que **preserva a marca**, não uma troca de identidade.

### 4.2 Mudar
| Item | Hoje | Proposta |
|---|---|---|
| Tipografia display/corpo | Inter via Google Fonts `<link>` | **Geist Sans** (ou Satoshi) self-hosted via `@fontsource`, com JetBrains Mono para código e números. O H1 ganha peso e tracking controlados, sem escala exagerada. |
| Papel das cores | Roxo em tudo (CTA, glow, bordas, badges) | **Roxo = marca** (logo, links, superfícies), **amarelo = sinal/anomalia** (o único acento "quente", usado só onde algo quebra), **verde = invariante ok**, vermelho só dentro dos traces. Sem glow e sem gradiente roxo. |
| Tema | Light com blocos escuros soltos | **Decidido: escuro "laboratório"** em todas as páginas de marketing (landing, pricing, cenários). A base é ink (`#141021`), coerente com os artefatos (trace, terminal, visualizer), que já são escuros. Superfícies elevadas usam `#1C1630`, o texto usa cream `#FCFBF8` e o texto secundário usa cream a 70%. Os docs seguem `prefers-color-scheme` com toggle manual. A página inteira fica num tema só, sem seções claras intercaladas. |
| Raios | 2/3/4/9/16px misturados | Regra única: controles 6px, painéis 12px, pills só em badges de status. |
| Cards | Tudo é card com borda | Cards só para artefatos (traces, código). O restante é agrupado por espaço e divisores. |
| Ícones | lucide + emojis | Só lucide (já é dependência), stroke 1.5 global, zero emojis. |
| Textura | Nenhuma | Grade de pontos muito sutil só atrás dos traces, reforçando a ideia de "osciloscópio/laboratório", em pseudo-elemento fixo. |

### 4.3 Imagens reais
- Screenshots reais do `chaossql ui` (Gantt swimlane, DSG, painel ddmin) e do comentário de PR da GitHub Action, capturados com seeds fixas. Nada de "fake screenshots" em `div`.
- Nova imagem OG por página (1200×630): headline + miniatura do trace. Gerada no build com Satori/`@vercel/og` ou estática.

---

## 5. Sistema de animação

Princípio: **toda animação explica causalidade entre transações** ou dá feedback de uma ação. Nada anima "porque fica bonito".

| Animação | Onde | O que comunica | Técnica |
|---|---|---|---|
| Trace ao vivo (swimlanes) | Hero | O bug acontecendo em ordem | Motion (`motion/react`), timeline dirigida pelos dados do trace JSON |
| Painel fixo com passos | Seção 3 | Narrativa passo a passo | `position: sticky` + `useScroll`/IntersectionObserver (sem `scroll` listener) |
| Colapso ddmin | Seção 5 | 50 ops → 3 | `layout`/`layoutId` do Motion, com as ops removidas saindo em stagger |
| Troca de cenário | Seção 4 | Mudança de estado | `AnimatePresence` + cross-fade de 200ms |
| Reveal de seções | Todas | Hierarquia na entrada | `whileInView` com `once: true`, y 16px, ease `[0.16,1,0.3,1]` |
| CTA / copiar | Botões | Feedback tátil | `:active` scale 0.98, check animado ao copiar |
| Contadores | Faixa de prova, ddmin | Números reais "chegando" | Count-up único ao entrar na viewport |

Regras técnicas:
- Animar só `transform`/`opacity`. Tudo coberto por `prefers-reduced-motion` (estado final estático).
- **Uma** biblioteca: Motion (~30 KB gz), carregada com lazy import nos componentes abaixo da dobra. Sem GSAP e sem Three.js.
- As animações de trace consomem o **mesmo formato JSON** que a CLI exporta, então dá para trocar o cenário sem reescrever a animação e fica tudo fiel ao produto.

---

## 6. Explicações: reescrita de conteúdo

1. **Guia de voz** (1 página, em `site/COPY.md`): frases curtas, verbos concretos, sem "VIP", sem "revolucionar", sem emoji, sem travessão decorativo, e números sempre com fonte.
2. **Três camadas de profundidade** em todo tópico: *uma frase de dor* → *um parágrafo de mecanismo* → *link "Teoria"* para `docs/THEORY.md` / Adya.
3. **FAQ orientada a objeções** (ex.: "Funciona com meu ORM?", "Roda contra produção?" [não; ambiente de teste], "Preciso de CGO?" [não], "Qual a diferença para testes de carga?", "E se meu banco já usa SERIALIZABLE?", "O que exatamente a auditoria entrega?").
4. **Página "Como funciona"** (pode ser uma âncora longa na landing ou `/how-it-works`) que concentra o conteúdo técnico hoje espalhado pelos pilares.
5. Paridade EN/PT total, com todas as strings em `i18n` (hoje há ternários `lang === 'pt'` espalhados). Extrair para dicionários `en.ts` (fonte da verdade) e `pt.ts` e adicionar um teste que falha se uma chave estiver faltando. **A copy é escrita primeiro em inglês** (mercado global) e depois adaptada para PT. Não é tradução literal do PT atual.

---

## 7. Marketing, SEO e medição

### 7.1 Medição (fazer primeiro)
- **Cloudflare Web Analytics** (sem cookies, sem banner e compatível com a restrição "zero tracking pesado" do plano original) para pageviews e origem.
- Eventos próprios, enviados a um endpoint do Worker (`/api/event`, com agregação em D1/Analytics Engine): `cta_hero_click`, `install_copy`, `scenario_tab_view`, `playground_open`, `waitlist_submit`, `audit_submit`, `github_click`, `pricing_toggle`.
- Painel simples com o funil: visita → interação com a demo → install/playground → lead.
- Guardar uma **baseline de 2 semanas antes** do novo layout entrar, para poder comparar.

### 7.2 SEO
- Corrigir `og:locale`, alinhar o OG com o idioma e atualizar o `softwareVersion` para a versão real (gerar a partir do `CHANGELOG` no build).
- **EN na raiz (`/`) e PT em `/pt/…`**, com `hreflang` en/pt-BR + `x-default` → `/`. Hoje o idioma padrão é PT e o idioma só muda o estado (os buscadores veem uma única versão). Detectar `Accept-Language` no Worker só para *sugerir* PT (banner discreto), nunca para redirecionar à força, porque isso prejudica o SEO.
- `og:locale=en_US` com alternate `pt_BR`, e `<html lang>` correto por rota.
- **Páginas de cenário indexáveis** (`/scenarios/lost-update`, `/scenarios/write-skew`…), cada uma respondendo a uma busca real ("lost update postgresql read committed", "write skew exemplo"). É o maior potencial de tráfego orgânico do projeto.
- FAQ com `FAQPage` JSON-LD.
- Pré-render estático (SSG) da landing e das páginas de cenário. Hoje é SPA com HTML inicial genérico, o que é ruim para LCP e para indexação.

### 7.3 Conversão
- Waitlist: **1 campo** (e-mail) + botão, com o banco pedido opcionalmente na tela de sucesso.
- Auditoria: formulário qualificado separado (banco, stack, volume de transações, prazo), com expectativa clara de resposta ("respondemos em até 2 dias úteis").
- Sticky CTA discreto no mobile depois da seção 3.
- Prova social progressiva: estrelas do GitHub agora; depois, 1–2 depoimentos reais de usuários da CLI ou da auditoria (pedir aos primeiros leads), e um estudo de caso escrito a partir de um bug real encontrado em projeto open source (é um ótimo conteúdo e gera backlinks).

### 7.4 Conteúdo de distribuição (fora do site, alimentando o site)
- Post técnico "Encontramos um lost update em [projeto OSS]" com o trace e o repro.
- GIF/vídeo curto (15s) da animação do hero para README, LinkedIn e X.
- Badge "Tested with ChaosSQL" para READMEs de usuários (backlink).

---

## 8. Performance e acessibilidade (critérios de aceite)

- **LCP < 2,0s** em 4G no mobile, **CLS < 0,05**, **INP < 200ms**. JS inicial da landing ≤ 120 KB gz (Motion e o trace em chunk lazy). O WASM (`chaossql.wasm`) nunca é carregado na landing.
- Fontes self-hosted com `font-display: swap` e preload da display font.
- WCAG AA: contraste de todos os CTAs e formulários, foco visível (já existe; manter), nav por teclado nos tabs de cenário (`role="tablist"` correto), `aria-live` no resultado do trace.
- Lighthouse ≥ 95 em Performance, Accessibility, Best Practices e SEO, verificado no CI.
- `make verify` continua passando. Conferir o impacto de `test_english_purity.js` sobre os novos dicionários de i18n.

---

## 9. Decisões tomadas (2026-09-22)

| Decisão | Escolha | Impacto no plano |
|---|---|---|
| Tema | Escuro "laboratório" | §4.2: tokens escuros como padrão; Fase 1 redefine `tokens.css` |
| Headline | "Your tests pass. Your balances don't add up." / "Seus testes passam. Seus saldos não fecham." | §2.3 |
| Mercado | **Global** | EN canônico em `/`, PT em `/pt`; copy escrita em EN primeiro; exemplos em USD; logos/prova social pensados para o público internacional |
| Auditoria | $1,490 público, 1 semana | Os entregáveis propostos em §3 (item 7) ainda precisam de confirmação antes da Fase 5 |
| Medição | Cloudflare Web Analytics + eventos próprios, sem cookies | Fase 0 |
| SSG | Aprovado | Fase 6: pré-render de landing, pricing e `/scenarios/:slug` em EN e PT |

**Ainda em aberto:** a lista final de entregáveis da auditoria (o que o cliente recebe, formato do relatório, se inclui sessão com o time) e o *site tag* do Cloudflare Web Analytics (gerado no painel da Cloudflare pela conta dona do domínio).

---

## 10. Fases de execução

Cada fase é um PR independente, com `make verify` verde e screenshot antes/depois.

### Fase 0: Fundamentos e medição (2–3 dias)
- [x] Cloudflare Web Analytics + endpoint `/api/event` + eventos da §7.1 nos CTAs atuais.
- [x] Corrigir os bugs de i18n (§1.3), o JSON-LD `softwareVersion` e o `og:locale`.
- [x] Trocar o idioma padrão para EN (sem mudar URLs ainda; as rotas `/pt` entram na Fase 6).
- [x] Remover emojis e revisar os badges de anomalia contra `SCENARIO_ACADEMIC_AUDIT.md`.
- [ ] Coletar a baseline de conversão (depende do deploy e do token do Web Analytics em `wrangler.toml`).

### Fase 1: Design system (3–4 dias)
- [x] Tokens novos em `tokens.css` (papéis de cor, regra de raios, escala tipográfica, tema `data-theme="lab"`).
- [x] Fontes self-hosted (Geist + JetBrains Mono via `@fontsource-variable`; Google Fonts removido).
- [x] Componentes base em `site/src/components/system/`: `Button` (3 variantes), `Section`/`SectionHeader`, `Terminal` (o `CodeBlock` dos docs passou a usá-lo), `Badge`, `Tabs` acessível. Preview em `npm run dev` → `/design-preview.html`.
- [x] Dicionários tipados em `site/src/i18n/` (EN canônico, PT com o mesmo tipo) + teste de paridade. Migrados nav, footer e textos comuns; os textos da landing e do pricing entram direto nos dicionários na reescrita das Fases 2 e 4.

### Fase 2: Copy (2–3 dias, em paralelo com a Fase 1)
- [ ] `site/COPY.md` (guia de voz).
- [ ] Texto final PT/EN de todas as seções da §3 + FAQ + planos.
- [ ] Revisão técnica das afirmações e números contra `evals/`.

### Observações da Fase 1
- A navegação atual quebra em duas linhas no desktop (8 itens); resolver na Fase 4 junto com a nova nav (§3.1).
- O bundle JS continua em ~179 KB gzip; tratar na Fase 6.

### Fase 3: Assets reais (2 dias)
- [ ] Exportar traces JSON com seeds fixas dos 10 exemplos (script em `tools/` para regenerar).
- [ ] Screenshots do `chaossql ui` e do comentário de PR da Action.
- [ ] OG images por página.

### Fase 4: Landing nova (5–7 dias)
- [ ] Hero com `TraceAnimation` (dados reais).
- [ ] Faixa de prova (estrelas do GitHub buscadas no build).
- [ ] Scroll-story "O bug em 30 segundos".
- [ ] Galeria de cenários com deep link para o Playground (`/playground?scenario=…`).
- [ ] Animação ddmin.
- [ ] Seção CI, comparação de planos, FAQ, CTA final.
- [ ] Nav e footer novos.
- [ ] Remover `ProjectCycle`, pilares, workflow e o lockup da landing.

### Fase 5: Pricing e formulários (2–3 dias)
- [ ] `/pricing` com o novo design, CTAs unificados e auditoria com entregáveis.
- [ ] Waitlist com 1 campo; formulário de auditoria qualificado.
- [ ] Estados de loading, erro e sucesso revisados.

### Fase 6: SEO e performance (3–4 dias)
- [ ] SSG da landing, do pricing e das páginas de cenário.
- [ ] EN na raiz, rotas `/pt/…`, `hreflang` + `x-default`, redirects das URLs antigas se necessário.
- [ ] Páginas `/scenarios/:slug` indexáveis + sitemap atualizado.
- [ ] Lighthouse CI com os budgets da §8.

### Fase 7: Lançamento e iteração (contínuo)
- [ ] Comparar o funil com a baseline após 2–4 semanas.
- [ ] Teste A/B da headline (2 variantes) via Worker, se o volume permitir.
- [ ] Post técnico + vídeo do hero para distribuição.

**Estimativa total:** cerca de 4 semanas de trabalho focado (Fases 0–6).

### Métricas de sucesso
| Métrica | Como medir | Meta inicial |
|---|---|---|
| Interação com a demo do hero/cenários | `scenario_tab_view` ou `cta_hero_click` / visitas | > 35% |
| Cópia do comando de instalação | `install_copy` / visitas | +50% vs baseline |
| Conversão da waitlist | `waitlist_submit` / visitas | +2× vs baseline |
| Leads de auditoria | `audit_submit` / mês | baseline → meta definida após a Fase 0 |
| Tráfego orgânico | Visitas em `/scenarios/*` | crescimento mês a mês |
| Core Web Vitals | CrUX / Lighthouse | todos "Good" |
