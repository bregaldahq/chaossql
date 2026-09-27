import type { RouteId } from './router';

export const SITE_ORIGIN = 'https://chaossql.bregalda.com';

// Per-route <title>/<meta description>. Keep in sync with ROUTE_META in
// worker.ts and site/_worker.js, which inject the same values server-side so
// crawlers see them without executing JavaScript.
export const ROUTE_META: Record<RouteId, { title: string; description: string; indexable: boolean }> = {
  landing: {
    title: "ChaosSQL | Deterministic SQL Concurrency Fuzzer for PostgreSQL, MySQL and SQLite",
    description:
      "Open-source (MIT) fuzzer that forces races between SQL transactions, detects lost updates, write skew and Adya isolation anomalies, and shrinks each failure to a minimal Go test.",
    indexable: true,
  },
  docs: {
    title: "Documentation | ChaosSQL",
    description:
      "ChaosSQL guide: installation, the chaos.yaml spec, SQL invariants, isolation levels, Adya anomaly classification and delta-debugging minimization.",
    indexable: true,
  },
  scenarios: {
    title: "SQL Concurrency Anomaly Scenarios | ChaosSQL",
    description:
      "Canonical SQL concurrency anomalies (lost update, write skew, G2, deadlock and more) with invariants and deterministic, seed-based reproduction.",
    indexable: true,
  },
  visualizer: {
    title: "Trace Visualizer | ChaosSQL",
    description:
      "Inspect interleaved concurrent transactions, per-worker timings and the delta-debugged minimal trace behind a SQL anomaly.",
    indexable: true,
  },
  matrix: {
    title: "Hermitage Isolation Level Matrix | ChaosSQL",
    description:
      "Which concurrency anomalies each isolation level allows in PostgreSQL, MySQL and SQLite, inspired by the Hermitage project.",
    indexable: true,
  },
  playground: {
    title: "WASM Playground: Test SQL Concurrency in Your Browser | ChaosSQL",
    description:
      "Run the ChaosSQL concurrency fuzzer in your browser with WebAssembly and reproduce lost updates, write skew and deadlocks without installing anything.",
    indexable: true,
  },
  pricing: {
    title: "Pricing | ChaosSQL Cloud and Concurrency Audits",
    description:
      "ChaosSQL Cloud plans for CI concurrency regression testing, plus one-week database concurrency audits by Studio Bregalda.",
    indexable: true,
  },
  dashboard: {
    title: "Cloud Dashboard | ChaosSQL",
    description:
      "ChaosSQL Cloud dashboard for CI runs, concurrency regressions and alerts.",
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
