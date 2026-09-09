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

          const webhookUrl = env.DISCORD_WEBHOOK_URL || env.WAITLIST_WEBHOOK_URL;
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
          } else {
            console.warn('DISCORD_WEBHOOK_URL or WAITLIST_WEBHOOK_URL environment variable is not configured');
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

    // Default: Serve static assets via Cloudflare Assets binding
    return env.ASSETS.fetch(request);
  },
};
