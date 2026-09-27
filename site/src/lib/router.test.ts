import { describe, expect, it } from 'vitest';

import { localizePath, pathFromLegacyHash, routeFromPath, scenarioSlugFromPath, splitLocale } from './router';
import { pageMeta } from './seo';
import { SCENARIO_SLUGS, SCENARIOS_DATA } from '../data/scenarios-data';
// @ts-expect-error plain ESM build script without type declarations
import { buildPagesWorker } from '../../scripts/build-pages-worker.mjs';
import { ROUTE_META } from './route-meta';
import { SITE_EVENTS } from './analytics';
import sitemap from '../../sitemap.xml?raw';
import publicSitemap from '../../public/sitemap.xml?raw';
import workerJs from '../../_worker.js?raw';
import workerTs from '../../../worker.ts?raw';
import wranglerToml from '../../../wrangler.toml?raw';

describe('routeFromPath', () => {
  it.each([
    ['/', 'landing'],
    ['/dashboard', 'dashboard'],
    ['/docs', 'docs'],
    ['/docs/', 'docs'],
    ['/scenarios', 'scenarios'],
    ['/visualizer', 'visualizer'],
    ['/matrix', 'matrix'],
    ['/playground', 'playground'],
    ['/pricing', 'pricing'],
    ['/unknown', 'landing'],
    ['/pt', 'landing'],
    ['/pt/pricing', 'pricing'],
    ['/scenarios/lost-update', 'scenarios'],
    ['/pt/scenarios/write-skew', 'scenarios'],
  ])('maps %s to %s', (path, expected) => {
    expect(routeFromPath(path)).toBe(expected);
  });
});

describe('pathFromLegacyHash', () => {
  it.each([
    ['#/docs?chapter=invariants', '/docs?chapter=invariants'],
    ['#/playground', '/playground'],
    ['#/', '/'],
    ['#quickstart', null],
    ['', null],
  ])('maps %s to %s', (hash, expected) => {
    expect(pathFromLegacyHash(hash)).toBe(expected);
  });
});

describe('language URLs', () => {
  it.each([
    ['/', 'en', '/'],
    ['/pricing', 'en', '/pricing'],
    ['/pt', 'pt', '/'],
    ['/pt/pricing', 'pt', '/pricing'],
    ['/ptx', 'en', '/ptx'],
  ])('splits %s', (pathname, lang, path) => {
    expect(splitLocale(pathname)).toEqual({ lang, path });
  });

  it.each([
    ['/', 'pt', '/pt'],
    ['/#story', 'pt', '/pt#story'],
    ['/pricing#audit', 'pt', '/pt/pricing#audit'],
    ['/playground?scenario=banking', 'pt', '/pt/playground?scenario=banking'],
    ['/pt/pricing', 'en', '/pricing'],
    ['/pt', 'en', '/'],
    ['/pt/docs', 'pt', '/pt/docs'],
  ])('localizes %s for %s', (href, lang, expected) => {
    expect(localizePath(href, lang as 'en' | 'pt')).toBe(expected);
  });

  it('reads scenario slugs in both languages', () => {
    expect(scenarioSlugFromPath('/scenarios/lost-update')).toBe('lost-update');
    expect(scenarioSlugFromPath('/pt/scenarios/write-skew')).toBe('write-skew');
    expect(scenarioSlugFromPath('/scenarios')).toBeNull();
  });

  it('gives every page reciprocal hreflang alternates with an English x-default', () => {
    const meta = pageMeta('pricing', 'pt', '/pricing');
    expect(meta.canonical).toBe('https://chaossql.bregalda.com/pt/pricing');
    expect(meta.alternates).toEqual({
      en: 'https://chaossql.bregalda.com/pricing',
      'pt-BR': 'https://chaossql.bregalda.com/pt/pricing',
      'x-default': 'https://chaossql.bregalda.com/pricing',
    });
    expect(pageMeta('landing', 'pt', '/').canonical).toBe('https://chaossql.bregalda.com/pt');
  });

  it('has a unique slug per scenario', () => {
    const slugs = SCENARIOS_DATA.map((s) => SCENARIO_SLUGS[s.id]);
    expect(slugs.every(Boolean)).toBe(true);
    expect(new Set(slugs).size).toBe(slugs.length);
  });
});

describe('sitemap.xml', () => {
  it('matches the public/ copy', () => {
    expect(publicSitemap).toBe(sitemap);
  });

  it.each(Object.entries(ROUTE_META))('lists %s in both languages only when indexable', (_route, meta) => {
    for (const lang of ['en', 'pt'] as const) {
      const loc = `<loc>https://chaossql.bregalda.com${localizePath(meta.path, lang)}</loc>`;
      expect(sitemap.includes(loc)).toBe(meta.indexable);
    }
  });

  it('lists every scenario page with its alternate', () => {
    for (const slug of Object.values(SCENARIO_SLUGS)) {
      expect(sitemap).toContain(`<loc>https://chaossql.bregalda.com/scenarios/${slug}</loc>`);
      expect(sitemap).toContain(`hreflang="pt-BR" href="https://chaossql.bregalda.com/pt/scenarios/${slug}"`);
    }
  });
});

describe('edge worker', () => {
  it('reads page metadata from the shared SEO module', () => {
    expect(workerTs).toContain("from './site/src/lib/seo'");
  });

  it('site/_worker.js is the current build of worker.ts', () => {
    expect(workerJs).toBe(buildPagesWorker());
  });
});

describe('wrangler.toml', () => {
  it('binds static assets as env.ASSETS for worker.ts', () => {
    expect(wranglerToml).toMatch(/^\[assets\][^[]*^binding = "ASSETS"$/m);
  });

  it('has no run_worker_first rule made redundant by a wildcard (wrangler rejects the deploy)', () => {
    const block = wranglerToml.match(/^run_worker_first = \[([^\]]*)\]/m)?.[1] ?? '';
    const rules = [...block.matchAll(/"([^"]+)"/g)].map((m) => m[1]);
    const wildcards = rules.filter((r) => r.endsWith('/*')).map((r) => r.slice(0, -1));
    for (const rule of rules) {
      if (rule.endsWith('/*')) continue;
      for (const prefix of wildcards) expect(rule.startsWith(prefix), `${rule} is covered by ${prefix}*`).toBe(false);
    }
  });

  it('routes API calls to the worker before the assets layer', () => {
    const runWorkerFirst = wranglerToml.match(/^run_worker_first = \[([^\]]*)\]/m)?.[1] ?? '';
    expect(runWorkerFirst).toContain('"/api/*"');
  });
});

describe('site analytics', () => {
  it('worker.ts accepts exactly the events the app sends', () => {
    for (const event of SITE_EVENTS) expect(workerTs).toContain(`'${event}',`);
  });
});

describe('open graph cards', () => {
  const cards = import.meta.glob('../../og/*.png', { query: '?url', import: 'default', eager: true });
  it.each(Object.entries(ROUTE_META))('%s points at a rendered card', (_route, meta) => {
    expect(Object.keys(cards)).toContain(`../..${meta.image}`);
  });
});
