// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { Button } from './Button';
import { Tabs } from './Tabs';
import { Terminal } from './Terminal';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe('Button', () => {
  it('renders a link when given an href and a typed button otherwise', () => {
    render(
      <>
        <Button href="/playground">Run it</Button>
        <Button variant="secondary">Install</Button>
      </>
    );
    expect(screen.getByRole('link', { name: 'Run it' }).getAttribute('href')).toBe('/playground');
    expect(screen.getByRole('button', { name: 'Install' }).getAttribute('type')).toBe('button');
  });
});

describe('Tabs', () => {
  const items = [
    { id: 'banking', label: 'Banking', content: <p>ledger</p> },
    { id: 'hospital', label: 'Hospital', content: <p>on call</p> },
    { id: 'booking', label: 'Booking', content: <p>seats</p> },
  ];

  it('exposes one tab stop and links tabs to panels', () => {
    render(<Tabs label="Scenarios" items={items} />);
    const tabs = screen.getAllByRole('tab');
    expect(tabs.map((t) => t.tabIndex)).toEqual([0, -1, -1]);
    const panel = screen.getByRole('tabpanel');
    expect(panel.getAttribute('aria-labelledby')).toBe(tabs[0].id);
    expect(tabs[0].getAttribute('aria-controls')).toBe(panel.id);
  });

  it('moves selection and focus with arrow keys, Home and End, wrapping around', () => {
    const onChange = vi.fn();
    render(<Tabs label="Scenarios" items={items} onChange={onChange} />);
    const tab = (name: string) => screen.getByRole('tab', { name });

    fireEvent.keyDown(tab('Banking'), { key: 'ArrowLeft' });
    expect(tab('Booking').getAttribute('aria-selected')).toBe('true');
    expect(document.activeElement).toBe(tab('Booking'));

    fireEvent.keyDown(tab('Booking'), { key: 'ArrowRight' });
    expect(tab('Banking').getAttribute('aria-selected')).toBe('true');

    fireEvent.keyDown(tab('Banking'), { key: 'End' });
    fireEvent.keyDown(tab('Booking'), { key: 'Home' });
    expect(screen.getByRole('tabpanel').textContent).toBe('ledger');
    expect(onChange.mock.calls.map((c) => c[0])).toEqual(['booking', 'banking', 'booking', 'banking']);
  });

  it('respects a controlled value', () => {
    render(<Tabs label="Scenarios" items={items} value="hospital" />);
    fireEvent.click(screen.getByRole('tab', { name: 'Booking' }));
    expect(screen.getByRole('tabpanel').textContent).toBe('on call');
  });
});

describe('Terminal', () => {
  it('copies the code, reports it and confirms in the chosen language', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
    const onCopy = vi.fn();
    render(<Terminal code="go install ./cmd/chaossql" title="shell" prompt lang="pt" onCopy={onCopy} />);

    fireEvent.click(screen.getByRole('button', { name: 'Copiar: shell' }));
    expect(await screen.findByText('Copiado')).toBeTruthy();
    expect(writeText).toHaveBeenCalledWith('go install ./cmd/chaossql');
    expect(onCopy).toHaveBeenCalledOnce();
  });

  it('does not report a copy the browser refused', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('denied'));
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true });
    const onCopy = vi.fn();
    render(<Terminal code="SELECT 1" language="sql" onCopy={onCopy} />);
    fireEvent.click(screen.getByRole('button', { name: 'Copy: sql' }));
    await Promise.resolve();
    expect(onCopy).not.toHaveBeenCalled();
    expect(screen.queryByText('Copied')).toBeNull();
  });
});
