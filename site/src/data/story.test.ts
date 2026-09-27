import { describe, expect, it } from 'vitest';
import { RECORDED_RUNS } from './traces';
import { lostUpdateStory } from './story';
import { format, messages } from '../i18n';

describe('lostUpdateStory', () => {
  it('reads the banking story straight from the recorded trace', () => {
    expect(lostUpdateStory(RECORDED_RUNS.banking)).toEqual({
      seed: 42,
      start: 1000,
      amountA: 19,
      afterA: 981,
      amountB: 11,
      afterB: 989,
      expected: 970,
      lost: 19,
      minimal: 2,
    });
  });

  it('fills every story placeholder in both languages', () => {
    const story = lostUpdateStory(RECORDED_RUNS.banking);
    for (const lang of ['en', 'pt'] as const) {
      const steps = messages[lang].story.steps;
      for (const step of Object.values(steps)) expect(() => format(step.body, story)).not.toThrow();
      expect(format(messages[lang].story.lead, story)).toContain('42');
    }
  });
});
