import { pageMeta, seoRouteForPath, splitLocale, type Lang } from './site/src/lib/seo';

interface Env {
  ASSETS: {
    fetch: (request: Request) => Promise<Response>;
  };
  DISCORD_WEBHOOK_URL?: string;
  WAITLIST_WEBHOOK_URL?: string;
  // Workers Analytics Engine dataset for cookieless site events (/api/event).
  SITE_EVENTS?: { writeDataPoint: (point: { blobs?: string[]; doubles?: number[]; indexes?: string[] }) => void };
  // Public Cloudflare Web Analytics site token; the beacon is injected into HTML
  // pages only when it is set.
  CF_WEB_ANALYTICS_TOKEN?: string;
}

interface WaitlistPayload {
  name?: string;
  email?: string;
  company?: string;
  database?: string;
  wantAudit?: boolean;
  plan?: string;
  billingCycle?: string;
  notes?: string;
  timeline?: string;
  source?: string;
  timestamp?: string;
}

// Only these hosts may ever be reached by the outbound webhook proxy. Without
// this allowlist /api/webhooks/test is an open SSRF relay: any caller could make
// our infrastructure POST to arbitrary internal or third-party endpoints.
const ALLOWED_WEBHOOK_HOSTS = new Set([
  'discord.com',
  'discordapp.com',
  'canary.discord.com',
  'ptb.discord.com',
  'hooks.slack.com',
]);

const ALLOWED_ORIGINS = new Set([
  'https://chaossql.bregalda.com',
  'http://localhost:5173',
  'http://127.0.0.1:5173',
]);

const RATE_LIMITS = {
  waitlist: { max: 5, windowMs: 60_000 },
  webhookTest: { max: 10, windowMs: 60_000 },
  siteEvent: { max: 60, windowMs: 60_000 },
};

// Best-effort throttle. Workers isolates are per-colo and short lived, so this
// caps trivial floods but is NOT a global limit — move to KV or a Durable
// Object if a hard guarantee is ever required.
const rateBuckets = new Map<string, number[]>();

function isRateLimited(key: string, limit: { max: number; windowMs: number }): boolean {
  const now = Date.now();
  const hits = (rateBuckets.get(key) ?? []).filter((t) => now - t < limit.windowMs);
  if (hits.length >= limit.max) {
    rateBuckets.set(key, hits);
    return true;
  }
  hits.push(now);
  rateBuckets.set(key, hits);
  if (rateBuckets.size > 10_000) rateBuckets.clear();
  return false;
}

function clientKey(request: Request, scope: string): string {
  const ip = request.headers.get('CF-Connecting-IP') || 'unknown';
  return `${scope}:${ip}`;
}

function corsHeaders(request: Request): Record<string, string> {
  const origin = request.headers.get('Origin');
  const allowed = origin && ALLOWED_ORIGINS.has(origin) ? origin : 'https://chaossql.bregalda.com';
  return {
    'Access-Control-Allow-Origin': allowed,
    'Vary': 'Origin',
  };
}

function isAllowedOrigin(request: Request): boolean {
  const origin = request.headers.get('Origin');
  // Same-origin browser POSTs and server-to-server callers may omit Origin.
  if (!origin) return true;
  return ALLOWED_ORIGINS.has(origin);
}

function jsonResponse(request: Request, body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json', ...corsHeaders(request) },
  });
}

function preflight(request: Request): Response {
  return new Response(null, {
    status: 204,
    headers: {
      ...corsHeaders(request),
      'Access-Control-Allow-Methods': 'POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
      'Access-Control-Max-Age': '86400',
    },
  });
}

/**
 * Validates a caller-supplied webhook destination. Rejects anything that is not
 * an https URL on an allowlisted provider host with a plausible webhook path.
 */
function parseWebhookTarget(raw: unknown): { url: string } | { error: string } {
  if (typeof raw !== 'string' || !raw.trim()) {
    return { error: 'Valid webhook url is required' };
  }

  let parsed: URL;
  try {
    parsed = new URL(raw.trim());
  } catch {
    return { error: 'Malformed webhook url' };
  }

  if (parsed.protocol !== 'https:') {
    return { error: 'Webhook url must use https' };
  }
  if (parsed.username || parsed.password) {
    return { error: 'Webhook url must not contain credentials' };
  }
  if (parsed.port && parsed.port !== '443') {
    return { error: 'Webhook url must use the default https port' };
  }

  const host = parsed.hostname.toLowerCase();
  if (!ALLOWED_WEBHOOK_HOSTS.has(host)) {
    return { error: 'Webhook host is not allowed' };
  }

  const isDiscord = host !== 'hooks.slack.com';
  const pathOk = isDiscord
    ? /^\/api\/webhooks\/\d+\/[\w-]+$/.test(parsed.pathname)
    : /^\/services\/[\w/-]+$/.test(parsed.pathname);
  if (!pathOk) {
    return { error: 'Webhook url does not match the expected provider format' };
  }

  return { url: parsed.toString() };
}

// Page metadata and language URLs come from the same module the app uses, so
// what crawlers see cannot drift from what the app sets. Pages prerendered at
// build time live in site/prerender/ (listed in /prerender/manifest.json) and
// are served at their canonical URL; every other section gets the app shell
// (index.html) with its metadata injected. index.html stays a plain shell
// because the dashboard and self-hosted servers use it as their SPA fallback.
// Detail pages: /scenarios/<slug> and /docs/<chapter>.
const DETAIL_PATH = /^\/(scenarios|docs)\/[a-z0-9-]+$/;

/** Unprefixed app path this request maps to, or null when it is not an app page. */
function appPath(pathname: string): { lang: Lang; path: string } | null {
  const { lang, path } = splitLocale(pathname);
  if (path === '/' || seoRouteForPath(path) !== null || DETAIL_PATH.test(path)) return { lang, path };
  return null;
}

let prerendered: Promise<Set<string>> | null = null;

/** Asset path of a prerendered page: "/" → "/prerender/home", "/pt/pricing" → "/prerender/pt/pricing". */
function prerenderAsset(pathname: string): string {
  return `/prerender${pathname === '/' ? '/home' : pathname}`;
}

/** Paths with a prerendered HTML file, read once per isolate from the build manifest. */
function prerenderedPaths(request: Request, env: Env): Promise<Set<string>> {
  prerendered ??= env.ASSETS.fetch(new Request(new URL('/prerender/manifest.json', request.url)))
    .then((res) => (res.ok ? res.json() : []))
    .then((paths: unknown) => new Set(Array.isArray(paths) ? paths.filter((p): p is string => typeof p === 'string') : []))
    .catch(() => {
      prerendered = null;
      return new Set<string>();
    });
  return prerendered;
}

// Keep in sync with SITE_EVENTS in site/src/lib/analytics.ts.
const SITE_EVENTS = new Set([
  'cta_click',
  'install_copy',
  'command_copy',
  'scenario_view',
  'lead_submit',
  'plan_select',
  'pricing_toggle',
  'outbound_click',
  'lang_switch',
]);
const EVENT_LABEL = /^[a-z0-9_:.-]{0,64}$/;
const EVENT_PATH = /^\/[\w\-/.]{0,127}$/;
const EVENT_REF = /^[a-z0-9.-]{0,64}$/;
const MAX_EVENT_BYTES = 1024;

/**
 * Records one cookieless site event. Only allowlisted, shape-checked fields are
 * stored: no IP, no user agent, no identifiers. Always answers 204 for
 * well-formed requests so the client never retries or surfaces errors.
 */
async function handleSiteEvent(request: Request, env: Env): Promise<Response> {
  if (!isAllowedOrigin(request)) {
    return jsonResponse(request, { error: 'Origin not allowed' }, 403);
  }
  const raw = await request.text();
  if (raw.length > MAX_EVENT_BYTES) {
    return jsonResponse(request, { error: 'Payload too large' }, 413);
  }
  let data: Record<string, unknown>;
  try {
    data = JSON.parse(raw);
  } catch {
    return jsonResponse(request, { error: 'Invalid JSON body' }, 400);
  }
  const event = typeof data.event === 'string' ? data.event : '';
  const label = typeof data.label === 'string' ? data.label : '';
  const path = typeof data.path === 'string' ? data.path : '/';
  const ref = typeof data.ref === 'string' ? data.ref.toLowerCase() : '';
  const lang = data.lang === 'pt' ? 'pt' : 'en';
  if (!SITE_EVENTS.has(event) || !EVENT_LABEL.test(label) || !EVENT_PATH.test(path) || !EVENT_REF.test(ref)) {
    return jsonResponse(request, { error: 'Invalid event' }, 400);
  }
  if (isRateLimited(clientKey(request, 'site-event'), RATE_LIMITS.siteEvent)) {
    return new Response(null, { status: 204, headers: corsHeaders(request) });
  }
  const country = (request as Request & { cf?: { country?: string } }).cf?.country ?? '';
  try {
    env.SITE_EVENTS?.writeDataPoint({
      indexes: [event],
      blobs: [event, label, path, lang, ref, country],
      doubles: [1],
    });
  } catch {
    console.error('Failed to write site event');
  }
  return new Response(null, { status: 204, headers: corsHeaders(request) });
}

const CF_BEACON_TOKEN = /^[a-f0-9]{32}$/;

// Appends the Cloudflare Web Analytics beacon (cookieless page views) to an HTML
// response. A missing or malformed token leaves the page untouched.
function withWebAnalytics(response: Response, env: Env): Response {
  const token = env.CF_WEB_ANALYTICS_TOKEN?.trim();
  const isHtml = (response.headers.get('Content-Type') || '').includes('text/html');
  if (!token || !CF_BEACON_TOKEN.test(token) || !isHtml) return response;
  const beacon = `<script defer src="https://static.cloudflareinsights.com/beacon.min.js" data-cf-beacon='${JSON.stringify({ token, spa: true })}'></script>`;
  return new HTMLRewriter()
    .on('body', { element: (el) => { el.append(beacon, { html: true }); } })
    .transform(response);
}

async function serveAppShell(request: Request, env: Env, lang: Lang, path: string): Promise<Response> {
  // Detail pages are all prerendered; one that reaches the shell has an unknown slug.
  const detail = DETAIL_PATH.exec(path);
  const route = detail ? (detail[1] as 'scenarios' | 'docs') : seoRouteForPath(path) ?? 'landing';
  const meta = pageMeta(route, lang, path);
  const shell = await env.ASSETS.fetch(new Request(new URL('/', request.url), request));
  if (!shell.ok) return shell;
  const rewritten = new HTMLRewriter()
    .on('html', { element: (el) => { el.setAttribute('lang', meta.htmlLang); } })
    .on('title', { element: (el) => { el.setInnerContent(meta.title); } })
    .on('meta[name="description"]', { element: (el) => { el.setAttribute('content', meta.description); } })
    .on('meta[name="robots"]', {
      // An unknown scenario or chapter slug is a 404 and must not be indexed.
      element: (el) => { el.setAttribute('content', detail ? 'noindex, follow' : meta.robots); },
    })
    .on('link[rel="canonical"]', { element: (el) => { el.setAttribute('href', meta.canonical); } })
    .on('meta[property="og:url"]', { element: (el) => { el.setAttribute('content', meta.canonical); } })
    .on('meta[property="og:title"]', { element: (el) => { el.setAttribute('content', meta.title); } })
    .on('meta[property="og:description"]', { element: (el) => { el.setAttribute('content', meta.description); } })
    .on('meta[property="og:locale"]', { element: (el) => { el.setAttribute('content', meta.ogLocale); } })
    .on('meta[property="og:locale:alternate"]', { element: (el) => { el.setAttribute('content', meta.ogLocaleAlternate); } })
    .on('meta[property="og:image"]', { element: (el) => { el.setAttribute('content', meta.image); } })
    .on('meta[name="twitter:title"]', { element: (el) => { el.setAttribute('content', meta.title); } })
    .on('meta[name="twitter:description"]', { element: (el) => { el.setAttribute('content', meta.description); } })
    .on('meta[name="twitter:image"]', { element: (el) => { el.setAttribute('content', meta.image); } })
    .on('link[rel="alternate"][hreflang]', {
      element: (el) => {
        const href = meta.alternates[el.getAttribute('hreflang') as keyof typeof meta.alternates];
        if (href) el.setAttribute('href', href);
      },
    })
    .transform(shell);
  if (!detail) return rewritten;
  return new Response(rewritten.body, { status: 404, headers: rewritten.headers });
}


export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === '/api/event') {
      if (request.method === 'OPTIONS') return preflight(request);
      if (request.method === 'POST') return handleSiteEvent(request, env);
      return new Response('Method Not Allowed', { status: 405, headers: corsHeaders(request) });
    }

    if (url.pathname === '/api/waitlist') {
      if (request.method === 'OPTIONS') return preflight(request);

      if (request.method === 'POST') {
        if (!isAllowedOrigin(request)) {
          return jsonResponse(request, { error: 'Origin not allowed' }, 403);
        }
        if (isRateLimited(clientKey(request, 'waitlist'), RATE_LIMITS.waitlist)) {
          return jsonResponse(request, { error: 'Too many requests, try again shortly' }, 429);
        }

        try {
          const data = (await request.json()) as WaitlistPayload;

          // Name is optional (the waitlist asks only for an email); when sent it must be text.
          if (data.name !== undefined && typeof data.name !== 'string') {
            return jsonResponse(request, { error: 'Name must be a string' }, 400);
          }

          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
          if (!data.email || typeof data.email !== 'string' || !emailRegex.test(data.email.trim())) {
            return jsonResponse(request, { error: 'Valid email is required' }, 400);
          }

          // No hardcoded fallback: the webhook is the only sink for leads, so a
          // missing secret must fail loudly instead of silently dropping them.
          const webhookUrl = env?.DISCORD_WEBHOOK_URL || env?.WAITLIST_WEBHOOK_URL;
          if (!webhookUrl) {
            console.error('DISCORD_WEBHOOK_URL is not configured; waitlist submission rejected');
            return jsonResponse(
              request,
              { error: 'Serviço de inscrição temporariamente indisponível. Tente novamente em instantes.' },
              503
            );
          }

          const wantAudit = data.wantAudit === true || data.plan?.toLowerCase() === 'audit';
          const auditText = wantAudit ? '🛡️ **SIM (VIP Advisory)**' : 'Não solicitado';
          const sourceText = data.source === 'pricing_page' ? '🏷️ Página de Pricing' : '🚀 Landing Page';

          const discordPayload = {
            username: 'ChaosSQL Lead Bot',
            avatar_url: 'https://chaossql.bregalda.com/brand/icone_bregalda.svg',
            embeds: [
              {
                title: '🚨 Novo Lead / Registro de Interesse!',
                description: 'Um novo desenvolvedor acabou de se registrar no **ChaosSQL SaaS**.',
                color: 0x4b2e83, // Bregalda Purple (#4B2E83)
                fields: [
                  { name: '👤 Nome', value: data.name?.trim().slice(0, 128) || 'Não informado', inline: true },
                  { name: '📧 E-mail', value: data.email.trim(), inline: true },
                  { name: '🏢 Empresa / Repo', value: data.company?.trim().slice(0, 128) || 'Não informada', inline: true },
                  { name: '🗄️ Banco Principal', value: data.database?.slice(0, 64) || 'PostgreSQL', inline: true },
                  { name: '🛡️ Concurrency Audit', value: auditText, inline: true },
                  { name: 'Plan', value: data.plan?.trim().slice(0, 128) || 'Not selected', inline: true },
                  { name: 'Billing cycle', value: data.billingCycle?.trim().slice(0, 32) || 'Not selected', inline: true },
                  { name: 'Timeline', value: data.timeline?.trim().slice(0, 128) || 'Not provided', inline: true },
                  { name: '📍 Origem', value: sourceText, inline: true },
                  { name: '📝 Desafio / Caso de Uso', value: data.notes?.trim().slice(0, 1000) || 'Nenhum detalhe adicional informado.', inline: false },
                ],
                footer: {
                  text: 'ChaosSQL Cloud Control Plane • Studio Bregalda',
                  icon_url: 'https://chaossql.bregalda.com/brand/icone_bregalda.svg',
                },
                timestamp: new Date().toISOString(),
              },
            ],
          };

          let dispatched = false;
          try {
            const discordRes = await fetch(webhookUrl, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'User-Agent': 'ChaosSQL-Lead-Dispatcher/1.0',
              },
              body: JSON.stringify(discordPayload),
            });

            if (!discordRes.ok) {
              // Never echo the response body: it can carry the webhook URL back.
              console.error('Discord webhook failed with status ' + discordRes.status);
            } else {
              dispatched = true;
            }
          } catch {
            console.error('Failed to dispatch Discord webhook');
          }

          if (!dispatched) {
            return jsonResponse(
              request,
              { error: 'Não foi possível registrar sua inscrição agora. Tente novamente em instantes.' },
              502
            );
          }

          return jsonResponse(
            request,
            {
              success: true,
              message: 'Inscrição registrada com sucesso!',
              lead: {
                name: data.name?.trim().slice(0, 128) ?? '',
                email: data.email.trim(),
                wantAudit,
                dispatched,
              },
            },
            200
          );
        } catch {
          return jsonResponse(request, { error: 'Invalid JSON body' }, 400);
        }
      }

      return new Response('Method Not Allowed', { status: 405, headers: corsHeaders(request) });
    }

    // Webhook Test Dispatcher Route
    if (url.pathname === '/api/webhooks/test') {
      if (request.method === 'OPTIONS') return preflight(request);

      if (request.method === 'POST') {
        if (!isAllowedOrigin(request)) {
          return jsonResponse(request, { error: 'Origin not allowed' }, 403);
        }
        if (isRateLimited(clientKey(request, 'webhook-test'), RATE_LIMITS.webhookTest)) {
          return jsonResponse(request, { error: 'Too many requests, try again shortly' }, 429);
        }

        try {
          const data = (await request.json()) as { target?: string; url?: string };

          const target = parseWebhookTarget(data.url);
          if ('error' in target) {
            return jsonResponse(request, { error: target.error }, 400);
          }

          const kind = data.target === 'discord' || data.target === 'slack' ? data.target : 'generic';
          let bodyPayload: unknown;

          if (kind === 'discord') {
            bodyPayload = {
              username: 'ChaosSQL Alert Bot',
              avatar_url: 'https://chaossql.bregalda.com/brand/icone_bregalda.svg',
              embeds: [
                {
                  title: '🚨 [TEST] Concurrency Regression Detected',
                  description: 'Notificação de teste em tempo real disparada a partir do ChaosSQL Concurrency Gate.',
                  color: 0xDC2626,
                  fields: [
                    { name: 'Repositório', value: '`acme/payments`', inline: true },
                    { name: 'Branch / PR', value: 'PR #104 (`fix/concurrent-settlement`)', inline: true },
                    { name: 'Anomalia', value: '**Deadlock Cycle (40P01)**', inline: true },
                    { name: 'Engine', value: 'PostgreSQL 16 (REPEATABLE READ)', inline: true },
                    { name: 'Status', value: '❌ FAILED (18/50 schedules abortados)', inline: true },
                    { name: 'Dashboard', value: '[Visualizar Trace & Repro ➔](https://chaossql.bregalda.com/dashboard)', inline: false },
                  ],
                  footer: {
                    text: 'ChaosSQL Concurrency Intelligence Engine • v1.6.0',
                    icon_url: 'https://chaossql.bregalda.com/brand/icone_bregalda.svg',
                  },
                  timestamp: new Date().toISOString(),
                },
              ],
            };
          } else if (kind === 'slack') {
            bodyPayload = {
              text: '🚨 *[TEST] ChaosSQL Alert:* Concurrency regression in `acme/payments` PR #104 (Deadlock Cycle)',
              blocks: [
                {
                  type: 'header',
                  text: { type: 'plain_text', text: '🚨 [TEST] Concurrency Regression Detected', emoji: true },
                },
                {
                  type: 'section',
                  fields: [
                    { type: 'mrkdwn', text: '*Repositório:*\n`acme/payments`' },
                    { type: 'mrkdwn', text: '*Branch / PR:*\n`fix/concurrent-settlement` (PR #104)' },
                    { type: 'mrkdwn', text: '*Anomalia:*\n*Deadlock Cycle (40P01)*' },
                    { type: 'mrkdwn', text: '*Engine:*\nPostgreSQL 16' },
                  ],
                },
              ],
            };
          } else {
            bodyPayload = {
              event: 'test_concurrency_alert',
              timestamp: new Date().toISOString(),
              message: 'Test webhook from ChaosSQL Dashboard',
            };
          }

          const resp = await fetch(target.url, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'User-Agent': 'ChaosSQL-Webhook-Dispatcher/1.5',
            },
            body: JSON.stringify(bodyPayload),
          });

          return jsonResponse(request, { success: resp.ok, status: resp.status }, 200);
        } catch {
          return jsonResponse(request, { error: 'Failed to dispatch webhook' }, 500);
        }
      }

      return new Response('Method Not Allowed', { status: 405, headers: corsHeaders(request) });
    }

    if (request.method === 'GET' || request.method === 'HEAD') {
      const trimmed = url.pathname.replace(/\/+$/, '');
      if (trimmed && trimmed !== url.pathname && appPath(trimmed)) {
        return Response.redirect(`${url.origin}${trimmed}${url.search}`, 301);
      }
      const page = appPath(url.pathname);
      if (page) {
        try {
          if ((await prerenderedPaths(request, env)).has(url.pathname)) {
            const asset = await env.ASSETS.fetch(new Request(new URL(prerenderAsset(url.pathname), request.url), request));
            if (asset.ok) return withWebAnalytics(asset, env);
          }
          return withWebAnalytics(await serveAppShell(request, env, page.lang, page.path), env);
        } catch {
          return new Response('Service Unavailable', { status: 503, headers: { 'Content-Type': 'text/plain' } });
        }
      }
    }

    // Default: Serve static assets via Cloudflare Assets binding. A thrown
    // asset lookup must never surface as a 5xx: crawlers treat a failing
    // robots.txt as "do not crawl" and drop the whole site from search.
    try {
      return await env.ASSETS.fetch(request);
    } catch {
      return new Response('Not Found', { status: 404, headers: { 'Content-Type': 'text/plain' } });
    }
  },
};
