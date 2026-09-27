import { useEffect } from 'react';
import type { Language } from '../i18n';
import { localizePath, navigate, splitLocale, useLocation } from './router';

export type { Language };

const STORAGE_KEY = 'chaossql_lang';

// English is the canonical language (global audience) and lives at the root;
// Portuguese lives under /pt. The URL decides the language, so every page is
// crawlable in both languages and the server-rendered HTML always matches.
export const DEFAULT_LANGUAGE: Language = 'en';

function isLanguage(value: unknown): value is Language {
  return value === 'pt' || value === 'en';
}

/** The visitor's explicit choice, if they ever switched language. */
export function getStoredLanguage(): Language | null {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return isLanguage(saved) ? saved : null;
  } catch {
    return null;
  }
}

export function setStoredLanguage(lang: Language): void {
  try {
    localStorage.setItem(STORAGE_KEY, lang);
  } catch {
    // Storage can be unavailable (private mode); the URL still carries the language.
  }
}

/** True when the browser prefers Portuguese and the visitor never chose a language. */
export function prefersPortuguese(): boolean {
  if (typeof navigator === 'undefined' || getStoredLanguage() !== null) return false;
  const preferred = navigator.languages?.length ? navigator.languages : [navigator.language];
  return preferred.some((l) => typeof l === 'string' && l.toLowerCase().startsWith('pt'));
}

export function useI18n() {
  const { pathname, search } = useLocation();
  const lang = splitLocale(pathname).lang;

  useEffect(() => {
    document.documentElement.lang = lang === 'pt' ? 'pt-BR' : 'en';
  }, [lang]);

  const setLang = (next: Language) => {
    setStoredLanguage(next);
    if (next !== lang) navigate(localizePath(splitLocale(pathname).path + search + window.location.hash, next));
  };

  return { lang, setLang };
}
