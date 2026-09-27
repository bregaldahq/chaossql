import { useEffect, useState } from 'react';
import { ArrowUpRight, Menu, X } from 'lucide-react';
import { track } from '../../lib/analytics';
import { messages, type Language } from '../../i18n';
import { Button } from '../system/Button';
import styles from './SiteNav.module.css';

export interface SiteNavProps {
  currentRoute: string;
  lang: Language;
  onLanguageChange: (lang: Language) => void;
}

const GITHUB_URL = 'https://github.com/bregaldahq/chaossql';

/**
 * Public navigation: five destinations, language, GitHub and the primary CTA.
 * Product tools (visualizer, matrix, dashboard) live in the footer. Links are
 * plain hrefs; App's click interceptor routes them without a reload.
 */
export function SiteNav({ currentRoute, lang, onLanguageChange }: SiteNavProps) {
  const m = messages[lang];
  const t = m.nav;
  const [open, setOpen] = useState(false);

  const links = [
    { id: 'how', href: '/#story', label: t.howItWorks },
    { id: 'scenarios', href: '/scenarios', label: t.items.scenarios },
    { id: 'playground', href: '/playground', label: t.items.playground },
    { id: 'docs', href: '/docs', label: t.items.docs },
    { id: 'pricing', href: '/pricing', label: t.items.pricing },
  ];

  useEffect(() => {
    setOpen(false);
  }, [currentRoute]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false);
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open]);

  const langSwitch = (
    <div className={styles.lang} role="group" aria-label={t.languageLabel}>
      {(['en', 'pt'] as const).map((code) => (
        <button
          key={code}
          type="button"
          aria-pressed={lang === code}
          onClick={() => {
            onLanguageChange(code);
            track('lang_switch', code);
          }}
        >
          {code.toUpperCase()}
        </button>
      ))}
    </div>
  );

  return (
    <header className={styles.nav} data-nav>
      <a href="/" className={styles.brand} aria-label={t.homeLabel}>
        <img src="/brand/icone_bregalda.svg" alt="" width="28" height="28" />
        <span>ChaosSQL</span>
      </a>

      <nav aria-label={t.mainLabel} className={styles.links}>
        {links.map((link) => (
          <a
            key={link.id}
            href={link.href}
            className={styles.link}
            aria-current={currentRoute === link.id ? 'page' : undefined}
          >
            {link.label}
          </a>
        ))}
      </nav>

      <div className={styles.actions}>
        {langSwitch}
        <a
          href={GITHUB_URL}
          target="_blank"
          rel="noreferrer"
          className={styles.github}
          onClick={() => track('outbound_click', 'github_nav')}
        >
          GitHub <ArrowUpRight size={14} aria-hidden="true" />
        </a>
        <Button href="/#story" className={styles.cta} onClick={() => track('cta_click', 'nav_story')}>
          {m.cta.primary}
        </Button>
        <button
          type="button"
          className={styles.menuToggle}
          onClick={() => setOpen((o) => !o)}
          aria-expanded={open}
          aria-controls="mobile-nav"
          aria-label={open ? t.closeMenu : t.openMenu}
        >
          {open ? <X size={20} aria-hidden="true" /> : <Menu size={20} aria-hidden="true" />}
        </button>
      </div>

      {open && (
        <div id="mobile-nav" className={styles.drawer}>
          <nav aria-label={t.mobileLabel}>
            {links.map((link) => (
              <a
                key={link.id}
                href={link.href}
                className={styles.drawerLink}
                aria-current={currentRoute === link.id ? 'page' : undefined}
                onClick={() => setOpen(false)}
              >
                {link.label}
              </a>
            ))}
          </nav>
          <div className={styles.drawerFoot}>
            {langSwitch}
            <a
              href={GITHUB_URL}
              target="_blank"
              rel="noreferrer"
              className={styles.github}
              onClick={() => track('outbound_click', 'github_nav_mobile')}
            >
              {t.githubRepository} <ArrowUpRight size={14} aria-hidden="true" />
            </a>
          </div>
        </div>
      )}
    </header>
  );
}
