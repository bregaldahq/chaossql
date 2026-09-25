---
name: chaossql-website-portal
description: The chaossql.bregalda.com portal in site/ — React 19 + Vite app with path routing, pages (landing, docs, scenarios, visualizer, matrix, playground, pricing, dashboard), bilingual pt/en i18n, the Cloud dashboard API client, the committed build output, route metadata kept in sync across four places, legacy vanilla files still used by tools, and frontend tests. Use when changing anything under site/ (except the WASM engine itself).
---

# Website Portal (`site/`)

The portal is the only place where Portuguese is allowed (it is bilingual
and defaults to `pt`); the English purity gate skips `site/`.

## When to use

- Changing pages, components, styles, data files, routing, SEO metadata, the
  dashboard, or the build.

## Structure

- Entry: `site/template.html` → `src/main.tsx` → `src/App.tsx` (nav, page by
  route, footer). Dev server rewrites `/` to `template.html` (`vite.config.ts`,
  port 3000).
- Routing (`src/lib/router.ts`): real paths `/`, `/dashboard`, `/docs`,
  `/scenarios`, `/visualizer`, `/matrix`, `/playground`, `/pricing`;
  `routeFromPath` uses the first segment; legacy `#/docs` hashes are migrated
  with `history.replaceState`; internal link clicks are intercepted
  (`interceptLinkClick`, `navigate`).
- Pages in `src/pages/*Page.tsx` with CSS modules; shared components in
  `src/components/{ui,docs,artifacts}`; design tokens in
  `src/styles/tokens.css` (Studio Bregalda identity; brand SVGs in
  `site/brand`, `site/public/brand`, `site/assets`).
- Content data: `src/data/docs.json` + `docs-content.ts`,
  `src/data/scenarios.json` + `scenarios-data.ts`.
- i18n (`src/lib/i18n.ts`): `pt | en`, default `pt`, persisted in
  `localStorage["chaossql_lang"]`, broadcast with a `languagechange` event.
- Playground: `src/lib/wasm-bridge.ts` → WASM worker (`chaossql-wasm-playground`).
- Dashboard (`DashboardPage.tsx` + `src/lib/cloud-api.ts`): the user enters an
  API base URL and token; `CloudAPI` calls `/v1/runs`, `/v1/runs/{id}`,
  `/v1/organizations/me/webhooks` (create uses events `regression,failure`),
  and member-token issuance, with `cache: no-store`, `credentials: omit`,
  `redirect: error`.
- Waitlist/contact forms post to `/api/waitlist` (`src/lib/lead-request.ts`,
  handled by the edge worker).

## Route metadata (keep four places in sync)

Per-route title/description/indexable live in `src/lib/route-meta.ts`
(client-side `applyRouteMeta`), `worker.ts` and `site/_worker.js` (server-side
injection into the app shell), plus `site/sitemap.xml`, `site/public/sitemap.xml`
and the `run_worker_first` list in `wrangler.toml`.
`src/lib/router.test.ts` imports all of them and fails on drift.

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
  `src/pages/CloudFlows.test.tsx`; plus `node tools/test_playground_ui.js`.
- Local: `cd site && npm run dev`; `make serve-site` serves `site/` statically on 8080.

## Gotchas

- `site/package.json` version is not the product version.
- Security headers and immutable caching for `/assets/*`, `/wasm/*`, `/brand/*`
  come from `site/_headers` (Pages) — hashed filenames are required for safe caching.
- Pricing on `PricingPage.tsx` duplicates `internal/server/billing.go`.

## Source map

- `site/template.html`
- `site/vite.config.ts`
- `site/package.json`
- `site/src/App.tsx`
- `site/src/main.tsx`
- `site/src/lib/router.ts`
- `site/src/lib/route-meta.ts`
- `site/src/lib/i18n.ts`
- `site/src/lib/cloud-api.ts`
- `site/src/lib/lead-request.ts`
- `site/src/pages/DashboardPage.tsx`
- `site/src/pages/PricingPage.tsx`
- `site/src/data/docs.json`
- `site/src/data/scenarios.json`
- `site/src/styles/tokens.css`
- `site/src/lib/router.test.ts`
- `site/index.html`
- `site/_headers`
- `site/sitemap.xml`
- `site/app.js`
- `specs/12_documentation_portal_and_branding_site.md`

## Related skills

- `chaossql-edge-worker`, `chaossql-wasm-playground`,
  `chaossql-control-plane-api`, `chaossql-plans-retention`, `chaossql-quality-gate`
