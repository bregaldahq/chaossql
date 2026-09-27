// Generated from worker.ts by site/scripts/build-pages-worker.mjs. Do not edit.

// src/lib/seo.ts
var SITE_ORIGIN = "https://chaossql.bregalda.com";
var PT_PREFIX = "/pt";
function splitLocale(pathname) {
  if (pathname === PT_PREFIX || pathname.startsWith(`${PT_PREFIX}/`)) {
    return { lang: "pt", path: pathname.slice(PT_PREFIX.length) || "/" };
  }
  return { lang: "en", path: pathname || "/" };
}
function localizePath(href, lang) {
  const [, rawPath = "/", rest = ""] = /^([^?#]*)(.*)$/.exec(href) ?? [];
  const { path } = splitLocale(rawPath || "/");
  if (lang === "en") return path + rest;
  return (path === "/" ? PT_PREFIX : PT_PREFIX + path) + rest;
}
var SEO_ROUTES = {
  landing: {
    path: "/",
    indexable: true,
    image: "/og/home.png",
    en: {
      title: "ChaosSQL | Deterministic SQL Concurrency Fuzzer for PostgreSQL, MySQL and SQLite",
      description: "Open-source (MIT) fuzzer that forces races between SQL transactions, detects lost updates, write skew and Adya isolation anomalies, and shrinks each failure to a minimal Go test."
    },
    pt: {
      title: "ChaosSQL | Fuzzer determin\xEDstico de concorr\xEAncia SQL para PostgreSQL, MySQL e SQLite",
      description: "Fuzzer open source (MIT) que for\xE7a corridas entre transa\xE7\xF5es SQL, detecta lost update, write skew e anomalias de isolamento de Adya e reduz cada falha a um teste Go m\xEDnimo."
    }
  },
  docs: {
    path: "/docs",
    indexable: true,
    image: "/og/docs.png",
    en: {
      title: "Documentation | ChaosSQL",
      description: "ChaosSQL guide: installation, the chaos.yaml spec, SQL invariants, isolation levels, Adya anomaly classification and delta-debugging minimization."
    },
    pt: {
      title: "Documenta\xE7\xE3o | ChaosSQL",
      description: "Guia do ChaosSQL: instala\xE7\xE3o, a especifica\xE7\xE3o chaos.yaml, invariantes SQL, n\xEDveis de isolamento, classifica\xE7\xE3o de anomalias de Adya e minimiza\xE7\xE3o por delta debugging."
    }
  },
  scenarios: {
    path: "/scenarios",
    indexable: true,
    image: "/og/scenarios.png",
    en: {
      title: "SQL Concurrency Anomaly Scenarios | ChaosSQL",
      description: "Canonical SQL concurrency anomalies (lost update, write skew, G2, deadlock and more) with invariants and deterministic, seed-based reproduction."
    },
    pt: {
      title: "Cen\xE1rios de anomalias de concorr\xEAncia SQL | ChaosSQL",
      description: "Anomalias can\xF4nicas de concorr\xEAncia SQL (lost update, write skew, G2, deadlock e outras) com invariantes e reprodu\xE7\xE3o determin\xEDstica por seed."
    }
  },
  visualizer: {
    path: "/visualizer",
    indexable: true,
    image: "/og/scenarios.png",
    en: {
      title: "Trace Visualizer | ChaosSQL",
      description: "Inspect interleaved concurrent transactions, per-worker timings and the delta-debugged minimal trace behind a SQL anomaly."
    },
    pt: {
      title: "Trace Visualizer | ChaosSQL",
      description: "Inspecione transa\xE7\xF5es concorrentes intercaladas, tempos por worker e o trace m\xEDnimo, reduzido por delta debugging, por tr\xE1s de uma anomalia SQL."
    }
  },
  matrix: {
    path: "/matrix",
    indexable: true,
    image: "/og/scenarios.png",
    en: {
      title: "Hermitage Isolation Level Matrix | ChaosSQL",
      description: "Which concurrency anomalies each isolation level allows in PostgreSQL, MySQL and SQLite, inspired by the Hermitage project."
    },
    pt: {
      title: "Matriz Hermitage de n\xEDveis de isolamento | ChaosSQL",
      description: "Quais anomalias de concorr\xEAncia cada n\xEDvel de isolamento permite no PostgreSQL, MySQL e SQLite, inspirado no projeto Hermitage."
    }
  },
  playground: {
    path: "/playground",
    indexable: true,
    image: "/og/playground.png",
    en: {
      title: "WASM Playground: Test SQL Concurrency in Your Browser | ChaosSQL",
      description: "Run the ChaosSQL concurrency fuzzer in your browser with WebAssembly and reproduce lost updates, write skew and deadlocks without installing anything."
    },
    pt: {
      title: "Playground WASM: teste concorr\xEAncia SQL no navegador | ChaosSQL",
      description: "Rode o fuzzer de concorr\xEAncia ChaosSQL no navegador com WebAssembly e reproduza lost update, write skew e deadlocks sem instalar nada."
    }
  },
  pricing: {
    path: "/pricing",
    indexable: true,
    image: "/og/pricing.png",
    en: {
      title: "Pricing | ChaosSQL Cloud and Concurrency Audits",
      description: "ChaosSQL Cloud plans for CI concurrency regression testing, plus one-week database concurrency audits by Studio Bregalda."
    },
    pt: {
      title: "Pre\xE7os | ChaosSQL Cloud e auditorias de concorr\xEAncia",
      description: "Planos do ChaosSQL Cloud para testes de regress\xE3o de concorr\xEAncia no CI, al\xE9m de auditorias de concorr\xEAncia de uma semana pelo Studio Bregalda."
    }
  },
  dashboard: {
    path: "/dashboard",
    indexable: false,
    image: "/og/home.png",
    en: {
      title: "Cloud Dashboard | ChaosSQL",
      description: "ChaosSQL Cloud dashboard for CI runs, concurrency regressions and alerts."
    },
    pt: {
      title: "Cloud Dashboard | ChaosSQL",
      description: "Painel do ChaosSQL Cloud para execu\xE7\xF5es de CI, regress\xF5es de concorr\xEAncia e alertas."
    }
  }
};
function seoRouteForPath(path) {
  const clean = path.replace(/\/+$/, "") || "/";
  for (const [id, route] of Object.entries(SEO_ROUTES)) {
    if (route.path === clean) return id;
  }
  return null;
}
function pageMeta(route, lang, path, override) {
  const seo = SEO_ROUTES[route];
  const clean = route === "landing" ? "/" : path.replace(/\/+$/, "") || seo.path;
  const url = (l) => {
    const localized = localizePath(clean, l);
    return `${SITE_ORIGIN}${localized === "/" ? "/" : localized}`;
  };
  return {
    lang,
    htmlLang: lang === "pt" ? "pt-BR" : "en",
    ogLocale: lang === "pt" ? "pt_BR" : "en_US",
    ogLocaleAlternate: lang === "pt" ? "en_US" : "pt_BR",
    title: override?.title ?? seo[lang].title,
    description: override?.description ?? seo[lang].description,
    robots: seo.indexable ? "index, follow, max-image-preview:large" : "noindex, follow",
    canonical: url(lang),
    image: `${SITE_ORIGIN}${seo.image}`,
    alternates: { en: url("en"), "pt-BR": url("pt"), "x-default": url("en") }
  };
}

// ../worker.ts
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
  webhookTest: { max: 10, windowMs: 6e4 },
  siteEvent: { max: 60, windowMs: 6e4 }
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
var SCENARIO_PATH = /^\/scenarios\/[a-z0-9-]+$/;
function appPath(pathname) {
  const { lang, path } = splitLocale(pathname);
  if (path === "/" || seoRouteForPath(path) !== null || SCENARIO_PATH.test(path)) return { lang, path };
  return null;
}
var prerendered = null;
function prerenderAsset(pathname) {
  return `/prerender${pathname === "/" ? "/home" : pathname}`;
}
function prerenderedPaths(request, env) {
  prerendered ??= env.ASSETS.fetch(new Request(new URL("/prerender/manifest.json", request.url))).then((res) => res.ok ? res.json() : []).then((paths) => new Set(Array.isArray(paths) ? paths.filter((p) => typeof p === "string") : [])).catch(() => {
    prerendered = null;
    return /* @__PURE__ */ new Set();
  });
  return prerendered;
}
var SITE_EVENTS = /* @__PURE__ */ new Set([
  "cta_click",
  "install_copy",
  "command_copy",
  "scenario_view",
  "lead_submit",
  "plan_select",
  "pricing_toggle",
  "outbound_click",
  "lang_switch"
]);
var EVENT_LABEL = /^[a-z0-9_:.-]{0,64}$/;
var EVENT_PATH = /^\/[\w\-/.]{0,127}$/;
var EVENT_REF = /^[a-z0-9.-]{0,64}$/;
var MAX_EVENT_BYTES = 1024;
async function handleSiteEvent(request, env) {
  if (!isAllowedOrigin(request)) {
    return jsonResponse(request, { error: "Origin not allowed" }, 403);
  }
  const raw = await request.text();
  if (raw.length > MAX_EVENT_BYTES) {
    return jsonResponse(request, { error: "Payload too large" }, 413);
  }
  let data;
  try {
    data = JSON.parse(raw);
  } catch {
    return jsonResponse(request, { error: "Invalid JSON body" }, 400);
  }
  const event = typeof data.event === "string" ? data.event : "";
  const label = typeof data.label === "string" ? data.label : "";
  const path = typeof data.path === "string" ? data.path : "/";
  const ref = typeof data.ref === "string" ? data.ref.toLowerCase() : "";
  const lang = data.lang === "pt" ? "pt" : "en";
  if (!SITE_EVENTS.has(event) || !EVENT_LABEL.test(label) || !EVENT_PATH.test(path) || !EVENT_REF.test(ref)) {
    return jsonResponse(request, { error: "Invalid event" }, 400);
  }
  if (isRateLimited(clientKey(request, "site-event"), RATE_LIMITS.siteEvent)) {
    return new Response(null, { status: 204, headers: corsHeaders(request) });
  }
  const country = request.cf?.country ?? "";
  try {
    env.SITE_EVENTS?.writeDataPoint({
      indexes: [event],
      blobs: [event, label, path, lang, ref, country],
      doubles: [1]
    });
  } catch {
    console.error("Failed to write site event");
  }
  return new Response(null, { status: 204, headers: corsHeaders(request) });
}
var CF_BEACON_TOKEN = /^[a-f0-9]{32}$/;
function withWebAnalytics(response, env) {
  const token = env.CF_WEB_ANALYTICS_TOKEN?.trim();
  const isHtml = (response.headers.get("Content-Type") || "").includes("text/html");
  if (!token || !CF_BEACON_TOKEN.test(token) || !isHtml) return response;
  const beacon = `<script defer src="https://static.cloudflareinsights.com/beacon.min.js" data-cf-beacon='${JSON.stringify({ token, spa: true })}'></script>`;
  return new HTMLRewriter().on("body", { element: (el) => {
    el.append(beacon, { html: true });
  } }).transform(response);
}
async function serveAppShell(request, env, lang, path) {
  const scenarioPage = SCENARIO_PATH.test(path);
  const route = scenarioPage ? "scenarios" : seoRouteForPath(path) ?? "landing";
  const meta = pageMeta(route, lang, path);
  const shell = await env.ASSETS.fetch(new Request(new URL("/", request.url), request));
  if (!shell.ok) return shell;
  const rewritten = new HTMLRewriter().on("html", { element: (el) => {
    el.setAttribute("lang", meta.htmlLang);
  } }).on("title", { element: (el) => {
    el.setInnerContent(meta.title);
  } }).on('meta[name="description"]', { element: (el) => {
    el.setAttribute("content", meta.description);
  } }).on('meta[name="robots"]', {
    // An unknown scenario slug is a 404 and must not be indexed.
    element: (el) => {
      el.setAttribute("content", scenarioPage ? "noindex, follow" : meta.robots);
    }
  }).on('link[rel="canonical"]', { element: (el) => {
    el.setAttribute("href", meta.canonical);
  } }).on('meta[property="og:url"]', { element: (el) => {
    el.setAttribute("content", meta.canonical);
  } }).on('meta[property="og:title"]', { element: (el) => {
    el.setAttribute("content", meta.title);
  } }).on('meta[property="og:description"]', { element: (el) => {
    el.setAttribute("content", meta.description);
  } }).on('meta[property="og:locale"]', { element: (el) => {
    el.setAttribute("content", meta.ogLocale);
  } }).on('meta[property="og:locale:alternate"]', { element: (el) => {
    el.setAttribute("content", meta.ogLocaleAlternate);
  } }).on('meta[property="og:image"]', { element: (el) => {
    el.setAttribute("content", meta.image);
  } }).on('meta[name="twitter:title"]', { element: (el) => {
    el.setAttribute("content", meta.title);
  } }).on('meta[name="twitter:description"]', { element: (el) => {
    el.setAttribute("content", meta.description);
  } }).on('meta[name="twitter:image"]', { element: (el) => {
    el.setAttribute("content", meta.image);
  } }).on('link[rel="alternate"][hreflang]', {
    element: (el) => {
      const href = meta.alternates[el.getAttribute("hreflang")];
      if (href) el.setAttribute("href", href);
    }
  }).transform(shell);
  if (!scenarioPage) return rewritten;
  return new Response(rewritten.body, { status: 404, headers: rewritten.headers });
}
var worker_default = {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname === "/api/event") {
      if (request.method === "OPTIONS") return preflight(request);
      if (request.method === "POST") return handleSiteEvent(request, env);
      return new Response("Method Not Allowed", { status: 405, headers: corsHeaders(request) });
    }
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
          if (data.name !== void 0 && typeof data.name !== "string") {
            return jsonResponse(request, { error: "Name must be a string" }, 400);
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
          const wantAudit = data.wantAudit === true || data.plan?.toLowerCase() === "audit";
          const auditText = wantAudit ? "\u{1F6E1}\uFE0F **SIM (VIP Advisory)**" : "N\xE3o solicitado";
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
                  { name: "\u{1F464} Nome", value: data.name?.trim().slice(0, 128) || "N\xE3o informado", inline: true },
                  { name: "\u{1F4E7} E-mail", value: data.email.trim(), inline: true },
                  { name: "\u{1F3E2} Empresa / Repo", value: data.company?.trim().slice(0, 128) || "N\xE3o informada", inline: true },
                  { name: "\u{1F5C4}\uFE0F Banco Principal", value: data.database?.slice(0, 64) || "PostgreSQL", inline: true },
                  { name: "\u{1F6E1}\uFE0F Concurrency Audit", value: auditText, inline: true },
                  { name: "Plan", value: data.plan?.trim().slice(0, 128) || "Not selected", inline: true },
                  { name: "Billing cycle", value: data.billingCycle?.trim().slice(0, 32) || "Not selected", inline: true },
                  { name: "Timeline", value: data.timeline?.trim().slice(0, 128) || "Not provided", inline: true },
                  { name: "\u{1F4CD} Origem", value: sourceText, inline: true },
                  { name: "\u{1F4DD} Desafio / Caso de Uso", value: data.notes?.trim().slice(0, 1e3) || "Nenhum detalhe adicional informado.", inline: false }
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
                name: data.name?.trim().slice(0, 128) ?? "",
                email: data.email.trim(),
                wantAudit,
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
                    { name: "Dashboard", value: "[Visualizar Trace & Repro \u2794](https://chaossql.bregalda.com/dashboard)", inline: false }
                  ],
                  footer: {
                    text: "ChaosSQL Concurrency Intelligence Engine \u2022 v1.6.0",
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
    if (request.method === "GET" || request.method === "HEAD") {
      const trimmed = url.pathname.replace(/\/+$/, "");
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
          return new Response("Service Unavailable", { status: 503, headers: { "Content-Type": "text/plain" } });
        }
      }
    }
    try {
      return await env.ASSETS.fetch(request);
    } catch {
      return new Response("Not Found", { status: 404, headers: { "Content-Type": "text/plain" } });
    }
  }
};
export {
  worker_default as default
};
