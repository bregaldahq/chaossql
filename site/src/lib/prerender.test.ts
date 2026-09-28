import { describe, expect, it } from 'vitest';
import manifest from '../../prerender/manifest.json';
import { SCENARIO_SLUGS } from '../data/scenarios-data';
import { CHAPTER_ORDER } from '../data/docs-content';
import { localizePath, splitLocale, SITE_ORIGIN } from './seo';

// The committed output of scripts/prerender.mjs (run by `npm run build`).
const files = import.meta.glob('../../prerender/**/*.html', { query: '?raw', import: 'default', eager: true }) as Record<
  string,
  string
>;
const fileFor = (path: string) => files[`../../prerender${path === '/' ? '/home' : path}.html`];

describe('prerendered pages', () => {
  it('cover the public pages and every scenario in both languages', () => {
    const pages = [
      '/', '/pricing', '/scenarios', '/docs', '/matrix',
      ...Object.values(SCENARIO_SLUGS).map((s) => `/scenarios/${s}`),
      ...CHAPTER_ORDER.map((id) => `/docs/${id}`),
    ];
    const expected = pages.flatMap((p) => [localizePath(p, 'en'), localizePath(p, 'pt')]);
    expect([...manifest].sort()).toEqual(expected.sort());
  });

  it.each(manifest)('%s carries its own language, canonical URL and content', (path) => {
    const html = fileFor(path);
    expect(html, path).toBeTruthy();
    const { lang } = splitLocale(path);
    expect(html).toMatch(new RegExp(`<html lang="${lang === 'pt' ? 'pt-BR' : 'en'}"`));
    expect(html).toContain(`<link rel="canonical" href="${SITE_ORIGIN}${path}" />`);
    expect(html).toContain('<div id="root" data-prerendered>');
    expect(html).toMatch(/<h1[^>]*>[^<]+/);
    // Suspense boundaries must be inline: outlined ones need JS to appear and shift the layout.
    expect(html).not.toContain('<template id="B:');
  });

  it.each(manifest)('%s renders in the lab theme from the first byte', (path) => {
    expect(fileFor(path)).toMatch(/<html [^>]*data-theme="lab"/);
  });

  it('links the lazy page stylesheet so prerendered markup is styled on first paint', () => {
    expect(fileFor('/pricing')).toMatch(/<link rel="stylesheet" crossorigin href="\/assets\/PricingPage-[\w-]+\.css">/);
    expect(fileFor('/scenarios/lost-update')).toMatch(/<link rel="stylesheet" crossorigin href="\/assets\/ScenariosPage-[\w-]+\.css">/);
  });

  it('gives every docs chapter its own title and rendered formulas', () => {
    const html = fileFor('/docs/academic-theory');
    expect(html).toMatch(/<title>[^<]+\| Docs \| ChaosSQL<\/title>/);
    expect(html).toContain('class="math');
    expect(html).not.toMatch(/\\(?:text|frac|mathbb)\{/);
  });

  it('ships FAQ structured data on the landing only', () => {
    expect(fileFor('/')).toContain('"@type":"FAQPage"');
    expect(fileFor('/pt')).toContain('"@type":"FAQPage"');
    expect(fileFor('/pricing')).not.toContain('FAQPage');
  });
});
