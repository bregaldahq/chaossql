#!/usr/bin/env node
// Deletes hashed build files in site/assets that the current index.html no
// longer reaches. `npm run build` copies new bundles next to the old ones and
// the deployed site is this directory, so without pruning every build leaves
// dead chunks behind. Unhashed files (wasm, legacy style.css) are never touched.

import { existsSync, readdirSync, readFileSync, rmSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const SITE = join(dirname(fileURLToPath(import.meta.url)), '..');
const ASSETS = join(SITE, 'assets');
const HASHED = /^[\w.-]+-[\w-]{8}\.(js|css|woff2|webp|png|svg)(\.map)?$/;
const REF = /[\w.-]+-[\w-]{8}\.(?:js|css|woff2|webp|png|svg)/g;

const reachable = new Set();
const prerendered = join(SITE, 'prerender');
const pages = (dir) =>
  existsSync(dir)
    ? readdirSync(dir, { withFileTypes: true }).flatMap((e) =>
        e.isDirectory() ? pages(join(dir, e.name)) : e.name.endsWith('.html') ? [join(dir, e.name)] : []
      )
    : [];
const queue = [join(SITE, 'index.html'), ...pages(prerendered)];
while (queue.length) {
  const text = readFileSync(queue.pop(), 'utf8');
  for (const name of text.match(REF) ?? []) {
    if (reachable.has(name) || !existsSync(join(ASSETS, name))) continue;
    reachable.add(name);
    reachable.add(`${name}.map`);
    if (/\.(js|css)$/.test(name)) queue.push(join(ASSETS, name));
  }
}

const stale = readdirSync(ASSETS).filter((name) => HASHED.test(name) && !reachable.has(name));
for (const name of stale) rmSync(join(ASSETS, name));
console.log(`prune-assets: kept ${reachable.size / 2} hashed files, removed ${stale.length}`);
