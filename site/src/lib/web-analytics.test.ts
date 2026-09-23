// @vitest-environment jsdom
import { afterEach, describe, expect, it } from 'vitest';
import { installWebAnalytics, WEB_ANALYTICS_TOKEN } from './web-analytics';

const beacons = () => document.querySelectorAll('script[data-cf-beacon]');

describe('web analytics', () => {
  afterEach(() => beacons().forEach((node) => node.remove()));

  it('loads the beacon once on the public site with SPA tracking', () => {
    expect(installWebAnalytics(document, 'chaossql.bregalda.com')).toBe(true);
    expect(installWebAnalytics(document, 'chaossql.bregalda.com')).toBe(false);
    expect(beacons()).toHaveLength(1);
    const script = beacons()[0] as HTMLScriptElement;
    expect(script.src).toBe('https://static.cloudflareinsights.com/beacon.min.js');
    expect(script.defer).toBe(true);
    expect(JSON.parse(script.getAttribute('data-cf-beacon')!)).toEqual({ token: WEB_ANALYTICS_TOKEN, spa: true });
  });

  it.each(['api.chaossql.bregalda.com', 'localhost', '127.0.0.1', 'chaossql.internal.example', 'evil-chaossql.bregalda.com.example'])(
    'never reports from %s (managed, self-hosted, or local)',
    (host) => {
      expect(installWebAnalytics(document, host)).toBe(false);
      expect(beacons()).toHaveLength(0);
    },
  );
});
