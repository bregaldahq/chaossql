import { useEffect, useState } from 'react';
import { localizePath, splitLocale } from './seo';

// Path-based routing (/docs, /playground, ...) so every section is a real,
// crawlable URL. English lives at the root and Portuguese under /pt, so each
// language is its own indexable page. Hash URLs (#/docs) from older links are
// migrated on load.

export const ROUTE_PATHS = {
  landing: '/',
  dashboard: '/dashboard',
  docs: '/docs',
  scenarios: '/scenarios',
  visualizer: '/visualizer',
  matrix: '/matrix',
  playground: '/playground',
  pricing: '/pricing',
} as const;

export type RouteId = keyof typeof ROUTE_PATHS;

const LOCATION_CHANGE = 'chaossql:locationchange';

export { localizePath, splitLocale } from './seo';

/** Scenario slug from "/scenarios/lost-update" (any language), or null. */
export function scenarioSlugFromPath(pathname: string): string | null {
  const match = /^\/scenarios\/([a-z0-9-]+)\/?$/.exec(splitLocale(pathname).path);
  return match ? match[1] : null;
}

export function routeFromPath(pathname: string): RouteId {
  const segment = splitLocale(pathname).path.replace(/\/+$/, '').split('/')[1] || '';
  const match = (Object.keys(ROUTE_PATHS) as RouteId[]).find(
    (id) => id !== 'landing' && ROUTE_PATHS[id] === `/${segment}`
  );
  return match ?? 'landing';
}

// Converts a legacy hash URL ("#/docs?chapter=x") into its path equivalent
// ("/docs?chapter=x"). Returns null when the hash is not a route.
export function pathFromLegacyHash(hash: string): string | null {
  if (!hash.startsWith('#/')) return null;
  return hash.slice(1);
}

export function migrateLegacyHash(): void {
  if (typeof window === 'undefined') return;
  const path = pathFromLegacyHash(window.location.hash);
  if (path !== null) window.history.replaceState(null, '', path);
}

/** Scrolls to the element named by a #hash. Returns false when it is not rendered yet. */
export function scrollToHash(hash: string): boolean {
  const id = decodeURIComponent(hash.replace(/^#/, ''));
  const target = id ? document.getElementById(id) : null;
  if (!target) return false;
  const reduce = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;
  target.scrollIntoView?.({ behavior: reduce ? 'auto' : 'smooth', block: 'start' });
  return true;
}

export function navigate(to: string, { replace = false } = {}): void {
  const url = new URL(to, window.location.origin);
  const samePage = url.pathname === window.location.pathname && url.search === window.location.search;
  if (samePage && url.hash) {
    // In-page anchor: update the URL and scroll, without a route change.
    if (url.hash !== window.location.hash) window.history.pushState(null, '', to);
    scrollToHash(url.hash);
    return;
  }
  if (samePage) return;
  if (replace) window.history.replaceState(null, '', to);
  else window.history.pushState(null, '', to);
  window.dispatchEvent(new Event(LOCATION_CHANGE));
}

function snapshot() {
  return { pathname: window.location.pathname, search: window.location.search };
}

// Location used while prerendering on the server, where there is no window.
let serverLocation = { pathname: '/', search: '' };
export function setServerLocation(pathname: string, search = ''): void {
  serverLocation = { pathname, search };
}

export function useLocation(): { pathname: string; search: string } {
  const [location, setLocation] = useState(() => (typeof window === 'undefined' ? serverLocation : snapshot()));
  useEffect(() => {
    const update = () => setLocation(snapshot());
    window.addEventListener('popstate', update);
    window.addEventListener(LOCATION_CHANGE, update);
    return () => {
      window.removeEventListener('popstate', update);
      window.removeEventListener(LOCATION_CHANGE, update);
    };
  }, []);
  return location;
}

export function useSearchParams(): URLSearchParams {
  return new URLSearchParams(useLocation().search);
}

// Routes same-origin <a href="/..."> clicks through the History API instead of
// a full page load, while leaving modified clicks and new-tab links alone.
export function interceptLinkClick(event: MouseEvent): void {
  if (event.defaultPrevented || event.button !== 0) return;
  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
  const anchor = (event.target as Element | null)?.closest?.('a');
  if (!anchor || (anchor.target && anchor.target !== '_self') || anchor.hasAttribute('download')) return;
  const href = anchor.getAttribute('href');
  if (!href || !href.startsWith('/') || href.startsWith('//')) return;
  const url = new URL(href, window.location.origin);
  const { path } = splitLocale(url.pathname);
  if (routeFromPath(url.pathname) === 'landing' && path !== '/') return;
  event.preventDefault();
  // Unprefixed links from a Portuguese page stay in Portuguese; explicit
  // language switches carry data-lang and are left alone.
  const current = splitLocale(window.location.pathname).lang;
  const unprefixed = splitLocale(url.pathname).lang === 'en';
  const target = current === 'pt' && unprefixed && !anchor.hasAttribute('data-lang') ? localizePath(href, 'pt') : href;
  const resolved = new URL(target, window.location.origin);
  navigate(resolved.pathname + resolved.search + resolved.hash);
}
