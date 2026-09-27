import { useState, type FormEvent } from 'react';
import { CheckCircle2 } from 'lucide-react';
import { submitLead } from '../../lib/lead-request';
import { track } from '../../lib/analytics';
import { format, messages, type Language } from '../../i18n';
import { Button } from '../system/Button';
import styles from './Form.module.css';

export const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export interface WaitlistPlanOption {
  id: string;
  label: string;
}

export interface WaitlistFormProps {
  lang: Language;
  /** Where the form lives, for the lead source and analytics. */
  source: 'landing_page' | 'pricing_page';
  /** When given, the visitor can say which Cloud plan they want (pricing page). */
  plans?: readonly WaitlistPlanOption[];
  plan?: string;
  onPlanChange?: (plan: string) => void;
  billingCycle?: 'monthly' | 'annual';
}

/** Cloud waitlist: one required field. Everything else is asked at onboarding. */
export function WaitlistForm({ lang, source, plans, plan = '', onPlanChange, billingCycle }: WaitlistFormProps) {
  const t = messages[lang].landingUi;
  const p = messages[lang].pricing.waitlist;
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState<'idle' | 'sending' | 'done'>('idle');
  const [error, setError] = useState('');
  const id = `waitlist-${source}`;

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!EMAIL_PATTERN.test(email.trim())) {
      setError(t.requiredError);
      return;
    }
    setError('');
    setStatus('sending');
    try {
      await submitLead({
        email: email.trim(),
        ...(plans ? { plan: plan || 'not_sure', billingCycle } : {}),
        source,
        timestamp: new Date().toISOString(),
      });
      track('lead_submit', plans ? `waitlist:${plan || 'not_sure'}` : 'waitlist');
      setStatus('done');
    } catch {
      setStatus('idle');
      setError(t.sendError);
    }
  };

  if (status === 'done') {
    return (
      <div className={styles.done} role="status">
        <CheckCircle2 size={22} aria-hidden="true" />
        <div>
          <p className={styles.doneTitle}>{t.successTitle}</p>
          <p>{format(t.successBody, { email: email.trim() })}</p>
        </div>
      </div>
    );
  }

  return (
    <form className={[styles.form, plans ? styles.withPlan : styles.single].join(' ')} onSubmit={onSubmit} noValidate>
      {plans && (
        <div className={styles.field}>
          <label htmlFor={`${id}-plan`}>{p.plan}</label>
          <select id={`${id}-plan`} value={plan} onChange={(e) => onPlanChange?.(e.target.value)}>
            <option value="">{p.notSure}</option>
            {plans.map((option) => (
              <option key={option.id} value={option.id}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
      )}
      <div className={styles.field}>
        <label htmlFor={`${id}-email`}>{t.emailLabel}</label>
        <input
          id={`${id}-email`}
          name="email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          aria-invalid={error !== ''}
          aria-describedby={`${id}-error`}
        />
      </div>
      <Button type="submit" size="lg" disabled={status === 'sending'} className={styles.submit}>
        {status === 'sending' ? t.submitting : t.submit}
      </Button>
      <p id={`${id}-error`} className={styles.error} role="alert">
        {error}
      </p>
    </form>
  );
}
