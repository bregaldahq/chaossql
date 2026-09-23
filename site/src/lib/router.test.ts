import { describe, expect, it } from 'vitest';

import { pathFromLegacyHash, routeFromPath, ROUTE_PATHS } from './router';
import { ROUTE_META } from './route-meta';
import sitemap from '../../sitemap.xml?raw';
import publicSitemap from '../../public/sitemap.xml?raw';
import workerJs from '../../_worker.js?raw';
import workerTs from '../../../worker.ts?raw';

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

describe('sitemap.xml', () => {
  it('matches the public/ copy', () => {
    expect(publicSitemap).toBe(sitemap);
  });

  it.each(Object.keys(ROUTE_PATHS) as (keyof typeof ROUTE_PATHS)[])('lists %s only when indexable', (route) => {
    const loc = `<loc>https://chaossql.bregalda.com${ROUTE_PATHS[route]}</loc>`;
    expect(sitemap.includes(loc)).toBe(ROUTE_META[route].indexable);
  });
});

describe('worker route metadata', () => {
  it.each([
    ['worker.ts', workerTs],
    ['site/_worker.js', workerJs],
  ])('%s injects the same metadata as the app', (_name, source) => {
    for (const [route, meta] of Object.entries(ROUTE_META)) {
      if (route === 'landing') continue;
      const path = ROUTE_PATHS[route as keyof typeof ROUTE_PATHS];
      expect(source.includes(`'${path}':`) || source.includes(`"${path}":`)).toBe(true);
      expect(source).toContain(JSON.stringify(meta.title));
      expect(source).toContain(JSON.stringify(meta.description));
    }
  });
});
