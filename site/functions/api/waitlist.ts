interface Env {
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

export async function onRequestPost(context: { request: Request; env: Env }) {
  const { request, env } = context;

  try {
    const data = (await request.json()) as WaitlistPayload;

    if (!data.name || typeof data.name !== 'string' || !data.name.trim()) {
      return new Response(JSON.stringify({ error: 'Name is required' }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!data.email || typeof data.email !== 'string' || !emailRegex.test(data.email.trim())) {
      return new Response(JSON.stringify({ error: 'Valid email is required' }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      });
    }

    // Optional dispatch to team webhook (Slack, Discord, Zapier, etc.)
    if (env.WAITLIST_WEBHOOK_URL) {
      try {
        await fetch(env.WAITLIST_WEBHOOK_URL, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            content: `🚨 **Novo Lead ChaosSQL SaaS Registrado!**\n- **Nome:** ${data.name}\n- **E-mail:** ${data.email}\n- **Empresa:** ${data.company || 'N/A'}\n- **Banco:** ${data.database || 'N/A'}\n- **Interesse em Concurrency Audit:** ${data.wantAudit ? '✅ SIM (VIP Advisory)' : 'Não'}\n- **Notas:** ${data.notes || 'N/A'}`,
          }),
        });
      } catch (webhookErr) {
        console.error('Failed to dispatch webhook notification:', webhookErr);
      }
    }

    return new Response(
      JSON.stringify({
        success: true,
        message: 'Waitlist application received successfully',
        lead: {
          name: data.name.trim(),
          email: data.email.trim(),
          wantAudit: Boolean(data.wantAudit),
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
  } catch (err) {
    return new Response(
      JSON.stringify({ error: 'Invalid JSON body or server error' }),
      {
        status: 400,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  }
}

export async function onRequestOptions() {
  return new Response(null, {
    status: 204,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
    },
  });
}
