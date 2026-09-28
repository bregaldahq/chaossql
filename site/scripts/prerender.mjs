#!/usr/bin/env node
// Prerenders the public pages in both languages into site/prerender/, writes
// the manifest the edge worker reads, and regenerates the sitemaps.
//
// Runs after the client build (index.html is the built shell) and the SSR
// build of src/entry-server.tsx into dist-ssr/. index.html itself is left as
// the plain shell: the dashboard (nginx) and the self-hosted Go server use it
// as their SPA fallback for every route.

import { mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const SITE = join(dirname(fileURLToPath(import.meta.url)), '..');
const OUT = join(SITE, 'prerender');
const ssr = await import(pathToFileURL(join(SITE, 'dist-ssr', 'entry-server.js')).href);
const { renderPage, SCENARIO_SLUGS, CHAPTER_ORDER, SEO_ROUTES, SITE_ORIGIN, LAB_ROUTES, localizePath, messages } = ssr;

// Pages whose content renders fully without a browser. The visualizer,
// playground and dashboard need client APIs and keep the injected shell.
const PAGES = [
  '/',
  '/pricing',
  '/scenarios',
  '/docs',
  '/matrix',
  ...Object.values(SCENARIO_SLUGS).map((slug) => `/scenarios/${slug}`),
  ...CHAPTER_ORDER.map((id) => `/docs/${id}`),
];
const LANGS = ['en', 'pt'];

const shell = readFileSync(join(SITE, 'index.html'), 'utf8');
const viteManifest = JSON.parse(readFileSync(join(SITE, 'dist', '.vite', 'manifest.json'), 'utf8'));

// Lazy pages ship their CSS and JS in their own chunks. Link them in the page's
// HTML so the prerendered markup is styled on first paint and hydrates sooner.
const PAGE_MODULES = {
  pricing: 'src/pages/PricingPage.tsx',
  scenarios: 'src/pages/ScenariosPage.tsx',
  docs: 'src/pages/DocsPage.tsx',
  matrix: 'src/pages/MatrixPage.tsx',
};

function chunkFiles(key, seen = new Set()) {
  const chunk = viteManifest[key];
  if (!chunk || seen.has(key)) return { css: [], js: [] };
  seen.add(key);
  // The entry chunk (main-*.js/css) is already linked by index.html.
  const own = chunk.isEntry ? { css: [], js: [] } : { css: [...(chunk.css ?? [])], js: [chunk.file] };
  for (const dep of chunk.imports ?? []) {
    const nested = chunkFiles(dep, seen);
    own.css.push(...nested.css);
    own.js.push(...nested.js);
  }
  return own;
}

// Latin subsets of the two self-hosted fonts: preloading them lets the first
// paint use the real fonts instead of swapping after the CSS is parsed.
const entry = Object.values(viteManifest).find((chunk) => chunk.isEntry);
const FONT_PRELOADS = (entry?.assets ?? [])
  .filter((file) => /(geist|jetbrains-mono)-latin-wght-normal-[\w-]+\.woff2$/.test(file))
  .map((file) => `    <link rel="preload" as="font" type="font/woff2" crossorigin href="/${file}">`)
  .join('\n');
if (!FONT_PRELOADS) throw new Error('prerender: latin font files not found in the Vite manifest');

function pageAssetTags(route) {
  const key = PAGE_MODULES[route];
  if (!key) return '';
  if (!viteManifest[key]) throw new Error(`prerender: ${key} is missing from the Vite manifest`);
  const { css, js } = chunkFiles(key);
  return [
    ...[...new Set(css)].map((file) => `    <link rel="stylesheet" crossorigin href="/${file}">`),
    ...[...new Set(js)].map((file) => `    <link rel="modulepreload" crossorigin href="/${file}">`),
  ].join('\n');
}

const escapeAttr = (value) => value.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;');
const escapeText = (value) => value.replace(/&/g, '&amp;').replace(/</g, '&lt;');

/** Sets `attr` on the first tag matching `tagPattern` (the tag may span lines). */
function setAttr(html, tagPattern, attr, value) {
  let found = false;
  const out = html.replace(tagPattern, (tag) => {
    found = true;
    return tag.replace(new RegExp(`${attr}="[^"]*"`), `${attr}="${escapeAttr(value)}"`);
  });
  if (!found) throw new Error(`prerender: no tag matches ${tagPattern}`);
  return out;
}

const meta = (name) => new RegExp(`<meta\\s+name="${name}"[^>]*>`);
const prop = (name) => new RegExp(`<meta\\s+property="${name.replace(/[:]/g, '\\:')}"[^>]*>`);

function faqJsonLd(lang) {
  const items = Object.values(messages[lang].faq.items).map((item) => ({
    '@type': 'Question',
    name: item.q,
    acceptedAnswer: { '@type': 'Answer', text: item.a },
  }));
  return `<script type="application/ld+json">${JSON.stringify({ '@context': 'https://schema.org', '@type': 'FAQPage', mainEntity: items })}</script>`;
}

function documentFor(page) {
  const { html: body, meta: m, route } = page;
  let doc = shell;
  const lab = LAB_ROUTES.has(route);
  doc = doc.replace(/<html lang="[^"]*"/, `<html lang="${m.htmlLang}"${lab ? ' data-theme="lab"' : ''}`);
  doc = doc.replace(/<title>[\s\S]*?<\/title>/, `<title>${escapeText(m.title)}</title>`);
  doc = setAttr(doc, meta('description'), 'content', m.description);
  doc = setAttr(doc, meta('robots'), 'content', m.robots);
  doc = setAttr(doc, meta('theme-color'), 'content', lab ? '#141021' : '#4B2E83');
  doc = setAttr(doc, /<link\s+rel="canonical"[^>]*>/, 'href', m.canonical);
  for (const [hreflang, href] of Object.entries(m.alternates)) {
    doc = setAttr(doc, new RegExp(`<link\\s+rel="alternate"\\s+hreflang="${hreflang}"[^>]*>`), 'href', href);
  }
  doc = setAttr(doc, prop('og:url'), 'content', m.canonical);
  doc = setAttr(doc, prop('og:title'), 'content', m.title);
  doc = setAttr(doc, prop('og:description'), 'content', m.description);
  doc = setAttr(doc, prop('og:locale'), 'content', m.ogLocale);
  doc = setAttr(doc, prop('og:locale:alternate'), 'content', m.ogLocaleAlternate);
  doc = setAttr(doc, prop('og:image'), 'content', m.image);
  doc = setAttr(doc, meta('twitter:title'), 'content', m.title);
  doc = setAttr(doc, meta('twitter:description'), 'content', m.description);
  doc = setAttr(doc, meta('twitter:image'), 'content', m.image);
  doc = doc.replace('</head>', `${FONT_PRELOADS}\n  </head>`);
  const assets = pageAssetTags(route);
  if (assets) doc = doc.replace('</head>', `${assets}\n  </head>`);
  if (route === 'landing') doc = doc.replace('</head>', `    ${faqJsonLd(m.lang)}\n  </head>`);
  const root = /<div id="root">[\s\S]*?<\/div>(\s*<\/body>)/;
  if (!root.test(doc)) throw new Error('prerender: #root not found in index.html');
  return doc.replace(root, (_match, end) => `<div id="root" data-prerendered>${body}</div>${end}`);
}

rmSync(OUT, { recursive: true, force: true });
const written = [];
for (const lang of LANGS) {
  for (const page of PAGES) {
    const pathname = localizePath(page, lang);
    const rendered = await renderPage(pathname);
    const file = join(OUT, `${pathname === '/' ? '/home' : pathname}.html`);
    mkdirSync(dirname(file), { recursive: true });
    writeFileSync(file, documentFor(rendered));
    written.push(pathname);
  }
}
writeFileSync(join(OUT, 'manifest.json'), JSON.stringify(written, null, 2) + '\n');

// Sitemap: every indexable page in both languages, with reciprocal alternates.
const urls = [
  ...Object.values(SEO_ROUTES)
    .filter((route) => route.indexable)
    .map((route) => route.path),
  ...Object.values(SCENARIO_SLUGS).map((slug) => `/scenarios/${slug}`),
  ...CHAPTER_ORDER.map((id) => `/docs/${id}`),
];
const abs = (path) => `${SITE_ORIGIN}${path}`;
const entries = urls.flatMap((path) =>
  LANGS.map((lang) => {
    const alternates = [
      `    <xhtml:link rel="alternate" hreflang="en" href="${abs(localizePath(path, 'en'))}" />`,
      `    <xhtml:link rel="alternate" hreflang="pt-BR" href="${abs(localizePath(path, 'pt'))}" />`,
      `    <xhtml:link rel="alternate" hreflang="x-default" href="${abs(localizePath(path, 'en'))}" />`,
    ].join('\n');
    return `  <url>\n    <loc>${abs(localizePath(path, lang))}</loc>\n${alternates}\n  </url>`;
  })
);
const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">
${entries.join('\n')}
</urlset>
`;
writeFileSync(join(SITE, 'sitemap.xml'), sitemap);
writeFileSync(join(SITE, 'public', 'sitemap.xml'), sitemap);

console.log(`prerender: ${written.length} pages, ${entries.length} sitemap entries`);
