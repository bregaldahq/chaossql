interface Env {
  ASSETS: {
    fetch: (request: Request) => Promise<Response>;
  };
  DISCORD_WEBHOOK_URL?: string;
  WAITLIST_WEBHOOK_URL?: string;
}

interface WaitlistPayload {
  name?: string;
  email?: string;
  company?: string;
  database?: string;
  wantAudit?: boolean;
  notes?: string;
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

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

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

          if (!data.name || typeof data.name !== 'string' || !data.name.trim()) {
            return jsonResponse(request, { error: 'Name is required' }, 400);
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

          const auditText = data.wantAudit ? '🛡️ **SIM (VIP Advisory)**' : 'Não solicitado';
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
                  { name: '👤 Nome', value: data.name.trim(), inline: true },
                  { name: '📧 E-mail', value: data.email.trim(), inline: true },
                  { name: '🏢 Empresa / Repo', value: data.company?.trim() || 'Não informada', inline: true },
                  { name: '🗄️ Banco Principal', value: data.database || 'PostgreSQL', inline: true },
                  { name: '🛡️ Concurrency Audit', value: auditText, inline: true },
                  { name: '📍 Origem', value: sourceText, inline: true },
                  { name: '📝 Desafio / Caso de Uso', value: data.notes?.trim() || 'Nenhum detalhe adicional informado.', inline: false },
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
                name: data.name.trim(),
                email: data.email.trim(),
                wantAudit: Boolean(data.wantAudit),
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
                    { name: 'Dashboard', value: '[Visualizar Trace & Repro ➔](https://chaossql.bregalda.com/#/dashboard)', inline: false },
                  ],
                  footer: {
                    text: 'ChaosSQL Concurrency Intelligence Engine • v1.5.0',
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

    // Default: Serve static assets via Cloudflare Assets binding
    return env.ASSETS.fetch(request);
  },
};
