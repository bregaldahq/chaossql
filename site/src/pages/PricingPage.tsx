import React, { useState } from 'react';
import {
  Check,
  ShieldAlert,
  GitPullRequest,
  Award,
  X
} from 'lucide-react';
import styles from './PricingPage.module.css';
import { submitLead } from '../lib/lead-request';
import { track } from '../lib/analytics';

interface PricingPageProps {
  lang: 'pt' | 'en';
}

export function PricingPage({ lang }: PricingPageProps) {
  const [annual, setAnnual] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedPlan, setSelectedPlan] = useState<string>('team');
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    company: '',
    database: 'PostgreSQL',
    notes: '',
  });
  const [submitted, setSubmitted] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const t = {
    badge: lang === 'pt' ? 'Precificação Baseada em Repositórios' : 'Repository-Based Pricing',
    title: lang === 'pt' ? 'Invista em integridade. Não pague por assento.' : 'Invest in integrity. Zero per-seat fees.',
    subtitle: lang === 'pt'
      ? 'Acreditamos que testes de concorrência devem ser adotados por todo o seu time. O preço é fixo pelo número de repositórios que sua engenharia protege.'
      : 'We believe concurrency testing belongs across your entire engineering organization. Fixed billing based on protected repositories, not developer seats.',
    monthly: lang === 'pt' ? 'Mensal' : 'Monthly',
    annually: lang === 'pt' ? 'Anual' : 'Annually',
    discountBadge: lang === 'pt' ? 'Economize 20%' : 'Save 20%',
    popularBadge: lang === 'pt' ? 'Mais Escolhido' : 'Most Popular',
    perMonth: lang === 'pt' ? '/mês' : '/mo',
    billedAnnually: lang === 'pt' ? 'faturado anualmente' : 'billed annually',
    billedMonthly: lang === 'pt' ? 'faturado mensalmente' : 'billed monthly',

    // Plans
    ossTitle: 'Open Source',
    ossPrice: '$0',
    ossDesc: lang === 'pt' ? 'Para projetos públicos, estudantes e bibliotecas abertas.' : 'For public open-source libraries, tooling and individual hackers.',
    ossCta: lang === 'pt' ? 'Usar CLI Gratuito' : 'Use Free CLI',

    devTitle: 'Cloud Developer',
    devPrice: '$0',
    devDesc: lang === 'pt' ? 'Para validar o ChaosSQL em 1 repositório privado.' : 'To evaluate ChaosSQL Cloud on 1 private repository.',
    devCta: lang === 'pt' ? 'Solicitar Acesso' : 'Request Access',

    teamTitle: 'Cloud Team',
    teamPriceMonthly: '$39',
    teamPriceAnnual: '$31',
    teamDesc: lang === 'pt' ? 'O padrão para equipes de produto que colocam código em produção.' : 'The standard for product engineering teams shipping to production.',
    teamCta: lang === 'pt' ? 'Solicitar Cloud Team' : 'Request Cloud Team',

    proTitle: 'Cloud Pro',
    proPriceMonthly: '$99',
    proPriceAnnual: '$79',
    proDesc: lang === 'pt' ? 'Para arquiteturas distribuídas, múltiplos microsserviços e alta escala.' : 'For multi-service architectures, regulated domains and scale.',
    proCta: lang === 'pt' ? 'Solicitar Cloud Pro' : 'Request Cloud Pro',

    // Audit Section
    auditTag: lang === 'pt' ? 'SERVIÇO DE ENGENHARIA' : 'ENGINEERING ENGAGEMENT',
    auditTitle: lang === 'pt' ? 'ChaosSQL Concurrency Safety Audit' : 'ChaosSQL Concurrency Safety Audit',
    auditSubtitle: lang === 'pt'
      ? 'Uma semana de análise aprofundada conduzida por engenheiros especialistas no seu banco de dados e transações críticas.'
      : 'A 1-week deep-dive investigation conducted by database concurrency experts on your most critical transactional workflows.',
    auditPrice: '$1,490',
    auditPriceNote: lang === 'pt' ? 'Taxa única por auditoria completa' : 'One-time investment per audited application',
    auditCta: lang === 'pt' ? 'Solicitar Auditoria de Concorrência' : 'Request Concurrency Audit',

    // Modal
    modalTitle: lang === 'pt' ? 'Solicitar Contato' : 'Request Contact',
    nameLabel: lang === 'pt' ? 'Seu Nome' : 'Your Name',
    emailLabel: lang === 'pt' ? 'E-mail Corporativo' : 'Work Email',
    companyLabel: lang === 'pt' ? 'Empresa' : 'Company',
    dbLabel: lang === 'pt' ? 'Banco de Dados Principal' : 'Primary Database',
    notesLabel: lang === 'pt' ? 'Detalhes ou Repositórios (Opcional)' : 'Details or Repositories (Optional)',
    submitBtn: lang === 'pt' ? 'Confirmar Solicitação' : 'Confirm Request',
    successTitle: lang === 'pt' ? 'Solicitação Registrada!' : 'Request Registered!',
    successDesc: lang === 'pt'
      ? 'Recebemos sua solicitação. Nossa equipe entrará em contato para discutir acesso e configuração.'
      : 'We received your request. Our team will contact you to discuss access and setup.',
  };

  const pt = lang === 'pt';
  const L = (ptText: string, enText: string) => (pt ? ptText : enText);

  const handleOpenModal = (planName: string) => {
    track('plan_select', planName.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, ''));
    setSelectedPlan(planName);
    setSubmitted(false);
    setSubmitError(null);
    setModalOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setSubmitError(null);
    try {
      const payload = {
        ...formData,
        plan: selectedPlan,
        billingCycle: annual ? 'annual' : 'monthly',
        wantAudit: selectedPlan.includes('Audit'),
        source: 'pricing_page',
        timestamp: new Date().toISOString(),
      };
      await submitLead(payload);
      track('lead_submit', `pricing:${selectedPlan.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '')}`);
      setSubmitted(true);
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : 'Request delivery failed. Please try again.');
    } finally { setSubmitting(false); }
  };

  return (
    <div className={styles.pricingPage} data-surface="light">
      <div className={styles.heroArea}>
        <div className={styles.badge}>{t.badge}</div>
        <h1 className={styles.mainTitle}>{t.title}</h1>
        <p className={styles.mainSubtitle}>{t.subtitle}</p>

        {/* Billing Switch */}
        <div className={styles.switchWrapper}>
          <span className={`${styles.switchLabel} ${!annual ? styles.switchLabelActive : ''}`}>
            {t.monthly}
          </span>
          <button
            type="button"
            className={`${styles.switchToggle} ${annual ? styles.switchToggleAnnual : ''}`}
            onClick={() => {
              track('pricing_toggle', annual ? 'monthly' : 'annual');
              setAnnual(!annual);
            }}
            aria-label={L('Alternar ciclo de faturamento', 'Toggle billing cycle')}
            aria-pressed={annual}
          >
            <span className={styles.toggleKnob} />
          </button>
          <span className={`${styles.switchLabel} ${annual ? styles.switchLabelActive : ''}`}>
            {t.annually}
          </span>
          <span className={styles.discountPill}>{t.discountBadge}</span>
        </div>
      </div>

      {/* Pricing Cards Grid */}
      <div className={styles.cardsGrid}>
        {/* OSS */}
        <div className={styles.planCard}>
          <div className={styles.planHeader}>
            <div className={styles.planName}>{t.ossTitle}</div>
            <p className={styles.planDesc}>{t.ossDesc}</p>
          </div>
          <div className={styles.priceRow}>
            <span className={styles.priceNum}>{t.ossPrice}</span>
            <span className={styles.pricePeriod}>{L('/para sempre', '/forever')}</span>
          </div>
          <ul className={styles.featureList}>
            <li><Check size={16} className={styles.checkIcon} /> {L('Repositórios Open Source ilimitados', 'Unlimited open-source repositories')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Motor determinístico Go completo', 'Full deterministic Go engine')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Algoritmo ddmin de redução causal', 'Causal ddmin trace reduction')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Exportação HTML, JUnit, OTLP e SARIF', 'HTML, JUnit, OTLP and SARIF export')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Licença permissiva MIT', 'Permissive MIT license')}</li>
          </ul>
          <a
            href="https://github.com/bregaldahq/chaossql"
            target="_blank"
            rel="noreferrer"
            className={styles.ctaSecondary}
            onClick={() => track('outbound_click', 'github_pricing_oss')}
          >
            {t.ossCta} →
          </a>
        </div>

        {/* Developer */}
        <div className={styles.planCard}>
          <div className={styles.planHeader}>
            <div className={styles.planName}>{t.devTitle}</div>
            <p className={styles.planDesc}>{t.devDesc}</p>
          </div>
          <div className={styles.priceRow}>
            <span className={styles.priceNum}>{t.devPrice}</span>
            <span className={styles.pricePeriod}>{t.perMonth}</span>
          </div>
          <ul className={styles.featureList}>
            <li><Check size={16} className={styles.checkIcon} /> <strong>{L('1 repositório privado', '1 private repository')}</strong></li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Histórico de 7 dias na nuvem', '7-day cloud history')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Verificação básica em PRs do GitHub', 'Basic GitHub PR checks')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Sanitização estrita de credenciais', 'Strict credential sanitization')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Usuários/devs ilimitados', 'Unlimited users')}</li>
          </ul>
          <button className={styles.ctaSecondary} onClick={() => handleOpenModal('Cloud Developer')}>
            {t.devCta}
          </button>
        </div>

        {/* Team (Popular) */}
        <div className={`${styles.planCard} ${styles.popularCard}`}>
          <div className={styles.popularBadge}>{t.popularBadge}</div>
          <div className={styles.planHeader}>
            <div className={styles.planName}>{t.teamTitle}</div>
            <p className={styles.planDesc}>{t.teamDesc}</p>
          </div>
          <div className={styles.priceRow}>
            <span className={styles.priceNum}>
              {annual ? t.teamPriceAnnual : t.teamPriceMonthly}
            </span>
            <span className={styles.pricePeriod}>{t.perMonth}</span>
          </div>
          <div className={styles.billingSub}>
            {annual ? t.billedAnnually : t.billedMonthly}
          </div>
          <ul className={styles.featureList}>
            <li><Check size={16} className={styles.checkIconPopular} /> <strong>{L('10 repositórios privados', '10 private repositories')}</strong></li>
            <li><Check size={16} className={styles.checkIconPopular} /> <strong>{L('Detecção de Regressões Concorrentes', 'Concurrency regression detection')}</strong></li>
            <li><Check size={16} className={styles.checkIconPopular} /> {L('Bot de Comentários no GitHub PR', 'GitHub PR comment bot')}</li>
            <li><Check size={16} className={styles.checkIconPopular} /> Concurrency Health Dashboard</li>
            <li><Check size={16} className={styles.checkIconPopular} /> {L('Histórico de 90 dias com baseline', '90-day history with baseline')}</li>
            <li><Check size={16} className={styles.checkIconPopular} /> {L('Reprodutores Go para download', 'Downloadable Go reproducers')}</li>
          </ul>
          <button className={styles.ctaPrimary} onClick={() => handleOpenModal('Cloud Team')}>
            {t.teamCta}
          </button>
        </div>

        {/* Pro */}
        <div className={styles.planCard}>
          <div className={styles.planHeader}>
            <div className={styles.planName}>{t.proTitle}</div>
            <p className={styles.planDesc}>{t.proDesc}</p>
          </div>
          <div className={styles.priceRow}>
            <span className={styles.priceNum}>
              {annual ? t.proPriceAnnual : t.proPriceMonthly}
            </span>
            <span className={styles.pricePeriod}>{t.perMonth}</span>
          </div>
          <div className={styles.billingSub}>
            {annual ? t.billedAnnually : t.billedMonthly}
          </div>
          <ul className={styles.featureList}>
            <li><Check size={16} className={styles.checkIcon} /> <strong>{L('30 repositórios privados', '30 private repositories')}</strong></li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Histórico de 365 dias (1 ano)', '365-day history')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Fuzzing noturno contínuo agendado', 'Scheduled nightly fuzzing')}</li>
            <li><Check size={16} className={styles.checkIcon} /> Webhooks (Slack, Discord, PagerDuty)</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Políticas de isolamento avançadas', 'Advanced isolation policies')}</li>
            <li><Check size={16} className={styles.checkIcon} /> {L('Suporte prioritário via Slack privado', 'Priority support in a private Slack channel')}</li>
          </ul>
          <button className={styles.ctaSecondary} onClick={() => handleOpenModal('Cloud Pro')}>
            {t.proCta}
          </button>
        </div>
      </div>

      {/* Dedicated Audit Section (Sections 22 & 23 of 90-Days Plan) */}
      <div className={styles.auditSection}>
        <div className={styles.auditInner}>
          <div className={styles.auditLeft}>
            <div className={styles.auditTag}>{t.auditTag}</div>
            <h2 className={styles.auditTitle}>{t.auditTitle}</h2>
            <p className={styles.auditSubtitle}>{t.auditSubtitle}</p>

            <div className={styles.auditDeliverables}>
              <div className={styles.deliverableItem}>
                <ShieldAlert size={20} color="#ef4444" />
                <div>
                  <strong>{L('Diagnóstico Hermitage Completo', 'Full Hermitage Diagnosis')}</strong>
                  <p>
                    {L(
                      'Mapeamento de anomalias (P4 Lost Update, A5B Write Skew, deadlocks) nos seus fluxos de banco reais.',
                      'Anomaly mapping (P4 lost update, A5B write skew, deadlocks) across your real database flows.'
                    )}
                  </p>
                </div>
              </div>
              <div className={styles.deliverableItem}>
                <GitPullRequest size={20} color="#4b2e83" />
                <div>
                  <strong>{L('Testes Reprodutores em Go Entregues', 'Go Reproduction Tests Delivered')}</strong>
                  <p>
                    {L(
                      'Testes reprodutores mínimos, reduzidos por delta-debugging, prontos para rodar no seu CI/CD.',
                      'Minimal, delta-debugged reproduction tests ready to run in your CI/CD.'
                    )}
                  </p>
                </div>
              </div>
              <div className={styles.deliverableItem}>
                <Award size={20} color="#22c55e" />
                <div>
                  <strong>{L('Relatório Executivo com Mitigações', 'Executive Report with Mitigations')}</strong>
                  <p>
                    {L(
                      'Correções SQL exatas (SELECT FOR UPDATE, locks por linha, versionamento otimista).',
                      'Exact SQL fixes (SELECT FOR UPDATE, row-level locking, optimistic versioning).'
                    )}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div className={styles.auditCard}>
            <div className={styles.auditCardHeader}>
              <div className={styles.auditDuration}>{L('Duração: 1 semana útil', 'Duration: 1 business week')}</div>
              <div className={styles.auditPriceRow}>
                <span className={styles.auditPriceNum}>{t.auditPrice}</span>
              </div>
              <div className={styles.auditPriceNote}>{t.auditPriceNote}</div>
            </div>

            <ul className={styles.auditList}>
              <li>✓ {L('Kickoff de 45 min com o arquiteto do time', '45-min kickoff with your architect')}</li>
              <li>✓ {L('Mapeamento das 5 transações mais críticas', 'Mapping of your 5 most critical transactions')}</li>
              <li>✓ {L('100.000+ escalonamentos de concorrência', '100,000+ concurrency schedules explored')}</li>
              <li>✓ {L('Relatório final entregue e apresentado ao time', 'Final report delivered and presented to your team')}</li>
              <li>✓ {L('3 meses de ChaosSQL Cloud Team inclusos', '3 months of ChaosSQL Cloud Team included')}</li>
            </ul>

            <button
              className={styles.auditCtaBtn}
              onClick={() => handleOpenModal('Concurrency Safety Audit ($1,490)')}
            >
              {t.auditCta} →
            </button>
          </div>
        </div>
      </div>

      {/* Contact request modal */}
      {modalOpen && (
        <div className={styles.modalBackdrop} onClick={() => setModalOpen(false)}>
          <div className={styles.modalCard} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div>
                <span className={styles.modalSelectedPlan}>{selectedPlan}</span>
                <h3 className={styles.modalTitle}>{t.modalTitle}</h3>
              </div>
              <button type="button" className={styles.modalClose} onClick={() => setModalOpen(false)} aria-label={L('Fechar', 'Close')}>
                <X size={20} />
              </button>
            </div>

            {submitted ? (
              <div className={styles.successState}>
                <div className={styles.successIcon}>✓</div>
                <h4>{t.successTitle}</h4>
                <p>{t.successDesc}</p>
                <button type="button" className={styles.closeBtnFinal} onClick={() => setModalOpen(false)}>
                  {L('Fechar', 'Close')}
                </button>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className={styles.formContent}>
                <p>{lang === 'pt' ? 'Envie uma solicitação de contato. Nenhuma assinatura ou cobrança é criada aqui.' : 'Send a contact request. This form does not create a subscription or charge.'}</p>
                {submitError && <p role="alert">{submitError}</p>}
                <div className={styles.inputGroup}>
                  <label htmlFor="pricing-name">{t.nameLabel}</label>
                  <input
                    id="pricing-name"
                    type="text"
                    required
                    placeholder={L('Seu nome', 'Your name')}
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  />
                </div>

                <div className={styles.inputGroup}>
                  <label htmlFor="pricing-email">{t.emailLabel}</label>
                  <input
                    id="pricing-email"
                    type="email"
                    required
                    placeholder={L('voce@empresa.com', 'you@company.com')}
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  />
                </div>

                <div className={styles.formRow}>
                  <div className={styles.inputGroup}>
                    <label htmlFor="pricing-company">{t.companyLabel}</label>
                    <input
                      id="pricing-company"
                      type="text"
                      required
                      placeholder={L('Nome da empresa', 'Company name')}
                      value={formData.company}
                      onChange={(e) => setFormData({ ...formData, company: e.target.value })}
                    />
                  </div>
                  <div className={styles.inputGroup}>
                    <label htmlFor="pricing-database">{t.dbLabel}</label>
                    <select
                      id="pricing-database"
                      value={formData.database}
                      onChange={(e) => setFormData({ ...formData, database: e.target.value })}
                    >
                      <option value="PostgreSQL">PostgreSQL</option>
                      <option value="MySQL">MySQL</option>
                      <option value="SQLite">SQLite (WAL)</option>
                      <option value="CockroachDB">CockroachDB</option>
                      <option value="Oracle/SQLServer">SQL Server / Oracle</option>
                    </select>
                  </div>
                </div>

                <div className={styles.inputGroup}>
                  <label htmlFor="pricing-notes">{t.notesLabel}</label>
                  <textarea
                    id="pricing-notes"
                    rows={3}
                    placeholder={L('Volume de transações, repositórios ou prazos de release...', 'Transaction volume, repositories or release deadlines...')}
                    value={formData.notes}
                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  />
                </div>

                <button type="submit" className={styles.modalSubmitBtn} disabled={submitting}>
                  {submitting ? (lang === 'pt' ? 'Enviando…' : 'Sending…') : t.submitBtn} →
                </button>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
