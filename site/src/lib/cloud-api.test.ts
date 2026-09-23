import { describe, expect, it } from 'vitest';
import { CloudAPI, mapRun } from './cloud-api';

describe('cloud API contract', () => {
  it('keeps absent legacy metadata unavailable and zero measurements intact', () => {
    const item = mapRun({ id: 'r', repo_id: 'repo', scenario_id: 's', status: 'failed', anomaly_type: 'P4', seed: 0, duration_ms: 0 });
    expect(item).toMatchObject({ seed: 0, durationMS: 0, schedulesCount: null, failedSchedules: null, isRegression: null, isolation: '—', repo: 'repo', scenario: 's' });
  });
  it('uses organization routes, string events, server IDs and private token issuance', async () => {
    const requests: { url: string; init: RequestInit }[] = [];
    const hook = { id: 'wh-server', target_type: 'slack', url: 'https://hooks.slack.com/***', events: 'regression,failure', active: true, created_at: '2026-09-17' };
    const api = new CloudAPI('https://cloud.example/', 'admin', async (input, init) => {
      const url = String(input); requests.push({ url, init: init! });
      if (url.endsWith('/tokens')) return Response.json({ token: 'new-member-token', role: 'member', organization_id: 'org1' });
      if (init?.method === 'DELETE' || url.endsWith('/test')) return new Response(null, { status: 204 });
      return Response.json(init?.method === 'POST' ? hook : [hook]);
    });
    expect((await api.webhooks())[0]).toMatchObject({ id: 'wh-server', target: 'slack', events: ['regression', 'failure'] });
    expect((await api.createWebhook('slack', 'https://hooks.slack.com/secret')).id).toBe('wh-server');
    await api.testWebhook('wh-server');
    await api.deleteWebhook('wh-server');
    expect((await api.createToken()).token).toBe('new-member-token');
    expect(requests.map(r => r.url)).toEqual([
      'https://cloud.example/v1/organizations/me/webhooks', 'https://cloud.example/v1/organizations/me/webhooks',
      'https://cloud.example/v1/organizations/me/webhooks/wh-server/test', 'https://cloud.example/v1/organizations/me/webhooks/wh-server',
      'https://cloud.example/v1/organizations/me/tokens',
    ]);
    expect(JSON.parse(requests[1].init.body as string)).toEqual({ target_type: 'slack', url: 'https://hooks.slack.com/secret', events: 'regression,failure' });
    expect(requests.every(r => new Headers(r.init.headers).get('Authorization') === 'Bearer admin')).toBe(true);
  });
  it.each([401, 403, 500])('propagates HTTP %s on mutations', async status => {
    const api = new CloudAPI('', 'token', async () => new Response('failure', { status }));
    await expect(api.createWebhook('generic', 'https://example.com')).rejects.toThrow(`HTTP ${status}`);
    await expect(api.deleteWebhook('real')).rejects.toThrow(`HTTP ${status}`);
    await expect(api.testWebhook('real')).rejects.toThrow(`HTTP ${status}`);
    await expect(api.createToken()).rejects.toThrow(`HTTP ${status}`);
  });
});
