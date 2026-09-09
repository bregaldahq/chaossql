import { useState } from 'react';
import { Menu, X } from 'lucide-react';
import styles from './SiteNav.module.css';

export interface SiteNavProps {
  currentRoute: string;
  lang: 'pt' | 'en';
  onLanguageChange: (lang: 'pt' | 'en') => void;
  onRouteChange: (route: string) => void;
}

export function SiteNav({
  currentRoute,
  lang,
  onLanguageChange,
  onRouteChange,
}: SiteNavProps) {
  const [mobileOpen, setMobileOpen] = useState(false);

  const navItems = [
    { id: 'landing', labelPt: 'Início', labelEn: 'Home', path: '#/' },
    { id: 'dashboard', labelPt: 'Cloud Dashboard ●', labelEn: 'Cloud Dashboard ●', path: '#/dashboard' },
    { id: 'docs', labelPt: 'Documentação', labelEn: 'Docs', path: '#/docs' },
    { id: 'scenarios', labelPt: 'Cenários (9)', labelEn: 'Scenarios (9)', path: '#/scenarios' },
    { id: 'visualizer', labelPt: 'Trace Visualizer', labelEn: 'Trace Visualizer', path: '#/visualizer' },
    { id: 'matrix', labelPt: 'Matriz Hermitage', labelEn: 'Hermitage Matrix', path: '#/matrix' },
    { id: 'playground', labelPt: 'Playground WASM', labelEn: 'WASM Playground', path: '#/playground' },
    { id: 'pricing', labelPt: 'Preços', labelEn: 'Pricing', path: '#/pricing' },
  ];

  const handleNavClick = (id: string, path: string) => {
    onRouteChange(id);
    window.location.hash = path;
    setMobileOpen(false);
  };

  return (
    <>
      <header className={styles.nav} data-nav data-surface="light">
        <a
          href="#/"
          onClick={(e) => {
            e.preventDefault();
            handleNavClick('landing', '#/');
          }}
          className={styles.brandGroup}
          aria-label="ChaosSQL by Studio Bregalda — Início"
        >
          <div className={styles.wordmark}>
            <img
              src="/brand/bregalda_wordmark.svg"
              alt="Bregalda"
              width="138"
              height="36"
            />
          </div>
          <span className={styles.productTag}>ChaosSQL v1.4</span>
        </a>

        <nav aria-label="Navegação principal" className={styles.desktopLinks}>
          {navItems.map((item) => {
            const isActive = currentRoute === item.id;
            return (
              <a
                key={item.id}
                href={item.path}
                onClick={(e) => {
                  e.preventDefault();
                  handleNavClick(item.id, item.path);
                }}
                className={`${styles.navLink} ${isActive ? styles.navLinkActive : ''}`}
                aria-current={isActive ? 'page' : undefined}
              >
                {lang === 'pt' ? item.labelPt : item.labelEn}
              </a>
            );
          })}
        </nav>

        <div className={styles.actionsGroup}>
          <div className={styles.langSwitch} role="group" aria-label="Seletor de idioma">
            <button
              type="button"
              className={`${styles.langBtn} ${lang === 'pt' ? styles.langBtnActive : ''}`}
              onClick={() => onLanguageChange('pt')}
              aria-pressed={lang === 'pt'}
            >
              PT
            </button>
            <button
              type="button"
              className={`${styles.langBtn} ${lang === 'en' ? styles.langBtnActive : ''}`}
              onClick={() => onLanguageChange('en')}
              aria-pressed={lang === 'en'}
            >
              EN
            </button>
          </div>

          <a
            href="https://github.com/bregaldahq/chaossql"
            target="_blank"
            rel="noreferrer"
            className={styles.navCta}
          >
            GitHub
            <span aria-hidden="true">↗</span>
          </a>

          <button
            type="button"
            className={styles.mobileMenuToggle}
            onClick={() => setMobileOpen(!mobileOpen)}
            aria-label={mobileOpen ? 'Fechar menu de navegação' : 'Abrir menu de navegação'}
            aria-expanded={mobileOpen}
          >
            {mobileOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>
      </header>

      {mobileOpen && (
        <div className={styles.mobileDrawerOpen} data-surface="light">
          <nav aria-label="Navegação mobile">
            {navItems.map((item) => {
              const isActive = currentRoute === item.id;
              return (
                <a
                  key={item.id}
                  href={item.path}
                  onClick={(e) => {
                    e.preventDefault();
                    handleNavClick(item.id, item.path);
                  }}
                  className={`${styles.mobileNavLink} ${isActive ? styles.mobileNavLinkActive : ''}`}
                >
                  <span>{lang === 'pt' ? item.labelPt : item.labelEn}</span>
                  <span aria-hidden="true">→</span>
                </a>
              );
            })}
          </nav>
          <a
            href="https://github.com/bregaldahq/chaossql"
            target="_blank"
            rel="noreferrer"
            className={styles.mobileCta}
          >
            GitHub Repository
            <span aria-hidden="true">↗</span>
          </a>
        </div>
      )}
    </>
  );
}
