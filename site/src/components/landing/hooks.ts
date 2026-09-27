import { useEffect, useState, type RefObject } from 'react';

export function usePrefersReducedMotion(): boolean {
  const query = '(prefers-reduced-motion: reduce)';
  const [reduce, setReduce] = useState(() => typeof window !== 'undefined' && !!window.matchMedia?.(query).matches);
  useEffect(() => {
    const mql = window.matchMedia?.(query);
    if (!mql) return;
    const update = () => setReduce(mql.matches);
    mql.addEventListener('change', update);
    return () => mql.removeEventListener('change', update);
  }, []);
  return reduce;
}

/** True while (or, with `once`, after) the element intersects the viewport. */
export function useInView(
  ref: RefObject<Element | null>,
  { once = false, rootMargin = '0px', threshold = 0 }: { once?: boolean; rootMargin?: string; threshold?: number } = {}
): boolean {
  const [inView, setInView] = useState(false);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (typeof IntersectionObserver === 'undefined') {
      setInView(true);
      return;
    }
    const observer = new IntersectionObserver(
      ([entry]) => {
        setInView(entry.isIntersecting);
        if (entry.isIntersecting && once) observer.disconnect();
      },
      { rootMargin, threshold }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [ref, once, rootMargin, threshold]);
  return inView;
}

const STARS_KEY = 'chaossql_stars';
const STARS_TTL_MS = 6 * 60 * 60 * 1000;

/** GitHub star count, cached per visitor for 6 hours. Null until known or on any failure. */
export function useGitHubStars(repo = 'bregaldahq/chaossql'): number | null {
  const [stars, setStars] = useState<number | null>(() => {
    try {
      const cached = JSON.parse(localStorage.getItem(STARS_KEY) ?? 'null');
      return cached && Date.now() - cached.at < STARS_TTL_MS ? cached.stars : null;
    } catch {
      return null;
    }
  });
  useEffect(() => {
    if (stars !== null || typeof fetch === 'undefined') return;
    const controller = new AbortController();
    fetch(`https://api.github.com/repos/${repo}`, { signal: controller.signal, credentials: 'omit' })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (typeof data?.stargazers_count !== 'number') return;
        setStars(data.stargazers_count);
        try {
          localStorage.setItem(STARS_KEY, JSON.stringify({ stars: data.stargazers_count, at: Date.now() }));
        } catch {
          // Storage can be unavailable (private mode); the count still shows.
        }
      })
      .catch(() => {});
    return () => controller.abort();
  }, [repo, stars]);
  return stars;
}
