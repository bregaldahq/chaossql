import { useState, type FormEvent } from 'react';
import { CheckCircle2 } from 'lucide-react';
import { submitLead } from '../../lib/lead-request';
import { track } from '../../lib/analytics';
import { format, messages, type Language } from '../../i18n';
import { Button } from '../system/Button';
import { EMAIL_PATTERN } from './WaitlistForm';
import styles from './Form.module.css';

export const AUDIT_PLAN = 'Concurrency Safety Audit ($1,490)';
const DATABASES = ['PostgreSQL', 'MySQL', 'SQLite', 'CockroachDB'] as const;
const TIMELINES = ['soon', 'quarter', 'exploring'] as const;

/** Qualified audit request: enough to prepare the kickoff, nothing more. */
export function AuditForm({ lang }: { lang: Language }) {
  const t = messages[lang].pricing.form;
  const errors = messages[lang].landingUi;
  const [values, setValues] = useState({
    name: '',
    email: '',
    company: '',
    database: 'PostgreSQL',
    timeline: 'soon' as (typeof TIMELINES)[number],
    notes: '',
  });
  const [status, setStatus] = useState<'idle' | 'sending' | 'done'>('idle');
  const [error, setError] = useState('');
  const set = (key: keyof typeof values) => (e: { target: { value: string } }) =>
    setValues((v) => ({ ...v, [key]: e.target.value }));

  const invalid = {
    name: !values.name.trim(),
    email: !EMAIL_PATTERN.test(values.email.trim()),
    company: !values.company.trim(),
  };

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (invalid.name || invalid.email || invalid.company) {
      setError(t.required);
      return;
    }
    setError('');
    setStatus('sending');
    try {
      await submitLead({
        name: values.name.trim(),
        email: values.email.trim(),
        company: values.company.trim(),
        database: values.database,
        timeline: messages.en.pricing.form.timelines[values.timeline],
        notes: values.notes.trim(),
        plan: AUDIT_PLAN,
        wantAudit: true,
        source: 'pricing_page',
        timestamp: new Date().toISOString(),
      });
      track('lead_submit', `audit:${values.timeline}`);
      setStatus('done');
    } catch {
      setStatus('idle');
      setError(errors.sendError);
    }
  };

  if (status === 'done') {
    return (
      <div className={styles.done} role="status">
        <CheckCircle2 size={22} aria-hidden="true" />
        <div>
          <p className={styles.doneTitle}>{t.successTitle}</p>
          <p>{format(t.successBody, { name: values.name.trim(), email: values.email.trim() })}</p>
        </div>
      </div>
    );
  }

  const showInvalid = error !== '';
  return (
    <form className={[styles.form, styles.stacked].join(' ')} onSubmit={onSubmit} noValidate aria-labelledby="audit-form-title">
      <h3 id="audit-form-title" className={styles.title}>
        {t.title}
      </h3>
      <div className={styles.pair}>
        <div className={styles.field}>
          <label htmlFor="audit-name">{t.name}</label>
          <input id="audit-name" autoComplete="name" value={values.name} onChange={set('name')} aria-invalid={showInvalid && invalid.name} />
        </div>
        <div className={styles.field}>
          <label htmlFor="audit-email">{t.email}</label>
          <input
            id="audit-email"
            type="email"
            autoComplete="email"
            value={values.email}
            onChange={set('email')}
            aria-invalid={showInvalid && invalid.email}
          />
        </div>
      </div>
      <div className={styles.pair}>
        <div className={styles.field}>
          <label htmlFor="audit-company">{t.company}</label>
          <input
            id="audit-company"
            autoComplete="organization"
            value={values.company}
            onChange={set('company')}
            aria-invalid={showInvalid && invalid.company}
          />
        </div>
        <div className={styles.field}>
          <label htmlFor="audit-database">{t.database}</label>
          <select id="audit-database" value={values.database} onChange={set('database')}>
            {DATABASES.map((db) => (
              <option key={db} value={db}>
                {db}
              </option>
            ))}
            <option value="Other">{t.other}</option>
          </select>
        </div>
      </div>
      <fieldset className={styles.choices}>
        <legend>{t.timeline}</legend>
        {TIMELINES.map((id) => (
          <label key={id} className={styles.choice}>
            <input type="radio" name="audit-timeline" value={id} checked={values.timeline === id} onChange={set('timeline')} />
            <span>{t.timelines[id]}</span>
          </label>
        ))}
      </fieldset>
      <div className={styles.field}>
        <label htmlFor="audit-notes">{t.notes}</label>
        <textarea id="audit-notes" rows={3} value={values.notes} onChange={set('notes')} aria-describedby="audit-notes-hint" maxLength={1000} />
        <p id="audit-notes-hint" className={styles.hint}>
          {t.notesHint}
        </p>
      </div>
      <Button type="submit" size="lg" disabled={status === 'sending'} className={styles.submit}>
        {status === 'sending' ? t.submitting : t.submit}
      </Button>
      <p className={styles.error} role="alert">
        {error}
      </p>
    </form>
  );
}
