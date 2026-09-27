import { describe, expect, it } from 'vitest';
import { RECORDED_RUNS } from './traces';
import { balanceAfter, lostUpdateStory, storyBeats } from './story';
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

describe('storyBeats', () => {
  it('locates the lost update milestones in the banking trace', () => {
    const run = RECORDED_RUNS.banking;
    const beats = storyBeats(run);
    expect(run.minimalTrace[beats.bothRead].sql).toMatch(/^SELECT balance/);
    expect(run.minimalTrace[beats.firstCommit].type).toBe('COMMIT');
    expect(run.minimalTrace[beats.lostWrite].sql).toContain('balance = 989');
    expect(beats.end).toBe(run.minimalTrace.length - 1);
  });

  it('tracks the stored balance through the trace', () => {
    const run = RECORDED_RUNS.banking;
    const beats = storyBeats(run);
    expect(balanceAfter(run, beats.bothRead, 1000)).toBe(1000);
    expect(balanceAfter(run, beats.firstCommit, 1000)).toBe(981);
    expect(balanceAfter(run, beats.end, 1000)).toBe(989);
  });
});
