import { useEffect, useState } from 'react';

// Path-based routing (/docs, /playground, ...) so every section is a real,
// crawlable URL. Hash URLs (#/docs) from older links are migrated on load.

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

export function routeFromPath(pathname: string): RouteId {
  const segment = pathname.replace(/\/+$/, '').split('/')[1] || '';
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

export function navigate(to: string, { replace = false } = {}): void {
  const current = window.location.pathname + window.location.search;
  if (to === current) return;
  if (replace) window.history.replaceState(null, '', to);
  else window.history.pushState(null, '', to);
  window.dispatchEvent(new Event(LOCATION_CHANGE));
}

function snapshot() {
  return { pathname: window.location.pathname, search: window.location.search };
}

export function useLocation(): { pathname: string; search: string } {
  const [location, setLocation] = useState(() =>
    typeof window === 'undefined' ? { pathname: '/', search: '' } : snapshot()
  );
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
  if (routeFromPath(url.pathname) === 'landing' && url.pathname !== '/') return;
  event.preventDefault();
  navigate(url.pathname + url.search);
}
