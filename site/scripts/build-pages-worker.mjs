#!/usr/bin/env node
// Generates site/_worker.js (Cloudflare Pages advanced mode) from ../worker.ts,
// bundling the shared SEO module it imports. The file used to be edited by hand
// and drifted (it missed /api/event); src/lib/router.test.ts now fails when the
// committed file differs from this build.

import { buildSync } from 'esbuild';
import { existsSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const SITE = join(dirname(fileURLToPath(import.meta.url)), '..');

export const PAGES_WORKER_BANNER = '// Generated from worker.ts by site/scripts/build-pages-worker.mjs. Do not edit.';

export function buildPagesWorker() {
  const result = buildSync({
    entryPoints: [join(SITE, '..', 'worker.ts')],
    absWorkingDir: SITE,
    bundle: true,
    format: 'esm',
    platform: 'neutral',
    target: 'es2022',
    banner: { js: PAGES_WORKER_BANNER },
    write: false,
  });
  return result.outputFiles[0].text;
}

if (import.meta.url === pathToFileURL(process.argv[1]).href) {
  // Docker builds copy only site/: there is no worker to bundle there.
  if (!existsSync(join(SITE, '..', 'worker.ts'))) {
    console.log('build-pages-worker: ../worker.ts not found, skipping');
    process.exit(0);
  }
  writeFileSync(join(SITE, '_worker.js'), buildPagesWorker());
  console.log('build-pages-worker: wrote site/_worker.js');
}
