// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { DashboardPage } from './DashboardPage';
import { PricingPage } from './PricingPage';
import { VisualizerPage } from './VisualizerPage';
import { CloudWaitlistSection } from '../components/ui/CloudWaitlistSection';
import { navigate } from '../lib/router';

afterEach(() => { cleanup(); vi.unstubAllGlobals(); window.history.replaceState(null, '', '/'); });

const run = { id: 'run-real', repo_id: 'repo-1', repo_full_name: 'team/service', scenario_id: 'scenario-1', scenario_name: 'actual_scenario', branch: 'feature', commit_sha: 'abcdef123', pr_number: 8, status: 'failed', anomaly_type: 'P4', is_regression: false, driver: 'postgres', seed: 0, duration_ms: 0, created_at: '2026-09-17T10:00:00Z' };

async function connect() {
  fireEvent.click(screen.getByRole('button', { name: /Live Cloud/i }));
  fireEvent.change(screen.getByLabelText('API token'), { target: { value: 'private-owner-token' } });
  fireEvent.click(screen.getByRole('button', { name: /Refresh/i }));
  await screen.findAllByText('team/service');
}

describe('live dashboard', () => {
  it('authenticates same-origin requests, displays real metadata and never substitutes demo evidence', async () => {
    const requests: [string, RequestInit | undefined][] = [];
    vi.stubGlobal('fetch', async (url: string, init?: RequestInit) => {
      requests.push([url, init]);
      return Response.json({ runs: [run] });
    });
    render(<DashboardPage lang="en" />);
    await connect();
    expect(requests[0][0]).toBe('/v1/runs');
    expect(new Headers(requests[0][1]?.headers).get('Authorization')).toBe('Bearer private-owner-token');
    expect(screen.getByText('actual_scenario')).toBeTruthy();
    expect(screen.getByText('0ms')).toBeTruthy();
    expect(screen.queryByText(/REGRESSION \(Base:/)).toBeNull();
    expect(screen.queryByText('100 sched')).toBeNull();
    expect(screen.queryByText('READ COMMITTED')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Inspect' }));
    expect(screen.getByText(/local CI artifacts/i)).toBeTruthy();
    expect(screen.queryByRole('button', { name: /Download/i })).toBeNull();
    expect(screen.queryByRole('link', { name: /Visualizer/i })).toBeNull();
    expect(localStorage.getItem('private-owner-token')).toBeNull();
  });

  it.each([401, 403, 500])('shows HTTP %s without retaining demo runs', async (status) => {
    vi.stubGlobal('fetch', async () => new Response('denied', { status }));
    render(<DashboardPage lang="en" />);
    fireEvent.click(screen.getByRole('button', { name: /Live Cloud/i }));
    fireEvent.change(screen.getByLabelText('API token'), { target: { value: 'token' } });
    fireEvent.click(screen.getByRole('button', { name: /Refresh/i }));
    expect(await screen.findByRole('alert')).toBeTruthy();
    expect(screen.queryByText('acme/payments')).toBeNull();
    expect(screen.queryByText(/not received runs yet/i)).toBeNull();
  });

  it('replaces data with an empty response and reports unavailable health', async () => {
    vi.stubGlobal('fetch', async () => Response.json({ runs: [] }));
    render(<DashboardPage lang="en" />);
    fireEvent.click(screen.getByRole('button', { name: /Live Cloud/i }));
    fireEvent.change(screen.getByLabelText('API token'), { target: { value: 'token' } });
    fireEvent.click(screen.getByRole('button', { name: /Refresh/i }));
    await screen.findByText('Connected to API');
    expect(screen.queryByText('100.0%')).toBeNull();
    expect(screen.queryByText('acme/payments')).toBeNull();
  });

  it('saves server webhook IDs, reports failed tests and deletes only after server success', async () => {
    const hook = { id: 'wh-real-id', target_type: 'discord', url: 'https://discord.com/***', events: 'regression,failure', active: true, created_at: '2026-09-17' };
    const requests: string[] = [];
    let failDelete = true;
    vi.stubGlobal('fetch', async (url: string, init: RequestInit) => {
      requests.push(`${init.method} ${url}`);
      if (url === '/v1/runs') return Response.json({ runs: [run] });
      if (url.endsWith('/test')) return new Response(null, { status: 502 });
      if (init.method === 'DELETE') return new Response(null, { status: failDelete ? 500 : 204 });
      if (init.method === 'POST') return Response.json(hook);
      return Response.json([]);
    });
    render(<DashboardPage lang="en" />);
    await connect();
    fireEvent.click(screen.getByRole('button', { name: 'Alerts & Webhooks' }));
    await waitFor(() => expect(screen.queryByText('Loading…')).toBeNull());
    fireEvent.change(screen.getByPlaceholderText('https://discord.com/api/webhooks/...'), { target: { value: 'https://discord.com/api/webhooks/test/secret' } });
    fireEvent.click(screen.getByRole('button', { name: /Save Webhook/ }));
    await screen.findByText('wh-real-id');
    fireEvent.click(screen.getByRole('button', { name: 'Test Alert' }));
    await screen.findByRole('alert');
    expect(screen.queryByText('Delivered!')).toBeNull();
    expect(requests).toContain('POST /v1/organizations/me/webhooks/wh-real-id/test');
    fireEvent.click(screen.getByRole('button', { name: 'Remove' }));
    await screen.findByText(/HTTP 500/);
    expect(screen.getByText('wh-real-id')).toBeTruthy();
    failDelete = false;
    fireEvent.click(screen.getByRole('button', { name: 'Remove' }));
    await waitFor(() => expect(screen.queryByText('wh-real-id')).toBeNull());
  });

  it.each([true, false])('creates only real CI tokens or explains administrator permission (allowed=%s)', async allowed => {
    vi.stubGlobal('fetch', async (url: string, init: RequestInit) => {
      if (url === '/v1/runs') return Response.json({ runs: [run] });
      expect(url).toBe('/v1/organizations/me/tokens');
      expect(JSON.parse(init.body as string)).toEqual({ name: 'CI Token' });
      return allowed ? Response.json({ token: 'new-ci-token', role: 'member', organization_id: 'org' }) : new Response(null, { status: 403 });
    });
    render(<DashboardPage lang="en" />);
    await connect();
    fireEvent.click(screen.getByRole('button', { name: /Connect Repository/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Create CI Token' }));
    if (allowed) expect(await screen.findByText('new-ci-token')).toBeTruthy();
    else {
      expect(await screen.findByRole('alert')).toHaveProperty('textContent', expect.stringContaining('Administrator permission required'));
      expect(screen.getByText(/chaossql server create-token/)).toBeTruthy();
      expect(screen.queryByRole('button', { name: 'Copy Token' })).toBeNull();
    }
  });

  it('opens a linked run using authenticated metadata even when outside the recent list', async () => {
    window.history.replaceState(null, '', '/dashboard?run=older-run');
    const requests: string[] = [];
    vi.stubGlobal('fetch', async (url: string, init: RequestInit) => {
      requests.push(url);
      expect(new Headers(init.headers).get('Authorization')).toBe('Bearer private-owner-token');
      return Response.json(url === '/v1/runs' ? { runs: [run] } : { run: { ...run, id: 'older-run' }, finding: { repro_code: 'NEVER RENDER THIS' } });
    });
    render(<DashboardPage lang="en" />);
    await connect();
    expect(await screen.findByText(/RUN FINDING EXPLORER \/\/ older-run/)).toBeTruthy();
    expect(requests).toContain('/v1/runs/older-run');
    expect(screen.queryByText('NEVER RENDER THIS')).toBeNull();
  });
});

it.each(['network', 'unacknowledged', 'HTTP'])('waitlist does not fake success on %s failure', async failure => {
  vi.stubGlobal('fetch', async () => {
    if (failure === 'network') throw new Error('offline');
    return Response.json({ success: true, lead: { dispatched: false } }, { status: failure === 'HTTP' ? 502 : 200 });
  });
  render(<CloudWaitlistSection lang="en" />);
  fireEvent.change(screen.getByLabelText(/Your Name/), { target: { value: 'Person' } });
  fireEvent.change(screen.getByLabelText(/Work Email/), { target: { value: 'p@example.com' } });
  fireEvent.submit(screen.getByRole('button', { name: 'Request Early Access' }).closest('form')!);
  expect(await screen.findByRole('alert')).toBeTruthy();
  expect(screen.queryByText('Registration Confirmed!')).toBeNull();
});

it('refuses to present a static trace as a linked CI finding', () => {
  window.history.replaceState(null, '', '/visualizer?scenario=private-scenario&seed=999');
  render(<VisualizerPage lang="en" />);
  expect(screen.getByText(/local CI artifacts/i)).toBeTruthy();
  expect(screen.queryByText('Inspecting CI Finding')).toBeNull();
  expect(screen.queryByRole('button', { name: 'Raw Trace (20 ops)' })).toBeNull();
});

it('reacts to CI links and demo links while the visualizer remains mounted', async () => {
  render(<VisualizerPage lang="en" />);
  expect(screen.getByRole('button', { name: 'Raw Trace (20 ops)' })).toBeTruthy();
  act(() => navigate('/visualizer?run_id=real-run'));
  expect(await screen.findByText(/local CI artifacts/i)).toBeTruthy();
  act(() => navigate('/visualizer'));
  expect(await screen.findByRole('button', { name: 'Raw Trace (20 ops)' })).toBeTruthy();
});

it('pricing waits for delivery acknowledgment before showing confirmation', async () => {
  let acknowledge: (response: Response) => void = () => {};
  vi.stubGlobal('fetch', () => new Promise<Response>(resolve => { acknowledge = resolve; }));
  render(<PricingPage lang="en" />);
  fireEvent.click(screen.getByRole('button', { name: 'Request Cloud Team' }));
  fireEvent.submit(screen.getByRole('button', { name: /Confirm Request/ }).closest('form')!);
  expect(screen.queryByText('Request Registered!')).toBeNull();
  acknowledge(Response.json({ success: true, lead: { dispatched: true } }));
  expect(await screen.findByText('Request Registered!')).toBeTruthy();
});

it('pricing reports delivery failures and preserves the requested plan and audit intent', async () => {
  let body: Record<string, unknown> = {};
  vi.stubGlobal('fetch', async (_url: string, init: RequestInit) => {
    body = JSON.parse(init.body as string);
    return new Response('delivery failed', { status: 502 });
  });
  render(<PricingPage lang="en" />);
  fireEvent.click(screen.getByRole('button', { name: /Concurrency Audit/i }));
  fireEvent.change(screen.getByPlaceholderText('Ricardo Bregalda'), { target: { value: 'Test Person' } });
  fireEvent.change(screen.getByPlaceholderText('ricardo@empresa.com'), { target: { value: 'test@example.com' } });
  fireEvent.change(screen.getByPlaceholderText('Acme Fintech'), { target: { value: 'Team' } });
  fireEvent.submit(screen.getByRole('button', { name: /Confirm Request/ }).closest('form')!);
  await waitFor(() => expect(screen.getByRole('alert')).toBeTruthy());
  expect(screen.queryByText('Request Registered!')).toBeNull();
  expect(body).toMatchObject({ billingCycle: 'annual', wantAudit: true, source: 'pricing_page' });
  expect(body.plan).toMatch(/Audit/);
});
