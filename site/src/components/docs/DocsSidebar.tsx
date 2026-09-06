import { useState } from 'react';
import { Search } from 'lucide-react';
import { DocChapter, CHAPTER_ORDER, ChapterId } from '../../data/docs-content';
import styles from './DocsSidebar.module.css';

export interface DocsSidebarProps {
  chapters: Record<string, DocChapter>;
  activeChapterId: string;
  onSelectChapter: (id: ChapterId) => void;
  lang?: 'pt' | 'en';
}

export function DocsSidebar({
  chapters,
  activeChapterId,
  onSelectChapter,
  lang = 'pt',
}: DocsSidebarProps) {
  const [query, setQuery] = useState('');

  const filteredChapterIds = CHAPTER_ORDER.filter((id) => {
    const ch = chapters[id];
    if (!ch) return false;
    const matchTitle = ch.title.toLowerCase().includes(query.toLowerCase());
    const matchCat = ch.category.toLowerCase().includes(query.toLowerCase());
    const matchSummary = ch.summary.toLowerCase().includes(query.toLowerCase());
    return matchTitle || matchCat || matchSummary;
  });

  // Agrupar por categorias
  const categories: Record<string, ChapterId[]> = {};
  filteredChapterIds.forEach((id) => {
    const cat = chapters[id]?.category || 'Geral';
    if (!categories[cat]) categories[cat] = [];
    categories[cat].push(id);
  });

  return (
    <aside className={styles.sidebar} aria-label="Navegação da documentação">
      <div className={styles.searchBox}>
        <input
          type="text"
          className={styles.searchInput}
          placeholder={lang === 'pt' ? 'Buscar tópicos (/)...' : 'Search topics (/)...'}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <Search size={14} className={styles.searchIcon} />
      </div>

      <nav className={styles.navGroup}>
        {Object.entries(categories).map(([category, ids]) => (
          <div key={category} style={{ marginBottom: 'var(--space-2)' }}>
            <div className={styles.categoryTitle}>{category}</div>
            {ids.map((id, idx) => {
              const ch = chapters[id];
              const isActive = activeChapterId === id;
              return (
                <a
                  key={id}
                  href={`#/docs?chapter=${id}`}
                  onClick={(e) => {
                    e.preventDefault();
                    onSelectChapter(id);
                  }}
                  className={`${styles.chapterLink} ${isActive ? styles.chapterLinkActive : ''}`}
                >
                  <span className="technical-label" style={{ fontSize: '0.68rem', color: 'var(--text-secondary)' }}>
                    0{idx + 1}
                  </span>
                  <span>{ch.title}</span>
                </a>
              );
            })}
          </div>
        ))}
      </nav>
    </aside>
  );
}
