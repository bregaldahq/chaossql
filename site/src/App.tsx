import { useEffect } from 'react';
import { SiteNav } from './components/ui/SiteNav';
import { SiteFooter } from './components/ui/SiteFooter';
import { LandingPage } from './pages/LandingPage';
import { DashboardPage } from './pages/DashboardPage';
import { PricingPage } from './pages/PricingPage';
import { DocsPage } from './pages/DocsPage';
import { ScenariosPage } from './pages/ScenariosPage';
import { MatrixPage } from './pages/MatrixPage';
import { VisualizerPage } from './pages/VisualizerPage';
import { PlaygroundPage } from './pages/PlaygroundPage';
import { useI18n } from './lib/i18n';
import { interceptLinkClick, navigate, routeFromPath, ROUTE_PATHS, RouteId, useLocation } from './lib/router';
import { applyRouteMeta } from './lib/route-meta';

export default function App() {
  const { lang, setLang } = useI18n();
  const { pathname } = useLocation();
  const route = routeFromPath(pathname);

  useEffect(() => {
    applyRouteMeta(route, pathname);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }, [route, pathname]);

  useEffect(() => {
    document.addEventListener('click', interceptLinkClick);
    return () => document.removeEventListener('click', interceptLinkClick);
  }, []);

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <SiteNav
        currentRoute={route}
        lang={lang}
        onLanguageChange={setLang}
        onRouteChange={(r) => navigate(ROUTE_PATHS[r as RouteId])}
      />

      <main style={{ flexGrow: 1 }}>
        {route === 'landing' && <LandingPage lang={lang} />}
        {route === 'dashboard' && <DashboardPage lang={lang} />}
        {route === 'docs' && <DocsPage lang={lang} />}
        {route === 'scenarios' && <ScenariosPage lang={lang} />}
        {route === 'visualizer' && <VisualizerPage lang={lang} />}
        {route === 'matrix' && <MatrixPage lang={lang} />}
        {route === 'playground' && <PlaygroundPage lang={lang} />}
        {route === 'pricing' && <PricingPage lang={lang} />}
      </main>

      <SiteFooter lang={lang} />
    </div>
  );
}
