// Turns the small LaTeX subset used in docs.json ($T_1 \xrightarrow{rw} T_2$,
// $$\mathbb{P}(...) \ge \frac{1}{n}$$) into plain HTML with Unicode symbols,
// sub/superscripts and fractions. A math library would add hundreds of KB for
// about a hundred short formulas; this keeps them readable and indexable.

const SYMBOLS: Record<string, string> = {
  to: '→', rightarrow: '→', Rightarrow: '⇒', leftarrow: '←', Leftrightarrow: '⇔', iff: '⟺', mapsto: '↦',
  in: '∈', notin: '∉', subset: '⊂', subseteq: '⊆', supset: '⊃', cup: '∪', cap: '∩', setminus: '∖', emptyset: '∅',
  forall: '∀', exists: '∃', land: '∧', lor: '∨', neg: '¬', bot: '⊥', top: '⊤',
  cdot: '·', times: '×', ast: '∗', star: '⋆', circ: '∘', pm: '±',
  ge: '≥', geq: '≥', le: '≤', leq: '≤', neq: '≠', ne: '≠', approx: '≈', equiv: '≡', sim: '∼', propto: '∝',
  prec: '≺', preceq: '⪯', succ: '≻', succeq: '⪰', ll: '≪', gg: '≫',
  dots: '…', ldots: '…', cdots: '⋯', infty: '∞', sum: '∑', prod: '∏', partial: '∂', nabla: '∇',
  alpha: 'α', beta: 'β', gamma: 'γ', delta: 'δ', Delta: 'Δ', epsilon: 'ε', lambda: 'λ', mu: 'μ', pi: 'π',
  sigma: 'σ', Sigma: 'Σ', tau: 'τ', phi: 'φ', omega: 'ω', Omega: 'Ω', theta: 'θ',
  langle: '⟨', rangle: '⟩', mid: '|', vert: '|', lvert: '|', rvert: '|',
  quad: ' ', qquad: '  ',
};
const BLACKBOARD: Record<string, string> = { N: 'ℕ', Z: 'ℤ', Q: 'ℚ', R: 'ℝ', C: 'ℂ', P: 'ℙ', E: '𝔼' };
const FUNCTIONS = new Set(['log', 'ln', 'exp', 'max', 'min', 'sup', 'inf', 'lim', 'deg', 'det', 'Pr']);

/** Reads one {group} or a single character starting at `i`; returns [text, next index]. */
function readArg(src: string, i: number): [string, number] {
  while (src[i] === ' ') i++;
  if (src[i] !== '{') return [src[i] ?? '', i + 1];
  let depth = 0;
  for (let j = i; j < src.length; j++) {
    if (src[j] === '{') depth++;
    else if (src[j] === '}' && --depth === 0) return [src.slice(i + 1, j), j + 1];
  }
  return [src.slice(i + 1), src.length];
}

const wrapFraction = (part: string) => (/[\s+\-·×]/.test(part) ? `(${part})` : part);

/** Converts one LaTeX formula (without the $ delimiters) to HTML. */
export function texToHtml(src: string): string {
  // Some Portuguese entries were saved with doubled backslashes (\\mathbb).
  src = src.replace(/\\\\(?=[A-Za-z{}%;,:!#&_ ])/g, '\\');
  let out = '';
  let i = 0;
  while (i < src.length) {
    const ch = src[i];
    if (ch === '\\') {
      const name = /^[A-Za-z]+/.exec(src.slice(i + 1))?.[0];
      if (!name) {
        // \{ \} \% \_ \, \; \: \! and similar
        const next = src[i + 1] ?? '';
        out += ',;:! '.includes(next) ? (next === '!' ? '' : ' ') : next;
        i += 2;
        continue;
      }
      i += name.length + 1;
      if (name === 'text' || name === 'mathrm' || name === 'textrm' || name === 'operatorname') {
        const [arg, next] = readArg(src, i);
        out += arg;
        i = next;
      } else if (name === 'mathbf' || name === 'textbf' || name === 'boldsymbol') {
        const [arg, next] = readArg(src, i);
        out += `<strong>${texToHtml(arg)}</strong>`;
        i = next;
      } else if (name === 'mathbb') {
        const [arg, next] = readArg(src, i);
        out += BLACKBOARD[arg] ?? arg;
        i = next;
      } else if (name === 'mathcal' || name === 'mathit' || name === 'textit') {
        const [arg, next] = readArg(src, i);
        out += `<em>${texToHtml(arg)}</em>`;
        i = next;
      } else if (name === 'frac' || name === 'dfrac' || name === 'tfrac') {
        const [num, afterNum] = readArg(src, i);
        const [den, afterDen] = readArg(src, afterNum);
        out += `${wrapFraction(texToHtml(num))}⁄${wrapFraction(texToHtml(den))}`;
        i = afterDen;
      } else if (name === 'xrightarrow' || name === 'xleftarrow') {
        const [label, next] = readArg(src, i);
        out += name === 'xrightarrow' ? ` —<sup>${texToHtml(label)}</sup>→ ` : ` ←<sup>${texToHtml(label)}</sup>— `;
        i = next;
      } else if (name === 'left' || name === 'right' || name === 'big' || name === 'Big' || name === 'bigl' || name === 'bigr') {
        // Sizing only: the delimiter that follows is kept as is.
      } else if (FUNCTIONS.has(name)) {
        out += name;
      } else {
        out += SYMBOLS[name] ?? name;
      }
      continue;
    }
    if (ch === '^' || ch === '_') {
      const [arg, next] = readArg(src, i + 1);
      const tag = ch === '^' ? 'sup' : 'sub';
      out += `<${tag}>${texToHtml(arg)}</${tag}>`;
      i = next;
      continue;
    }
    if (ch === '{' || ch === '}') {
      i++;
      continue;
    }
    out += ch;
    i++;
  }
  return out;
}

/**
 * Replaces $$display$$ and $inline$ formulas in an HTML fragment, leaving
 * <pre> and <code> blocks untouched (shell variables and `${{ }}` stay as is).
 */
export function renderTexInHtml(html: string): string {
  return html
    .split(/(<(?:pre|code)\b[^>]*>[\s\S]*?<\/(?:pre|code)>)/)
    .map((part, index) => {
      if (index % 2 === 1) return part;
      return part
        .replace(/\$\$([\s\S]+?)\$\$/g, (_m, tex: string) => `<span class="math math-block">${texToHtml(tex.trim())}</span>`)
        .replace(/\$([^$\n]+?)\$/g, (_m, tex: string) => `<span class="math">${texToHtml(tex)}</span>`);
    })
    .join('');
}
