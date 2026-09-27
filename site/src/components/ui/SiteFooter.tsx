import { messages } from '../../i18n';
import styles from './SiteFooter.module.css';

export interface SiteFooterProps {
  lang?: 'pt' | 'en';
}

export function SiteFooter({ lang = 'en' }: SiteFooterProps) {
  const t = messages[lang].footer;
  return (
    <footer className={styles.footer} data-surface="light">
      <div className={styles.footerInner}>
        <div className={styles.lockup}>
          <img
            src="/brand/bregalda_secondary_lockup.svg"
            alt="Bregalda · Software Engineering"
            height="28"
          />
        </div>

        <nav aria-label={t.label} className={styles.metaLinks}>
          <a href="/docs">{t.docs}</a>
          <a href="/scenarios">{t.scenarios}</a>
          <a href="/playground">{t.playground}</a>
          <a
            href="https://github.com/bregaldahq/chaossql"
            target="_blank"
            rel="noreferrer"
          >
            GitHub ↗
          </a>
        </nav>

        <p className={styles.copyright}>
          © {new Date().getFullYear()} Studio Bregalda. {t.rights}
        </p>
      </div>
    </footer>
  );
}
