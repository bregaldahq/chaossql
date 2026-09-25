import { useState } from 'react';
import { Menu, X } from 'lucide-react';
import styles from './SiteNav.module.css';
import { track } from '../../lib/analytics';
import { messages } from '../../i18n';

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
  const t = messages[lang].nav;

  const navItems = [
    { id: 'landing', path: '/' },
    { id: 'dashboard', path: '/dashboard' },
    { id: 'docs', path: '/docs' },
    { id: 'scenarios', path: '/scenarios' },
    { id: 'visualizer', path: '/visualizer' },
    { id: 'matrix', path: '/matrix' },
    { id: 'playground', path: '/playground' },
    { id: 'pricing', path: '/pricing' },
  ] as const;

  const handleNavClick = (id: string) => {
    onRouteChange(id);
    setMobileOpen(false);
  };

  return (
    <>
      <header className={styles.nav} data-nav data-surface="light">
        <a
          href="/"
          onClick={(e) => {
            e.preventDefault();
            handleNavClick('landing');
          }}
          className={styles.brandGroup}
          aria-label={t.homeLabel}
        >
          <div className={styles.wordmark}>
            <img
              src="/brand/bregalda_wordmark.svg"
              alt="Bregalda"
              width="138"
              height="36"
            />
          </div>
          <span className={styles.productTag}>ChaosSQL</span>
        </a>

        <nav aria-label={t.mainLabel} className={styles.desktopLinks}>
          {navItems.map((item) => {
            const isActive = currentRoute === item.id;
            return (
              <a
                key={item.id}
                href={item.path}
                onClick={(e) => {
                  e.preventDefault();
                  handleNavClick(item.id);
                }}
                className={`${styles.navLink} ${isActive ? styles.navLinkActive : ''}`}
                aria-current={isActive ? 'page' : undefined}
              >
                {t.items[item.id]}
              </a>
            );
          })}
        </nav>

        <div className={styles.actionsGroup}>
          <div className={styles.langSwitch} role="group" aria-label={t.languageLabel}>
            <button
              type="button"
              className={`${styles.langBtn} ${lang === 'pt' ? styles.langBtnActive : ''}`}
              onClick={() => {
                onLanguageChange('pt');
                track('lang_switch', 'pt');
              }}
              aria-pressed={lang === 'pt'}
            >
              PT
            </button>
            <button
              type="button"
              className={`${styles.langBtn} ${lang === 'en' ? styles.langBtnActive : ''}`}
              onClick={() => {
                onLanguageChange('en');
                track('lang_switch', 'en');
              }}
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
            onClick={() => track('outbound_click', 'github_nav')}
          >
            GitHub
            <span aria-hidden="true">↗</span>
          </a>

          <button
            type="button"
            className={styles.mobileMenuToggle}
            onClick={() => setMobileOpen(!mobileOpen)}
            aria-label={mobileOpen ? t.closeMenu : t.openMenu}
            aria-expanded={mobileOpen}
          >
            {mobileOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>
      </header>

      {mobileOpen && (
        <div className={styles.mobileDrawerOpen} data-surface="light">
          <nav aria-label={t.mobileLabel}>
            {navItems.map((item) => {
              const isActive = currentRoute === item.id;
              return (
                <a
                  key={item.id}
                  href={item.path}
                  onClick={(e) => {
                    e.preventDefault();
                    handleNavClick(item.id);
                  }}
                  className={`${styles.mobileNavLink} ${isActive ? styles.mobileNavLinkActive : ''}`}
                >
                  <span>{t.items[item.id]}</span>
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
            onClick={() => track('outbound_click', 'github_nav_mobile')}
          >
            {t.githubRepository}
            <span aria-hidden="true">↗</span>
          </a>
        </div>
      )}
    </>
  );
}
