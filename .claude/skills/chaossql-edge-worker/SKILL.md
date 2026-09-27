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
| `site/_worker.js` | bundled JavaScript copy of `worker.ts` for Cloudflare Pages advanced mode (`pages deploy site`); no script in the repo regenerates it |
| `functions/api/waitlist.ts`, `site/functions/api/waitlist.ts` | identical Pages Function versions of the waitlist endpoint |

Any behavior change must be applied to all applicable copies;
`tools/test_waitlist.mjs` transpiles and exercises all four with esbuild, and
`site/src/lib/router.test.ts` checks route metadata across `worker.ts`,
`site/_worker.js` and `wrangler.toml`.

## Endpoints (`worker.ts`)

- `POST /api/waitlist` (+ `OPTIONS` preflight): origin must be in
  `ALLOWED_ORIGINS` (`https://chaossql.bregalda.com`, `localhost:5173`,
  `127.0.0.1:5173`) → 403 otherwise; rate limit 5/min per client key → 429;
  requires non-empty `name` and a valid `email` → 400; forwards the lead to
  `env.DISCORD_WEBHOOK_URL || env.WAITLIST_WEBHOOK_URL`, and fails loudly
  when neither secret is configured (no hard-coded fallback).
- `POST /api/webhooks/test`: same origin check, 10/min limit; the URL must be
  https, without credentials, default port, and its host in
  `ALLOWED_WEBHOOK_HOSTS` (Discord variants, `hooks.slack.com`); sends a
  Discord/Slack/generic test message and returns `{success, status}`.
- Section URLs (`/docs`, `/playground`, ...): `serveAppShell` fetches the SPA
  shell from `ASSETS` and injects the route's `<title>`, description,
  canonical URL and robots directive from `ROUTE_META`; trailing slashes are
  normalized, then the Web Analytics beacon is appended (below).
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
- `site/_worker.js` drifts silently if you edit only `worker.ts`. It has
  already drifted: it has no `/api/event` endpoint and no analytics beacon, so
  a Pages deployment of `site/` loses site events and page-view analytics.
  Only the waitlist and route-metadata tests cover both copies.

## Tests

- `node --test tools/test_waitlist.mjs tools/test_site_events.mjs` (part of
  `make verify`; `test_site_events.mjs` covers `worker.ts` only)
- `cd site && npx vitest run src/lib/router.test.ts`

## Source map

- `worker.ts`
- `site/_worker.js`
- `functions/api/waitlist.ts`
- `site/functions/api/waitlist.ts`
- `wrangler.toml`
- `site/wrangler.toml`
- `site/_redirects`
- `tools/test_waitlist.mjs`
- `tools/test_site_events.mjs`
- `site/src/lib/analytics.ts`
- `.github/workflows/deploy-pages.yml`
- `.github/workflows/static-pages.yml`

## Related skills

- `chaossql-website-portal`, `chaossql-quality-gate`
