import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { test } from 'node:test';
import { transform } from '../site/node_modules/esbuild/lib/main.js';

const sources = ['worker.ts', 'site/_worker.js', 'functions/api/waitlist.ts', 'site/functions/api/waitlist.ts'];

for (const source of sources) {
  const raw = await readFile(new URL(`../${source}`, import.meta.url), 'utf8');
  const { code } = await transform(raw, { loader: source.endsWith('.ts') ? 'ts' : 'js', format: 'esm' });
  const module = await import(`data:text/javascript;base64,${Buffer.from(code + `\n//# sourceURL=${source}`).toString('base64')}`);
  function submit(payload, env = { DISCORD_WEBHOOK_URL: 'https://discord.example.invalid/test' }) {
    const request = new Request('https://chaossql.bregalda.com/api/waitlist', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'CF-Connecting-IP': '192.0.2.15' },
      body: JSON.stringify({ name: 'Test User', email: 'test@example.invalid', ...payload }),
    });
    return module.default ? module.default.fetch(request, env) : module.onRequestPost({ request, env });
  }

  test(`${source}: pricing lead preserves plan and billing choice`, async (t) => {
    let delivered;
    t.mock.method(globalThis, 'fetch', async (_url, init) => {
      delivered = JSON.parse(init.body);
      return new Response(null, { status: 204 });
    });
    const response = await submit({ plan: 'Pro', billingCycle: 'annual', wantAudit: false, source: 'pricing_page' });
    assert.equal(response.status, 200);
    const fields = delivered.embeds[0].fields.map((field) => field.value);
    assert.ok(fields.includes('Pro'), 'the requested plan must reach the lead sink');
    assert.ok(fields.includes('annual'), 'the requested billing cycle must reach the lead sink');
    assert.equal((await response.json()).lead.dispatched, true);
  });

  test(`${source}: audit plan carries audit intent`, async (t) => {
    t.mock.method(globalThis, 'fetch', async () => new Response(null, { status: 204 }));
    const response = await submit({ plan: 'audit', billingCycle: 'one-time', source: 'pricing_page' });
    assert.equal(response.status, 200);
    assert.equal((await response.json()).lead.wantAudit, true);
  });

  test(`${source}: failed delivery is not acknowledged as success`, async (t) => {
    t.mock.method(globalThis, 'fetch', async () => new Response('unavailable', { status: 503 }));
    const response = await submit({ plan: 'Team', billingCycle: 'monthly' });
    assert.equal(response.status, 502);
    assert.notEqual((await response.json()).success, true);
  });

  test(`${source}: missing lead sink rejects submission`, async (t) => {
    t.mock.method(globalThis, 'fetch', async () => { throw new Error('must not send without configuration'); });
    const response = await submit({ plan: 'Team' }, {});
    assert.equal(response.status, 503);
  });
}
