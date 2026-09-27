// Canonical copy (English first: global audience). pt.ts must mirror these keys;
// the Messages type makes a missing or extra key a compile error.
export const en = {
  common: {
    copy: 'Copy',
    copied: 'Copied',
    close: 'Close',
  },
  nav: {
    homeLabel: 'ChaosSQL by Studio Bregalda, home',
    mainLabel: 'Main navigation',
    mobileLabel: 'Mobile navigation',
    languageLabel: 'Language selector',
    openMenu: 'Open navigation menu',
    closeMenu: 'Close navigation menu',
    githubRepository: 'GitHub Repository',
    items: {
      landing: 'Home',
      dashboard: 'Cloud Dashboard',
      docs: 'Docs',
      scenarios: 'Scenarios',
      visualizer: 'Trace Visualizer',
      matrix: 'Hermitage Matrix',
      playground: 'WASM Playground',
      pricing: 'Pricing',
    },
  },
  footer: {
    label: 'Footer links',
    docs: 'Documentation',
    scenarios: 'Scenarios',
    playground: 'WASM Playground',
    rights: 'All rights reserved.',
  },
};

type DeepStrings<T> = { [K in keyof T]: T[K] extends string ? string : DeepStrings<T[K]> };
export type Messages = DeepStrings<typeof en>;
