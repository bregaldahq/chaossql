// Shared helpers for the build-time renderers (OG cards, trace GIF).
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import { homedir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const SITE = join(dirname(fileURLToPath(import.meta.url)), '..');

/** Chromium for playwright-core: CHROMIUM_PATH, else the newest cached headless shell. */
export function chromiumPath() {
  if (process.env.CHROMIUM_PATH) return process.env.CHROMIUM_PATH;
  const cache = join(homedir(), '.cache/ms-playwright');
  const shells = existsSync(cache) ? readdirSync(cache).filter((d) => d.startsWith('chromium_headless_shell-')).sort() : [];
  for (const dir of shells.reverse()) {
    for (const sub of readdirSync(join(cache, dir))) {
      const exe = join(cache, dir, sub, 'chrome-headless-shell');
      if (existsSync(exe)) return exe;
    }
  }
  throw new Error('No Chromium found: install one with `npx playwright install chromium-headless-shell` or set CHROMIUM_PATH');
}

// Inlined as data URLs: pages loaded with setContent cannot fetch file:// fonts.
export function fontUrl(pkg, file) {
  const bytes = readFileSync(join(SITE, 'node_modules/@fontsource-variable', pkg, 'files', file));
  return `data:font/woff2;base64,${bytes.toString('base64')}`;
}

