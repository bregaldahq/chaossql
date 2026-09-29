import { describe, expect, it } from 'vitest';
import docs from './docs.json';

// Guards the docs against claims the code does not back (see site/COPY.md).
type Chapter = { id: string; title: string; summary: string; content: string };
const chapters = (['en', 'pt'] as const).flatMap((lang) =>
  Object.values(docs[lang] as Record<string, Chapter>).map((chapter) => ({ lang, chapter }))
);
const prose = (chapter: Chapter) =>
  [chapter.title, chapter.summary, chapter.content.replace(/<(pre|code)\b[^>]*>[\s\S]*?<\/\1>/g, '')].join('\n');

describe('docs copy', () => {
  it.each(chapters.map(({ lang, chapter }) => [`${lang}/${chapter.id}`, chapter] as const))('%s has no banned claims', (_name, chapter) => {
    const text = prose(chapter);
    expect(text).not.toMatch(/[—–]/);
    expect(text).not.toMatch(/seamless|100% (?:resilient|segura|of false|dos falsos)|free of false negatives|falsos negativos/i);
    expect(text).not.toMatch(/causal closure|fecho causal|Causal[- ](?:Delta|\$ddmin)|Minimiza\S+ Causal|Causal Minimization/i);
    expect(text).not.toMatch(/adopts the <em>Probabilistic|adota o algoritmo|PCT-SQL\):/);
    expect(text).not.toMatch(/\b\d+ (?:milliseconds|milissegundos)|Within milliseconds|milissegundos,|\b\d+ Flags\b|\b\d+ flags\b/i);
    expect(text).not.toMatch(/13[.,]9|&lt; 100ms|under 50ms|&lt; 50ms/i);
  });

  it('only documents flags that exist', () => {
    // --config, --ddmin and --output were once listed for run but were never implemented.
    for (const { lang, chapter } of chapters) {
      expect(chapter.content, `${lang}/${chapter.id}`).not.toMatch(/--(?:config|ddmin|output)\b/);
      expect(chapter.content, `${lang}/${chapter.id}`).not.toMatch(/chaossql run [^\n<]*--duration/);
    }
  });

  it('documents the real exit codes: a violation exits 1 and no trustworthy answer exits 2', () => {
    for (const lang of ['en', 'pt'] as const) {
      const cli = (docs[lang] as Record<string, Chapter>)['cli-reference'].content;
      expect(cli, lang).toMatch(/<code>1<\/code>:[^\n]*<code>--fail-on violation<\/code>/);
      expect(cli, lang).toMatch(/<code>2<\/code>:[^\n]*<code>--fail-on regression<\/code>/);
      expect(cli, lang).not.toMatch(/<code>0<\/code>:[^\n]*<code>violation<\/code>/);
    }
  });

  it('documents every run flag', () => {
    for (const lang of ['en', 'pt'] as const) {
      const cli = (docs[lang] as Record<string, Chapter>)['cli-reference'].content;
      for (const flag of ['--driver', '--dsn', '--isolation', '--fail-on']) {
        expect(cli, `${lang} ${flag}`).toContain(`<td><code>${flag}</code></td>`);
      }
    }
  });
});
