import { Pause, Play } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { RECORDED_RUNS } from '../../data/traces';
import { balanceAfter, lostUpdateStory, storyBeats } from '../../data/story';
import { messages, type Language } from '../../i18n';
import { Badge } from '../system/Badge';
import { TraceLanes } from './TraceLanes';
import { useInView, usePrefersReducedMotion } from './hooks';
import styles from './TracePanel.module.css';

const RUN = RECORDED_RUNS.banking;
const STORY = lostUpdateStory(RUN);
const BEATS = storyBeats(RUN);
const STEP_MS = 700;
const HOLD_MS = 2800;

export const money = (n: number) => `$${n.toLocaleString('en-US')}`;

export interface TracePanelProps {
  lang: Language;
  /** Controlled position (scroll story). Omit to autoplay (hero). */
  active?: number;
  /** Extra content under the readout, e.g. the repro file in the last story step. */
  footer?: React.ReactNode;
  className?: string;
}

/**
 * The recorded banking run: lanes, stored balance, and the invariant verdict.
 * Autoplays when uncontrolled, with a pause button (WCAG 2.2.2), stops while
 * off screen, and shows the final state when motion is reduced.
 */
export function TracePanel({ lang, active: controlled, footer, className }: TracePanelProps) {
  const t = messages[lang].landingUi;
  const ref = useRef<HTMLDivElement>(null);
  const visible = useInView(ref);
  const reduce = usePrefersReducedMotion();
  const [paused, setPaused] = useState(false);
  const [tick, setTick] = useState(-1);
  const autoplay = controlled === undefined;
  const running = autoplay && !reduce && !paused && visible;

  useEffect(() => {
    if (!running) return;
    const delay = tick >= BEATS.end ? HOLD_MS : STEP_MS;
    const timer = setTimeout(() => setTick((i) => (i >= BEATS.end ? -1 : i + 1)), delay);
    return () => clearTimeout(timer);
  }, [running, tick]);

  const active = controlled ?? (autoplay && (reduce || (paused && tick < 0)) ? BEATS.end : tick);
  const balance = balanceAfter(RUN, active, STORY.start);
  const finished = active >= BEATS.end;

  return (
    <div ref={ref} className={[styles.panel, className].filter(Boolean).join(' ')}>
      <div className={styles.head}>
        <span className={styles.title}>banking_lost_update</span>
        <span className={styles.meta}>
          seed {STORY.seed} · {RUN.isolation}
        </span>
        {autoplay && !reduce && (
          <button
            type="button"
            className={styles.toggle}
            onClick={() => setPaused((p) => !p)}
            aria-label={paused ? t.play : t.pause}
          >
            {paused ? <Play size={14} aria-hidden="true" /> : <Pause size={14} aria-hidden="true" />}
          </button>
        )}
      </div>

      <TraceLanes trace={RUN.minimalTrace} active={active} lostIndex={BEATS.lostWrite} label={t.traceLabel} />

      <dl className={styles.readout}>
        <div>
          <dt>{t.balance}</dt>
          <dd className={finished ? styles.bad : undefined}>{money(balance)}</dd>
        </div>
        <div>
          <dt>{t.expected}</dt>
          <dd>{money(STORY.expected)}</dd>
        </div>
        <div className={styles.verdict}>
          <dt>{t.invariant}</dt>
          <dd>
            {finished ? (
              <Badge tone="signal">
                {t.invariantFails}: {money(balance)} ≠ {money(STORY.expected)}
              </Badge>
            ) : (
              <span className={styles.pending}>{t.invariantPending}</span>
            )}
          </dd>
        </div>
      </dl>
      {footer}
    </div>
  );
}
