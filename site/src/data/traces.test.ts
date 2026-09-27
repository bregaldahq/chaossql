import { describe, expect, it } from 'vitest';
import { RECORDED_RUNS } from './traces';

describe.each(Object.entries(RECORDED_RUNS))('recorded run %s', (_name, run) => {
  it('is a shrunk violation recorded from its example', () => {
    expect(run.status).toBe('violation');
    expect(run.source).toBe(`examples/${run.scenario}/chaos.yaml`);
    expect(run.isolation).toBe('READ_UNCOMMITTED');
  });

  it('has shrink numbers consistent with its operations', () => {
    expect(run.rawOps).toHaveLength(run.shrink.originalOps);
    expect(run.minimalOps).toHaveLength(run.shrink.minimalOps);
    const pct = ((run.shrink.originalOps - run.shrink.minimalOps) / run.shrink.originalOps) * 100;
    expect(run.shrink.reductionPct).toBeCloseTo(pct, 1);
    expect(run.shrink.reductionPct).toBeGreaterThanOrEqual(80);
  });

  it('keeps only minimal operations, each taken from the full schedule', () => {
    const raw = new Set(run.rawOps.map((op) => op.id));
    const minimal = new Set(run.minimalOps.map((op) => op.id));
    for (const id of minimal) expect(raw.has(id)).toBe(true);
    for (const event of run.minimalTrace) expect(minimal.has(event.op)).toBe(true);
    for (const id of minimal) {
      const events = run.minimalTrace.filter((e) => e.op === id);
      expect(events[0].type).toBe('BEGIN');
      expect(events[events.length - 1].type).toBe('COMMIT');
    }
  });
});
