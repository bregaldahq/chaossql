// worker.ts
var DEFAULT_DISCORD_WEBHOOK = "https://discord.com/api/webhooks/1547260093618327612/nQo6Orm496uN4AW0i0vCui2UyllzWibNzT2sH6R_mqgUCmDLNUG2XunB98D-RGwGNKeX";
var worker_default = {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname === "/api/waitlist") {
      if (request.method === "OPTIONS") {
        return new Response(null, {
          status: 204,
          headers: {
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "POST, OPTIONS",
            "Access-Control-Allow-Headers": "Content-Type"
          }
        });
      }
      if (request.method === "POST") {
        try {
          const data = await request.json();
          if (!data.name || typeof data.name !== "string" || !data.name.trim()) {
            return new Response(JSON.stringify({ error: "Name is required" }), {
              status: 400,
              headers: { "Content-Type": "application/json", "Access-Control-Allow-Origin": "*" }
            });
          }
          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
          if (!data.email || typeof data.email !== "string" || !emailRegex.test(data.email.trim())) {
            return new Response(JSON.stringify({ error: "Valid email is required" }), {
              status: 400,
              headers: { "Content-Type": "application/json", "Access-Control-Allow-Origin": "*" }
            });
          }
          const webhookUrl = env?.DISCORD_WEBHOOK_URL || env?.WAITLIST_WEBHOOK_URL || DEFAULT_DISCORD_WEBHOOK;
          let dispatched = false;
          if (webhookUrl) {
            try {
              const auditText = data.wantAudit ? "\u{1F6E1}\uFE0F **SIM (VIP Advisory)**" : "N\xE3o solicitado";
              const sourceText = data.source === "pricing_page" ? "\u{1F3F7}\uFE0F P\xE1gina de Pricing" : "\u{1F680} Landing Page";
              const discordPayload = {
                username: "ChaosSQL Lead Bot",
                avatar_url: "https://chaossql.bregalda.com/brand/icone_bregalda.svg",
                embeds: [
                  {
                    title: "\u{1F6A8} Novo Lead / Registro de Interesse!",
                    description: "Um novo desenvolvedor acabou de se registrar no **ChaosSQL SaaS**.",
                    color: 4927107,
                    // Bregalda Purple (#4B2E83)
                    fields: [
                      { name: "\u{1F464} Nome", value: data.name.trim(), inline: true },
                      { name: "\u{1F4E7} E-mail", value: data.email.trim(), inline: true },
                      { name: "\u{1F3E2} Empresa / Repo", value: data.company?.trim() || "N\xE3o informada", inline: true },
                      { name: "\u{1F5C4}\uFE0F Banco Principal", value: data.database || "PostgreSQL", inline: true },
                      { name: "\u{1F6E1}\uFE0F Concurrency Audit", value: auditText, inline: true },
                      { name: "\u{1F4CD} Origem", value: sourceText, inline: true },
                      { name: "\u{1F4DD} Desafio / Caso de Uso", value: data.notes?.trim() || "Nenhum detalhe adicional informado.", inline: false }
                    ],
                    footer: {
                      text: "ChaosSQL Cloud Control Plane \u2022 Studio Bregalda",
                      icon_url: "https://chaossql.bregalda.com/brand/icone_bregalda.svg"
                    },
                    timestamp: (/* @__PURE__ */ new Date()).toISOString()
                  }
                ]
              };
              const discordRes = await fetch(webhookUrl, {
                method: "POST",
                headers: {
                  "Content-Type": "application/json",
                  "User-Agent": "ChaosSQL-Lead-Dispatcher/1.0"
                },
                body: JSON.stringify(discordPayload)
              });
              if (!discordRes.ok) {
                const errText = await discordRes.text();
                console.error("Discord webhook failed with status " + discordRes.status + ": " + errText);
              } else {
                dispatched = true;
              }
            } catch (dispatchErr) {
              console.error("Failed to dispatch Discord webhook:", dispatchErr);
            }
          }
          return new Response(
            JSON.stringify({
              success: true,
              message: "Inscri\xE7\xE3o registrada com sucesso!",
              lead: {
                name: data.name.trim(),
                email: data.email.trim(),
                wantAudit: Boolean(data.wantAudit),
                dispatched
              }
            }),
            {
              status: 200,
              headers: {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*"
              }
            }
          );
        } catch (err) {
          return new Response(JSON.stringify({ error: err?.message || "Invalid JSON body" }), {
            status: 400,
            headers: {
              "Content-Type": "application/json",
              "Access-Control-Allow-Origin": "*"
            }
          });
        }
      }
      return new Response("Method Not Allowed", {
        status: 405,
        headers: { "Access-Control-Allow-Origin": "*" }
      });
    }
    return env.ASSETS.fetch(request);
  }
};
export {
  worker_default as default
};
