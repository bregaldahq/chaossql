import { useState, useEffect } from 'react';
import { navigate, useSearchParams } from '../lib/router';
import { DOCS_DATA, CHAPTER_ORDER, ChapterId } from '../data/docs-content';
import { DocsSidebar } from '../components/docs/DocsSidebar';
import { DocsContent } from '../components/docs/DocsContent';
import styles from './DocsPage.module.css';

export interface DocsPageProps {
  lang?: 'pt' | 'en';
}

export function DocsPage({ lang = 'pt' }: DocsPageProps) {
  const [activeChapterId, setActiveChapterId] = useState<ChapterId>('getting-started');

  const chapterParam = useSearchParams().get('chapter');
  useEffect(() => {
    if (chapterParam && CHAPTER_ORDER.includes(chapterParam as ChapterId)) {
      setActiveChapterId(chapterParam as ChapterId);
    }
  }, [chapterParam]);

  const handleSelectChapter = (id: ChapterId) => {
    setActiveChapterId(id);
    navigate(`/docs?chapter=${id}`);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const chapters = DOCS_DATA[lang] || DOCS_DATA['pt'];
  const currentChapter = chapters[activeChapterId] || chapters['getting-started'];

  const currentIndex = CHAPTER_ORDER.indexOf(activeChapterId);
  const prevChapterId = currentIndex > 0 ? CHAPTER_ORDER[currentIndex - 1] : undefined;
  const nextChapterId = currentIndex < CHAPTER_ORDER.length - 1 ? CHAPTER_ORDER[currentIndex + 1] : undefined;

  const prevChapter = prevChapterId ? { id: prevChapterId, title: chapters[prevChapterId]?.title || '' } : undefined;
  const nextChapter = nextChapterId ? { id: nextChapterId, title: chapters[nextChapterId]?.title || '' } : undefined;

  return (
    <div className={styles.pageContainer} data-surface="light">
      <div className={styles.docsLayout}>
        <DocsSidebar
          chapters={chapters}
          activeChapterId={activeChapterId}
          onSelectChapter={handleSelectChapter}
          lang={lang}
        />
        <DocsContent
          chapter={currentChapter}
          prevChapter={prevChapter}
          nextChapter={nextChapter}
          onNavigate={handleSelectChapter}
          lang={lang}
        />
      </div>
    </div>
  );
}
