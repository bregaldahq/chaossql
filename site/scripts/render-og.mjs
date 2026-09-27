#!/usr/bin/env node
// Renders the Open Graph cards (1200x630) in site/og/ with a headless Chromium.
// The swimlane on each card is drawn from the recorded banking trace in
// src/data/traces/, so the preview shows a real lost update, not an illustration.
//
// Usage: npm run og   (set CHROMIUM_PATH to use a specific browser binary)

import { chromium } from 'playwright-core';
import { readFileSync, mkdirSync } from 'node:fs';
import { chromiumPath, fontUrl } from './render-lib.mjs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const SITE = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const OUT = join(SITE, 'og');

export const CARDS = {
  home: { title: "Your tests pass. Your balances don't add up.", kicker: 'Deterministic SQL concurrency fuzzer' },
  scenarios: { title: 'Real concurrency bugs, replayable from a seed.', kicker: 'Anomaly scenarios' },
  docs: { title: 'From chaos.yaml to a minimal failing test.', kicker: 'Documentation' },
  playground: { title: 'Watch a lost update happen in your browser.', kicker: 'WASM playground' },
  pricing: { title: 'Free CLI. Cloud for CI. Audits before launch.', kicker: 'Pricing' },
};

function escapeHtml(text) {
  return text.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]);
}

// Two lanes, one block per statement, in recorded order. The last write of the
// trace is the one that silently overwrites the other transaction: it is marked.
function swimlane(trace) {
  const workers = [...new Set(trace.map((e) => e.worker))];
  const step = 100 / trace.length;
  const isWrite = (e) => /^\s*(UPDATE|INSERT)/i.test(e.sql);
  const lostIndex = trace.findLastIndex((e) => /^\s*UPDATE/i.test(e.sql));
  const lanes = workers
    .map((w, lane) => {
      const blocks = trace
        .map((e, i) => {
          if (e.worker !== w) return '';
          const kind = e.type === 'COMMIT' ? 'commit' : isWrite(e) ? 'write' : e.type === 'BEGIN' ? 'begin' : 'read';
          const lost = i === lostIndex ? ' lost' : '';
          return `<div class="block ${kind}${lost}" style="left:calc(${(i * step).toFixed(3)}% + 3px);width:calc(${step.toFixed(3)}% - 6px)"></div>`;
        })
        .join('');
      return `<div class="lane"><span class="tx">T${lane + 1}</span><div class="track">${blocks}</div></div>`;
    })
    .join('');
  return `<div class="lanes">${lanes}</div>`;
}

function page({ title, kicker }, trace) {
  return `<!doctype html><html><head><meta charset="utf-8"><style>
@font-face { font-family: Geist; src: url(${fontUrl('geist', 'geist-latin-wght-normal.woff2')}) format('woff2'); font-weight: 100 900; }
@font-face { font-family: JBM; src: url(${fontUrl('jetbrains-mono', 'jetbrains-mono-latin-wght-normal.woff2')}) format('woff2'); font-weight: 100 800; }
* { box-sizing: border-box; margin: 0; }
body { width: 1200px; height: 630px; background: #141021; color: #fcfbf8; font-family: Geist; padding: 64px 80px; display: flex; flex-direction: column; }
.top { display: flex; align-items: center; justify-content: space-between; font-family: JBM; font-size: 22px; color: #949197; }
.brand { color: #fcfbf8; font-family: Geist; font-weight: 640; font-size: 30px; letter-spacing: -0.02em; }
h1 { margin-top: 56px; font-size: 68px; line-height: 1.06; font-weight: 640; letter-spacing: -0.035em; max-width: 980px; }
.lanes { margin-top: auto; display: grid; gap: 14px; }
.lane { display: flex; align-items: center; gap: 20px; }
.tx { font-family: JBM; font-size: 20px; color: #949197; width: 40px; }
.track { position: relative; height: 26px; flex: 1; border-bottom: 1px solid #302c3b; }
.block { position: absolute; top: 0; height: 26px; border-radius: 6px; }
.begin { background: #302c3b; } .read { background: #474450; } .write { background: #a58ee8; } .commit { background: #22c55e; }
.lost { background: #f5c400; }
.foot { margin-top: 26px; display: flex; justify-content: space-between; font-family: JBM; font-size: 20px; color: #949197; }
.foot b { color: #f5c400; font-weight: 500; }
</style></head><body>
<div class="top"><span class="brand">ChaosSQL</span><span>${escapeHtml(kicker)}</span></div>
<h1>${escapeHtml(title)}</h1>
${swimlane(trace)}
<div class="foot"><span>chaossql.bregalda.com</span><span><b>lost update</b> caught and shrunk to 2 transactions</span></div>
</body></html>`;
}

const trace = JSON.parse(readFileSync(join(SITE, 'src/data/traces/banking_lost_update.json'), 'utf8')).minimalTrace;
mkdirSync(OUT, { recursive: true });
const browser = await chromium.launch({ executablePath: chromiumPath() });
try {
  const tab = await browser.newPage({ viewport: { width: 1200, height: 630 } });
  for (const [name, card] of Object.entries(CARDS)) {
    await tab.setContent(page(card, trace), { waitUntil: 'load' });
    await tab.evaluate(() => document.fonts.ready);
    await tab.screenshot({ path: join(OUT, `${name}.png`) });
    console.log(`og/${name}.png`);
  }
} finally {
  await browser.close();
}
