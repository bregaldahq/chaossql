import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { test } from 'node:test';
import { transform } from '../site/node_modules/esbuild/lib/main.js';

const sources = ['worker.ts', 'site/_worker.js', 'functions/api/waitlist.ts', 'site/functions/api/waitlist.ts'];

for (const source of sources) {
  const raw = await readFile(new URL(`../${source}`, import.meta.url), 'utf8');
  const { code } = await transform(raw, { loader: source.endsWith('.ts') ? 'ts' : 'js', format: 'esm' });
  const module = await import(`data:text/javascript;base64,${Buffer.from(code + `\n//# sourceURL=${source}`).toString('base64')}`);
  // A distinct client per request keeps the per-IP rate limit out of these tests.
  let client = 0;
  function submit(payload, env = { DISCORD_WEBHOOK_URL: 'https://discord.example.invalid/test' }) {
    const request = new Request('https://chaossql.bregalda.com/api/waitlist', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'CF-Connecting-IP': `192.0.2.${++client}` },
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

  test(`${source}: email-only waitlist lead is accepted`, async (t) => {
    let delivered;
    t.mock.method(globalThis, 'fetch', async (_url, init) => {
      delivered = JSON.parse(init.body);
      return new Response(null, { status: 204 });
    });
    const response = await submit({ name: undefined, source: 'landing_page' });
    assert.equal(response.status, 200);
    assert.equal((await response.json()).lead.dispatched, true);
    // The name field still renders, with the sink's placeholder instead of a value.
    const [nameField] = delivered.embeds[0].fields;
    assert.ok(nameField.value.length > 0 && nameField.value !== 'undefined');
    assert.ok(delivered.embeds[0].fields.every((field) => typeof field.value === 'string' && field.value.length > 0));
  });

  test(`${source}: non-string name is rejected`, async (t) => {
    t.mock.method(globalThis, 'fetch', async () => { throw new Error('must not send an invalid lead'); });
    const response = await submit({ name: 42 });
    assert.equal(response.status, 400);
  });

  test(`${source}: audit timeline reaches the sink and free text stays under Discord limits`, async (t) => {
    let delivered;
    t.mock.method(globalThis, 'fetch', async (_url, init) => {
      delivered = JSON.parse(init.body);
      return new Response(null, { status: 204 });
    });
    const response = await submit({ plan: 'audit', timeline: 'Launch in 6 weeks', notes: 'x'.repeat(5000), company: 'c'.repeat(500) });
    assert.equal(response.status, 200);
    const fields = delivered.embeds[0].fields;
    assert.ok(fields.some((field) => field.value === 'Launch in 6 weeks'));
    for (const field of fields) assert.ok(field.value.length <= 1024, `${field.name} exceeds 1024 characters`);
  });

  test(`${source}: missing lead sink rejects submission`, async (t) => {
    t.mock.method(globalThis, 'fetch', async () => { throw new Error('must not send without configuration'); });
    const response = await submit({ plan: 'Team' }, {});
    assert.equal(response.status, 503);
  });
}
