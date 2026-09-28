import docsJson from './docs.json';
import { renderTexInHtml } from '../lib/tex';

export interface DocChapter {
  id: string;
  title: string;
  category: string;
  summary: string;
  content: string;
}

export interface DocsData {
  pt: Record<string, DocChapter>;
  en: Record<string, DocChapter>;
}

const RAW_DOCS = docsJson as unknown as DocsData;

/**
 * Chapter HTML as stored in docs.json, prepared for the page: LaTeX formulas
 * become readable HTML (see lib/tex.ts) and the inline onclick handlers of the
 * legacy copy buttons are removed (DocsContent handles copying).
 */
function prepare(chapter: DocChapter, lang: 'pt' | 'en'): DocChapter {
  const copy = lang === 'pt' ? 'Copiar' : 'Copy';
  const content = renderTexInHtml(chapter.content)
    // Legacy buttons used inline onclick handlers and only some blocks had one:
    // drop them and give every code block the same button.
    .replace(/<button class="copy-code-btn"[^>]*>[\s\S]*?<\/button>\s*/g, '')
    .replace(/<div class="code-container">/g, `<div class="code-container"><button type="button" class="copy-code-btn">${copy}</button>`);
  return { ...chapter, summary: renderTexInHtml(chapter.summary), content };
}

export const DOCS_DATA: DocsData = {
  en: Object.fromEntries(Object.entries(RAW_DOCS.en).map(([id, ch]) => [id, prepare(ch, 'en')])),
  pt: Object.fromEntries(Object.entries(RAW_DOCS.pt).map(([id, ch]) => [id, prepare(ch, 'pt')])),
};

export const CHAPTER_ORDER = [
  'getting-started',
  'dsl-spec',
  'cli-reference',
  'trace-visualizer',
  'cicd-sarif',
  'drivers',
  'go-sdk',
  'academic-theory',
] as const;

export type ChapterId = (typeof CHAPTER_ORDER)[number];

export function isChapterId(id: string | null | undefined): id is ChapterId {
  return !!id && (CHAPTER_ORDER as readonly string[]).includes(id);
}

/** A chapter in the given language, or undefined for an unknown id. */
export function docChapter(id: string | null, lang: 'pt' | 'en'): DocChapter | undefined {
  return isChapterId(id) ? DOCS_DATA[lang][id] : undefined;
}
