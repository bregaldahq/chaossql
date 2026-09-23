// @vitest-environment jsdom
// @vitest-environment-options {"url": "https://chaossql.bregalda.com/pricing"}
import { afterEach, describe, expect, it, vi } from 'vitest';
import { track } from './analytics';
import { getStoredLanguage } from './i18n';

function stubBeacon() {
  const calls: Array<{ url: string; body: Record<string, unknown> }> = [];
  const sendBeacon = vi.fn((url: string, blob: Blob) => {
    void blob.text().then((text) => calls.push({ url, body: JSON.parse(text) }));
    return true;
  });
  Object.defineProperty(navigator, 'sendBeacon', { value: sendBeacon, configurable: true });
  return { calls, sendBeacon };
}

afterEach(() => {
  vi.restoreAllMocks();
  Object.defineProperty(navigator, 'globalPrivacyControl', { value: undefined, configurable: true });
  localStorage.clear();
});

describe('track', () => {
  it('posts the event, a sanitized label, path and language', async () => {
    const { calls } = stubBeacon();
    document.documentElement.lang = 'pt-BR';
    track('plan_select', 'Cloud_Team');
    await vi.waitFor(() => expect(calls).toHaveLength(1));
    expect(calls[0]).toEqual({
      url: '/api/event',
      body: { event: 'plan_select', label: 'cloud_team', path: '/pricing', lang: 'pt', ref: '' },
    });
  });

  it('drops labels that do not match the allowed shape', async () => {
    const { calls } = stubBeacon();
    track('cta_click', 'someone@example.com');
    await vi.waitFor(() => expect(calls).toHaveLength(1));
    expect(calls[0].body.label).toBe('');
  });

  it('sends nothing when Global Privacy Control is on', () => {
    const { sendBeacon } = stubBeacon();
    Object.defineProperty(navigator, 'globalPrivacyControl', { value: true, configurable: true });
    track('cta_click', 'hero_playground');
    expect(sendBeacon).not.toHaveBeenCalled();
  });
});

describe('getStoredLanguage', () => {
  it('defaults to English for a global audience', () => {
    vi.spyOn(navigator, 'languages', 'get').mockReturnValue(['en-US', 'de']);
    expect(getStoredLanguage()).toBe('en');
  });

  it('picks Portuguese when the browser prefers it', () => {
    vi.spyOn(navigator, 'languages', 'get').mockReturnValue(['pt-BR', 'en']);
    expect(getStoredLanguage()).toBe('pt');
  });

  it('keeps an explicit choice over the browser preference', () => {
    vi.spyOn(navigator, 'languages', 'get').mockReturnValue(['pt-BR']);
    localStorage.setItem('chaossql_lang', 'en');
    expect(getStoredLanguage()).toBe('en');
  });
});
