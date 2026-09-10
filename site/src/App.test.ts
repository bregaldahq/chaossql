import { describe, expect, it } from 'vitest';

import { routeFromHash } from './App';

describe('routeFromHash', () => {
  it.each([
    ['', 'landing'],
    ['#/dashboard', 'dashboard'],
    ['#/docs/getting-started', 'docs'],
    ['#/scenarios', 'scenarios'],
    ['#/visualizer', 'visualizer'],
    ['#/matrix', 'matrix'],
    ['#/playground', 'playground'],
    ['#/pricing', 'pricing'],
    ['#/unknown', 'landing'],
  ])('maps %s to %s', (hash, expected) => {
    expect(routeFromHash(hash)).toBe(expected);
  });
});
