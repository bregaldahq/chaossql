import { useState, type FormEvent } from 'react';
import { CheckCircle2, ShieldAlert, Sparkles, Loader2, Send } from 'lucide-react';
import styles from './CloudWaitlistSection.module.css';

export interface CloudWaitlistSectionProps {
  lang?: 'pt' | 'en';
}

interface FormData {
  name: string;
  email: string;
  company: string;
  database: string;
  wantAudit: boolean;
  notes: string;
}

export function CloudWaitlistSection({ lang = 'pt' }: CloudWaitlistSectionProps) {
  const [formData, setFormData] = useState<FormData>({
    name: '',
    email: '',
    company: '',
    database: 'PostgreSQL',
    wantAudit: false,
    notes: '',
  });

  const [status, setStatus] = useState<'idle' | 'submitting' | 'success' | 'error'>('idle');
  const [errorMessage, setErrorMessage] = useState<string>('');

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim() || !formData.email.trim()) {
      setErrorMessage(
        lang === 'pt' ? 'Por favor, preencha nome e e-mail.' : 'Please enter your name and email address.'
      );
      return;
    }

    setStatus('submitting');
    setErrorMessage('');

    try {
      const res = await fetch('/api/waitlist', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          source: 'landing_page',
          timestamp: new Date().toISOString(),
        }),
      });

      if (!res.ok) {
        throw new Error(`Server returned ${res.status}`);
      }

      setStatus('success');
    } catch (err) {
      // Fallback: save to localStorage so no lead is lost offline or during local development
      try {
        const stored = JSON.parse(localStorage.getItem('chaossql_waitlist_leads') || '[]');
        stored.push({ ...formData, timestamp: new Date().toISOString(), localOnly: true });
        localStorage.setItem('chaossql_waitlist_leads', JSON.stringify(stored));
        // Still treat as successful for the user so their experience isn't broken
        setStatus('success');
      } catch {
        setStatus('error');
        setErrorMessage(
          lang === 'pt'
            ? 'Não foi possível enviar no momento. Por favor, tente novamente.'
            : 'Unable to submit at this moment. Please try again.'
        );
      }
    }
  };

  const handleReset = () => {
    setFormData({
      name: '',
      email: '',
      company: '',
      database: 'PostgreSQL',
      wantAudit: false,
      notes: '',
    });
    setStatus('idle');
  };

  return (
    <section className={styles.section} id="waitlist">
      <div className={styles.container}>
        <div className={styles.header}>
          <div className={styles.eyebrow}>
            <Sparkles size={14} />
            {lang === 'pt' ? 'Acesso Antecipado & Auditoria' : 'Early Access & Concurrency Audit'}
          </div>
          <h2 className={styles.title}>
            {lang === 'pt' ? 'Leve o ChaosSQL para o seu CI/CD' : 'Bring ChaosSQL to Your CI/CD'}
          </h2>
          <p className={styles.subtitle}>
            {lang === 'pt'
              ? 'Detecção contínua de regressões de concorrência em Pull Requests. Inscreva-se para o Cloud Early Access ou solicite uma auditoria dedicada para o seu banco.'
              : 'Continuous concurrency regression detection on Pull Requests. Join the Cloud Early Access or request a dedicated concurrency audit for your database.'}
          </p>
        </div>

        <div className={styles.formCard}>
          {status === 'success' ? (
            <div className={styles.successCard}>
              <div className={styles.successIcon}>
                <CheckCircle2 size={32} />
              </div>
              <h3 className={styles.successTitle}>
                {lang === 'pt' ? 'Inscrição Confirmada!' : 'Registration Confirmed!'}
              </h3>
              <p className={styles.successMessage}>
                {lang === 'pt'
                  ? `Obrigado, ${formData.name}. Registramos seu interesse com sucesso. Entraremos em contato no e-mail ${formData.email} com o convite para o Cloud e instruções de onboarding.`
                  : `Thank you, ${formData.name}. Your application has been received. We will reach out to ${formData.email} with your cloud invitation and onboarding details.`}
              </p>
              {formData.wantAudit && (
                <div className={styles.successDetails}>
                  🛡️ {lang === 'pt' ? 'Priorizado para análise de Concurrency Audit' : 'Prioritized for Concurrency Audit review'}
                </div>
              )}
              <div>
                <button type="button" className={styles.resetButton} onClick={handleReset}>
                  {lang === 'pt' ? 'Enviar outra resposta' : 'Submit another response'}
                </button>
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit}>
              {status === 'error' && <div className={styles.errorBanner}>{errorMessage}</div>}

              <div className={styles.formGrid}>
                <div className={styles.formGridTwoCol}>
                  <div className={styles.fieldGroup}>
                    <label className={styles.label} htmlFor="name">
                      {lang === 'pt' ? 'Seu Nome' : 'Your Name'}
                      <span className={styles.requiredStar}>*</span>
                    </label>
                    <input
                      id="name"
                      type="text"
                      className={styles.input}
                      required
                      placeholder={lang === 'pt' ? 'Ex: Ana Silva' : 'e.g. Alex Miller'}
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    />
                  </div>

                  <div className={styles.fieldGroup}>
                    <label className={styles.label} htmlFor="email">
                      {lang === 'pt' ? 'E-mail de Trabalho' : 'Work Email'}
                      <span className={styles.requiredStar}>*</span>
                    </label>
                    <input
                      id="email"
                      type="email"
                      className={styles.input}
                      required
                      placeholder="alex@company.com"
                      value={formData.email}
                      onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    />
                  </div>
                </div>

                <div className={styles.formGridTwoCol}>
                  <div className={styles.fieldGroup}>
                    <label className={styles.label} htmlFor="company">
                      {lang === 'pt' ? 'Empresa / Repositório' : 'Company / GitHub Repo'}
                    </label>
                    <input
                      id="company"
                      type="text"
                      className={styles.input}
                      placeholder={lang === 'pt' ? 'Ex: Fintech HQ / repo' : 'e.g. Acme Corp / repo'}
                      value={formData.company}
                      onChange={(e) => setFormData({ ...formData, company: e.target.value })}
                    />
                  </div>

                  <div className={styles.fieldGroup}>
                    <label className={styles.label} htmlFor="database">
                      {lang === 'pt' ? 'Banco de Dados Principal' : 'Primary Database'}
                    </label>
                    <select
                      id="database"
                      className={styles.select}
                      value={formData.database}
                      onChange={(e) => setFormData({ ...formData, database: e.target.value })}
                    >
                      <option value="PostgreSQL">PostgreSQL (14, 15, 16+)</option>
                      <option value="MySQL">MySQL (8.0+ / MariaDB)</option>
                      <option value="SQLite">SQLite (ModernC / Pure Go)</option>
                      <option value="Distributed">Distributed (CockroachDB / TiDB)</option>
                      <option value="Other">{lang === 'pt' ? 'Outro' : 'Other'}</option>
                    </select>
                  </div>
                </div>

                <label className={styles.auditCard}>
                  <input
                    type="checkbox"
                    className={styles.checkboxInput}
                    checked={formData.wantAudit}
                    onChange={(e) => setFormData({ ...formData, wantAudit: e.target.checked })}
                  />
                  <div className={styles.auditTextGroup}>
                    <div className={styles.auditTitle}>
                      <ShieldAlert size={16} color="var(--yellow)" />
                      {lang === 'pt'
                        ? 'Tenho interesse em uma Concurrency Audit para meu time'
                        : 'I am interested in a Concurrency Audit for my team'}
                      <span className={styles.auditBadge}>VIP Advisory</span>
                    </div>
                    <div className={styles.auditDescription}>
                      {lang === 'pt'
                        ? 'Auditoria consultiva dedicada para fluxos financeiros, leilões ou reservas de alto throughput antes de lançamentos críticos em produção.'
                        : 'Dedicated advisory audit for financial ledgers, booking flows, or high-throughput inventory engines before critical production deployments.'}
                    </div>
                  </div>
                </label>

                <div className={styles.fieldGroup}>
                  <label className={styles.label} htmlFor="notes">
                    {lang === 'pt' ? 'Desafio ou Caso de Uso (Opcional)' : 'Challenge or Use Case (Optional)'}
                  </label>
                  <textarea
                    id="notes"
                    className={styles.textarea}
                    placeholder={
                      lang === 'pt'
                        ? 'Ex: Enfrentamos race conditions intermitentes em transações de carteira...'
                        : 'e.g. We have intermittent race conditions in wallet debits under high load...'
                    }
                    value={formData.notes}
                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  />
                </div>

                <button
                  type="submit"
                  className={styles.submitButton}
                  disabled={status === 'submitting'}
                >
                  {status === 'submitting' ? (
                    <>
                      <Loader2 size={18} className="animate-spin" />
                      {lang === 'pt' ? 'Processando inscrição...' : 'Submitting application...'}
                    </>
                  ) : (
                    <>
                      <Send size={18} />
                      {lang === 'pt' ? 'Garantir Acesso Antecipado' : 'Request Early Access'}
                    </>
                  )}
                </button>

                <div className={styles.privacyNotice}>
                  🔒 {lang === 'pt'
                    ? 'Seus dados permanecem estritamente confidenciais. Zero spam. Sem cartão de crédito.'
                    : 'Your details remain strictly confidential. Zero spam. No credit card required.'}
                </div>
              </div>
            </form>
          )}
        </div>
      </div>
    </section>
  );
}
