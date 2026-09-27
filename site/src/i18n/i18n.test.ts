import { describe, expect, it } from 'vitest';
import { messages } from './index';

function flatten(obj: object, prefix = ''): Record<string, unknown> {
  return Object.entries(obj).reduce<Record<string, unknown>>((acc, [key, value]) => {
    const path = prefix ? `${prefix}.${key}` : key;
    if (value && typeof value === 'object') Object.assign(acc, flatten(value, path));
    else acc[path] = value;
    return acc;
  }, {});
}

describe('i18n dictionaries', () => {
  const en = flatten(messages.en);
  const pt = flatten(messages.pt);

  it('have exactly the same keys', () => {
    expect(Object.keys(pt).sort()).toEqual(Object.keys(en).sort());
  });

  it.each(Object.entries({ en, pt }))('%s has no empty strings', (_lang, dict) => {
    for (const [key, value] of Object.entries(dict)) {
      expect(typeof value, key).toBe('string');
      expect((value as string).trim(), key).not.toBe('');
    }
  });

  it.each(Object.entries({ en, pt }))('%s copy never uses em or en dashes', (_lang, dict) => {
    for (const [key, value] of Object.entries(dict)) expect(value, key).not.toMatch(/[–—]/);
  });
});
