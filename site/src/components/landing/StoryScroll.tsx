import { FileCode2 } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { RECORDED_RUNS } from '../../data/traces';
import { lostUpdateStory, storyBeats } from '../../data/story';
import { format, messages, type Language } from '../../i18n';
import { SectionHeader } from '../system/Section';
import { money, TracePanel } from './TracePanel';
import styles from './StoryScroll.module.css';

const RUN = RECORDED_RUNS.banking;
const STORY = lostUpdateStory(RUN);
const BEATS = storyBeats(RUN);
const STEPS = ['read', 'write', 'lost', 'repro'] as const;
type Step = (typeof STEPS)[number];
const BEAT_OF: Record<Step, number> = {
  read: BEATS.bothRead,
  write: BEATS.firstCommit,
  lost: BEATS.lostWrite,
  repro: BEATS.end,
};

const VALUES = {
  seed: STORY.seed,
  start: money(STORY.start),
  amountA: money(STORY.amountA),
  afterA: money(STORY.afterA),
  amountB: money(STORY.amountB),
  afterB: money(STORY.afterB),
  expected: money(STORY.expected),
  lost: money(STORY.lost),
  minimal: STORY.minimal,
};

/**
 * "A lost update in four steps": the text scrolls, the recorded run stays in
 * view and advances to the milestone each step describes.
 */
export function StoryScroll({ lang }: { lang: Language }) {
  const m = messages[lang];
  const [step, setStep] = useState<Step>('read');
  const refs = useRef<Partial<Record<Step, HTMLLIElement | null>>>({});

  useEffect(() => {
    if (typeof IntersectionObserver === 'undefined') return;
    // A step becomes active when it crosses a band just below the middle of
    // the viewport, which stays clear of the pinned panel on phones.
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) setStep((entry.target as HTMLElement).dataset.step as Step);
        }
      },
      { rootMargin: '-55% 0px -35% 0px' }
    );
    for (const el of Object.values(refs.current)) if (el) observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return (
    <div className={styles.story}>
      <SectionHeader id="story-title" title={m.story.title} lead={format(m.story.lead, VALUES)} />
      <div className={styles.layout}>
        <ol className={styles.steps}>
          {STEPS.map((id) => (
            <li
              key={id}
              data-step={id}
              ref={(el) => {
                refs.current[id] = el;
              }}
              className={[styles.step, step === id ? styles.active : ''].join(' ')}
              aria-current={step === id ? 'step' : undefined}
            >
              <h3 className={styles.stepTitle}>{m.story.steps[id].title}</h3>
              <p className={styles.stepBody}>{format(m.story.steps[id].body, VALUES)}</p>
            </li>
          ))}
        </ol>
        <div className={styles.stage}>
          <TracePanel
            lang={lang}
            active={BEAT_OF[step]}
            footer={
              <p className={[styles.repro, step === 'repro' ? styles.reproShown : ''].join(' ')}>
                <FileCode2 size={16} aria-hidden="true" />
                <code>repro_test.go</code>
                <span>
                  {format(m.landingUi.shrinkFromTo, { original: RUN.shrink.originalOps, minimal: RUN.shrink.minimalOps })}
                </span>
              </p>
            }
          />
        </div>
      </div>
      <div className={styles.terms}>
        <h3 className={styles.termsTitle}>{m.landingUi.termsTitle}</h3>
        <dl className={styles.termList}>
        <div>
          <dt>{m.landingUi.termInvariant}</dt>
          <dd>{m.story.glossary.invariant}</dd>
        </div>
        <div>
          <dt>{m.landingUi.termIsolation}</dt>
          <dd>{m.story.glossary.isolation}</dd>
        </div>
        <div>
          <dt>{m.landingUi.termSeed}</dt>
          <dd>{m.story.glossary.seed}</dd>
        </div>
        </dl>
      </div>
    </div>
  );
}
