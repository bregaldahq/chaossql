#!/usr/bin/env node
// Renders docs/media/lost-update.gif: the recorded banking run, one frame per
// statement, ending on the failing invariant. Same data as the landing hero
// (src/data/traces/banking_lost_update.json), for the README and social posts.
//
// Usage: npm run gif   (set CHROMIUM_PATH to use a specific browser binary)

import { chromium } from 'playwright-core';
import gifenc from 'gifenc';
import pngjs from 'pngjs';
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromiumPath, fontUrl } from './render-lib.mjs';

// Both packages are CommonJS.
const { GIFEncoder, applyPalette, quantize } = gifenc;
const { PNG } = pngjs;

const SITE = join(dirname(fileURLToPath(import.meta.url)), '..');
const OUT = join(SITE, '..', 'docs', 'media', 'lost-update.gif');
const WIDTH = 880;
const HEIGHT = 330;

const run = JSON.parse(readFileSync(join(SITE, 'src/data/traces/banking_lost_update.json'), 'utf8'));
const trace = run.minimalTrace;
const lanes = new Map();
for (const e of trace) if (!lanes.has(e.worker)) lanes.set(e.worker, lanes.size + 1);

// Story numbers, derived like src/data/story.ts does.
const writes = trace
  .map((e, i) => ({ i, op: e.op, value: /UPDATE accounts SET balance = (\d+)/i.exec(e.sql)?.[1] }))
  .filter((w) => w.value);
const amount = (op) => Number(run.minimalOps.find((o) => o.id === op).params.amount);
const start = Number(writes[0].value) + amount(writes[0].op);
const expected = start - amount(writes[0].op) - amount(writes[1].op);
const lostIndex = writes[1].i;
const money = (n) => `$${n.toLocaleString('en-US')}`;

const escapeHtml = (text) => text.replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' })[c]);

function frame(active) {
  const step = 100 / trace.length;
  let balance = start;
  for (const e of trace.slice(0, active + 1)) {
    const m = /UPDATE accounts SET balance = (\d+)/i.exec(e.sql);
    if (m) balance = Number(m[1]);
  }
  const done = active >= trace.length - 1;
  const current = trace[active];
  const lanesHtml = [...lanes.entries()]
    .map(([worker, lane]) => {
      const blocks = trace
        .map((e, i) => {
          if (e.worker !== worker) return '';
          const kind = e.type === 'COMMIT' ? 'commit' : /^\s*(UPDATE|INSERT)/i.test(e.sql) ? 'write' : e.type === 'BEGIN' ? 'begin' : 'read';
          const state = i > active ? 'upcoming' : i === active ? 'current' : '';
          const lost = i === lostIndex && i <= active ? 'lost' : '';
          return `<div class="block ${kind} ${state} ${lost}" style="left:calc(${(i * step).toFixed(3)}% + 3px);width:calc(${step.toFixed(3)}% - 6px)"></div>`;
        })
        .join('');
      return `<div class="lane"><span class="tx">T${lane}</span><div class="track">${blocks}</div></div>`;
    })
    .join('');
  return `<!doctype html><html><head><meta charset="utf-8"><style>
@font-face { font-family: Geist; src: url(${fontUrl('geist', 'geist-latin-wght-normal.woff2')}) format('woff2'); font-weight: 100 900; }
@font-face { font-family: JBM; src: url(${fontUrl('jetbrains-mono', 'jetbrains-mono-latin-wght-normal.woff2')}) format('woff2'); font-weight: 100 800; }
* { box-sizing: border-box; margin: 0; }
body { width: ${WIDTH}px; height: ${HEIGHT}px; background: #141021; color: #fcfbf8; font-family: Geist; padding: 36px 40px; display: flex; flex-direction: column; gap: 22px; }
.head { display: flex; gap: 14px; font-family: JBM; font-size: 14px; color: #949197; } .head b { color: #fcfbf8; font-weight: 500; }
.lanes { display: grid; gap: 14px; } .lane { display: flex; align-items: center; gap: 16px; }
.tx { font-family: JBM; font-size: 14px; color: #949197; width: 28px; }
.track { position: relative; height: 30px; flex: 1; border-bottom: 1px solid #302c3b; }
.block { position: absolute; top: 0; height: 30px; border-radius: 6px; }
.begin { background: #3a3547; } .read { background: #6c6878; } .write { background: #a58ee8; } .commit { background: #22c55e; } .lost { background: #f5c400; }
.upcoming { background: transparent; box-shadow: inset 0 0 0 1px #302c3b; }
.current { box-shadow: 0 0 0 2px #141021, 0 0 0 3px #fcfbf8; }
.now { font-family: JBM; font-size: 15px; color: #bbb9bc; min-height: 20px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.now .t { color: #949197; margin-right: 12px; } .now .l { color: #f5c400; }
.readout { display: flex; gap: 44px; padding-top: 18px; border-top: 1px solid #302c3b; margin-top: auto; }
.readout dt { font-size: 13px; color: #949197; margin-bottom: 4px; } .readout dd { font-family: JBM; font-size: 22px; }
.bad { color: #f5c400; }
.verdict { font-family: JBM; font-size: 14px; padding: 5px 12px; border-radius: 999px; background: rgba(245, 196, 0, 0.14); color: #f5c400; }
.pending { font-size: 14px; color: #949197; }
</style></head><body>
<div class="head"><b>banking_lost_update</b><span>seed ${run.seed} · ${run.isolation}</span></div>
<div class="lanes">${lanesHtml}</div>
<p class="now">${current ? `<span class="t">T${lanes.get(current.worker)}</span><span class="${active === lostIndex ? 'l' : ''}">${escapeHtml(current.sql)}</span>` : '&nbsp;'}</p>
<dl class="readout">
  <div><dt>Stored balance</dt><dd class="${done ? 'bad' : ''}">${money(balance)}</dd></div>
  <div><dt>Expected</dt><dd>${money(expected)}</dd></div>
  <div><dt>Invariant</dt><dd>${done ? `<span class="verdict">fails: ${money(balance)} ≠ ${money(expected)}</span>` : '<span class="pending">checked when both commit</span>'}</dd></div>
</dl>
</body></html>`;
}

const browser = await chromium.launch({ executablePath: chromiumPath() });
try {
  const page = await browser.newPage({ viewport: { width: WIDTH, height: HEIGHT } });
  const gif = GIFEncoder();
  // Empty lanes, then one frame per statement, then a long hold on the verdict.
  const frames = [-1, ...trace.map((_, i) => i)];
  for (const active of frames) {
    await page.setContent(frame(active), { waitUntil: 'load' });
    await page.evaluate(() => document.fonts.ready);
    const png = PNG.sync.read(await page.screenshot());
    const palette = quantize(png.data, 64);
    const delay = active === -1 ? 900 : active === trace.length - 1 ? 4000 : active === lostIndex ? 1400 : 700;
    gif.writeFrame(applyPalette(png.data, palette), png.width, png.height, { palette, delay });
  }
  gif.finish();
  mkdirSync(dirname(OUT), { recursive: true });
  writeFileSync(OUT, gif.bytes());
  console.log(`render-trace-gif: wrote docs/media/lost-update.gif (${frames.length} frames, ${Math.round(gif.bytes().length / 1024)} KB)`);
} finally {
  await browser.close();
}
