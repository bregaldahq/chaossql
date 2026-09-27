import assert from 'node:assert/strict';
import { test } from 'node:test';
import { build } from '../site/node_modules/esbuild/lib/main.js';

// Records what the worker asks HTMLRewriter to change instead of parsing HTML.
// Each handler runs once per selector; hreflang links run once per language.
class FakeRewriter {
  constructor() {
    this.handlers = [];
  }
  on(selector, handler) {
    this.handlers.push([selector, handler]);
    return this;
  }
  transform(response) {
    const changes = {};
    for (const [selector, handler] of this.handlers) {
      const langs = selector.includes('hreflang') ? ['en', 'pt-BR', 'x-default'] : [null];
      for (const hreflang of langs) {
        const key = hreflang ? `${selector}=${hreflang}` : selector;
        handler.element({
          getAttribute: (name) => (name === 'hreflang' ? hreflang : null),
          setAttribute: (name, value) => (changes[`${key} ${name}`] = value),
          setInnerContent: (value) => (changes[`${key} text`] = value),
          append: () => {},
        });
      }
    }
    return new Response(JSON.stringify(changes), { status: response.status, headers: { 'Content-Type': 'text/html' } });
  }
}
globalThis.HTMLRewriter = FakeRewriter;

const bundled = await build({
  entryPoints: [new URL('../worker.ts', import.meta.url).pathname],
  bundle: true,
  format: 'esm',
  platform: 'neutral',
  write: false,
});
const worker = (await import(`data:text/javascript;base64,${Buffer.from(bundled.outputFiles[0].text).toString('base64')}`)).default;

const PRERENDERED = ['/', '/pt', '/pricing', '/pt/pricing', '/scenarios/lost-update'];
const env = {
  ASSETS: {
    fetch: async (request) => {
      const { pathname } = new URL(request.url);
      if (pathname === '/prerender/manifest.json') return Response.json(PRERENDERED);
      if (pathname === '/') return new Response('<html></html>', { headers: { 'Content-Type': 'text/html' } });
      return new Response(`asset:${pathname}`, { headers: { 'Content-Type': 'text/plain' } });
    },
  },
};

const get = (path) => worker.fetch(new Request(`https://chaossql.bregalda.com${path}`), env);

test('prerendered pages are served at their canonical URL from site/prerender', async () => {
  for (const path of PRERENDERED) {
    const res = await get(path);
    assert.equal(res.status, 200, path);
    assert.equal(await res.text(), `asset:/prerender${path === '/' ? '/home' : path}`);
  }
});

test('other sections get the shell with Portuguese metadata under /pt', async () => {
  const res = await get('/pt/docs');
  assert.equal(res.status, 200);
  const changes = await res.json();
  assert.equal(changes['html lang'], 'pt-BR');
  // The Portuguese title comes from site/src/lib/seo.ts; here it only has to
  // differ from the English one while keeping the brand suffix.
  assert.notEqual(changes['title text'], 'Documentation | ChaosSQL');
  assert.match(changes['title text'], /\| ChaosSQL$/);
  assert.equal(changes['link[rel="canonical"] href'], 'https://chaossql.bregalda.com/pt/docs');
  assert.equal(changes['link[rel="alternate"][hreflang]=en href'], 'https://chaossql.bregalda.com/docs');
  assert.equal(changes['link[rel="alternate"][hreflang]=pt-BR href'], 'https://chaossql.bregalda.com/pt/docs');
  assert.equal(changes['link[rel="alternate"][hreflang]=x-default href'], 'https://chaossql.bregalda.com/docs');
});

test('English sections keep English metadata and indexing rules', async () => {
  const changes = await (await get('/dashboard')).json();
  assert.equal(changes['html lang'], 'en');
  assert.equal(changes['title text'], 'Cloud Dashboard | ChaosSQL');
  assert.equal(changes['meta[name="robots"] content'], 'noindex, follow');
});

test('an unknown scenario slug is a noindex 404', async () => {
  const res = await get('/scenarios/not-a-scenario');
  assert.equal(res.status, 404);
  assert.equal((await res.json())['meta[name="robots"] content'], 'noindex, follow');
});

test('trailing slashes redirect to the canonical path', async () => {
  for (const [from, to] of [['/pricing/', '/pricing'], ['/pt/', '/pt'], ['/pt/docs/', '/pt/docs']]) {
    const res = await get(from);
    assert.equal(res.status, 301, from);
    assert.equal(res.headers.get('Location'), `https://chaossql.bregalda.com${to}`);
  }
});

test('non-page paths fall through to static assets', async () => {
  assert.equal(await (await get('/robots.txt')).text(), 'asset:/robots.txt');
  assert.equal(await (await get('/pt-br-notes')).text(), 'asset:/pt-br-notes');
});
