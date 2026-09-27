import { useEffect, useState } from 'react';
import { X } from 'lucide-react';
import { prefersPortuguese, setStoredLanguage } from '../../lib/i18n';
import { track } from '../../lib/analytics';
import styles from './LanguageHint.module.css';

/**
 * Offers the Portuguese version to visitors whose browser prefers Portuguese.
 * A suggestion, never a redirect: crawlers and shared links must always get
 * the language in the URL. Hidden until mounted, so prerendered HTML matches.
 */
export function LanguageHint({ lang, ptHref }: { lang: 'en' | 'pt'; ptHref: string }) {
  const [show, setShow] = useState(false);
  useEffect(() => {
    setShow(lang === 'en' && prefersPortuguese());
  }, [lang]);
  if (!show) return null;

  return (
    <div className={styles.hint} lang="pt-BR" role="region" aria-label="Idioma">
      <p>
        Esta página também está em português.{' '}
        <a
          href={ptHref}
          data-lang="pt"
          onClick={() => {
            setStoredLanguage('pt');
            track('lang_switch', 'pt_hint');
          }}
        >
          Ver em português
        </a>
      </p>
      <button
        type="button"
        aria-label="Continuar em inglês"
        onClick={() => {
          setStoredLanguage('en');
          setShow(false);
        }}
      >
        <X size={16} aria-hidden="true" />
      </button>
    </div>
  );
}
