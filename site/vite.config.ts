import { defineConfig, Plugin } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

function templateHtmlRewrite(): Plugin {
  return {
    name: 'template-html-rewrite',
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        // Serve the source template for every app route (/, /docs, /pricing...).
        // Without this, deep links fall back to the committed, prebuilt index.html
        // and dev shows a stale bundle.
        const path = (req.url ?? '').split('?')[0];
        const isAppRoute = path === '/index.html' || (!path.includes('.') && !path.startsWith('/@') && !path.startsWith('/src/') && !path.startsWith('/node_modules/'));
        if (isAppRoute) {
          req.url = '/template.html';
        }
        next();
      });
    },
  };
}

// https://vitejs.dev/config/
// The client build starts from template.html; the SSR build (used only by
// scripts/prerender.mjs) starts from src/entry-server.tsx.
export default defineConfig(({ isSsrBuild }) => ({
  plugins: [react(), templateHtmlRewrite()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 3000,
    host: true,
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    // scripts/prerender.mjs reads dist/.vite/manifest.json to link each lazy
    // page's CSS in its prerendered HTML (no flash of unstyled content).
    manifest: !isSsrBuild,
    target: 'esnext',
    rollupOptions: isSsrBuild
      ? undefined
      : {
          input: {
            main: fileURLToPath(new URL('./template.html', import.meta.url)),
          },
        },
  },
}));
