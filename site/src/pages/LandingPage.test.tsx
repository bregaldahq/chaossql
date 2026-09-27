// @vitest-environment jsdom
import { afterEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { LandingPage } from './LandingPage';
import { messages } from '../i18n';

afterEach(cleanup);

describe.each(['en', 'pt'] as const)('LandingPage (%s)', (lang) => {
  const m = messages[lang];

  it('renders the hero promise and its two actions', () => {
    render(<LandingPage lang={lang} />);
    expect(screen.getByRole('heading', { level: 1, name: m.hero.title })).toBeTruthy();
    const primary = screen.getAllByRole('link', { name: m.cta.primary });
    expect(primary[0].getAttribute('href')).toBe('/#story');
    expect(screen.getAllByRole('button', { name: new RegExp(m.landingUi.installTitle) }).length).toBeGreaterThan(0);
  });

  it('never leaks an unfilled {placeholder}', () => {
    const { container } = render(<LandingPage lang={lang} />);
    expect(container.textContent).not.toMatch(/\{\w+\}/);
  });

  it('tells the real lost update story from the recorded run', () => {
    const { container } = render(<LandingPage lang={lang} />);
    const text = container.textContent ?? '';
    expect(text).toContain('$989');
    expect(text).toContain('$970');
    expect(text).toContain('42');
  });

  it('links every scenario to the playground preset and every audit CTA to pricing', () => {
    render(<LandingPage lang={lang} />);
    expect(screen.getByRole('link', { name: m.cta.playground }).getAttribute('href')).toBe('/playground?scenario=banking');
    expect(screen.getByRole('link', { name: m.cta.audit }).getAttribute('href')).toBe('/pricing#audit');
  });

  it('answers every FAQ entry', () => {
    const { container } = render(<LandingPage lang={lang} />);
    expect(container.querySelectorAll('details')).toHaveLength(Object.keys(m.faq.items).length);
  });

  it('does not claim that CI blocks merges by itself', () => {
    const { container } = render(<LandingPage lang={lang} />);
    expect(container.textContent).not.toMatch(/blocks (the )?merge|bloqueia o merge caso/i);
  });
});
