export interface PublicRun {
  id: string;
  repo_id: string;
  scenario_id: string;
  repo_full_name?: string | null;
  scenario_name?: string | null;
  driver?: string | null;
  branch?: string;
  pr_number?: number;
  commit_sha?: string;
  status: string;
  anomaly_type?: string;
  is_regression?: boolean | null;
  seed?: number | null;
  duration_ms?: number | null;
  total_schedules?: number | null;
  failed_schedules?: number | null;
  created_at?: string;
}

export function mapRun(run: PublicRun) {
  return {
    id: run.id, repo: run.repo_full_name || run.repo_id, scenario: run.scenario_name || run.scenario_id,
    branch: run.branch || '—', prNumber: run.pr_number, commitSHA: run.commit_sha || '',
    status: run.status, anomalyType: run.anomaly_type || '—', anomalyName: run.anomaly_type || '—',
    isRegression: run.is_regression ?? null, driver: run.driver || '—', isolation: '—',
    schedulesCount: run.total_schedules ?? null, failedSchedules: run.failed_schedules ?? null,
    seed: run.seed ?? null, durationMS: run.duration_ms ?? null,
    timestamp: run.created_at ? new Date(run.created_at).toLocaleString() : '—',
  };
}

interface PublicWebhook {
  id: string; target_type: 'discord' | 'slack' | 'generic'; url: string;
  events: string; active: boolean; created_at: string;
}

export function mapWebhook(hook: PublicWebhook) {
  return { id: hook.id, target: hook.target_type, url: hook.url, events: hook.events.split(',').filter(Boolean), active: hook.active, createdAt: hook.created_at };
}
export type WebhookItem = ReturnType<typeof mapWebhook>;

export class CloudAPI {
  constructor(private baseURL: string, private token: string, private fetcher: typeof fetch = fetch) {}

  private async request(path: string, method = 'GET', body?: unknown): Promise<Response> {
    if (!this.token.trim()) throw new Error('API token required');
    const response = await this.fetcher(`${this.baseURL.trim().replace(/\/+$/, '')}${path}`, {
      method, headers: { Authorization: `Bearer ${this.token.trim()}`, ...(body === undefined ? {} : { 'Content-Type': 'application/json' }) },
      body: body === undefined ? undefined : JSON.stringify(body), cache: 'no-store', credentials: 'omit', redirect: 'error',
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}${response.status === 401 ? ': Invalid API token' : response.status === 403 ? ': Administrator permission required' : ': Request failed'}`);
    return response;
  }

  async runs() {
    const body: { runs: PublicRun[] } = await (await this.request('/v1/runs')).json();
    if (!Array.isArray(body.runs)) throw new Error('Invalid runs response');
    return body.runs.map(mapRun);
  }
  async run(id: string) {
    const body: { run: PublicRun } = await (await this.request(`/v1/runs/${encodeURIComponent(id)}`)).json();
    if (!body.run || body.run.id !== id) throw new Error('Invalid run response');
    return mapRun(body.run);
  }
  async webhooks() {
    const body: PublicWebhook[] = await (await this.request('/v1/organizations/me/webhooks')).json();
    if (!Array.isArray(body)) throw new Error('Invalid webhooks response');
    return body.map(mapWebhook);
  }
  async createWebhook(target: WebhookItem['target'], url: string) {
    const body: PublicWebhook = await (await this.request('/v1/organizations/me/webhooks', 'POST', { target_type: target, url, events: 'regression,failure' })).json();
    return mapWebhook(body);
  }
  async deleteWebhook(id: string) { await this.request(`/v1/organizations/me/webhooks/${encodeURIComponent(id)}`, 'DELETE'); }
  async testWebhook(id: string) { await this.request(`/v1/organizations/me/webhooks/${encodeURIComponent(id)}/test`, 'POST'); }
  async createToken(): Promise<{ token: string; role: 'member'; organization_id: string }> {
    return (await this.request('/v1/organizations/me/tokens', 'POST', { name: 'CI Token' })).json();
  }
}
