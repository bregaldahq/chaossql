import { describe, expect, it } from 'vitest';
import { format, messages } from './index';

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

  it('uses the same placeholders in both languages', () => {
    const names = (text: unknown) => [...String(text).matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort();
    for (const key of Object.keys(en)) expect(names(pt[key]), key).toEqual(names(en[key]));
  });
});

describe('format', () => {
  it('fills placeholders and refuses to leave one empty', () => {
    expect(format('{a} of {b}', { a: 2, b: 20 })).toBe('2 of 20');
    expect(() => format('{missing}', {})).toThrow(/missing/);
  });
});
