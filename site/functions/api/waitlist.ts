interface Env {
  DISCORD_WEBHOOK_URL?: string;
  WAITLIST_WEBHOOK_URL?: string;
  ENVIRONMENT?: string;
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

const ALLOWED_ORIGINS = new Set([
  'https://chaossql.bregalda.com',
  'http://localhost:5173',
  'http://127.0.0.1:5173',
]);

const RATE_LIMIT = { max: 5, windowMs: 60_000 };

// Best-effort throttle, scoped to a single isolate — see worker.ts for the
// same caveat. Promote to KV or a Durable Object if a hard limit is needed.
const rateBuckets = new Map<string, number[]>();

function isRateLimited(ip: string): boolean {
  const now = Date.now();
  const hits = (rateBuckets.get(ip) ?? []).filter((t) => now - t < RATE_LIMIT.windowMs);
  if (hits.length >= RATE_LIMIT.max) {
    rateBuckets.set(ip, hits);
    return true;
  }
  hits.push(now);
  rateBuckets.set(ip, hits);
  if (rateBuckets.size > 10_000) rateBuckets.clear();
  return false;
}

function corsHeaders(request: Request): Record<string, string> {
  const origin = request.headers.get('Origin');
  const allowed = origin && ALLOWED_ORIGINS.has(origin) ? origin : 'https://chaossql.bregalda.com';
  return {
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': allowed,
    'Vary': 'Origin',
  };
}

function jsonResponse(request: Request, body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), { status, headers: corsHeaders(request) });
}

export async function onRequestPost(context: { request: Request; env: Env }) {
  const { request, env } = context;

  const origin = request.headers.get('Origin');
  if (origin && !ALLOWED_ORIGINS.has(origin)) {
    return jsonResponse(request, { error: 'Origin not allowed' }, 403);
  }

  if (isRateLimited(request.headers.get('CF-Connecting-IP') || 'unknown')) {
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

      if (discordRes.ok) {
        dispatched = true;
      } else {
        // Never echo the response body: it can carry the webhook URL back.
        console.error('Discord webhook failed with status ' + discordRes.status);
      }
    } catch {
      console.error('Discord dispatch failed');
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

export async function onRequestOptions(context: { request: Request }) {
  const { request } = context;
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
