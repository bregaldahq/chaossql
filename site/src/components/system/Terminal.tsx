import { useEffect, useRef, useState } from 'react';
import { Check, Copy } from 'lucide-react';
import { messages, type Language } from '../../i18n';
import styles from './Terminal.module.css';

export interface TerminalProps {
  code: string;
  /** Header label: a filename, or the language when omitted. */
  title?: string;
  language?: string;
  /** Show a "$" prompt before each line (shell commands). */
  prompt?: boolean;
  copyable?: boolean;
  /** Called after a successful copy, e.g. to record an analytics event. */
  onCopy?: () => void;
  lang?: Language;
  className?: string;
}

/**
 * The one code surface of the site (commands, SQL, YAML, Go). Always dark:
 * it carries data-theme="lab" so it renders the same on light legacy pages.
 */
export function Terminal({
  code,
  title,
  language = 'bash',
  prompt = false,
  copyable = true,
  onCopy,
  lang = 'en',
  className,
}: TerminalProps) {
  const t = messages[lang].common;
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => () => clearTimeout(timer.current), []);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(code);
    } catch {
      return;
    }
    onCopy?.();
    setCopied(true);
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setCopied(false), 2000);
  };

  const lines = code.replace(/\n$/, '').split('\n');

  return (
    <div data-theme="lab" className={[styles.terminal, className].filter(Boolean).join(' ')}>
      <div className={styles.header}>
        <span className={styles.title}>{title ?? language}</span>
        {copyable && (
          <button
            type="button"
            className={styles.copy}
            onClick={handleCopy}
            aria-label={copied ? t.copied : `${t.copy}: ${title ?? language}`}
          >
            {copied ? <Check size={13} aria-hidden="true" /> : <Copy size={13} aria-hidden="true" />}
            <span aria-live="polite">{copied ? t.copied : t.copy}</span>
          </button>
        )}
      </div>
      <pre className={styles.body} data-language={language}>
        <code>
          {prompt
            ? lines.map((line, i) => (
                <span key={i} className={styles.line}>
                  <span className={styles.prompt} aria-hidden="true">
                    $
                  </span>
                  {line}
                </span>
              ))
            : code}
        </code>
      </pre>
    </div>
  );
}
