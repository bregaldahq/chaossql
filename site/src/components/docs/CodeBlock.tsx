import { Terminal } from '../system/Terminal';
import styles from './CodeBlock.module.css';

export interface CodeBlockProps {
  code: string;
  language?: string;
  filename?: string;
  lang?: 'pt' | 'en';
}

/** Docs-flavored code block: the shared Terminal with document spacing. */
export function CodeBlock({ code, language = 'bash', filename, lang = 'en' }: CodeBlockProps) {
  return <Terminal className={styles.container} code={code} language={language} title={filename} lang={lang} />;
}
