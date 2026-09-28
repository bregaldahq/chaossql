import { describe, expect, it } from 'vitest';
import index from '../../index.html?raw';

// Guards the committed build: the landing's JavaScript is what every visitor
// downloads first. Content JSON (docs, scenarios) belongs in lazy page chunks.
const bundles = import.meta.glob('../../assets/main-*.js', { query: '?raw', import: 'default', eager: true }) as Record<string, string>;
const mainFile = /assets\/(main-[\w-]+\.js)/.exec(index)?.[1];
const main = mainFile ? bundles[`../../assets/${mainFile}`] ?? '' : '';

async function gzipSize(text: string): Promise<number> {
  const stream = new Blob([text]).stream().pipeThrough(new CompressionStream('gzip'));
  return (await new Response(stream).arrayBuffer()).byteLength;
}

describe('main bundle', () => {
  it('is the file index.html references', () => {
    expect(main.length).toBeGreaterThan(0);
  });

  it('stays under 120 KB gzip', async () => {
    expect(await gzipSize(main)).toBeLessThan(120 * 1024);
  });

  it('carries no docs or scenario content', () => {
    // Phrases that only exist in docs.json and scenarios.json.
    expect(main).not.toContain('Formal Theory & Mathematical Foundation');
    expect(main).not.toContain('Hospital Write Skew');
  });
});
