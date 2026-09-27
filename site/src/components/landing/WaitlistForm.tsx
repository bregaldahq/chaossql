import { useState, type FormEvent } from 'react';
import { CheckCircle2 } from 'lucide-react';
import { submitLead } from '../../lib/lead-request';
import { track } from '../../lib/analytics';
import { format, messages, type Language } from '../../i18n';
import { Button } from '../system/Button';
import styles from './WaitlistForm.module.css';

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/** Cloud waitlist: the two fields the lead endpoint requires, nothing more. */
export function WaitlistForm({ lang }: { lang: Language }) {
  const t = messages[lang].landingUi;
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [status, setStatus] = useState<'idle' | 'sending' | 'done'>('idle');
  const [error, setError] = useState('');

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (!name.trim() || !EMAIL.test(email.trim())) {
      setError(t.requiredError);
      return;
    }
    setError('');
    setStatus('sending');
    try {
      await submitLead({
        name: name.trim(),
        email: email.trim(),
        source: 'landing_page',
        timestamp: new Date().toISOString(),
      });
      track('lead_submit', 'waitlist');
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
          <p>{format(t.successBody, { name: name.trim(), email: email.trim() })}</p>
        </div>
      </div>
    );
  }

  return (
    <form className={styles.form} onSubmit={onSubmit} noValidate>
      <div className={styles.field}>
        <label htmlFor="waitlist-name">{t.nameLabel}</label>
        <input
          id="waitlist-name"
          name="name"
          autoComplete="name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          aria-invalid={error !== '' && !name.trim()}
        />
      </div>
      <div className={styles.field}>
        <label htmlFor="waitlist-email">{t.emailLabel}</label>
        <input
          id="waitlist-email"
          name="email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          aria-invalid={error !== '' && !EMAIL.test(email.trim())}
        />
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
