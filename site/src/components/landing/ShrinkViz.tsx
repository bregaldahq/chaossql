import { useRef, type CSSProperties } from 'react';
import type { RecordedRun } from '../../data/traces';
import { format, messages, type Language } from '../../i18n';
import { useInView } from './hooks';
import styles from './ShrinkViz.module.css';

/**
 * The failing run's transactions, in schedule order. Once in view, every
 * transaction delta debugging removed fades out, one after another, leaving
 * the minimal set. Reduced motion shows the end state (CSS).
 */
export function ShrinkViz({ run, lang }: { run: RecordedRun; lang: Language }) {
  const t = messages[lang].landingUi;
  const ref = useRef<HTMLDivElement>(null);
  const shown = useInView(ref, { once: true, threshold: 0.4 });
  const keep = new Set(run.minimalOps.map((op) => op.id));
  let pruneOrder = 0;

  return (
    <div ref={ref} className={[styles.viz, shown ? styles.run : ''].join(' ')}>
      <p className="sr-only">{format(t.shrinkLabel, { minimal: run.shrink.minimalOps })}</p>
      <ol className={styles.ops} aria-hidden="true">
        {run.rawOps.map((op) => {
          const kept = keep.has(op.id);
          return (
            <li
              key={op.id}
              className={kept ? styles.kept : styles.pruned}
              style={{ '--i': kept ? 0 : pruneOrder++ } as CSSProperties}
            >
              <span>#{op.id}</span>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
