import assert from 'node:assert/strict';
import { test } from 'node:test';
import { eventsQuery, funnel, STAGES, summarize } from './site_funnel.mjs';

// Every label the site sends (site/src) must land in the stage it describes.
const rows = [
  { event: 'cta_click', label: 'hero_story', lang: 'en', ref: 'news.ycombinator.com', n: '40' },
  { event: 'cta_click', label: 'nav_story', lang: 'pt', ref: '', n: '5' },
  { event: 'scenario_view', label: 'inventory', lang: 'en', ref: '', n: '12' },
  { event: 'install_copy', label: 'hero', lang: 'en', ref: 'news.ycombinator.com', n: '9' },
  { event: 'install_copy', label: 'final', lang: 'pt', ref: '', n: '2' },
  { event: 'cta_click', label: 'scenario_playground:banking', lang: 'en', ref: '', n: '6' },
  { event: 'plan_select', label: 'team', lang: 'en', ref: '', n: '3' },
  { event: 'lead_submit', label: 'waitlist', lang: 'en', ref: '', n: '2' },
  { event: 'lead_submit', label: 'waitlist:team', lang: 'en', ref: '', n: '1' },
  { event: 'lead_submit', label: 'audit:soon', lang: 'pt', ref: '', n: '1' },
  { event: 'cta_click', label: 'plans_audit', lang: 'en', ref: '', n: '4' },
];

test('summarize maps site labels to funnel stages', () => {
  const { stages, byLang, topReferrers, totalEvents } = summarize(rows);
  const count = Object.fromEntries(stages.map((s) => [s.id, s.count]));
  assert.deepEqual(count, { story: 45, scenarios: 12, install: 11, playground: 6, pricing: 3, waitlist: 3, audit: 1 });
  assert.deepEqual(byLang, { en: 77, pt: 8 });
  assert.deepEqual(topReferrers, [['news.ycombinator.com', 49]]);
  assert.equal(totalEvents, 85);
  assert.equal(stages.length, STAGES.length);
});

test('eventsQuery reads the site dataset and bounds the period', () => {
  assert.match(eventsQuery(7), /FROM chaossql_site_events/);
  assert.match(eventsQuery(30), /INTERVAL '30' DAY/);
  assert.throws(() => eventsQuery(0));
  assert.throws(() => eventsQuery(365));
  assert.throws(() => eventsQuery(1.5));
});

test('funnel queries Analytics Engine and Web Analytics with the token', async (t) => {
  const calls = [];
  t.mock.method(globalThis, 'fetch', async (url, init) => {
    calls.push({ url, init });
    if (url.endsWith('/analytics_engine/sql')) return Response.json({ data: rows });
    return Response.json({ data: { viewer: { accounts: [{ rumPageloadEventsAdaptiveGroups: [{ count: 900 }] }] } } });
  });
  const report = await funnel({ accountId: 'acc', token: 'tok', days: 7, siteTag: 'site' });
  assert.equal(report.pageViews, 900);
  assert.equal(report.stages[0].count, 45);
  assert.equal(calls[0].url, 'https://api.cloudflare.com/client/v4/accounts/acc/analytics_engine/sql');
  assert.equal(calls[0].init.headers.Authorization, 'Bearer tok');
  assert.equal(calls[1].url, 'https://api.cloudflare.com/client/v4/graphql');
});

test('funnel surfaces API errors instead of printing zeros', async (t) => {
  t.mock.method(globalThis, 'fetch', async () => new Response('forbidden', { status: 403 }));
  await assert.rejects(() => funnel({ accountId: 'acc', token: 'bad', days: 7 }), /answered 403/);
});
