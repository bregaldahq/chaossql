import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App';
import { migrateLegacyHash } from './lib/router';
import '@fontsource-variable/geist';
import '@fontsource-variable/jetbrains-mono';
import './styles/tokens.css';
import './styles/globals.css';

// Rewrite old #/section links to /section before the first render.
migrateLegacyHash();

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
