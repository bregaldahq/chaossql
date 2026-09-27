// Single source of truth for page metadata and language URLs. Pure module (no
// DOM, no React): the app (route-meta.ts), the build-time prerender and the
// edge worker (worker.ts, bundled into site/_worker.js) all import it, so the
// title a crawler sees can never drift from the one the app sets.

export const SITE_ORIGIN = 'https://chaossql.bregalda.com';

export type Lang = 'en' | 'pt';
export const LANGS: readonly Lang[] = ['en', 'pt'];

const PT_PREFIX = '/pt';

/** Splits "/pt/pricing" into { lang: 'pt', path: '/pricing' }. */
export function splitLocale(pathname: string): { lang: Lang; path: string } {
  if (pathname === PT_PREFIX || pathname.startsWith(`${PT_PREFIX}/`)) {
    return { lang: 'pt', path: pathname.slice(PT_PREFIX.length) || '/' };
  }
  return { lang: 'en', path: pathname || '/' };
}

/** Prefixes an internal href ("/pricing#audit") for a language ("/pt/pricing#audit"). */
export function localizePath(href: string, lang: Lang): string {
  const [, rawPath = '/', rest = ''] = /^([^?#]*)(.*)$/.exec(href) ?? [];
  const { path } = splitLocale(rawPath || '/');
  if (lang === 'en') return path + rest;
  return (path === '/' ? PT_PREFIX : PT_PREFIX + path) + rest;
}

interface Copy {
  title: string;
  description: string;
}

export interface RouteSeo {
  path: string;
  indexable: boolean;
  image: string;
  en: Copy;
  pt: Copy;
}

export const SEO_ROUTES = {
  landing: {
    path: '/',
    indexable: true,
    image: '/og/home.png',
    en: {
      title: 'ChaosSQL | Deterministic SQL Concurrency Fuzzer for PostgreSQL, MySQL and SQLite',
      description:
        'Open-source (MIT) fuzzer that forces races between SQL transactions, detects lost updates, write skew and Adya isolation anomalies, and shrinks each failure to a minimal Go test.',
    },
    pt: {
      title: 'ChaosSQL | Fuzzer determinístico de concorrência SQL para PostgreSQL, MySQL e SQLite',
      description:
        'Fuzzer open source (MIT) que força corridas entre transações SQL, detecta lost update, write skew e anomalias de isolamento de Adya e reduz cada falha a um teste Go mínimo.',
    },
  },
  docs: {
    path: '/docs',
    indexable: true,
    image: '/og/docs.png',
    en: {
      title: 'Documentation | ChaosSQL',
      description:
        'ChaosSQL guide: installation, the chaos.yaml spec, SQL invariants, isolation levels, Adya anomaly classification and delta-debugging minimization.',
    },
    pt: {
      title: 'Documentação | ChaosSQL',
      description:
        'Guia do ChaosSQL: instalação, a especificação chaos.yaml, invariantes SQL, níveis de isolamento, classificação de anomalias de Adya e minimização por delta debugging.',
    },
  },
  scenarios: {
    path: '/scenarios',
    indexable: true,
    image: '/og/scenarios.png',
    en: {
      title: 'SQL Concurrency Anomaly Scenarios | ChaosSQL',
      description:
        'Canonical SQL concurrency anomalies (lost update, write skew, G2, deadlock and more) with invariants and deterministic, seed-based reproduction.',
    },
    pt: {
      title: 'Cenários de anomalias de concorrência SQL | ChaosSQL',
      description:
        'Anomalias canônicas de concorrência SQL (lost update, write skew, G2, deadlock e outras) com invariantes e reprodução determinística por seed.',
    },
  },
  visualizer: {
    path: '/visualizer',
    indexable: true,
    image: '/og/scenarios.png',
    en: {
      title: 'Trace Visualizer | ChaosSQL',
      description:
        'Inspect interleaved concurrent transactions, per-worker timings and the delta-debugged minimal trace behind a SQL anomaly.',
    },
    pt: {
      title: 'Trace Visualizer | ChaosSQL',
      description:
        'Inspecione transações concorrentes intercaladas, tempos por worker e o trace mínimo, reduzido por delta debugging, por trás de uma anomalia SQL.',
    },
  },
  matrix: {
    path: '/matrix',
    indexable: true,
    image: '/og/scenarios.png',
    en: {
      title: 'Hermitage Isolation Level Matrix | ChaosSQL',
      description:
        'Which concurrency anomalies each isolation level allows in PostgreSQL, MySQL and SQLite, inspired by the Hermitage project.',
    },
    pt: {
      title: 'Matriz Hermitage de níveis de isolamento | ChaosSQL',
      description:
        'Quais anomalias de concorrência cada nível de isolamento permite no PostgreSQL, MySQL e SQLite, inspirado no projeto Hermitage.',
    },
  },
  playground: {
    path: '/playground',
    indexable: true,
    image: '/og/playground.png',
    en: {
      title: 'WASM Playground: Test SQL Concurrency in Your Browser | ChaosSQL',
      description:
        'Run the ChaosSQL concurrency fuzzer in your browser with WebAssembly and reproduce lost updates, write skew and deadlocks without installing anything.',
    },
    pt: {
      title: 'Playground WASM: teste concorrência SQL no navegador | ChaosSQL',
      description:
        'Rode o fuzzer de concorrência ChaosSQL no navegador com WebAssembly e reproduza lost update, write skew e deadlocks sem instalar nada.',
    },
  },
  pricing: {
    path: '/pricing',
    indexable: true,
    image: '/og/pricing.png',
    en: {
      title: 'Pricing | ChaosSQL Cloud and Concurrency Audits',
      description:
        'ChaosSQL Cloud plans for CI concurrency regression testing, plus one-week database concurrency audits by Studio Bregalda.',
    },
    pt: {
      title: 'Preços | ChaosSQL Cloud e auditorias de concorrência',
      description:
        'Planos do ChaosSQL Cloud para testes de regressão de concorrência no CI, além de auditorias de concorrência de uma semana pelo Studio Bregalda.',
    },
  },
  dashboard: {
    path: '/dashboard',
    indexable: false,
    image: '/og/home.png',
    en: {
      title: 'Cloud Dashboard | ChaosSQL',
      description: 'ChaosSQL Cloud dashboard for CI runs, concurrency regressions and alerts.',
    },
    pt: {
      title: 'Cloud Dashboard | ChaosSQL',
      description: 'Painel do ChaosSQL Cloud para execuções de CI, regressões de concorrência e alertas.',
    },
  },
} satisfies Record<string, RouteSeo>;

export type SeoRouteId = keyof typeof SEO_ROUTES;

/** Route id for an unprefixed section path ("/pricing"), or null. */
export function seoRouteForPath(path: string): SeoRouteId | null {
  const clean = path.replace(/\/+$/, '') || '/';
  for (const [id, route] of Object.entries(SEO_ROUTES) as [SeoRouteId, RouteSeo][]) {
    if (route.path === clean) return id;
  }
  return null;
}

export interface PageMeta {
  lang: Lang;
  htmlLang: string;
  ogLocale: string;
  ogLocaleAlternate: string;
  title: string;
  description: string;
  robots: string;
  canonical: string;
  image: string;
  /** hreflang → absolute URL, including x-default (English). */
  alternates: Record<'en' | 'pt-BR' | 'x-default', string>;
}

/**
 * Everything the <head> needs for one page. `path` is the unprefixed path
 * ("/scenarios/lost-update"); `override` replaces the route's title and
 * description (scenario pages).
 */
export function pageMeta(route: SeoRouteId, lang: Lang, path: string, override?: Partial<Copy>): PageMeta {
  const seo: RouteSeo = SEO_ROUTES[route];
  const clean = route === 'landing' ? '/' : path.replace(/\/+$/, '') || seo.path;
  const url = (l: Lang) => {
    const localized = localizePath(clean, l);
    return `${SITE_ORIGIN}${localized === '/' ? '/' : localized}`;
  };
  return {
    lang,
    htmlLang: lang === 'pt' ? 'pt-BR' : 'en',
    ogLocale: lang === 'pt' ? 'pt_BR' : 'en_US',
    ogLocaleAlternate: lang === 'pt' ? 'en_US' : 'pt_BR',
    title: override?.title ?? seo[lang].title,
    description: override?.description ?? seo[lang].description,
    robots: seo.indexable ? 'index, follow, max-image-preview:large' : 'noindex, follow',
    canonical: url(lang),
    image: `${SITE_ORIGIN}${seo.image}`,
    alternates: { en: url('en'), 'pt-BR': url('pt'), 'x-default': url('en') },
  };
}
