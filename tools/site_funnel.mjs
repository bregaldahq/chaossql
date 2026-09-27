#!/usr/bin/env node
// Prints the marketing funnel of chaossql.bregalda.com for a period, from the
// cookieless events worker.ts writes to Workers Analytics Engine
// (dataset chaossql_site_events) and, when a site tag is given, page views
// from Cloudflare Web Analytics.
//
// Usage:
//   CLOUDFLARE_ACCOUNT_ID=... CLOUDFLARE_API_TOKEN=... node tools/site_funnel.mjs [days=7] [--json]
// Optional: CF_WEB_ANALYTICS_SITE_TAG=... to include page views.
// The token needs "Account Analytics: Read".

const DATASET = 'chaossql_site_events';
const API = 'https://api.cloudflare.com/client/v4';

// Funnel stages in order. Each stage counts the events matching `match`.
export const STAGES = [
  { id: 'story', title: 'Opened the lost update story', match: (e) => e.event === 'cta_click' && /(^|_)story$/.test(e.label) },
  { id: 'scenarios', title: 'Switched scenario tabs', match: (e) => e.event === 'scenario_view' },
  { id: 'install', title: 'Copied the install command', match: (e) => e.event === 'install_copy' },
  { id: 'playground', title: 'Opened a scenario in the playground', match: (e) => e.event === 'cta_click' && e.label.startsWith('scenario_playground:') },
  { id: 'pricing', title: 'Selected a plan', match: (e) => e.event === 'plan_select' },
  { id: 'waitlist', title: 'Joined the Cloud waitlist', match: (e) => e.event === 'lead_submit' && e.label.startsWith('waitlist') },
  { id: 'audit', title: 'Requested an audit', match: (e) => e.event === 'lead_submit' && e.label.startsWith('audit') },
];

export function eventsQuery(days) {
  if (!Number.isInteger(days) || days < 1 || days > 90) throw new Error('days must be an integer between 1 and 90');
  return `SELECT blob1 AS event, blob2 AS label, blob4 AS lang, blob5 AS ref, SUM(_sample_interval) AS n
FROM ${DATASET}
WHERE timestamp > NOW() - INTERVAL '${days}' DAY
GROUP BY event, label, lang, ref
FORMAT JSON`;
}

/** Aggregates Analytics Engine rows into funnel stages, languages and referrers. */
export function summarize(rows) {
  const events = rows.map((r) => ({ event: r.event, label: r.label ?? '', lang: r.lang, ref: r.ref ?? '', n: Number(r.n) }));
  const total = (pred) => events.filter(pred).reduce((sum, e) => sum + e.n, 0);
  const stages = STAGES.map((s) => ({ id: s.id, title: s.title, count: total(s.match) }));
  const byLang = { en: total((e) => e.lang === 'en'), pt: total((e) => e.lang === 'pt') };
  const referrers = new Map();
  for (const e of events) if (e.ref) referrers.set(e.ref, (referrers.get(e.ref) ?? 0) + e.n);
  const topReferrers = [...referrers.entries()].sort((a, b) => b[1] - a[1]).slice(0, 5);
  return { stages, byLang, topReferrers, totalEvents: total(() => true) };
}

async function cf(path, { token, body, json = false }) {
  const res = await fetch(`${API}${path}`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}`, ...(json ? { 'Content-Type': 'application/json' } : {}) },
    body,
  });
  if (!res.ok) throw new Error(`Cloudflare API ${path} answered ${res.status}: ${(await res.text()).slice(0, 300)}`);
  return res.json();
}

export async function pageViews({ accountId, token, siteTag, days }) {
  const since = new Date(Date.now() - days * 86400_000).toISOString().slice(0, 10);
  const query = `query($account: String!, $site: String!, $since: Date!) {
    viewer { accounts(filter: { accountTag: $account }) {
      rumPageloadEventsAdaptiveGroups(limit: 1, filter: { siteTag: $site, date_geq: $since }) { count }
    } }
  }`;
  const data = await cf('/graphql', {
    token,
    json: true,
    body: JSON.stringify({ query, variables: { account: accountId, site: siteTag, since } }),
  });
  if (data.errors?.length) throw new Error(`Web Analytics: ${data.errors[0].message}`);
  return data.data?.viewer?.accounts?.[0]?.rumPageloadEventsAdaptiveGroups?.[0]?.count ?? 0;
}

export async function funnel({ accountId, token, days, siteTag }) {
  const result = await cf(`/accounts/${accountId}/analytics_engine/sql`, { token, body: eventsQuery(days) });
  const summary = summarize(result.data ?? []);
  const views = siteTag ? await pageViews({ accountId, token, siteTag, days }) : null;
  return { days, pageViews: views, ...summary };
}

function print(report) {
  const pct = (n) => (report.pageViews ? `${((n / report.pageViews) * 100).toFixed(1)}%` : '');
  console.log(`chaossql.bregalda.com, last ${report.days} days`);
  if (report.pageViews !== null) console.log(`  Page views                              ${String(report.pageViews).padStart(7)}`);
  for (const s of report.stages) console.log(`  ${s.title.padEnd(40)}${String(s.count).padStart(7)}  ${pct(s.count)}`);
  console.log(`  Events by language                      en ${report.byLang.en}, pt ${report.byLang.pt}`);
  if (report.topReferrers.length) {
    console.log('  Top referrers');
    for (const [ref, n] of report.topReferrers) console.log(`    ${ref.padEnd(38)}${String(n).padStart(7)}`);
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const accountId = process.env.CLOUDFLARE_ACCOUNT_ID;
  const token = process.env.CLOUDFLARE_API_TOKEN;
  if (!accountId || !token) {
    console.error('Set CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN (Account Analytics: Read).');
    process.exit(2);
  }
  const days = Number(process.argv.find((a) => /^\d+$/.test(a)) ?? 7);
  const report = await funnel({ accountId, token, days, siteTag: process.env.CF_WEB_ANALYTICS_SITE_TAG });
  if (process.argv.includes('--json')) console.log(JSON.stringify(report, null, 2));
  else print(report);
}
