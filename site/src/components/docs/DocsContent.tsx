import { useRef, type MouseEvent } from 'react';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import { DocChapter } from '../../data/docs-content';
import { messages, type Language } from '../../i18n';
import styles from './DocsContent.module.css';

interface ChapterLink {
  id: string;
  title: string;
  href: string;
}

export interface DocsContentProps {
  chapter: DocChapter;
  prev?: ChapterLink;
  next?: ChapterLink;
  lang?: Language;
}

/** One chapter: title, summary and the chapter HTML from docs.json. */
export function DocsContent({ chapter, prev, next, lang = 'en' }: DocsContentProps) {
  const m = messages[lang];
  const timers = useRef(new WeakMap<Element, ReturnType<typeof setTimeout>>());

  // Every code block in the chapter HTML carries a .copy-code-btn (added in
  // docs-content.ts); one delegated handler copies the block next to it.
  const onClick = async (event: MouseEvent<HTMLDivElement>) => {
    const button = (event.target as Element).closest('.copy-code-btn');
    const code = button?.closest('.code-container')?.querySelector('pre');
    if (!button || !code) return;
    try {
      await navigator.clipboard.writeText(code.textContent ?? '');
    } catch {
      return;
    }
    button.textContent = m.common.copied;
    clearTimeout(timers.current.get(button));
    timers.current.set(button, setTimeout(() => (button.textContent = m.common.copy), 2000));
  };

  return (
    <article className={styles.article} aria-labelledby="chapter-title">
      <p className={styles.category}>{chapter.category}</p>
      <h1 id="chapter-title" className={styles.title}>
        {chapter.title}
      </h1>
      <p className={styles.summary} dangerouslySetInnerHTML={{ __html: chapter.summary }} />
      <div className={styles.prose} onClick={onClick} dangerouslySetInnerHTML={{ __html: chapter.content }} />
      <nav className={styles.pager} aria-label={m.docs.chaptersLabel}>
        {prev ? (
          <a href={prev.href} className={styles.pagerLink} rel="prev">
            <span className={styles.pagerLabel}>
              <ArrowLeft size={14} aria-hidden="true" /> {m.docs.previous}
            </span>
            <span className={styles.pagerTitle}>{prev.title}</span>
          </a>
        ) : (
          <span />
        )}
        {next ? (
          <a href={next.href} className={[styles.pagerLink, styles.pagerNext].join(' ')} rel="next">
            <span className={styles.pagerLabel}>
              {m.docs.next} <ArrowRight size={14} aria-hidden="true" />
            </span>
            <span className={styles.pagerTitle}>{next.title}</span>
          </a>
        ) : (
          <span />
        )}
      </nav>
    </article>
  );
}
