import { useState } from 'react';
import { Check, Copy } from 'lucide-react';
import styles from './CodeBlock.module.css';

export interface CodeBlockProps {
  code: string;
  language?: string;
  filename?: string;
}

export function CodeBlock({ code, language = 'bash', filename }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className={styles.container} data-surface="dark">
      <div className={styles.header}>
        <span>{filename || language}</span>
        <button
          type="button"
          className={styles.copyBtn}
          onClick={handleCopy}
          aria-label="Copiar código"
        >
          {copied ? (
            <>
              <Check size={12} /> Copiado
            </>
          ) : (
            <>
              <Copy size={12} /> Copiar
            </>
          )}
        </button>
      </div>
      <pre className={styles.codeArea}>
        <code>{code}</code>
      </pre>
    </div>
  );
}
