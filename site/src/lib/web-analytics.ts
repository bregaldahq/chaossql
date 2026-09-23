// Cloudflare Web Analytics for the public marketing site only. The same bundle
// also serves the managed dashboard and self-hosted or air-gapped installs,
// which must never report visits (or run URLs) to our analytics account.
// The token is public by design: it is embedded in every tracked page.

export const WEB_ANALYTICS_TOKEN = '50a8ee7b1e1047c7951cf92911c0422e';
export const WEB_ANALYTICS_HOSTS: readonly string[] = ['chaossql.bregalda.com'];
const BEACON_SRC = 'https://static.cloudflareinsights.com/beacon.min.js';

export function installWebAnalytics(doc: Document = document, hostname: string = window.location.hostname): boolean {
  if (!WEB_ANALYTICS_HOSTS.includes(hostname)) return false;
  if (doc.querySelector('script[data-cf-beacon]')) return false;
  const script = doc.createElement('script');
  script.defer = true;
  script.src = BEACON_SRC;
  // spa: count History API route changes (/docs, /pricing) as page views.
  script.setAttribute('data-cf-beacon', JSON.stringify({ token: WEB_ANALYTICS_TOKEN, spa: true }));
  doc.head.appendChild(script);
  return true;
}
