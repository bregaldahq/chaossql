import { en, type Messages } from './en';
import { pt } from './pt';

export type Language = 'pt' | 'en';
export type { Messages };

export const messages: Record<Language, Messages> = { en, pt };

/** Copy for the given language. Components take `lang` as a prop, as before. */
export function useMessages(lang: Language): Messages {
  return messages[lang];
}

/** Fills {name} placeholders. A missing value is a bug, so it throws. */
export function format(template: string, values: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (_match, key: string) => {
    if (!(key in values)) throw new Error(`Missing value for {${key}} in "${template}"`);
    return String(values[key]);
  });
}
