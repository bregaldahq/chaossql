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

export interface StoryBeats {
  /** Index of the last read before anyone writes: both transactions saw the same balance. */
  bothRead: number;
  /** Index of the first COMMIT: transaction A's write is durable. */
  firstCommit: number;
  /** Index of the write that silently overwrites A. */
  lostWrite: number;
  /** Last event of the trace. */
  end: number;
}

const isBalanceWrite = (sql: string) => /^\s*UPDATE accounts SET balance/i.test(sql);

/** Milestones of the lost update, located in the recorded trace. */
export function storyBeats(run: RecordedRun): StoryBeats {
  const trace = run.minimalTrace;
  const firstWrite = trace.findIndex((e) => isBalanceWrite(e.sql));
  const beats = {
    bothRead: firstWrite - 1,
    firstCommit: trace.findIndex((e) => e.type === 'COMMIT'),
    lostWrite: trace.map((e) => isBalanceWrite(e.sql)).lastIndexOf(true),
    end: trace.length - 1,
  };
  if (beats.bothRead < 0 || beats.firstCommit < 0 || beats.lostWrite <= beats.firstCommit) {
    throw new Error(`${run.scenario}: trace does not have the shape of a lost update`);
  }
  return beats;
}

/** Balance stored in the row after event `index` (reads do not change it). */
export function balanceAfter(run: RecordedRun, index: number, start: number): number {
  let balance = start;
  for (const event of run.minimalTrace.slice(0, index + 1)) {
    const match = /UPDATE accounts SET balance = (\d+)/i.exec(event.sql);
    if (match) balance = Number(match[1]);
  }
  return balance;
}
