import { messages, type Language } from '../../i18n';
import styles from './SiteFooter.module.css';

export interface SiteFooterProps {
  lang?: Language;
}

const GITHUB_URL = 'https://github.com/bregaldahq/chaossql';

export function SiteFooter({ lang = 'en' }: SiteFooterProps) {
  const m = messages[lang];
  const t = m.footer;
  const groups = [
    {
      title: t.product,
      links: [
        { href: '/#story', label: m.nav.howItWorks },
        { href: '/scenarios', label: t.scenarios },
        { href: '/pricing', label: t.pricing },
      ],
    },
    {
      title: t.tools,
      links: [
        { href: '/playground', label: t.playground },
        { href: '/visualizer', label: t.visualizer },
        { href: '/matrix', label: t.matrix },
        { href: '/dashboard', label: t.dashboard },
      ],
    },
    {
      title: t.resources,
      links: [
        { href: '/docs', label: t.docs },
        { href: GITHUB_URL, label: 'GitHub', external: true },
        { href: `${GITHUB_URL}/releases`, label: t.releases, external: true },
      ],
    },
  ];

  return (
    <footer className={styles.footer}>
      <div className={styles.inner}>
        <div className={styles.brand}>
          <a href="/" className={styles.logo}>
            <img src="/brand/icone_bregalda.svg" alt="" width="28" height="28" />
            <span>ChaosSQL</span>
          </a>
          <p>{t.builtBy}</p>
        </div>
        <nav aria-label={t.label} className={styles.groups}>
          {groups.map((group) => (
            <div key={group.title}>
              <h2 className={styles.groupTitle}>{group.title}</h2>
              <ul>
                {group.links.map((link) => (
                  <li key={link.href}>
                    <a href={link.href} {...('external' in link ? { target: '_blank', rel: 'noreferrer' } : {})}>
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </nav>
        <p className={styles.copyright}>
          © {new Date().getFullYear()} Studio Bregalda. {t.rights}
        </p>
      </div>
    </footer>
  );
}
