import { StrictMode } from 'react';
import { createRoot, hydrateRoot } from 'react-dom/client';
import '@fontsource-variable/geist';
import '@fontsource-variable/jetbrains-mono';
import './styles/tokens.css';
import './styles/globals.css';
import App from './App';
import { migrateLegacyHash } from './lib/router';
import { installWebAnalytics } from './lib/web-analytics';

// Rewrite old #/section links to /section before the first render.
migrateLegacyHash();
installWebAnalytics();

const root = document.getElementById('root')!;
const app = (
  <StrictMode>
    <App />
  </StrictMode>
);

// Prerendered pages (scripts/prerender.mjs) already contain the markup: attach
// to it. The shell used for other pages only has a static fallback: replace it.
if (root.dataset.prerendered !== undefined) hydrateRoot(root, app);
else createRoot(root).render(app);
