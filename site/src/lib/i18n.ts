import { useState, useEffect } from 'react';

export type Language = 'pt' | 'en';

const STORAGE_KEY = 'chaossql_lang';

export function getStoredLanguage(): Language {
  if (typeof window === 'undefined') return 'pt';
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved === 'pt' || saved === 'en') return saved;
  } catch (_) {
    // ignore
  }
  return 'pt';
}

export function setStoredLanguage(lang: Language): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, lang);
    window.dispatchEvent(new CustomEvent('languagechange', { detail: lang }));
  } catch (_) {
    // ignore
  }
}

export function useI18n() {
  const [lang, setLangState] = useState<Language>(getStoredLanguage);

  useEffect(() => {
    const handleLangChange = (e: Event) => {
      const customEvent = e as CustomEvent<Language>;
      if (customEvent.detail && (customEvent.detail === 'pt' || customEvent.detail === 'en')) {
        setLangState(customEvent.detail);
      }
    };

    window.addEventListener('languagechange', handleLangChange);
    window.addEventListener('storage', (e) => {
      if (e.key === STORAGE_KEY && (e.newValue === 'pt' || e.newValue === 'en')) {
        setLangState(e.newValue as Language);
      }
    });

    return () => {
      window.removeEventListener('languagechange', handleLangChange);
    };
  }, []);

  const setLang = (newLang: Language) => {
    setLangState(newLang);
    setStoredLanguage(newLang);
  };

  return { lang, setLang };
}
