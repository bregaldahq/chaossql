import { lazy, Suspense, useEffect } from 'react';
import { SiteNav } from './components/ui/SiteNav';
import { SiteFooter } from './components/ui/SiteFooter';
import { LandingPage } from './pages/LandingPage';
import { useI18n } from './lib/i18n';
import { interceptLinkClick, routeFromPath, RouteId, scrollToHash, useLocation } from './lib/router';
import { applyRouteMeta } from './lib/route-meta';

// The landing ships in the main bundle; every other page loads on demand so
// first-time visitors do not download the dashboard, docs or WASM playground.
const DashboardPage = lazy(() => import('./pages/DashboardPage').then((m) => ({ default: m.DashboardPage })));
const PricingPage = lazy(() => import('./pages/PricingPage').then((m) => ({ default: m.PricingPage })));
const DocsPage = lazy(() => import('./pages/DocsPage').then((m) => ({ default: m.DocsPage })));
const ScenariosPage = lazy(() => import('./pages/ScenariosPage').then((m) => ({ default: m.ScenariosPage })));
const MatrixPage = lazy(() => import('./pages/MatrixPage').then((m) => ({ default: m.MatrixPage })));
const VisualizerPage = lazy(() => import('./pages/VisualizerPage').then((m) => ({ default: m.VisualizerPage })));
const PlaygroundPage = lazy(() => import('./pages/PlaygroundPage').then((m) => ({ default: m.PlaygroundPage })));

// Pages rendered in the dark "lab" theme. The rest keep the legacy light look
// until they are migrated.
const LAB_ROUTES: ReadonlySet<RouteId> = new Set(['landing', 'pricing']);

export default function App() {
  const { lang, setLang } = useI18n();
  const { pathname } = useLocation();
  const route = routeFromPath(pathname);
  const lab = LAB_ROUTES.has(route);

  useEffect(() => {
    const root = document.documentElement;
    if (lab) root.dataset.theme = 'lab';
    else delete root.dataset.theme;
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', lab ? '#141021' : '#4B2E83');
  }, [lab]);

  useEffect(() => {
    applyRouteMeta(route, pathname);
    // Land on the requested section (/pricing#audit) or at the top of the new
    // page. Lazy pages render after this effect, so wait briefly for the anchor.
    const hash = window.location.hash;
    if (!hash || !scrollToHash(hash)) window.scrollTo({ top: 0 });
    if (!hash || document.getElementById(hash.slice(1))) return;
    const observer = new MutationObserver(() => {
      if (scrollToHash(hash)) observer.disconnect();
    });
    observer.observe(document.body, { childList: true, subtree: true });
    const timeout = setTimeout(() => observer.disconnect(), 5000);
    return () => {
      observer.disconnect();
      clearTimeout(timeout);
    };
  }, [route, pathname]);

  useEffect(() => {
    document.addEventListener('click', interceptLinkClick);
    return () => document.removeEventListener('click', interceptLinkClick);
  }, []);

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <a className="skip-link" href="#main">
        {lang === 'pt' ? 'Pular para o conteúdo' : 'Skip to content'}
      </a>
      <SiteNav currentRoute={route} lang={lang} onLanguageChange={setLang} />

      <main id="main" style={{ flexGrow: 1 }}>
        <Suspense fallback={<div style={{ minHeight: '60vh' }} aria-busy="true" />}>
          {route === 'landing' && <LandingPage lang={lang} />}
          {route === 'dashboard' && <DashboardPage lang={lang} />}
          {route === 'docs' && <DocsPage lang={lang} />}
          {route === 'scenarios' && <ScenariosPage lang={lang} />}
          {route === 'visualizer' && <VisualizerPage lang={lang} />}
          {route === 'matrix' && <MatrixPage lang={lang} />}
          {route === 'playground' && <PlaygroundPage lang={lang} />}
          {route === 'pricing' && <PricingPage lang={lang} />}
        </Suspense>
      </main>

      <SiteFooter lang={lang} />
    </div>
  );
}
