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
export default defineConfig({
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
    target: 'esnext',
    rollupOptions: {
      input: {
        main: fileURLToPath(new URL('./template.html', import.meta.url)),
      },
    },
  },
});
