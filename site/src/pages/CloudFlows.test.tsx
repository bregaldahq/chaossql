// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { DashboardPage } from './DashboardPage';
import { PricingPage } from './PricingPage';
import { VisualizerPage } from './VisualizerPage';
import { WaitlistForm } from '../components/forms/WaitlistForm';
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
    fireEvent.click(screen.getByRole('button', { name: 'Create CI token' }));
    if (allowed) expect(await screen.findByText('new-ci-token')).toBeTruthy();
    else {
      expect(await screen.findByRole('alert')).toHaveProperty('textContent', expect.stringContaining('Administrator permission required'));
      expect(screen.getByText(/chaossql server create-token/)).toBeTruthy();
      expect(screen.queryByRole('button', { name: 'Copy token' })).toBeNull();
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
    expect(await screen.findByText(/RUN DETAILS \/\/ older-run/)).toBeTruthy();
    expect(requests).toContain('/v1/runs/older-run');
    expect(screen.queryByText('NEVER RENDER THIS')).toBeNull();
  });
});

it.each(['network', 'unacknowledged', 'HTTP'])('waitlist does not fake success on %s failure', async failure => {
  vi.stubGlobal('fetch', async () => {
    if (failure === 'network') throw new Error('offline');
    return Response.json({ success: true, lead: { dispatched: false } }, { status: failure === 'HTTP' ? 502 : 200 });
  });
  render(<WaitlistForm lang="en" source="landing_page" />);
  fireEvent.change(screen.getByLabelText('Work email'), { target: { value: 'p@example.com' } });
  fireEvent.submit(screen.getByRole('button', { name: 'Join the waitlist' }).closest('form')!);
  expect(await screen.findByText('We could not send this right now. Please try again.')).toBeTruthy();
  expect(screen.queryByText('You are on the list')).toBeNull();
});

it('waitlist needs only an email and confirms after the lead is acknowledged', async () => {
  let body: Record<string, unknown> = {};
  vi.stubGlobal('fetch', async (_url: string, init: RequestInit) => {
    body = JSON.parse(init.body as string);
    return Response.json({ success: true, lead: { dispatched: true } });
  });
  render(<WaitlistForm lang="en" source="landing_page" />);
  expect(screen.queryByLabelText('Name')).toBeNull();
  fireEvent.change(screen.getByLabelText('Work email'), { target: { value: 'p@example.com' } });
  fireEvent.submit(screen.getByRole('button', { name: 'Join the waitlist' }).closest('form')!);
  expect(await screen.findByText('You are on the list')).toBeTruthy();
  expect(body).toMatchObject({ email: 'p@example.com', source: 'landing_page' });
  expect(body).not.toHaveProperty('name');
});

it('waitlist rejects an invalid email without sending', () => {
  const fetchSpy = vi.fn();
  vi.stubGlobal('fetch', fetchSpy);
  render(<WaitlistForm lang="en" source="landing_page" />);
  fireEvent.change(screen.getByLabelText('Work email'), { target: { value: 'not-an-email' } });
  fireEvent.submit(screen.getByRole('button', { name: 'Join the waitlist' }).closest('form')!);
  expect(screen.getByText('Please enter a valid email.')).toBeTruthy();
  expect(fetchSpy).not.toHaveBeenCalled();
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

it('pricing plan buttons preselect the plan and send it with the billing cycle', async () => {
  let body: Record<string, unknown> = {};
  vi.stubGlobal('fetch', async (_url: string, init: RequestInit) => {
    body = JSON.parse(init.body as string);
    return Response.json({ success: true, lead: { dispatched: true } });
  });
  render(<PricingPage lang="en" />);
  fireEvent.click(screen.getByRole('button', { name: 'Monthly' }));
  expect(screen.getByText('$39')).toBeTruthy();
  fireEvent.click(screen.getAllByRole('button', { name: 'Join the Cloud waitlist' })[1]);
  expect((screen.getByLabelText('Plan you are interested in') as HTMLSelectElement).value).toBe('team');
  const waitlist = screen.getByRole('heading', { name: 'Join the Cloud waitlist' }).parentElement!;
  fireEvent.change(within(waitlist).getByLabelText('Work email'), { target: { value: 'lead@example.com' } });
  fireEvent.submit(within(waitlist).getByRole('button', { name: 'Join the waitlist' }).closest('form')!);
  expect(await screen.findByText('You are on the list')).toBeTruthy();
  expect(body).toMatchObject({ email: 'lead@example.com', plan: 'team', billingCycle: 'monthly', source: 'pricing_page' });
});

it('pricing lists only features the Cloud actually provides', () => {
  const { container } = render(<PricingPage lang="en" />);
  const text = container.textContent ?? '';
  for (const claim of [/PagerDuty/i, /nightly fuzzing/i, /downloadable/i, /most popular/i, /isolation policies/i, /Apache/i]) {
    expect(text).not.toMatch(claim);
  }
});

it('audit form requires the qualifying fields before sending', () => {
  const fetchSpy = vi.fn();
  vi.stubGlobal('fetch', fetchSpy);
  render(<PricingPage lang="en" />);
  fireEvent.submit(screen.getByRole('button', { name: 'Book the audit' }).closest('form')!);
  expect(screen.getByText('Please fill in your name, a valid work email and your company.')).toBeTruthy();
  expect(fetchSpy).not.toHaveBeenCalled();
});

it('audit form reports delivery failures and sends the audit intent and timeline', async () => {
  let body: Record<string, unknown> = {};
  vi.stubGlobal('fetch', async (_url: string, init: RequestInit) => {
    body = JSON.parse(init.body as string);
    return new Response('delivery failed', { status: 502 });
  });
  render(<PricingPage lang="en" />);
  fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Test Person' } });
  fireEvent.change(screen.getByLabelText('Work email', { selector: '#audit-email' }), { target: { value: 'test@example.com' } });
  fireEvent.change(screen.getByLabelText('Company'), { target: { value: 'Team' } });
  fireEvent.click(screen.getByLabelText('In 1 to 3 months'));
  fireEvent.submit(screen.getByRole('button', { name: 'Book the audit' }).closest('form')!);
  await waitFor(() => expect(screen.getByText('We could not send this right now. Please try again.')).toBeTruthy());
  expect(screen.queryByText('Request received')).toBeNull();
  expect(body).toMatchObject({ wantAudit: true, source: 'pricing_page', timeline: 'In 1 to 3 months', company: 'Team' });
  expect(body.plan).toMatch(/Audit/);
});

describe.each(['en', 'pt'] as const)('dashboard demo in %s', (lang) => {
  it('shows relative times and labels in the page language', () => {
    render(<DashboardPage lang={lang} />);
    const text = document.body.textContent ?? '';
    if (lang === 'en') {
      expect(text).not.toMatch(/atrás|passados|Baseline Verificado|Livre de Anomalias|Sem anomalias|Quebra vs/);
      expect(text).toMatch(/12 min\. ago|12 min ago/);
    } else {
      expect(text).toMatch(/há 12 min/);
      expect(text).toContain('Baseline verificado');
      expect(text).not.toMatch(/\bpassed\b|with anomalies|RUN FINDING EXPLORER/);
    }
    cleanup();
  });
});

it('labels the demo write skew run with its Adya class A5B', () => {
  render(<DashboardPage lang="en" />);
  expect(document.body.textContent).toContain('Write Skew (A5B)');
  expect(document.body.textContent).not.toContain('Write Skew (A5A)');
});
