import { useEffect, useRef, useState } from 'react';
import { Search } from 'lucide-react';
import { DocChapter, CHAPTER_ORDER } from '../../data/docs-content';
import { localizePath } from '../../lib/router';
import { messages, type Language } from '../../i18n';
import styles from './DocsSidebar.module.css';

export interface DocsSidebarProps {
  chapters: Record<string, DocChapter>;
  activeChapterId: string | null;
  lang?: Language;
}

/** Chapter list grouped by category, with a filter. Links are real URLs. */
export function DocsSidebar({ chapters, activeChapterId, lang = 'en' }: DocsSidebarProps) {
  const t = messages[lang].docs;
  const [query, setQuery] = useState('');
  const q = query.trim().toLowerCase();
  const navRef = useRef<HTMLElement>(null);

  // On phones the list is a horizontal row: bring the current chapter into view.
  useEffect(() => {
    const nav = navRef.current;
    const active = nav?.querySelector<HTMLElement>('[aria-current="page"]');
    if (!nav || !active || nav.scrollWidth <= nav.clientWidth) return;
    nav.scrollLeft = active.offsetLeft - (nav.clientWidth - active.offsetWidth) / 2;
  }, [activeChapterId]);

  const visible = CHAPTER_ORDER.filter((id) => {
    const ch = chapters[id];
    return ch && (!q || [ch.title, ch.category, ch.summary].some((text) => text.toLowerCase().includes(q)));
  });
  const categories = new Map<string, string[]>();
  for (const id of visible) {
    const category = chapters[id].category;
    categories.set(category, [...(categories.get(category) ?? []), id]);
  }

  return (
    <aside className={styles.sidebar}>
      <a href={localizePath('/docs', lang)} className={styles.home} aria-current={activeChapterId === null ? 'page' : undefined}>
        {t.allChapters}
      </a>
      <label className={styles.search}>
        <span className="sr-only">{t.search}</span>
        <Search size={14} aria-hidden="true" />
        <input type="search" placeholder={t.search} value={query} onChange={(e) => setQuery(e.target.value)} />
      </label>
      <nav ref={navRef} aria-label={t.chaptersLabel}>
        {visible.length === 0 && <p className={styles.empty}>{t.noResults}</p>}
        {[...categories.entries()].map(([category, ids]) => (
          <div key={category} className={styles.group}>
            <p className={styles.category}>{category}</p>
            <ul>
              {ids.map((id) => (
                <li key={id}>
                  <a
                    href={localizePath(`/docs/${id}`, lang)}
                    className={styles.link}
                    aria-current={activeChapterId === id ? 'page' : undefined}
                  >
                    {chapters[id].title}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  );
}
