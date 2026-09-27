import type { RouteId } from './router';
import { scenarioSlugFromPath } from './router';
import { pageMeta, SEO_ROUTES, SITE_ORIGIN, splitLocale, type Lang, type PageMeta } from './seo';
import { scenarioBySlug } from '../data/scenarios-data';

export { SITE_ORIGIN };

// Per-route metadata lives in seo.ts, shared with the prerender and the worker.
export const ROUTE_META = SEO_ROUTES;

/** Title and description of a scenario page (/scenarios/<slug>). */
export function scenarioMeta(slug: string | null, lang: Lang): { title: string; description: string } | undefined {
  const scenario = scenarioBySlug(slug);
  if (!scenario) return undefined;
  const summary = (scenario.summary[lang] || scenario.summary.en).replace(/\s+/g, ' ').trim();
  return {
    title: `${scenario.name[lang] || scenario.name.en} (${scenario.code}) | ${lang === 'pt' ? 'Cenários' : 'Scenarios'} | ChaosSQL`,
    description: summary.length > 160 ? `${summary.slice(0, 157).trimEnd()}...` : summary,
  };
}

/** Metadata for a full pathname ("/pt/scenarios/write-skew"). */
export function metaForPath(route: RouteId, pathname: string): PageMeta {
  const { lang, path } = splitLocale(pathname);
  const override = route === 'scenarios' ? scenarioMeta(scenarioSlugFromPath(pathname), lang) : undefined;
  return pageMeta(route, lang, path, override);
}

function setAttr(selector: string, attr: string, value: string) {
  document.head.querySelector(selector)?.setAttribute(attr, value);
}

export function applyRouteMeta(route: RouteId, pathname: string): void {
  const meta = metaForPath(route, pathname);
  document.title = meta.title;
  setAttr('meta[name="description"]', 'content', meta.description);
  setAttr('meta[name="robots"]', 'content', meta.robots);
  setAttr('link[rel="canonical"]', 'href', meta.canonical);
  setAttr('meta[property="og:url"]', 'content', meta.canonical);
  setAttr('meta[property="og:title"]', 'content', meta.title);
  setAttr('meta[property="og:description"]', 'content', meta.description);
  setAttr('meta[property="og:locale"]', 'content', meta.ogLocale);
  setAttr('meta[property="og:locale:alternate"]', 'content', meta.ogLocaleAlternate);
  setAttr('meta[property="og:image"]', 'content', meta.image);
  setAttr('meta[name="twitter:title"]', 'content', meta.title);
  setAttr('meta[name="twitter:description"]', 'content', meta.description);
  setAttr('meta[name="twitter:image"]', 'content', meta.image);
  for (const [hreflang, href] of Object.entries(meta.alternates)) {
    setAttr(`link[rel="alternate"][hreflang="${hreflang}"]`, 'href', href);
  }
}
