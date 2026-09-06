import styles from './SiteFooter.module.css';

export interface SiteFooterProps {
  lang?: 'pt' | 'en';
}

export function SiteFooter({ lang = 'pt' }: SiteFooterProps) {
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

        <nav aria-label="Links de rodapé" className={styles.metaLinks}>
          <a href="#/docs">{lang === 'pt' ? 'Documentação' : 'Documentation'}</a>
          <a href="#/scenarios">{lang === 'pt' ? 'Cenários' : 'Scenarios'}</a>
          <a href="#/playground">Playground WASM</a>
          <a
            href="https://github.com/bregaldahq/chaossql"
            target="_blank"
            rel="noreferrer"
          >
            GitHub ↗
          </a>
        </nav>

        <p className={styles.copyright}>
          © {new Date().getFullYear()} Studio Bregalda. All rights reserved.
        </p>
      </div>
    </footer>
  );
}
