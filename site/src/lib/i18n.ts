import { useState, useEffect } from 'react';

import type { Language } from '../i18n';

export type { Language };

const STORAGE_KEY = 'chaossql_lang';

// English is the canonical language (global audience). Portuguese is used when
// the visitor chose it before, or when the browser prefers Portuguese.
export const DEFAULT_LANGUAGE: Language = 'en';

function isLanguage(value: unknown): value is Language {
  return value === 'pt' || value === 'en';
}

function browserLanguage(): Language {
  if (typeof navigator === 'undefined') return DEFAULT_LANGUAGE;
  const preferred = navigator.languages?.length ? navigator.languages : [navigator.language];
  return preferred.some((l) => typeof l === 'string' && l.toLowerCase().startsWith('pt')) ? 'pt' : DEFAULT_LANGUAGE;
}

export function getStoredLanguage(): Language {
  if (typeof window === 'undefined') return DEFAULT_LANGUAGE;
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (isLanguage(saved)) return saved;
  } catch (_) {
    // ignore
  }
  return browserLanguage();
}

export function setStoredLanguage(lang: Language): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, lang);
  } catch (_) {
    // ignore
  }
  window.dispatchEvent(new CustomEvent('languagechange', { detail: lang }));
}

export function useI18n() {
  const [lang, setLangState] = useState<Language>(getStoredLanguage);

  useEffect(() => {
    document.documentElement.lang = lang === 'pt' ? 'pt-BR' : 'en';
  }, [lang]);

  useEffect(() => {
    const handleLangChange = (e: Event) => {
      const detail = (e as CustomEvent<Language>).detail;
      if (isLanguage(detail)) setLangState(detail);
    };
    const handleStorage = (e: StorageEvent) => {
      if (e.key === STORAGE_KEY && isLanguage(e.newValue)) setLangState(e.newValue);
    };

    window.addEventListener('languagechange', handleLangChange);
    window.addEventListener('storage', handleStorage);
    return () => {
      window.removeEventListener('languagechange', handleLangChange);
      window.removeEventListener('storage', handleStorage);
    };
  }, []);

  const setLang = (newLang: Language) => {
    setLangState(newLang);
    setStoredLanguage(newLang);
  };

  return { lang, setLang };
}
