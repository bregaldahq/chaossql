import { ArrowLeft, ArrowRight } from 'lucide-react';
import { DocChapter, ChapterId } from '../../data/docs-content';
import styles from './DocsContent.module.css';

export interface DocsContentProps {
  chapter: DocChapter;
  prevChapter?: { id: ChapterId; title: string };
  nextChapter?: { id: ChapterId; title: string };
  onNavigate: (id: ChapterId) => void;
  lang?: 'pt' | 'en';
}

export function DocsContent({
  chapter,
  prevChapter,
  nextChapter,
  onNavigate,
}: DocsContentProps) {
  return (
    <article className={styles.contentPane} data-surface="light">
      <div className={styles.articleInner}>
        <div className={styles.breadcrumbs}>
          <span>ChaosSQL</span>
          <span>/</span>
          <span>Docs</span>
          <span>/</span>
          <span>{chapter.category}</span>
        </div>

        <span className={styles.badge}>{chapter.category}</span>

        <h1 className={styles.title}>{chapter.title}</h1>

        <div className={styles.summaryBox}>
          <p style={{ margin: 0 }}>{chapter.summary}</p>
        </div>

        <div
          className={styles.htmlContent}
          dangerouslySetInnerHTML={{ __html: chapter.content }}
        />

        <div className={styles.footerNav}>
          {prevChapter ? (
            <button
              type="button"
              className={styles.navBtn}
              onClick={() => onNavigate(prevChapter.id)}
            >
              <ArrowLeft size={16} />
              <span>{prevChapter.title}</span>
            </button>
          ) : (
            <div />
          )}

          {nextChapter ? (
            <button
              type="button"
              className={styles.navBtn}
              onClick={() => onNavigate(nextChapter.id)}
            >
              <span>{nextChapter.title}</span>
              <ArrowRight size={16} />
            </button>
          ) : (
            <div />
          )}
        </div>
      </div>
    </article>
  );
}
