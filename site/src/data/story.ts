import type { RecordedRun } from './traces';

export type LostUpdateStory = {
  seed: number;
  start: number;
  amountA: number;
  afterA: number;
  amountB: number;
  afterB: number;
  expected: number;
  lost: number;
  minimal: number;
};

/**
 * Derives the numbers of the landing story from a recorded lost update, so the
 * copy can never drift from what the engine actually did. "A" is the first
 * transaction to write the balance, "B" the one that overwrites it.
 */
export function lostUpdateStory(run: RecordedRun): LostUpdateStory {
  const writes = run.minimalTrace
    .map((e) => ({ op: e.op, match: /UPDATE accounts SET balance = (\d+)/i.exec(e.sql) }))
    .filter((w): w is { op: number; match: RegExpExecArray } => w.match !== null);
  if (writes.length !== 2) throw new Error(`${run.scenario}: expected two balance writes, got ${writes.length}`);

  const amount = (op: number) => {
    const value = Number(run.minimalOps.find((o) => o.id === op)?.params.amount);
    if (!Number.isFinite(value)) throw new Error(`${run.scenario}: op ${op} has no amount`);
    return value;
  };
  const [a, b] = writes;
  const amountA = amount(a.op);
  const amountB = amount(b.op);
  const afterA = Number(a.match[1]);
  const afterB = Number(b.match[1]);
  const start = afterA + amountA;
  if (afterB + amountB !== start) throw new Error(`${run.scenario}: both writes must start from the same read`);

  return {
    seed: run.seed,
    start,
    amountA,
    afterA,
    amountB,
    afterB,
    expected: start - amountA - amountB,
    lost: amountA,
    minimal: run.shrink.minimalOps,
  };
}
