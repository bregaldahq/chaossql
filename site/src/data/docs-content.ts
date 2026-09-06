import docsJson from './docs.json';

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

export const DOCS_DATA = docsJson as unknown as DocsData;

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
