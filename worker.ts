const DEFAULT_DISCORD_WEBHOOK = "https://discord.com/api/webhooks/1547260093618327612/nQo6Orm496uN4AW0i0vCui2UyllzWibNzT2sH6R_mqgUCmDLNUG2XunB98D-RGwGNKeX";

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

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    // Handle CORS preflight for API routes
    if (url.pathname === '/api/waitlist') {
      if (request.method === 'OPTIONS') {
        return new Response(null, {
          status: 204,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'POST, OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type',
          },
        });
      }

      if (request.method === 'POST') {
        try {
          const data = (await request.json()) as WaitlistPayload;

          if (!data.name || typeof data.name !== 'string' || !data.name.trim()) {
            return new Response(JSON.stringify({ error: 'Name is required' }), {
              status: 400,
              headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
            });
          }

          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
          if (!data.email || typeof data.email !== 'string' || !emailRegex.test(data.email.trim())) {
            return new Response(JSON.stringify({ error: 'Valid email is required' }), {
              status: 400,
              headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
            });
          }

          const webhookUrl = env?.DISCORD_WEBHOOK_URL || env?.WAITLIST_WEBHOOK_URL || DEFAULT_DISCORD_WEBHOOK;
          let dispatched = false;

          if (webhookUrl) {
            try {
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

              const discordRes = await fetch(webhookUrl, {
                method: 'POST',
                headers: {
                  'Content-Type': 'application/json',
                  'User-Agent': 'ChaosSQL-Lead-Dispatcher/1.0',
                },
                body: JSON.stringify(discordPayload),
              });

              if (!discordRes.ok) {
                const errText = await discordRes.text();
                console.error('Discord webhook failed with status ' + discordRes.status + ': ' + errText);
              } else {
                dispatched = true;
              }
            } catch (dispatchErr) {
              console.error('Failed to dispatch Discord webhook:', dispatchErr);
            }
          }

          return new Response(
            JSON.stringify({
              success: true,
              message: 'Inscrição registrada com sucesso!',
              lead: {
                name: data.name.trim(),
                email: data.email.trim(),
                wantAudit: Boolean(data.wantAudit),
                dispatched,
              },
            }),
            {
              status: 200,
              headers: {
                'Content-Type': 'application/json',
                'Access-Control-Allow-Origin': '*',
              },
            }
          );
        } catch (err: any) {
          return new Response(JSON.stringify({ error: err?.message || 'Invalid JSON body' }), {
            status: 400,
            headers: {
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*',
            },
          });
        }
      }

      return new Response('Method Not Allowed', {
        status: 405,
        headers: { 'Access-Control-Allow-Origin': '*' },
      });
    }


    // Webhook Test Dispatcher Route
    if (url.pathname === '/api/webhooks/test') {
      if (request.method === 'OPTIONS') {
        return new Response(null, {
          status: 204,
          headers: {
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'POST, OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type',
          },
        });
      }

      if (request.method === 'POST') {
        try {
          const data = (await request.json()) as {
            target?: string;
            url?: string;
          };

          const targetUrl = data.url;
          if (!targetUrl || !targetUrl.startsWith('http')) {
            return new Response(JSON.stringify({ error: 'Valid webhook url is required' }), {
              status: 400,
              headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
            });
          }

          const target = data.target || 'generic';
          let bodyPayload: any;

          if (target === 'discord') {
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
          } else if (target === 'slack') {
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

          const resp = await fetch(targetUrl, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'User-Agent': 'ChaosSQL-Webhook-Dispatcher/1.5',
            },
            body: JSON.stringify(bodyPayload),
          });

          return new Response(JSON.stringify({ success: true, status: resp.status }), {
            status: 200,
            headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
          });
        } catch (err: any) {
          return new Response(JSON.stringify({ error: err?.message || 'Failed to dispatch webhook' }), {
            status: 500,
            headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' },
          });
        }
      }

      return new Response('Method Not Allowed', {
        status: 405,
        headers: { 'Access-Control-Allow-Origin': '*' },
      });
    }

    // Default: Serve static assets via Cloudflare Assets binding
    return env.ASSETS.fetch(request);
  },
};
