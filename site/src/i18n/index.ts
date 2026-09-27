import { en, type Messages } from './en';
import { pt } from './pt';

export type Language = 'pt' | 'en';
export type { Messages };

export const messages: Record<Language, Messages> = { en, pt };

/** Copy for the given language. Components take `lang` as a prop, as before. */
export function useMessages(lang: Language): Messages {
  return messages[lang];
}
