import type { RouteId } from './router';
import { pageMeta, SEO_ROUTES, SITE_ORIGIN, splitLocale, type Lang, type PageMeta } from './seo';

export { SITE_ORIGIN };

// Per-route metadata lives in seo.ts, shared with the prerender and the worker.
export const ROUTE_META = SEO_ROUTES;

type Copy = { title: string; description: string };
type MetaOverride = (pathname: string, lang: Lang) => Copy | undefined;

// Detail pages (/scenarios/<slug>, /docs/<chapter>) take their title from
// their content files. Those files are large, so they must stay in the lazy
// page chunks: each page module registers its override when it loads (see
// src/lib/detail-meta.ts), and this module never imports content data.
const overrides: Partial<Record<RouteId, MetaOverride>> = {};

export function registerMetaOverride(route: RouteId, override: MetaOverride): void {
  overrides[route] = override;
}

/** Shortens a plain-text summary to a meta description. */
export function metaDescription(text: string): string {
  const plain = text.replace(/<[^>]+>/g, '').replace(/\s+/g, ' ').trim();
  return plain.length > 160 ? `${plain.slice(0, 157).trimEnd()}...` : plain;
}

/** Metadata for a full pathname ("/pt/scenarios/write-skew"). */
export function metaForPath(route: RouteId, pathname: string): PageMeta {
  const { lang, path } = splitLocale(pathname);
  return pageMeta(route, lang, path, overrides[route]?.(pathname, lang));
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
