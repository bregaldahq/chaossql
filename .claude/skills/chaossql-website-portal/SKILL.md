---
name: chaossql-website-portal
description: The chaossql.bregalda.com portal in site/ — React 19 + Vite app with path routing, pages (landing, docs, scenarios, visualizer, matrix, playground, pricing, dashboard), bilingual pt/en i18n, the Cloud dashboard API client, the committed build output, route metadata kept in sync across four places, legacy vanilla files still used by tools, and frontend tests. Use when changing anything under site/ (except the WASM engine itself).
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
  lists the routes rendered in the dark lab theme (today only `landing`): App
  sets `data-theme="lab"` on `<html>` and the `theme-color` meta for them. The dev server (`vite.config.ts`, port 3000) serves
  `template.html` for every extension-less app route, so deep links in dev
  never load the committed prebuilt bundle. `design-preview.html`
  (`src/dev/design-preview.tsx`) is a dev-only styleguide of the lab design
  system; it is not a build input and is excluded by `.assetsignore`.
- Routing (`src/lib/router.ts`): real paths `/`, `/dashboard`, `/docs`,
  `/scenarios`, `/visualizer`, `/matrix`, `/playground`, `/pricing`;
  `routeFromPath` uses the first segment; legacy `#/docs` hashes are migrated
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
  `src/data/scenarios.json` + `scenarios-data.ts`.
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
- i18n (`src/lib/i18n.ts`): `pt | en`. A stored choice in
  `localStorage["chaossql_lang"]` wins; otherwise `pt` when any browser
  language starts with `pt`, else `DEFAULT_LANGUAGE = 'en'`. Changes are
  broadcast with a `languagechange` event, synced across tabs via `storage`,
  and set `<html lang>` (`pt-BR` / `en`).
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
- Waitlist/contact forms post to `/api/waitlist` (`src/lib/lead-request.ts`,
  handled by the edge worker, which requires `name` and `email`); the landing
  uses `WaitlistForm`, the pricing page its own modal.

## Route metadata (keep four places in sync)

Per-route title/description/indexable/image live in `src/lib/route-meta.ts`
(client-side `applyRouteMeta`), `worker.ts` and `site/_worker.js` (server-side
injection into the app shell), plus `site/sitemap.xml`, `site/public/sitemap.xml`
and the `run_worker_first` list in `wrangler.toml`.
`src/lib/router.test.ts` imports all of them and fails on drift, including a
missing `site/og/<name>.png` for any route `image`.

## Open Graph cards

`npm run og` (`scripts/render-og.mjs`, dev dependency `playwright-core`)
renders `site/og/{home,scenarios,docs,playground,pricing}.png` at 1200x630
with the lab palette and a swimlane drawn from the recorded banking trace. It
needs a Chromium binary: the newest Playwright headless shell in
`~/.cache/ms-playwright`, or `CHROMIUM_PATH`. Fonts are inlined as data URLs
(pages loaded with `setContent` cannot fetch `file://` fonts). Re-render and
commit the PNGs when a card title or the banking trace changes.

## Build output is committed

`npm run build` = `tsc --noEmit && vite build && cp dist/template.html index.html && cp -rf dist/assets/* assets/ && rm -rf dist`.
The deployed site is the `site/` directory itself (Cloudflare Pages
`pages deploy site`, GitHub Pages fallback, Workers assets `./site`), so after
source changes you must rebuild and commit `site/index.html` and the new
hashed `site/assets/main-*.js|css` — and delete the previous hashed bundles
(the copy step never removes them).

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
- Local: `cd site && npm run dev`; `make serve-site` serves `site/` statically on 8080.

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
- The GitHub Action does not fail a job on a violation, and PR comments and
  `is-regression` require Cloud: marketing copy must not claim that CI blocks
  merges by itself (see `site/COPY.md`).

## Source map

- `site/template.html`
- `site/vite.config.ts`
- `site/package.json`
- `site/src/App.tsx`
- `site/src/main.tsx`
- `site/src/lib/router.ts`
- `site/src/lib/route-meta.ts`
- `site/src/lib/i18n.ts`
- `site/src/lib/analytics.ts`
- `site/src/lib/cloud-api.ts`
- `site/src/lib/lead-request.ts`
- `site/src/pages/DashboardPage.tsx`
- `site/src/pages/PricingPage.tsx`
- `site/src/data/docs.json`
- `site/src/data/scenarios.json`
- `site/src/styles/tokens.css`
- `site/src/pages/LandingPage.tsx`
- `site/src/components/landing/TracePanel.tsx`
- `site/src/components/landing/StoryScroll.tsx`
- `site/src/components/landing/ShrinkViz.tsx`
- `site/src/components/landing/WaitlistForm.tsx`
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
- `site/design-preview.html`
- `site/COPY.md`
- `tools/export_site_traces.mjs`
- `site/src/lib/router.test.ts`
- `site/index.html`
- `site/_headers`
- `site/sitemap.xml`
- `site/app.js`
- `specs/12_documentation_portal_and_branding_site.md`

## Related skills

- `chaossql-edge-worker`, `chaossql-wasm-playground`,
  `chaossql-control-plane-api`, `chaossql-plans-retention`, `chaossql-quality-gate`
