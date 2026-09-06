import { defineConfig, Plugin } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

function devHtmlRewrite(): Plugin {
  return {
    name: 'dev-html-rewrite',
    configureServer(server) {
      server.middlewares.use((req, _res, next) => {
        if (req.url === '/' || req.url === '/index.html') {
          req.url = '/index.dev.html';
        }
        next();
      });
    },
    generateBundle(_, bundle) {
      if (bundle['index.dev.html']) {
        const item = bundle['index.dev.html'];
        item.fileName = 'index.html';
        bundle['index.html'] = item;
        delete bundle['index.dev.html'];
      }
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), devHtmlRewrite()],
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
        main: fileURLToPath(new URL('./index.dev.html', import.meta.url)),
      },
    },
  },
});
