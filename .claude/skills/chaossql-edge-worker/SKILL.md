---
name: chaossql-edge-worker
description: The Cloudflare edge code for the portal — worker.ts (Workers deploy via root wrangler.toml), its compiled twin site/_worker.js (Pages advanced mode), and the Pages Function waitlist.ts duplicated in functions/ and site/functions/ — covering /api/waitlist lead capture, /api/event cookieless analytics and the Web Analytics beacon, the /api/webhooks/test relay with host allow-list, origin checks, best-effort rate limits, secrets, and server-side route metadata injection. Use when changing any of these files or portal deployment.
---

# Edge Worker and Waitlist Function

## When to use

- Changing `worker.ts`, `site/_worker.js`, `functions/api/waitlist.ts`,
  `site/functions/api/waitlist.ts`, `wrangler.toml`, or `site/wrangler.toml`.
- Lead capture or webhook test from the portal fails.

## Four copies of the same logic

| File | Runtime |
| :--- | :--- |
| `worker.ts` | Cloudflare Worker (root `wrangler.toml`: `main = ./worker.ts`, assets `./site` bound as `ASSETS`, SPA fallback, `run_worker_first` for section URLs) |
| `site/_worker.js` | generated bundle of `worker.ts` (and the `site/src/lib/seo.ts` it imports) for Cloudflare Pages advanced mode (`pages deploy site`), written by `site/scripts/build-pages-worker.mjs` at the end of `npm run build`; never edit it by hand |
| `functions/api/waitlist.ts`, `site/functions/api/waitlist.ts` | identical Pages Function versions of the waitlist endpoint |

Edit `worker.ts` and rebuild the site (or run
`node site/scripts/build-pages-worker.mjs`); `site/src/lib/router.test.ts`
fails when the committed `site/_worker.js` differs from a fresh build. The two
Pages Functions are still hand-kept copies of the waitlist endpoint.
`tools/test_waitlist.mjs` bundles and exercises all four with esbuild.

## Endpoints (`worker.ts`)

- `POST /api/waitlist` (+ `OPTIONS` preflight): origin must be in
  `ALLOWED_ORIGINS` (`https://chaossql.bregalda.com`, `localhost:5173`,
  `127.0.0.1:5173`) → 403 otherwise; rate limit 5/min per client key → 429;
  requires a valid `email` → 400; `name` is optional (the landing waitlist is
  email only) but must be a string when sent → 400. Optional fields: `company`,
  `database`, `plan`, `billingCycle`, `timeline`, `notes`, `wantAudit`,
  `source`. Free text is truncated (name/company/timeline 128, database 64,
  notes 1000) because Discord rejects embed field values over 1024
  characters, which would drop the lead. Forwards the lead to
  `env.DISCORD_WEBHOOK_URL || env.WAITLIST_WEBHOOK_URL`, and fails loudly
  when neither secret is configured (no hard-coded fallback).
- `POST /api/webhooks/test`: same origin check, 10/min limit; the URL must be
  https, without credentials, default port, and its host in
  `ALLOWED_WEBHOOK_HOSTS` (Discord variants, `hooks.slack.com`); sends a
  Discord/Slack/generic test message and returns `{success, status}`.
- Page URLs (`appPath`): `/`, every section path, `/scenarios/<slug>`, each
  also under `/pt`. Trailing slashes redirect (301) to the canonical path.
  If the path is in `/prerender/manifest.json` (read once per isolate), the
  worker serves `site/prerender/<path>.html` (`/` → `/prerender/home`) as is;
  otherwise `serveAppShell` fetches the `index.html` shell and injects
  `<html lang>`, title, description, robots, canonical, Open Graph/Twitter
  tags, `og:locale` and hreflang alternates from `pageMeta` in
  `site/src/lib/seo.ts` (the module the app uses). An unknown scenario slug
  gets the shell with `noindex` and status 404. The Web Analytics beacon is
  appended to both (below).
- `POST /api/event` (+ `OPTIONS`): cookieless site events. Origin must be
  allowed (403); body ≤ 1024 bytes (413); `event` must be in `SITE_EVENTS`
  (keep in sync with `site/src/lib/analytics.ts`), `label`/`path`/`ref` must
  match their patterns, `lang` is coerced to `pt|en` (400 otherwise). Over the
  60/min limit it silently answers 204. Writes one Analytics Engine data
  point `{indexes: [event], blobs: [event, label, path, lang, ref, country], doubles: [1]}`
  to the `SITE_EVENTS` binding (dataset `chaossql_site_events`); no IP, user
  agent or identifiers are stored. Always 204 for well-formed requests.
- `/` → `serveLanding` (static landing + beacon); 503 on asset failure.
- Everything else → static assets.

`withWebAnalytics` appends the Cloudflare Web Analytics beacon
(`static.cloudflareinsights.com/beacon.min.js`, `spa: true`) to HTML responses
of `/` and the section shells when `CF_WEB_ANALYTICS_TOKEN` is 32 lowercase
hex characters; otherwise the page is untouched.

## Secrets and config

- `DISCORD_WEBHOOK_URL` must be an encrypted secret
  (`npx wrangler secret put DISCORD_WEBHOOK_URL`), never a `[vars]` entry.
- `CF_WEB_ANALYTICS_TOKEN` is a public site token set in `[vars]` of the
  root `wrangler.toml` (it ships in page HTML by design).
- `[[analytics_engine_datasets]]` binds `SITE_EVENTS`; query it with the
  Analytics Engine SQL API (example query in `wrangler.toml`).
- `run_worker_first` includes `/` so the landing page gets the beacon.
- Local dev reads `.dev.vars` (gitignored).

## Gotchas

- Rate limiting is per-isolate memory — best effort, not global.
- This relay is unrelated to the control plane's webhooks
  (`chaossql-webhooks-outbox`), which have their own SSRF protection.
- `site/_worker.js` used to be hand-edited and drifted (it lacked
  `/api/event`); it is now generated and a test pins it to `worker.ts`.
- `index.html` must stay a plain shell: other deployments (nginx, the Go
  `spaFileServer`) use it as their SPA fallback, which is why prerendered
  pages live under `site/prerender/` and only this worker maps them to URLs.
- Paths under `/pt` and `/scenarios/*` must be in `run_worker_first`,
  otherwise the assets layer answers with the SPA fallback and the Portuguese
  metadata and prerendered HTML are never served.

## Tests

- `node --test tools/test_waitlist.mjs tools/test_site_events.mjs tools/test_worker_routes.mjs`
  (part of `make verify`). `test_worker_routes.mjs` bundles `worker.ts` with a
  fake `HTMLRewriter` and covers prerendered pages, Portuguese shell metadata,
  the 404 for unknown slugs and trailing-slash redirects.
- `cd site && npx vitest run src/lib/router.test.ts`

## Source map

- `worker.ts`
- `site/_worker.js`
- `functions/api/waitlist.ts`
- `site/functions/api/waitlist.ts`
- `wrangler.toml`
- `site/wrangler.toml`
- `site/src/lib/seo.ts`
- `site/scripts/build-pages-worker.mjs`
- `tools/test_worker_routes.mjs`
- `site/_redirects`
- `tools/test_waitlist.mjs`
- `tools/test_site_events.mjs`
- `site/src/lib/analytics.ts`
- `.github/workflows/deploy-pages.yml`
- `.github/workflows/static-pages.yml`

## Related skills

- `chaossql-website-portal`, `chaossql-quality-gate`
