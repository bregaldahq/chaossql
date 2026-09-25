// Cookieless product analytics. Events carry no personal data: only the event
// name, a short label, the current path and the UI language. They are posted to
// the worker's /api/event endpoint, which aggregates them in Workers Analytics
// Engine. Page views come from Cloudflare Web Analytics (injected by worker.ts).

export const SITE_EVENTS = [
  'cta_click',
  'install_copy',
  'command_copy',
  'scenario_view',
  'lead_submit',
  'plan_select',
  'pricing_toggle',
  'outbound_click',
  'lang_switch',
] as const;

export type SiteEvent = (typeof SITE_EVENTS)[number];

const ENDPOINT = '/api/event';
const LABEL_PATTERN = /^[a-z0-9_:.-]{0,64}$/;

function trackingDisabled(): boolean {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') return true;
  const host = window.location.hostname;
  if (host === 'localhost' || host === '127.0.0.1') return true;
  // Honor Global Privacy Control even though nothing personal is collected.
  return (navigator as Navigator & { globalPrivacyControl?: boolean }).globalPrivacyControl === true;
}

function referrerHost(): string {
  try {
    if (!document.referrer) return '';
    const host = new URL(document.referrer).hostname;
    return host === window.location.hostname ? '' : host.slice(0, 64);
  } catch {
    return '';
  }
}

function currentLang(): string {
  return document.documentElement.lang.startsWith('pt') ? 'pt' : 'en';
}

export function track(event: SiteEvent, label = ''): void {
  if (trackingDisabled()) return;
  const normalized = label.toLowerCase();
  const body = JSON.stringify({
    event,
    label: LABEL_PATTERN.test(normalized) ? normalized : '',
    path: window.location.pathname.slice(0, 128),
    lang: currentLang(),
    ref: referrerHost(),
  });
  try {
    const blob = new Blob([body], { type: 'application/json' });
    if (navigator.sendBeacon?.(ENDPOINT, blob)) return;
    void fetch(ENDPOINT, { method: 'POST', body, keepalive: true, headers: { 'Content-Type': 'application/json' } }).catch(() => {});
  } catch {
    // Analytics must never break the page.
  }
}
