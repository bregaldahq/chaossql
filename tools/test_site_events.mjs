import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { test } from 'node:test';
import { transform } from '../site/node_modules/esbuild/lib/main.js';

const raw = await readFile(new URL('../worker.ts', import.meta.url), 'utf8');
const { code } = await transform(raw, { loader: 'ts', format: 'esm' });
const worker = (await import(`data:text/javascript;base64,${Buffer.from(code).toString('base64')}`)).default;

let ipCounter = 0;
function post(body, { env = {}, origin = 'https://chaossql.bregalda.com', ip } = {}) {
  const headers = { 'Content-Type': 'application/json', 'CF-Connecting-IP': ip ?? `198.51.100.${++ipCounter}` };
  if (origin) headers.Origin = origin;
  const request = new Request('https://chaossql.bregalda.com/api/event', {
    method: 'POST',
    headers,
    body: typeof body === 'string' ? body : JSON.stringify(body),
  });
  return worker.fetch(request, env);
}

function sink() {
  const points = [];
  return { points, env: { SITE_EVENTS: { writeDataPoint: (p) => points.push(p) } } };
}

const valid = { event: 'cta_click', label: 'hero_playground', path: '/', lang: 'en', ref: 'news.ycombinator.com' };

test('stores an allowlisted event with only the expected fields', async () => {
  const { points, env } = sink();
  const res = await post(valid, { env });
  assert.equal(res.status, 204);
  assert.equal(points.length, 1);
  assert.deepEqual(points[0].indexes, ['cta_click']);
  assert.deepEqual(points[0].blobs.slice(0, 5), ['cta_click', 'hero_playground', '/', 'en', 'news.ycombinator.com']);
  assert.deepEqual(points[0].doubles, [1]);
});

test('never stores the client IP or user agent', async () => {
  const { points, env } = sink();
  await post(valid, { env, ip: '203.0.113.77' });
  assert.ok(!JSON.stringify(points).includes('203.0.113.77'));
});

test('rejects events outside the allowlist', async () => {
  const { points, env } = sink();
  const res = await post({ ...valid, event: 'page_scrape' }, { env });
  assert.equal(res.status, 400);
  assert.equal(points.length, 0);
});

test('rejects free-form labels that could carry personal data', async () => {
  const { points, env } = sink();
  const res = await post({ ...valid, label: 'someone@example.com' }, { env });
  assert.equal(res.status, 400);
  assert.equal(points.length, 0);
});

test('rejects oversized payloads', async () => {
  const { env } = sink();
  const res = await post(JSON.stringify({ ...valid, pad: 'x'.repeat(2048) }), { env });
  assert.equal(res.status, 413);
});

test('rejects foreign origins', async () => {
  const { points, env } = sink();
  const res = await post(valid, { env, origin: 'https://evil.example' });
  assert.equal(res.status, 403);
  assert.equal(points.length, 0);
});

test('accepts events when the dataset binding is missing', async () => {
  const res = await post(valid, { env: {} });
  assert.equal(res.status, 204);
});

test('throttles floods from one client without erroring', async () => {
  const { points, env } = sink();
  for (let i = 0; i < 70; i++) {
    const res = await post(valid, { env, ip: '192.0.2.200' });
    assert.equal(res.status, 204);
  }
  assert.equal(points.length, 60);
});
