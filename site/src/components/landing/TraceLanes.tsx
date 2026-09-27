import type { CSSProperties } from 'react';
import type { TraceEvent } from '../../data/traces';
import styles from './TraceLanes.module.css';

export interface TraceLanesProps {
  trace: readonly TraceEvent[];
  /** Index of the latest event shown; later events render as upcoming. -1 shows none. */
  active: number;
  /** Event that silently overwrites another transaction; marked once shown. */
  lostIndex?: number;
  /** Accessible summary of what the lanes show. */
  label: string;
  className?: string;
}

type Kind = 'begin' | 'read' | 'write' | 'commit' | 'rollback';

function kindOf(event: TraceEvent): Kind {
  if (event.type === 'BEGIN') return 'begin';
  if (event.type === 'COMMIT') return 'commit';
  if (event.type === 'ROLLBACK') return 'rollback';
  return /^\s*(UPDATE|INSERT|DELETE)/i.test(event.sql) ? 'write' : 'read';
}

/** Lane number per worker, in order of first appearance (T1, T2...). */
export function laneOf(trace: readonly TraceEvent[]): Map<number, number> {
  const lanes = new Map<number, number>();
  for (const e of trace) if (!lanes.has(e.worker)) lanes.set(e.worker, lanes.size + 1);
  return lanes;
}

/**
 * One row per transaction, one block per statement, in the order the database
 * executed them. Rendered from a recorded run, never from hand-made data.
 */
export function TraceLanes({ trace, active, lostIndex, label, className }: TraceLanesProps) {
  const lanes = laneOf(trace);
  const current = trace[active];

  return (
    <div className={[styles.root, className].filter(Boolean).join(' ')}>
      <div
        className={styles.grid}
        style={{ '--steps': trace.length, '--lanes': lanes.size } as CSSProperties}
        role="img"
        aria-label={label}
      >
        {[...lanes.values()].map((lane) => (
          <span key={`label-${lane}`} className={styles.laneLabel} style={{ gridRow: lane }}>
            T{lane}
          </span>
        ))}
        {[...lanes.values()].map((lane) => (
          <span key={`rail-${lane}`} className={styles.rail} style={{ gridRow: lane }} aria-hidden="true" />
        ))}
        {trace.map((event, i) => {
          const state = i > active ? 'upcoming' : i === active ? 'current' : 'done';
          const lost = lostIndex === i && i <= active;
          return (
            <span
              key={i}
              aria-hidden="true"
              className={[styles.block, styles[kindOf(event)], styles[state], lost ? styles.lost : ''].join(' ')}
              style={{ gridRow: lanes.get(event.worker), gridColumn: i + 2 }}
            />
          );
        })}
      </div>
      <p className={styles.now} aria-hidden="true">
        {current ? (
          <>
            <span className={styles.nowTx}>T{lanes.get(current.worker)}</span>
            <code className={lostIndex === active ? styles.nowLost : undefined}>{current.sql}</code>
          </>
        ) : (
          <code>&nbsp;</code>
        )}
      </p>
      <ol className="sr-only">
        {trace.slice(0, active + 1).map((event, i) => (
          <li key={i}>
            T{lanes.get(event.worker)}: {event.sql}
          </li>
        ))}
      </ol>
    </div>
  );
}
