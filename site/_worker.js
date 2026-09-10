// worker.ts
var ALLOWED_WEBHOOK_HOSTS = /* @__PURE__ */ new Set([
  "discord.com",
  "discordapp.com",
  "canary.discord.com",
  "ptb.discord.com",
  "hooks.slack.com"
]);
var ALLOWED_ORIGINS = /* @__PURE__ */ new Set([
  "https://chaossql.bregalda.com",
  "http://localhost:5173",
  "http://127.0.0.1:5173"
]);
var RATE_LIMITS = {
  waitlist: { max: 5, windowMs: 6e4 },
  webhookTest: { max: 10, windowMs: 6e4 }
};
var rateBuckets = /* @__PURE__ */ new Map();
function isRateLimited(key, limit) {
  const now = Date.now();
  const hits = (rateBuckets.get(key) ?? []).filter((t) => now - t < limit.windowMs);
  if (hits.length >= limit.max) {
    rateBuckets.set(key, hits);
    return true;
  }
  hits.push(now);
  rateBuckets.set(key, hits);
  if (rateBuckets.size > 1e4) rateBuckets.clear();
  return false;
}
function clientKey(request, scope) {
  const ip = request.headers.get("CF-Connecting-IP") || "unknown";
  return `${scope}:${ip}`;
}
function corsHeaders(request) {
  const origin = request.headers.get("Origin");
  const allowed = origin && ALLOWED_ORIGINS.has(origin) ? origin : "https://chaossql.bregalda.com";
  return {
    "Access-Control-Allow-Origin": allowed,
    "Vary": "Origin"
  };
}
function isAllowedOrigin(request) {
  const origin = request.headers.get("Origin");
  if (!origin) return true;
  return ALLOWED_ORIGINS.has(origin);
}
function jsonResponse(request, body, status) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json", ...corsHeaders(request) }
  });
}
function preflight(request) {
  return new Response(null, {
    status: 204,
    headers: {
      ...corsHeaders(request),
      "Access-Control-Allow-Methods": "POST, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type",
      "Access-Control-Max-Age": "86400"
    }
  });
}
function parseWebhookTarget(raw) {
  if (typeof raw !== "string" || !raw.trim()) {
    return { error: "Valid webhook url is required" };
  }
  let parsed;
  try {
    parsed = new URL(raw.trim());
  } catch {
    return { error: "Malformed webhook url" };
  }
  if (parsed.protocol !== "https:") {
    return { error: "Webhook url must use https" };
  }
  if (parsed.username || parsed.password) {
    return { error: "Webhook url must not contain credentials" };
  }
  if (parsed.port && parsed.port !== "443") {
    return { error: "Webhook url must use the default https port" };
  }
  const host = parsed.hostname.toLowerCase();
  if (!ALLOWED_WEBHOOK_HOSTS.has(host)) {
    return { error: "Webhook host is not allowed" };
  }
  const isDiscord = host !== "hooks.slack.com";
  const pathOk = isDiscord ? /^\/api\/webhooks\/\d+\/[\w-]+$/.test(parsed.pathname) : /^\/services\/[\w/-]+$/.test(parsed.pathname);
  if (!pathOk) {
    return { error: "Webhook url does not match the expected provider format" };
  }
  return { url: parsed.toString() };
}
var worker_default = {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname === "/api/waitlist") {
      if (request.method === "OPTIONS") return preflight(request);
      if (request.method === "POST") {
        if (!isAllowedOrigin(request)) {
          return jsonResponse(request, { error: "Origin not allowed" }, 403);
        }
        if (isRateLimited(clientKey(request, "waitlist"), RATE_LIMITS.waitlist)) {
          return jsonResponse(request, { error: "Too many requests, try again shortly" }, 429);
        }
        try {
          const data = await request.json();
          if (!data.name || typeof data.name !== "string" || !data.name.trim()) {
            return jsonResponse(request, { error: "Name is required" }, 400);
          }
          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
          if (!data.email || typeof data.email !== "string" || !emailRegex.test(data.email.trim())) {
            return jsonResponse(request, { error: "Valid email is required" }, 400);
          }
          const webhookUrl = env?.DISCORD_WEBHOOK_URL || env?.WAITLIST_WEBHOOK_URL;
          if (!webhookUrl) {
            console.error("DISCORD_WEBHOOK_URL is not configured; waitlist submission rejected");
            return jsonResponse(
              request,
              { error: "Servi\xE7o de inscri\xE7\xE3o temporariamente indispon\xEDvel. Tente novamente em instantes." },
              503
            );
          }
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
          let dispatched = false;
          try {
            const discordRes = await fetch(webhookUrl, {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                "User-Agent": "ChaosSQL-Lead-Dispatcher/1.0"
              },
              body: JSON.stringify(discordPayload)
            });
            if (!discordRes.ok) {
              console.error("Discord webhook failed with status " + discordRes.status);
            } else {
              dispatched = true;
            }
          } catch {
            console.error("Failed to dispatch Discord webhook");
          }
          if (!dispatched) {
            return jsonResponse(
              request,
              { error: "N\xE3o foi poss\xEDvel registrar sua inscri\xE7\xE3o agora. Tente novamente em instantes." },
              502
            );
          }
          return jsonResponse(
            request,
            {
              success: true,
              message: "Inscri\xE7\xE3o registrada com sucesso!",
              lead: {
                name: data.name.trim(),
                email: data.email.trim(),
                wantAudit: Boolean(data.wantAudit),
                dispatched
              }
            },
            200
          );
        } catch {
          return jsonResponse(request, { error: "Invalid JSON body" }, 400);
        }
      }
      return new Response("Method Not Allowed", { status: 405, headers: corsHeaders(request) });
    }
    if (url.pathname === "/api/webhooks/test") {
      if (request.method === "OPTIONS") return preflight(request);
      if (request.method === "POST") {
        if (!isAllowedOrigin(request)) {
          return jsonResponse(request, { error: "Origin not allowed" }, 403);
        }
        if (isRateLimited(clientKey(request, "webhook-test"), RATE_LIMITS.webhookTest)) {
          return jsonResponse(request, { error: "Too many requests, try again shortly" }, 429);
        }
        try {
          const data = await request.json();
          const target = parseWebhookTarget(data.url);
          if ("error" in target) {
            return jsonResponse(request, { error: target.error }, 400);
          }
          const kind = data.target === "discord" || data.target === "slack" ? data.target : "generic";
          let bodyPayload;
          if (kind === "discord") {
            bodyPayload = {
              username: "ChaosSQL Alert Bot",
              avatar_url: "https://chaossql.bregalda.com/brand/icone_bregalda.svg",
              embeds: [
                {
                  title: "\u{1F6A8} [TEST] Concurrency Regression Detected",
                  description: "Notifica\xE7\xE3o de teste em tempo real disparada a partir do ChaosSQL Concurrency Gate.",
                  color: 14427686,
                  fields: [
                    { name: "Reposit\xF3rio", value: "`acme/payments`", inline: true },
                    { name: "Branch / PR", value: "PR #104 (`fix/concurrent-settlement`)", inline: true },
                    { name: "Anomalia", value: "**Deadlock Cycle (40P01)**", inline: true },
                    { name: "Engine", value: "PostgreSQL 16 (REPEATABLE READ)", inline: true },
                    { name: "Status", value: "\u274C FAILED (18/50 schedules abortados)", inline: true },
                    { name: "Dashboard", value: "[Visualizar Trace & Repro \u2794](https://chaossql.bregalda.com/#/dashboard)", inline: false }
                  ],
                  footer: {
                    text: "ChaosSQL Concurrency Intelligence Engine \u2022 v1.5.0",
                    icon_url: "https://chaossql.bregalda.com/brand/icone_bregalda.svg"
                  },
                  timestamp: (/* @__PURE__ */ new Date()).toISOString()
                }
              ]
            };
          } else if (kind === "slack") {
            bodyPayload = {
              text: "\u{1F6A8} *[TEST] ChaosSQL Alert:* Concurrency regression in `acme/payments` PR #104 (Deadlock Cycle)",
              blocks: [
                {
                  type: "header",
                  text: { type: "plain_text", text: "\u{1F6A8} [TEST] Concurrency Regression Detected", emoji: true }
                },
                {
                  type: "section",
                  fields: [
                    { type: "mrkdwn", text: "*Reposit\xF3rio:*\n`acme/payments`" },
                    { type: "mrkdwn", text: "*Branch / PR:*\n`fix/concurrent-settlement` (PR #104)" },
                    { type: "mrkdwn", text: "*Anomalia:*\n*Deadlock Cycle (40P01)*" },
                    { type: "mrkdwn", text: "*Engine:*\nPostgreSQL 16" }
                  ]
                }
              ]
            };
          } else {
            bodyPayload = {
              event: "test_concurrency_alert",
              timestamp: (/* @__PURE__ */ new Date()).toISOString(),
              message: "Test webhook from ChaosSQL Dashboard"
            };
          }
          const resp = await fetch(target.url, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "User-Agent": "ChaosSQL-Webhook-Dispatcher/1.5"
            },
            body: JSON.stringify(bodyPayload)
          });
          return jsonResponse(request, { success: resp.ok, status: resp.status }, 200);
        } catch {
          return jsonResponse(request, { error: "Failed to dispatch webhook" }, 500);
        }
      }
      return new Response("Method Not Allowed", { status: 405, headers: corsHeaders(request) });
    }
    return env.ASSETS.fetch(request);
  }
};
export {
  worker_default as default
};
