import { useState, useEffect } from 'react';
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

export function routeFromHash(hash: string): string {
  if (hash.startsWith('#/dashboard')) return 'dashboard';
  if (hash.startsWith('#/docs')) return 'docs';
  if (hash.startsWith('#/scenarios')) return 'scenarios';
  if (hash.startsWith('#/visualizer')) return 'visualizer';
  if (hash.startsWith('#/matrix')) return 'matrix';
  if (hash.startsWith('#/playground')) return 'playground';
  if (hash.startsWith('#/pricing')) return 'pricing';
  return 'landing';
}

function parseRoute(): string {
  if (typeof window === 'undefined') return 'landing';
  return routeFromHash(window.location.hash || '');
}

export default function App() {
  const { lang, setLang } = useI18n();
  const [route, setRoute] = useState<string>(parseRoute);

  useEffect(() => {
    const handleHashChange = () => {
      const nextRoute = parseRoute();
      setRoute(nextRoute);
      window.scrollTo({ top: 0, behavior: 'smooth' });
    };

    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, []);

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <SiteNav
        currentRoute={route}
        lang={lang}
        onLanguageChange={setLang}
        onRouteChange={(r) => {
          setRoute(r);
          window.scrollTo({ top: 0, behavior: 'smooth' });
        }}
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
