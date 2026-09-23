import type { RouteId } from './router';

export const SITE_ORIGIN = 'https://chaossql.bregalda.com';

// Per-route <title>/<meta description>. Keep in sync with ROUTE_META in
// worker.ts and site/_worker.js, which inject the same values server-side so
// crawlers see them without executing JavaScript.
export const ROUTE_META: Record<RouteId, { title: string; description: string; indexable: boolean }> = {
  landing: {
    title: 'ChaosSQL — SQL Concurrency Fuzzer | Testes Determinísticos de Concorrência e Isolamento',
    description:
      'ChaosSQL é um fuzzer open source (MIT) de concorrência SQL para PostgreSQL, MySQL e SQLite: detecta lost update, write skew e anomalias de isolamento de Adya com execuções determinísticas e delta-debugging.',
    indexable: true,
  },
  docs: {
    title: 'Documentação — ChaosSQL SQL Concurrency Fuzzer',
    description:
      'Guia do ChaosSQL: instalação, especificação chaos.yaml, invariantes, níveis de isolamento, classificação de anomalias de Adya e minimização por delta-debugging.',
    indexable: true,
  },
  scenarios: {
    title: 'Cenários de Anomalias de Concorrência — ChaosSQL',
    description:
      'Cenários canônicos de anomalias de concorrência SQL: lost update, write skew, G2, deadlock e mais, com invariantes e reprodução determinística.',
    indexable: true,
  },
  visualizer: {
    title: 'Trace Visualizer — ChaosSQL',
    description:
      'Visualize interleavings de transações concorrentes, timings por worker e o trace minimizado por delta-debugging de uma anomalia SQL.',
    indexable: true,
  },
  matrix: {
    title: 'Matriz Hermitage de Níveis de Isolamento — ChaosSQL',
    description:
      'Matriz de anomalias por nível de isolamento em PostgreSQL, MySQL e SQLite, inspirada no projeto Hermitage.',
    indexable: true,
  },
  playground: {
    title: 'Playground WASM — Teste Concorrência SQL no Navegador | ChaosSQL',
    description:
      'Execute o fuzzer de concorrência ChaosSQL direto no navegador via WebAssembly e reproduza lost update, write skew e deadlocks sem instalar nada.',
    indexable: true,
  },
  pricing: {
    title: 'Planos e Preços — ChaosSQL Cloud',
    description:
      'Planos do ChaosSQL Cloud e auditorias de concorrência de banco de dados pelo Studio Bregalda.',
    indexable: true,
  },
  dashboard: {
    title: 'Cloud Dashboard — ChaosSQL',
    description: 'Painel do ChaosSQL Cloud para acompanhar execuções de CI, regressões de concorrência e alertas.',
    indexable: false,
  },
};

function setMeta(selector: string, attr: string, value: string) {
  const el = document.head.querySelector(selector);
  if (el) el.setAttribute(attr, value);
}

export function applyRouteMeta(route: RouteId, pathname: string): void {
  const meta = ROUTE_META[route];
  const canonical = `${SITE_ORIGIN}${route === 'landing' ? '/' : pathname.replace(/\/+$/, '')}`;
  document.title = meta.title;
  setMeta('meta[name="description"]', 'content', meta.description);
  setMeta('meta[name="robots"]', 'content', meta.indexable ? 'index, follow, max-image-preview:large' : 'noindex, follow');
  setMeta('link[rel="canonical"]', 'href', canonical);
  setMeta('meta[property="og:url"]', 'content', canonical);
}
