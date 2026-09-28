import { useEffect } from 'react';
import { docChapterFromPath, localizePath, navigate, useLocation } from '../lib/router';
import { DOCS_DATA, CHAPTER_ORDER, docChapter, isChapterId } from '../data/docs-content';
import { messages, type Language } from '../i18n';
import { DocsSidebar } from '../components/docs/DocsSidebar';
import { DocsContent } from '../components/docs/DocsContent';
import styles from './DocsPage.module.css';

export interface DocsPageProps {
  lang?: Language;
}

/**
 * /docs is the chapter index; /docs/<chapter> is one indexable page per
 * chapter. Old links with ?chapter=<id> are moved to the chapter URL.
 */
export function DocsPage({ lang = 'en' }: DocsPageProps) {
  const { pathname, search } = useLocation();
  const t = messages[lang].docs;
  const legacy = new URLSearchParams(search).get('chapter');

  useEffect(() => {
    if (isChapterId(legacy)) navigate(localizePath(`/docs/${legacy}`, lang), { replace: true });
  }, [legacy, lang]);

  const chapterId = docChapterFromPath(pathname);
  const chapter = docChapter(chapterId, lang);
  const chapters = DOCS_DATA[lang];
  const index = chapter ? CHAPTER_ORDER.indexOf(chapter.id as (typeof CHAPTER_ORDER)[number]) : -1;
  const link = (id: string) => ({ id, title: chapters[id].title, href: localizePath(`/docs/${id}`, lang) });

  return (
    <div className={styles.page}>
      <div className={styles.layout}>
        <DocsSidebar chapters={chapters} activeChapterId={chapter?.id ?? null} lang={lang} />
        {chapter ? (
          <DocsContent
            chapter={chapter}
            lang={lang}
            prev={index > 0 ? link(CHAPTER_ORDER[index - 1]) : undefined}
            next={index < CHAPTER_ORDER.length - 1 ? link(CHAPTER_ORDER[index + 1]) : undefined}
          />
        ) : (
          <section className={styles.index} aria-labelledby="docs-title">
            <h1 id="docs-title" className={styles.title}>
              {t.title}
            </h1>
            <p className={styles.lead}>{t.lead}</p>
            <ol className={styles.chapters}>
              {CHAPTER_ORDER.map((id) => (
                <li key={id}>
                  <a href={localizePath(`/docs/${id}`, lang)} className={styles.chapterCard}>
                    <span className={styles.category}>{chapters[id].category}</span>
                    <span className={styles.chapterTitle}>{chapters[id].title}</span>
                    <span className={styles.chapterSummary} dangerouslySetInnerHTML={{ __html: chapters[id].summary }} />
                  </a>
                </li>
              ))}
            </ol>
          </section>
        )}
      </div>
    </div>
  );
}
