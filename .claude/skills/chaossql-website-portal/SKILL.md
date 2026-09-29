---
name: chaossql-website-portal
description: The chaossql.bregalda.com portal in site/ — React 19 + Vite app with path routing, pages (landing, docs, scenarios, visualizer, matrix, playground, pricing, dashboard), bilingual pt/en i18n, the Cloud dashboard API client, the committed build output, page metadata from one shared module, English/Portuguese URLs, build-time prerendering, legacy vanilla files still used by tools, and frontend tests. Use when changing anything under site/ (except the WASM engine itself).
---

# Website Portal (`site/`)

The portal is the only place where Portuguese is allowed (it is bilingual;
English is canonical); the English purity gate skips `site/`.

## When to use

- Changing pages, components, styles, data files, routing, SEO metadata, the
  dashboard, or the build.

## Structure

- Entry: `site/template.html` → `src/main.tsx` → `src/App.tsx` (nav, page by
  route, footer). The landing is in the main bundle; every other page is a
  `React.lazy` chunk behind a `Suspense` fallback. `LAB_ROUTES` in `App.tsx`
  lists the routes rendered in the dark lab theme (every route): App
  sets `data-theme="lab"` on `<html>` and the `theme-color` meta for them. The dev server (`vite.config.ts`, port 3000) serves
  `template.html` for every extension-less app route, so deep links in dev
  never load the committed prebuilt bundle. `design-preview.html`
  (`src/dev/design-preview.tsx`) is a dev-only styleguide of the lab design
  system; it is not a build input and is excluded by `.assetsignore`.
- Routing (`src/lib/router.ts`): real paths `/`, `/dashboard`, `/docs`,
  `/scenarios`, `/scenarios/<slug>`, `/visualizer`, `/matrix`, `/playground`,
  `/pricing`, each also under `/pt` (Portuguese). `splitLocale` and
  `localizePath` (from `src/lib/seo.ts`) strip or add the `/pt` prefix;
  `routeFromPath` uses the first segment after it; unprefixed internal links
  clicked on a Portuguese page stay in Portuguese unless the anchor has
  `data-lang`. Scenario slugs are the public anomaly names in `SCENARIO_SLUGS`
  (`src/data/scenarios-data.ts`; changing one breaks indexed URLs); legacy `#/docs` hashes are migrated
  with `history.replaceState`; internal link clicks are intercepted
  (`interceptLinkClick`, `navigate`). Hashes are kept: `navigate('/#story')`
  from another page changes route and App scrolls to the anchor after render
  (`scrollToHash`); on the same page it only scrolls. `/playground?scenario=<id>`
  opens that preset (ids from `PLAYGROUND_PRESETS`).
- Pages in `src/pages/*Page.tsx` with CSS modules; shared components in
  `src/components/{ui,docs,artifacts}`; design tokens in
  `src/styles/tokens.css` (Studio Bregalda identity; brand SVGs in
  `site/brand`, `site/public/brand`, `site/assets`).
- Landing (`src/pages/LandingPage.tsx`, lab theme): hero with the autoplaying
  `TracePanel` (pause button, stops off screen, final state under reduced
  motion), proof strip (GitHub stars shown only from 50 up), `StoryScroll`
  (`#story`, sticky panel driven by an IntersectionObserver band), scenario
  `Tabs` (`#scenarios`), `ShrinkViz` + report crop, CI section with the
  `action.yml` workflow, plans (`#plans`), FAQ (`<details>`), final CTA with
  `WaitlistForm` (`#waitlist`). Components in `src/components/landing/`;
  animation is CSS plus IntersectionObserver, no animation library.
- Every page now uses the semantic tokens; the legacy palette tokens
  (`--cream`, `--ink`, `--purple`, `--border-subtle`, `--text-secondary`)
  remain in `tokens.css` only for the vanilla files in `site/app.js` and
  `site/assets/style.css`. Do not use them in new CSS: `--cream` meant both a
  background and a text color, which is why each page was migrated by
  property.
- Shared chrome (`SiteNav`, `SiteFooter`) uses only semantic tokens, which
  have light values in `:root`, so it renders in both themes. Nav links:
  How it works, Scenarios, Playground, Docs, Pricing; the visualizer, matrix
  and dashboard are linked from the footer.
- Lab design system (marketing redesign, plan in
  `docs/superpowers/plans/2026-09-22-marketing-site-redesign.md`): pages opt in
  with `data-theme="lab"`, which defines semantic tokens (`--surface-*`,
  `--text-*`, `--line*`, `--brand-text`, `--signal`, `--ok`, `--danger`), one
  radius rule (`--r-control`, `--r-panel`, `--r-pill`), type, layout and motion
  tokens. Components in `src/components/system/`: `Button`, `Badge`,
  `Section`/`SectionHeader`, `Tabs` (WAI-ARIA, roving tabindex) and `Terminal`
  (the single code surface; docs `CodeBlock` renders it). Fonts are
  self-hosted Geist and JetBrains Mono (`@fontsource-variable`, imported in
  `main.tsx`); there is no Google Fonts request.
- Content data: `src/data/docs.json` + `docs-content.ts`,
  `src/data/scenarios.json` + `scenarios-data.ts` (typed from the JSON
  without a cast).
- Docs (`src/pages/DocsPage.tsx`, lab theme): `/docs` is the chapter index,
  `/docs/<chapter>` one page per chapter (`CHAPTER_ORDER` ids; old
  `?chapter=` links are replaced with the chapter URL). `docs-content.ts`
  prepares each chapter once: `renderTexInHtml` (`src/lib/tex.ts`) turns the
  `$...$` / `$$...$$` LaTeX subset into HTML with Unicode symbols and
  sub/superscripts, outside `<pre>`/`<code>` only (it also accepts the doubled
  backslashes some Portuguese entries have), and every `.code-container` gets
  one `.copy-code-btn` handled by a delegated click in `DocsContent` (the
  legacy inline `onclick` handlers are stripped). Chapter HTML is styled with
  `:global` rules in `DocsContent.module.css`.
- Docs copy follows `site/COPY.md` like the rest of the site and must match
  the code: `run` exits 1 on a finding (`--fail-on`) and 2 on any error; the
  flag table lists only the flags registered in `cmd/chaossql/main.go`; the
  engine uses seeded jitter (PCT-style scheduling is proxy only); the shrinker
  is memoized ddmin plus a 1-minimality audit on the failure signature, with
  no causal closure. `src/data/docs-copy.test.ts` guards dashes, removed
  claims and flags that do not exist; when a CLI flag or exit code changes,
  update `docs.json` (en and pt) in the same commit.
- Copy (`src/i18n/`): `en.ts` is canonical, `pt.ts` is typed as `Messages`
  so a missing or extra key fails `tsc`; `format(template, values)` fills
  `{name}` placeholders and throws on a missing value. `src/i18n/i18n.test.ts`
  checks key parity, identical placeholders in both languages, no empty
  strings and no em/en dashes. Voice rules and the claims register (the source
  of every number or promise in the copy) are in `site/COPY.md`.
- Recorded runs (`src/data/traces/*.json`, typed by `src/data/traces.ts`):
  real engine output for `banking_lost_update`, `inventory_oversell` and
  `hospital_write_skew` on SQLite `READ_UNCOMMITTED`, regenerated with
  `node tools/export_site_traces.mjs [cli]`. `src/data/story.ts` derives the
  landing's lost update numbers from the banking trace, so the copy cannot
  drift from what the engine did. Real report crops live in `src/media/`.
- i18n (`src/lib/i18n.ts`): the URL decides the language (English at the
  root, Portuguese under `/pt`), so every page is indexable in both and the
  prerendered HTML always matches. `setLang` navigates to the other language's
  URL and stores the choice in `localStorage["chaossql_lang"]`.
  `LanguageHint` only *suggests* `/pt` to browsers that prefer Portuguese and
  never chose a language; there is no automatic redirect.
- Analytics (`src/lib/analytics.ts`): `track(event, label)` sends cookieless
  events (`cta_click`, `install_copy`, `command_copy`, `scenario_view`,
  `lead_submit`, `plan_select`, `pricing_toggle`, `outbound_click`,
  `lang_switch`) to `/api/event` with `navigator.sendBeacon`, falling back to
  `fetch(..., keepalive)`. The event list must match `SITE_EVENTS` in
  `worker.ts` (`chaossql-edge-worker`).
- Playground: `src/lib/wasm-bridge.ts` → WASM worker (`chaossql-wasm-playground`).
- Dashboard (`DashboardPage.tsx` + `src/lib/cloud-api.ts`): the user enters an
  API base URL and token; `CloudAPI` calls `/v1/runs`, `/v1/runs/{id}`,
  `/v1/organizations/me/webhooks` (create uses events `regression,failure`),
  and member-token issuance, with `cache: no-store`, `credentials: omit`,
  `redirect: error`.
- Lead forms (`src/components/forms/`) post to `/api/waitlist` through
  `src/lib/lead-request.ts` (edge worker): `WaitlistForm` asks only for an
  email (landing `#waitlist`; on pricing it adds a plan select that the plan
  buttons preselect, plus the billing cycle); `AuditForm` (pricing `#audit`)
  requires name, work email and company and sends database, timeline, notes,
  `wantAudit: true` and a plan containing "Audit". Both show success only
  after the lead is acknowledged.
- Pricing (`src/pages/PricingPage.tsx`, lab theme): open source row, Cloud
  plans from `CLOUD_PLANS` with a monthly/annual switch, shared Cloud
  features, Enterprise on request, the audit offer with its form in place, and
  the waitlist. No modal.

## Page metadata (one source)

`src/lib/seo.ts` is a pure module (no DOM, no React) with `SEO_ROUTES`
(path, indexable, OG image, English and Portuguese title/description) and
`pageMeta(route, lang, path, override?)`, which returns the title,
description, robots, canonical, `og:locale` and reciprocal hreflang
alternates (`en`, `pt-BR`, `x-default` → English). Three consumers import it:
`src/lib/route-meta.ts` (`applyRouteMeta` in the browser). Scenario and
docs chapter titles come from overrides registered by `src/lib/detail-meta.ts`,
which only the lazy `ScenariosPage`/`DocsPage` modules and `entry-server.tsx`
import; those pages re-apply the metadata once loaded. `route-meta.ts` must
never import content data: it is in the main bundle, `scripts/prerender.mjs` (through the SSR bundle) and
`worker.ts` (bundled into `site/_worker.js`). A new route needs an entry in
`SEO_ROUTES`, `ROUTE_PATHS`, the `run_worker_first` list in `wrangler.toml`
and an OG card; the sitemap is regenerated by the build. `router.test.ts`
covers paths, alternates, slugs, sitemap and OG cards.

## Prerendering (SSG)

`npm run build` also builds `src/entry-server.tsx` for Node and runs
`scripts/prerender.mjs`, which renders `/`, `/pricing`, `/scenarios`,
`/docs`, `/matrix`, every `/scenarios/<slug>` and every `/docs/<chapter>` in
both languages with React
`prerender` (waits for lazy pages; `progressiveChunkSize: Infinity` keeps
every Suspense boundary inline) and writes `site/prerender/<path>.html`
(`/` → `home.html`) plus `site/prerender/manifest.json`, `sitemap.xml` and
`public/sitemap.xml`. Each file gets its language's `<head>` from `pageMeta`,
`data-theme="lab"` for `LAB_ROUTES`, font preloads, the lazy page's CSS and
`modulepreload` (from `dist/.vite/manifest.json`) and, on the landing,
`FAQPage` JSON-LD. `main.tsx` hydrates when `#root` has `data-prerendered`,
otherwise renders from scratch. The edge worker serves these files at the
canonical URLs; `robots.txt` disallows `/prerender/`.

`index.html` stays the plain shell on purpose: the dashboard image (nginx,
`Dockerfile.dashboard`) and the Go `spaFileServer` of the managed/self-hosted
server use it as the fallback for every route. Visualizer, playground and
dashboard are never prerendered (they need browser APIs at render).

Hydration rules: anything that differs between server and browser must be
read in an effect (see `src/components/landing/hooks.ts`: reduced motion and
GitHub stars start as `false`/`null`); `TracePanel` starts on its final state
so the prerendered hero already shows the failing invariant.

## Reading the funnel

`node tools/site_funnel.mjs [days] [--json]` (needs `CLOUDFLARE_ACCOUNT_ID`
and a `CLOUDFLARE_API_TOKEN` with Account Analytics: Read; optional
`CF_WEB_ANALYTICS_SITE_TAG` for page views) queries the `chaossql_site_events`
Analytics Engine dataset and prints the stages: story CTA, scenario tabs,
install copies, playground opens, plan selections, waitlist and audit leads,
plus events per language and top referrers. `STAGES` matches the labels the
site sends with `track()`; add a stage when a new CTA is instrumented.
`tools/test_site_funnel.mjs` covers the label mapping and API calls.

## README animation

`npm run gif` (`scripts/render-trace-gif.mjs`, dev dependencies `gifenc` and
`pngjs`, shared helpers in `scripts/render-lib.mjs`) renders
`docs/media/lost-update.gif` from the recorded banking trace, one frame per
statement, for the README and social posts. Re-render when the trace changes.

## Open Graph cards

`npm run og` (`scripts/render-og.mjs`, dev dependency `playwright-core`)
renders `site/og/{home,scenarios,docs,playground,pricing}.png` at 1200x630
with the lab palette and a swimlane drawn from the recorded banking trace. It
needs a Chromium binary: the newest Playwright headless shell in
`~/.cache/ms-playwright`, or `CHROMIUM_PATH`. Fonts are inlined as data URLs
(pages loaded with `setContent` cannot fetch `file://` fonts). Re-render and
commit the PNGs when a card title or the banking trace changes.

## Build output is committed

`npm run build` = client build, copy `index.html` and assets, SSR build into
`dist-ssr/`, `scripts/prerender.mjs`, remove `dist*`,
`scripts/prune-assets.mjs`, `scripts/build-pages-worker.mjs` (writes
`site/_worker.js` from `worker.ts`; skipped when `../worker.ts` is absent, as
in the Docker builds that copy only `site/`).
The deployed site is the `site/` directory itself (Cloudflare Pages
`pages deploy site`, GitHub Pages fallback, Workers assets `./site`), so after
source changes you must rebuild and commit `site/index.html` and the hashed
files in `site/assets/` (main bundle, one lazy chunk per page, fonts, images).
Commit `site/prerender/`, both sitemaps and `site/_worker.js` too.
`scripts/prune-assets.mjs` deletes hashed files that `index.html` and the
prerendered pages no longer reach (following imports through JS and CSS);
stage those deletions too.
Unhashed files (`chaossql.wasm`, `style.css`, `wasm-*.js`) are never pruned.

## Legacy files still in use

`site/app.js`, `site/docs-data.js`, `site/assets/style.css` (the pre-React
vanilla portal) are required by `make check-harness` and exercised by
`tools/test_playground_ui.js`, `tools/audit_i18n.js`,
`tools/e2e_senior_qa_audit.js`. `site_legacy/` is an older snapshot.

## Tests and commands

- `make test-frontend` → `cd site && npm ci && npm run verify`
  (typecheck, build into `.verify-dist`, `vitest run`).
- Tests: `src/lib/router.test.ts`, `src/lib/cloud-api.test.ts`,
  `src/lib/analytics.test.ts`, `src/components/system/system.test.tsx`,
  `src/i18n/i18n.test.ts`, `src/data/traces.test.ts`, `src/data/story.test.ts`,
  `src/pages/LandingPage.test.tsx`,
  `src/pages/CloudFlows.test.tsx`; plus `node tools/test_playground_ui.js`.
- `src/lib/bundle.test.ts` keeps the committed main bundle under 120 KB gzip
  and free of docs/scenario content (a content import from shared code once
  added 43 KB and failed the Lighthouse budget).
- `src/lib/prerender.test.ts` checks the committed prerendered pages
  (coverage in both languages, `lang`, canonical, H1, no outlined Suspense
  boundaries, lab theme, lazy CSS links, FAQ JSON-LD).
- Local: `cd site && npm run dev`; `npm run preview:static` serves the
  committed build the way the worker routes it (`scripts/serve-static.mjs`,
  gzip included); `make serve-site` serves `site/` as plain files on 8080.
- Lighthouse CI (`.github/workflows/lighthouse.yml`, `site/lighthouserc.json`)
  asserts performance >= 90, accessibility/best practices/SEO >= 95 and
  CLS <= 0.1 on `/`, `/pt`, `/pricing` and `/scenarios/lost-update`.

## Gotchas

- `site/package.json` version is not the product version.
- Security headers and immutable caching for `/assets/*`, `/wasm/*`, `/brand/*`
  come from `site/_headers` (Pages) — hashed filenames are required for safe caching.
- Pricing on `PricingPage.tsx` duplicates `internal/server/billing.go`.
- Only three examples are exported as recorded runs on purpose:
  `read_skew_financial_audit` flips between P4 and A5A across runs and
  `ticket_booking_anti_dependency` also fails on serial histories. Durations
  and full-run invariant values vary between regenerations; schedules, shrink
  results and minimal traces do not.
- A violation fails the GitHub Action step (`fail-on`, default `violation`),
  but it only blocks a merge when the check is required in branch protection;
  PR comments and `is-regression` still require Cloud (see `site/COPY.md`).

## Source map

- `site/template.html`
- `site/vite.config.ts`
- `site/package.json`
- `site/src/App.tsx`
- `site/src/main.tsx`
- `site/src/lib/router.ts`
- `site/src/lib/route-meta.ts`
- `site/src/lib/seo.ts`
- `site/src/lib/tex.ts`
- `site/src/lib/detail-meta.ts`
- `site/src/lib/bundle.test.ts`
- `site/src/data/docs-content.ts`
- `site/src/pages/DocsPage.tsx`
- `site/src/components/docs/DocsContent.tsx`
- `site/src/components/docs/DocsSidebar.tsx`
- `site/src/entry-server.tsx`
- `site/scripts/prerender.mjs`
- `site/scripts/serve-static.mjs`
- `site/scripts/build-pages-worker.mjs`
- `site/src/components/ui/LanguageHint.tsx`
- `site/src/pages/ScenariosPage.tsx`
- `site/src/data/scenarios-data.ts`
- `site/src/lib/prerender.test.ts`
- `site/prerender/manifest.json`
- `site/lighthouserc.json`
- `.github/workflows/lighthouse.yml`
- `site/src/lib/i18n.ts`
- `site/src/lib/analytics.ts`
- `site/src/lib/cloud-api.ts`
- `site/src/lib/lead-request.ts`
- `site/src/pages/DashboardPage.tsx`
- `site/src/pages/PricingPage.tsx`
- `site/src/data/docs.json`
- `site/src/data/docs-copy.test.ts`
- `site/src/data/scenarios.json`
- `site/src/styles/tokens.css`
- `site/src/pages/LandingPage.tsx`
- `site/src/components/landing/TracePanel.tsx`
- `site/src/components/landing/StoryScroll.tsx`
- `site/src/components/landing/ShrinkViz.tsx`
- `site/src/components/forms/WaitlistForm.tsx`
- `site/src/components/forms/AuditForm.tsx`
- `site/src/components/landing/InstallButton.tsx`
- `site/src/components/ui/SiteNav.tsx`
- `site/src/components/ui/SiteFooter.tsx`
- `site/src/pages/PlaygroundPage.tsx`
- `site/src/components/system/Tabs.tsx`
- `site/src/components/system/Terminal.tsx`
- `site/src/components/system/Button.tsx`
- `site/src/i18n/en.ts`
- `site/src/i18n/pt.ts`
- `site/src/i18n/index.ts`
- `site/src/data/traces.ts`
- `site/src/data/story.ts`
- `site/scripts/render-og.mjs`
- `site/scripts/prune-assets.mjs`
- `site/design-preview.html`
- `site/COPY.md`
- `tools/export_site_traces.mjs`
- `tools/site_funnel.mjs`
- `site/scripts/render-trace-gif.mjs`
- `site/scripts/render-lib.mjs`
- `docs/media/lost-update.gif`
- `site/src/lib/router.test.ts`
- `site/index.html`
- `site/_headers`
- `site/sitemap.xml`
- `site/app.js`
- `specs/12_documentation_portal_and_branding_site.md`

## Related skills

- `chaossql-edge-worker`, `chaossql-wasm-playground`,
  `chaossql-control-plane-api`, `chaossql-plans-retention`, `chaossql-quality-gate`
