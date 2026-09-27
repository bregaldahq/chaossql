import { useRef } from 'react';
import { ArrowRight, ArrowUpRight, Play, Star } from 'lucide-react';
import { track } from '../lib/analytics';
import { RECORDED_RUNS, type RecordedRun } from '../data/traces';
import { format, messages, type Language } from '../i18n';
import { Badge } from '../components/system/Badge';
import { Button } from '../components/system/Button';
import { Section, SectionHeader } from '../components/system/Section';
import { Tabs } from '../components/system/Tabs';
import { Terminal } from '../components/system/Terminal';
import { useGitHubStars, useInView } from '../components/landing/hooks';
import { laneOf } from '../components/landing/TraceLanes';
import { TracePanel } from '../components/landing/TracePanel';
import { StoryScroll } from '../components/landing/StoryScroll';
import { ShrinkViz } from '../components/landing/ShrinkViz';
import { WaitlistForm } from '../components/forms/WaitlistForm';
import { INSTALL_CMD, InstallButton } from '../components/landing/InstallButton';
import reportDdmin from '../media/report-ddmin.webp';
import styles from './LandingPage.module.css';

export interface LandingPageProps {
  lang?: Language;
}

const GITHUB_URL = 'https://github.com/bregaldahq/chaossql';
const CLOUD_FROM_PRICE = '$39';
const AUDIT_PRICE = '$1,490';
const SCENARIOS = ['banking', 'inventory', 'hospital'] as const;
// Below this, a star count reads as negative social proof: show a plain link.
const MIN_STARS_SHOWN = 50;

// The workflow users copy: inputs as declared in action.yml.
const ACTION_YAML = `name: Concurrency Guard
on:
  pull_request:
    branches: [main]

permissions:
  contents: read
  pull-requests: write

jobs:
  concurrency:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: bregaldahq/chaossql@v1.6.0
        with:
          spec-path: chaos/transfers.yaml
          seed: '42'
          export-junit: chaossql-junit.xml
          export-repro: 'true'
          # Optional, ChaosSQL Cloud: baseline against main and PR comments
          cloud-token: \${{ secrets.CHAOSSQL_CLOUD_TOKEN }}
`;

/** Fades a block in the first time it enters the viewport (CSS handles reduced motion). */
function Reveal({ children, className }: { children: React.ReactNode; className?: string }) {
  const ref = useRef<HTMLDivElement>(null);
  const shown = useInView(ref, { once: true, rootMargin: '0px 0px -10% 0px' });
  return (
    <div ref={ref} className={[styles.reveal, shown ? styles.revealShown : '', className].filter(Boolean).join(' ')}>
      {children}
    </div>
  );
}

function traceText(run: RecordedRun): string {
  const lanes = laneOf(run.minimalTrace);
  return run.minimalTrace.map((e) => `T${lanes.get(e.worker)}  ${e.sql}`).join('\n');
}

export function LandingPage({ lang = 'en' }: LandingPageProps) {
  const m = messages[lang];
  const stars = useGitHubStars();
  const banking = RECORDED_RUNS.banking;
  const anomalyName = (type: string) => (type.startsWith('A5B') ? m.landingUi.anomalyA5B : m.landingUi.anomalyP4);

  return (
    <div className={styles.page}>
      {/* 1. Hero: promise, two actions, and the bug happening on the right. */}
      <Section spacing="tight" width="wide" labelledBy="hero-title" className={styles.heroSection}>
        <div className={styles.hero}>
          <div className={styles.heroCopy}>
            <h1 id="hero-title" className={styles.heroTitle}>
              {m.hero.title}
            </h1>
            <p className={styles.heroLead}>{m.hero.lead}</p>
            <div className={styles.actions}>
              <Button href="/#story" size="lg" icon={<Play />} onClick={() => track('cta_click', 'hero_story')}>
                {m.cta.primary}
              </Button>
              <InstallButton lang={lang} placement="hero" />
            </div>
          </div>
          <TracePanel lang={lang} className={styles.heroPanel} />
        </div>
      </Section>

      {/* 2. Proof strip: facts a reader can check. */}
      <div className={styles.proof}>
        <ul className={styles.proofList}>
          <li>{m.proof.engines}</li>
          <li>{m.proof.license}</li>
          <li>{m.proof.binary}</li>
          <li>
            <a href={GITHUB_URL} target="_blank" rel="noreferrer" onClick={() => track('outbound_click', 'github_proof')}>
              {stars !== null && stars >= MIN_STARS_SHOWN ? (
                <>
                  <Star size={14} aria-hidden="true" /> {format(m.proof.stars, { stars: stars.toLocaleString(lang === 'pt' ? 'pt-BR' : 'en-US') })}
                </>
              ) : (
                <>
                  GitHub <ArrowUpRight size={14} aria-hidden="true" />
                </>
              )}
            </a>
          </li>
        </ul>
      </div>

      {/* 3. The bug, step by step, on the recorded run. */}
      <Section width="wide" id="story" labelledBy="story-title">
        <StoryScroll lang={lang} />
      </Section>

      {/* 4. Other domains, same failure mode. */}
      <Section width="wide" id="scenarios" labelledBy="scenarios-title" className={styles.scenariosSection}>
        <Reveal>
          <SectionHeader id="scenarios-title" title={m.scenarios.title} lead={m.scenarios.lead} />
          <Tabs
            label={m.scenarios.title}
            onChange={(id) => track('scenario_view', id)}
            items={SCENARIOS.map((id) => {
              const run = RECORDED_RUNS[id];
              const copy = m.scenarios[id];
              return {
                id,
                label: copy.name,
                content: (
                  <div className={styles.scenario}>
                    <div className={styles.scenarioCopy}>
                      <Badge tone="signal">{anomalyName(run.anomalyType)}</Badge>
                      <p className={styles.scenarioPain}>{copy.pain}</p>
                      <div>
                        <p className={styles.scenarioLabel}>{m.landingUi.scenarioInvariant}</p>
                        <p className={styles.scenarioInvariant}>{copy.invariant}</p>
                      </div>
                      <Button
                        href={`/playground?scenario=${id}`}
                        variant="secondary"
                        trailingIcon={<ArrowRight />}
                        onClick={() => track('cta_click', `scenario_playground:${id}`)}
                      >
                        {m.cta.playground}
                      </Button>
                    </div>
                    <figure className={styles.scenarioTrace}>
                      <Terminal
                        lang={lang}
                        title={m.landingUi.scenarioTrace}
                        language="sql"
                        code={traceText(run)}
                        onCopy={() => track('command_copy', `trace:${id}`)}
                      />
                      <figcaption>{format(m.landingUi.scenarioSource, { seed: run.seed, source: run.source })}</figcaption>
                    </figure>
                  </div>
                ),
              };
            })}
          />
          <p className={styles.more}>
            <a href="/scenarios">
              {m.cta.allScenarios} <ArrowRight size={14} aria-hidden="true" />
            </a>
          </p>
        </Reveal>
      </Section>

      {/* 5. Delta debugging: from a noisy failure to the transactions that matter. */}
      <Section width="wide" labelledBy="shrink-title" className={styles.band}>
        <Reveal>
          <SectionHeader
            id="shrink-title"
            title={format(m.shrink.title, { original: banking.shrink.originalOps, minimal: banking.shrink.minimalOps })}
            lead={m.shrink.lead}
          />
          <div className={styles.shrink}>
            <ShrinkViz run={banking} lang={lang} />
            <dl className={styles.stats}>
              <div>
                <dt>{m.shrink.statReduction}</dt>
                <dd>{Math.round(banking.shrink.reductionPct)}%</dd>
              </div>
              <div>
                <dt>{m.shrink.statTrials}</dt>
                <dd>{banking.shrink.trials}</dd>
              </div>
              <div>
                <dt>{m.shrink.statTime}</dt>
                <dd>&lt; 1 s</dd>
              </div>
            </dl>
          </div>
          <p className={styles.note}>
            {m.shrink.note}{' '}
            <a href="https://github.com/bregaldahq/chaossql/blob/main/evals/01_shrinking_ratio.md" target="_blank" rel="noreferrer">
              {m.cta.howWeMeasure}
            </a>
          </p>
          <figure className={styles.report}>
            <img src={reportDdmin} alt={m.landingUi.reportCaption} width={1800} height={788} loading="lazy" decoding="async" />
            <figcaption>{m.landingUi.reportCaption}</figcaption>
          </figure>
        </Reveal>
      </Section>

      {/* 6. CI: the real workflow, and what it does and does not do. */}
      <Section width="wide" labelledBy="ci-title">
        <Reveal className={styles.ci}>
          <div className={styles.ciCopy}>
            <h2 id="ci-title" className={styles.h2}>
              {m.ci.title}
            </h2>
            <p className={styles.lead}>{m.ci.lead}</p>
            <p className={styles.caveat}>{m.ci.gate}</p>
          </div>
          <Terminal
            lang={lang}
            title={m.landingUi.ciFile}
            language="yaml"
            code={ACTION_YAML}
            onCopy={() => track('command_copy', 'ci_yaml')}
          />
        </Reveal>
      </Section>

      {/* 7. Three ways in, with the audit as the one to act on before a launch. */}
      <Section width="wide" id="plans" labelledBy="plans-title" className={styles.band}>
        <Reveal>
          <SectionHeader id="plans-title" title={m.plans.title} />
          <div className={styles.plans}>
            <div className={styles.plan}>
              <h3>{m.plans.oss.name}</h3>
              <p className={styles.price}>{m.plans.oss.price}</p>
              <p>{m.plans.oss.body}</p>
              <InstallButton lang={lang} placement="plans" />
            </div>
            <div className={styles.plan}>
              <h3>{m.plans.cloud.name}</h3>
              <p className={styles.price}>{format(m.plans.cloud.price, { price: CLOUD_FROM_PRICE })}</p>
              <p>{m.plans.cloud.body}</p>
              <Button href="/#waitlist" variant="secondary" size="lg" onClick={() => track('cta_click', 'plans_waitlist')}>
                {m.cta.waitlist}
              </Button>
            </div>
            <div className={[styles.plan, styles.planAudit].join(' ')}>
              <h3>{m.plans.audit.name}</h3>
              <p className={styles.price}>{format(m.plans.audit.price, { price: AUDIT_PRICE })}</p>
              <p>{m.plans.audit.body}</p>
              <Button href="/pricing#audit" size="lg" onClick={() => track('cta_click', 'plans_audit')}>
                {m.cta.audit}
              </Button>
            </div>
          </div>
          <p className={styles.more}>
            <a href="/pricing">
              {m.plans.compare} <ArrowRight size={14} aria-hidden="true" />
            </a>
          </p>
        </Reveal>
      </Section>

      {/* 8. Objections, then the last call to action. */}
      <Section width="wide" labelledBy="faq-title">
        <Reveal className={styles.faqLayout}>
          <h2 id="faq-title" className={styles.h2}>
            {m.faq.title}
          </h2>
          <div className={styles.faq}>
            {Object.entries(m.faq.items).map(([id, item]) => (
              <details key={id} className={styles.faqItem}>
                <summary>{item.q}</summary>
                <p>{item.a}</p>
              </details>
            ))}
          </div>
        </Reveal>
      </Section>

      <Section width="wide" id="waitlist" labelledBy="final-title" className={styles.finalSection}>
        <Reveal className={styles.final}>
          <h2 id="final-title" className={styles.finalTitle}>
            {m.final.title}
          </h2>
          <p className={styles.lead}>{m.final.lead}</p>
          <Terminal lang={lang} title="shell" prompt code={INSTALL_CMD} onCopy={() => track('install_copy', 'final')} />
          <div className={styles.waitlist}>
            <h3>{m.landingUi.waitlistTitle}</h3>
            <p>{m.landingUi.waitlistLead}</p>
            <WaitlistForm lang={lang} source="landing_page" />
          </div>
          <p className={styles.auditLink}>
            <a href="/pricing#audit" onClick={() => track('cta_click', 'final_audit')}>
              {m.landingUi.auditCta} <ArrowRight size={14} aria-hidden="true" />
            </a>
          </p>
        </Reveal>
      </Section>
    </div>
  );
}
