// Server entry for build-time prerendering (scripts/prerender.mjs). Renders
// the same <App /> the browser hydrates, for one pathname at a time.
import { StrictMode } from 'react';
import { prerender } from 'react-dom/static';
import App from './App';
import { routeFromPath, setServerLocation } from './lib/router';
import { metaForPath } from './lib/route-meta';
// Detail page titles (scenarios, docs chapters) for the prerendered <head>.
import './lib/detail-meta';

export { SCENARIO_SLUGS } from './data/scenarios-data';
export { CHAPTER_ORDER } from './data/docs-content';
export { SEO_ROUTES, SITE_ORIGIN, localizePath, splitLocale } from './lib/seo';
export { messages } from './i18n';

/** Routes rendered in the dark lab theme; keep in sync with LAB_ROUTES in App.tsx. */
export { LAB_ROUTES } from './App';

export async function renderPage(pathname: string): Promise<{ html: string; meta: ReturnType<typeof metaForPath>; route: string }> {
  setServerLocation(pathname);
  // prerender (unlike renderToString) waits for lazy pages and Suspense.
  const { prelude } = await prerender(
    <StrictMode>
      <App />
    </StrictMode>,
    // Emit every Suspense boundary in place. By default large boundaries are
    // outlined and swapped in by inline scripts, which shows an empty page
    // without JavaScript and shifts the layout when the swap runs.
    { progressiveChunkSize: Number.POSITIVE_INFINITY }
  );
  const html = await new Response(prelude).text();
  const route = routeFromPath(pathname);
  return { html, meta: metaForPath(route, pathname), route };
}
