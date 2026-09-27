#!/usr/bin/env node
// Serves site/ the way the edge worker does, without Cloudflare: prerendered
// pages at their canonical URLs (/, /pt/pricing, ...), static files as they
// are, and the index.html shell for every other route. Used by Lighthouse CI
// and for local previews: `npm run preview:static` then open the printed URL.

import { createServer } from 'node:http';
import { existsSync, readFileSync, statSync } from 'node:fs';
import { dirname, extname, join, normalize } from 'node:path';
import { gzipSync } from 'node:zlib';
import { fileURLToPath } from 'node:url';

const SITE = join(dirname(fileURLToPath(import.meta.url)), '..');
const PORT = Number(process.argv[2] ?? 4173);
const manifest = new Set(JSON.parse(readFileSync(join(SITE, 'prerender', 'manifest.json'), 'utf8')));
const TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.webp': 'image/webp',
  '.woff2': 'font/woff2',
  '.wasm': 'application/wasm',
  '.xml': 'application/xml',
  '.txt': 'text/plain',
};

createServer((req, res) => {
  const pathname = decodeURIComponent(new URL(req.url ?? '/', 'http://localhost').pathname);
  const staticFile = join(SITE, normalize(pathname));
  let file = join(SITE, 'index.html');
  if (manifest.has(pathname)) file = join(SITE, 'prerender', `${pathname === '/' ? '/home' : pathname}.html`);
  else if (staticFile.startsWith(SITE) && existsSync(staticFile) && statSync(staticFile).isFile()) file = staticFile;
  else if (pathname.startsWith('/api/')) {
    res.writeHead(204).end();
    return;
  }
  const headers = { 'Content-Type': TYPES[extname(file)] ?? 'application/octet-stream' };
  if (pathname.startsWith('/assets/')) headers['Cache-Control'] = 'public, max-age=31536000, immutable';
  let body = readFileSync(file);
  // Compress text like Cloudflare does, so Lighthouse measures realistic sizes.
  if (/gzip/.test(req.headers['accept-encoding'] ?? '') && /text|javascript|json|xml|svg/.test(headers['Content-Type'])) {
    body = gzipSync(body);
    headers['Content-Encoding'] = 'gzip';
  }
  res.writeHead(200, headers).end(body);
}).listen(PORT, '127.0.0.1', () => console.log(`serve-static listening on http://127.0.0.1:${PORT}`));
