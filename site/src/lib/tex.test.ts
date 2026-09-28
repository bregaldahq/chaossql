import { describe, expect, it } from 'vitest';
import { renderTexInHtml, texToHtml } from './tex';
import { DOCS_DATA } from '../data/docs-content';

describe('texToHtml', () => {
  it.each([
    ['W_0, W_1, \\dots', 'W<sub>0</sub>, W<sub>1</sub>, …'],
    ['T_1 \\xrightarrow{rw} T_2', 'T<sub>1</sub>  —<sup>rw</sup>→  T<sub>2</sub>'],
    ['\\mu\\text{s}', 'μs'],
    ['k^{d-1}', 'k<sup>d-1</sup>'],
    ['\\frac{1}{n \\cdot k}', '1⁄(n · k)'],
    ['\\mathbb{P}(\\text{Detection}) \\ge 0.9', 'ℙ(Detection) ≥ 0.9'],
    ['\\mathbf{95.02\\%}', '<strong>95.02%</strong>'],
    ['C^* \\setminus \\{op\\}', 'C<sup>*</sup> ∖ {op}'],
    ['\\left(1 - x\\right)^{300}', '(1 - x)<sup>300</sup>'],
  ])('renders %s', (tex, html) => {
    expect(texToHtml(tex)).toBe(html);
  });

  it('accepts the doubled backslashes some Portuguese entries use', () => {
    expect(texToHtml('\\\\mathbb{P}(\\\\text{Detecção}) \\\\ge 1')).toBe(texToHtml('\\mathbb{P}(\\text{Detecção}) \\ge 1'));
  });
});

describe('renderTexInHtml', () => {
  it('converts inline and display formulas but never code', () => {
    const html = '<p>Workers $W_0$ and</p>$$x^2$$<pre><code>echo $HOME and $PATH</code></pre><code>${{ secrets.TOKEN }}</code>';
    expect(renderTexInHtml(html)).toBe(
      '<p>Workers <span class="math">W<sub>0</sub></span> and</p><span class="math math-block">x<sup>2</sup></span><pre><code>echo $HOME and $PATH</code></pre><code>${{ secrets.TOKEN }}</code>'
    );
  });

  it('leaves no raw LaTeX in any docs chapter', () => {
    for (const lang of ['en', 'pt'] as const) {
      for (const chapter of Object.values(DOCS_DATA[lang])) {
        const withoutCode = chapter.content.replace(/<(pre|code)\b[^>]*>[\s\S]*?<\/\1>/g, '');
        expect(withoutCode, `${lang}/${chapter.id}`).not.toMatch(/\$[^$\s]|\\(?:text|frac|mathbb|xrightarrow|cdot|dots)\b/);
      }
    }
  });
});
