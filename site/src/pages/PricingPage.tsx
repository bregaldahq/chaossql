import { useState } from 'react';
import { ArrowRight, Check } from 'lucide-react';
import { track } from '../lib/analytics';
import { scrollToHash } from '../lib/router';
import { messages, type Language } from '../i18n';
import { Badge } from '../components/system/Badge';
import { Button } from '../components/system/Button';
import { Section, SectionHeader } from '../components/system/Section';
import { InstallButton } from '../components/landing/InstallButton';
import { AuditForm } from '../components/forms/AuditForm';
import { WaitlistForm } from '../components/forms/WaitlistForm';
import styles from './PricingPage.module.css';

interface PricingPageProps {
  lang: Language;
}

type Cycle = 'monthly' | 'annual';

// Prices mirror internal/server/billing.go (see the chaossql-plans-retention skill).
const CLOUD_PLANS = [
  { id: 'developer', monthly: 0, annual: 0 },
  { id: 'team', monthly: 39, annual: 31 },
  { id: 'pro', monthly: 99, annual: 79 },
] as const;

type CloudPlanId = (typeof CLOUD_PLANS)[number]['id'];

export function PricingPage({ lang }: PricingPageProps) {
  const m = messages[lang];
  const t = m.pricing;
  const [cycle, setCycle] = useState<Cycle>('annual');
  const [plan, setPlan] = useState('');

  const planName = (id: CloudPlanId) => `Cloud ${t[id].name}`;
  const waitlistPlans = [
    ...CLOUD_PLANS.map((p) => ({ id: p.id, label: planName(p.id) })),
    { id: 'enterprise', label: 'Enterprise' },
  ];

  const chooseCycle = (next: Cycle) => {
    if (next === cycle) return;
    setCycle(next);
    track('pricing_toggle', next);
  };

  const joinWaitlist = (id: string) => {
    setPlan(id);
    track('plan_select', id);
    history.replaceState(null, '', '#waitlist');
    scrollToHash('#waitlist');
  };

  return (
    <div className={styles.page}>
      <Section width="wide" labelledBy="pricing-title" className={styles.hero}>
        <h1 id="pricing-title" className={styles.title}>
          {t.title}
        </h1>
        <p className={styles.lead}>{t.lead}</p>
      </Section>

      {/* Open source: the base everyone starts from. */}
      <Section width="wide" spacing="tight" labelledBy="oss-title">
        <div className={styles.oss}>
          <div className={styles.ossCopy}>
            <h2 id="oss-title" className={styles.planName}>
              {t.oss.name}
            </h2>
            <p className={styles.price}>
              {t.oss.price} <span>{t.oss.period}</span>
            </p>
            <p>{t.oss.body}</p>
            <InstallButton lang={lang} placement="pricing_oss" />
          </div>
          <ul className={styles.features}>
            {Object.entries(t.oss.features).map(([id, text]) => (
              <li key={id}>
                <Check size={16} aria-hidden="true" />
                {text}
              </li>
            ))}
          </ul>
        </div>
      </Section>

      {/* Cloud plans: same features, different limits. */}
      <Section width="wide" labelledBy="cloud-title">
        <div className={styles.cloudHead}>
          <SectionHeader id="cloud-title" title={t.cloudTitle} lead={t.cloudLead} />
          <div className={styles.cycle} role="group" aria-label={t.billingLabel}>
            {(['monthly', 'annual'] as const).map((c) => (
              <button key={c} type="button" aria-pressed={cycle === c} onClick={() => chooseCycle(c)}>
                {t[c]}
              </button>
            ))}
            <Badge tone="ok">{t.save}</Badge>
          </div>
        </div>

        <div className={styles.plans}>
          {CLOUD_PLANS.map((p) => {
            const copy = t[p.id];
            const price = p[cycle];
            return (
              <div key={p.id} className={styles.plan}>
                <h3 className={styles.planName}>{planName(p.id)}</h3>
                <p className={styles.planBody}>{copy.body}</p>
                <p className={styles.price}>
                  ${price} <span>{t.perMonth}</span>
                </p>
                <p className={styles.billed}>{price === 0 ? ' ' : cycle === 'annual' ? t.billedAnnually : t.billedMonthly}</p>
                <ul className={styles.limits}>
                  <li>{copy.repos}</li>
                  <li>{copy.history}</li>
                  {'support' in copy && <li>{copy.support}</li>}
                </ul>
                <Button variant={p.id === 'team' ? 'primary' : 'secondary'} onClick={() => joinWaitlist(p.id)}>
                  {m.cta.waitlist}
                </Button>
              </div>
            );
          })}
        </div>

        <div className={styles.included}>
          <h3>{t.includedTitle}</h3>
          <ul className={styles.features}>
            {Object.entries(t.included).map(([id, text]) => (
              <li key={id}>
                <Check size={16} aria-hidden="true" />
                {text}
              </li>
            ))}
          </ul>
          <p className={styles.enterprise}>
            {t.enterprise}{' '}
            <button type="button" className={styles.link} onClick={() => joinWaitlist('enterprise')}>
              {t.enterpriseCta} <ArrowRight size={14} aria-hidden="true" />
            </button>
          </p>
          <p className={styles.note}>{t.earlyAccess}</p>
        </div>
      </Section>

      {/* The audit: the offer to act on before a launch, with its form in place. */}
      <Section width="wide" id="audit" labelledBy="audit-title" className={styles.band}>
        <div className={styles.audit}>
          <div className={styles.auditCopy}>
            <Badge tone="signal">{t.audit.name}</Badge>
            <h2 id="audit-title" className={styles.auditTitle}>
              {t.audit.title}
            </h2>
            <p className={styles.lead}>{t.audit.lead}</p>
            <p className={styles.price}>
              {t.audit.price} <span>{t.audit.priceNote}</span>
            </p>
            <p className={styles.billed}>{t.audit.duration}</p>
            <ul className={styles.features}>
              {Object.entries(t.audit.deliverables).map(([id, text]) => (
                <li key={id}>
                  <Check size={16} aria-hidden="true" />
                  {text}
                </li>
              ))}
            </ul>
          </div>
          <div className={styles.formPanel}>
            <AuditForm lang={lang} />
          </div>
        </div>
      </Section>

      <Section width="wide" id="waitlist" labelledBy="waitlist-title">
        <div className={styles.waitlist}>
          <h2 id="waitlist-title" className={styles.planName}>
            {t.waitlist.title}
          </h2>
          <p className={styles.note}>{t.earlyAccess}</p>
          <WaitlistForm
            lang={lang}
            source="pricing_page"
            plans={waitlistPlans}
            plan={plan}
            onPlanChange={setPlan}
            billingCycle={cycle}
          />
        </div>
      </Section>
    </div>
  );
}
