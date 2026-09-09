import React, { useState } from 'react';
import {
  Check,
  ShieldAlert,
  GitPullRequest,
  Award,
  X
} from 'lucide-react';
import styles from './PricingPage.module.css';

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
    devCta: lang === 'pt' ? 'Começar Grátis' : 'Start Free',

    teamTitle: 'Cloud Team',
    teamPriceMonthly: '$39',
    teamPriceAnnual: '$31',
    teamDesc: lang === 'pt' ? 'O padrão para equipes de produto que colocam código em produção.' : 'The standard for product engineering teams shipping to production.',
    teamCta: lang === 'pt' ? 'Contratar Cloud Team' : 'Get Cloud Team',

    proTitle: 'Cloud Pro',
    proPriceMonthly: '$99',
    proPriceAnnual: '$79',
    proDesc: lang === 'pt' ? 'Para arquiteturas distribuídas, múltiplos microsserviços e alta escala.' : 'For multi-service architectures, regulated domains and scale.',
    proCta: lang === 'pt' ? 'Contratar Cloud Pro' : 'Get Cloud Pro',

    // Audit Section
    auditTag: lang === 'pt' ? 'SERVIÇO DE ENGENHARIA VIP' : 'VIP ENGINEERING ENGAGEMENT',
    auditTitle: lang === 'pt' ? 'ChaosSQL Concurrency Safety Audit' : 'ChaosSQL Concurrency Safety Audit',
    auditSubtitle: lang === 'pt'
      ? 'Uma semana de análise aprofundada conduzida por engenheiros especialistas no seu banco de dados e transações críticas.'
      : 'A 1-week deep-dive investigation conducted by database concurrency experts on your most critical transactional workflows.',
    auditPrice: '$1,490',
    auditPriceNote: lang === 'pt' ? 'Taxa única por auditoria completa' : 'One-time investment per audited application',
    auditCta: lang === 'pt' ? 'Agendar Auditoria de Concorrência' : 'Book Concurrency Audit',

    // Modal
    modalTitle: lang === 'pt' ? 'Ativação & Contratação' : 'Activation & Checkout',
    nameLabel: lang === 'pt' ? 'Seu Nome' : 'Your Name',
    emailLabel: lang === 'pt' ? 'E-mail Corporativo' : 'Work Email',
    companyLabel: lang === 'pt' ? 'Empresa' : 'Company',
    dbLabel: lang === 'pt' ? 'Banco de Dados Principal' : 'Primary Database',
    notesLabel: lang === 'pt' ? 'Detalhes ou Repositórios (Opcional)' : 'Details or Repositories (Optional)',
    submitBtn: lang === 'pt' ? 'Confirmar Solicitação' : 'Confirm Request',
    successTitle: lang === 'pt' ? 'Solicitação Registrada!' : 'Request Registered!',
    successDesc: lang === 'pt'
      ? 'Nossa equipe de engenharia entrará em contato em até 4 horas úteis com o seu token de acesso e instruções de setup.'
      : 'Our engineering team will reach out within 4 business hours with your access token and setup onboarding.',
  };

  const handleOpenModal = (planName: string) => {
    setSelectedPlan(planName);
    setSubmitted(false);
    setModalOpen(true);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const payload = {
        ...formData,
        plan: selectedPlan,
        billingCycle: annual ? 'annual' : 'monthly',
        timestamp: new Date().toISOString(),
      };
      // Local storage fallback + webhook
      const existing = JSON.parse(localStorage.getItem('chaossql_leads') || '[]');
      existing.push(payload);
      localStorage.setItem('chaossql_leads', JSON.stringify(existing));

      if (typeof window !== 'undefined' && (window as any).fetch) {
        fetch('/api/waitlist', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        }).catch(() => {});
      }
    } catch (_) {}
    setSubmitted(true);
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
            onClick={() => setAnnual(!annual)}
            aria-label="Alternar ciclo de faturamento"
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
            <span className={styles.pricePeriod}>/permanente</span>
          </div>
          <ul className={styles.featureList}>
            <li><Check size={16} className={styles.checkIcon} /> Repositórios Open Source ilimitados</li>
            <li><Check size={16} className={styles.checkIcon} /> Motor determinístico Go completo</li>
            <li><Check size={16} className={styles.checkIcon} /> Algoritmo ddmin de redução causal</li>
            <li><Check size={16} className={styles.checkIcon} /> Exportação HTML, JUnit, OTLP e SARIF</li>
            <li><Check size={16} className={styles.checkIcon} /> Licença permissiva Apache 2.0</li>
          </ul>
          <a href="https://github.com/bregaldahq/chaossql" target="_blank" rel="noreferrer" className={styles.ctaSecondary}>
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
            <li><Check size={16} className={styles.checkIcon} /> <strong>1 repositório privado</strong></li>
            <li><Check size={16} className={styles.checkIcon} /> Histórico de 7 dias na nuvem</li>
            <li><Check size={16} className={styles.checkIcon} /> Verificação básica em PRs do GitHub</li>
            <li><Check size={16} className={styles.checkIcon} /> Sanitização estrita de credenciais</li>
            <li><Check size={16} className={styles.checkIcon} /> Usuários/devs ilimitados</li>
          </ul>
          <button className={styles.ctaSecondary} onClick={() => handleOpenModal('Cloud Developer ($0)')}>
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
            <li><Check size={16} className={styles.checkIconPopular} /> <strong>10 repositórios privados</strong></li>
            <li><Check size={16} className={styles.checkIconPopular} /> <strong>Detecção de Regressões Concorrentes</strong></li>
            <li><Check size={16} className={styles.checkIconPopular} /> Bot de Comentários no GitHub PR</li>
            <li><Check size={16} className={styles.checkIconPopular} /> Concurrency Health Dashboard</li>
            <li><Check size={16} className={styles.checkIconPopular} /> Histórico de 90 dias com baseline</li>
            <li><Check size={16} className={styles.checkIconPopular} /> Reprodutores Go para download</li>
          </ul>
          <button className={styles.ctaPrimary} onClick={() => handleOpenModal('Cloud Team ($39/mo)')}>
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
            <li><Check size={16} className={styles.checkIcon} /> <strong>30 repositórios privados</strong></li>
            <li><Check size={16} className={styles.checkIcon} /> Histórico de 365 dias (1 ano)</li>
            <li><Check size={16} className={styles.checkIcon} /> Fuzzing noturno contínuo agendado</li>
            <li><Check size={16} className={styles.checkIcon} /> Webhooks (Slack, Discord, PagerDuty)</li>
            <li><Check size={16} className={styles.checkIcon} /> Políticas de isolamento avançadas</li>
            <li><Check size={16} className={styles.checkIcon} /> Suporte prioritário via Slack privado</li>
          </ul>
          <button className={styles.ctaSecondary} onClick={() => handleOpenModal('Cloud Pro ($99/mo)')}>
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
                  <strong>Diagnóstico Hermitage Completo</strong>
                  <p>Mapeamento de anomalias (P4 Lost Update, A5A Write Skew, Deadlocks) nos seus fluxos de banco reais.</p>
                </div>
              </div>
              <div className={styles.deliverableItem}>
                <GitPullRequest size={20} color="#4b2e83" />
                <div>
                  <strong>Testes Reprodutores em Go Entregues</strong>
                  <p>Scripts reprodutores minimalistas delta-debugged prontos para rodar no seu CI/CD.</p>
                </div>
              </div>
              <div className={styles.deliverableItem}>
                <Award size={20} color="#22c55e" />
                <div>
                  <strong>Relatório Executivo com Mitigações Cirúrgicas</strong>
                  <p>Instruções exatas de SQL (`SELECT FOR UPDATE`, row-level locking, versionamento otimista).</p>
                </div>
              </div>
            </div>
          </div>

          <div className={styles.auditCard}>
            <div className={styles.auditCardHeader}>
              <div className={styles.auditDuration}>Duração: 1 semana útil</div>
              <div className={styles.auditPriceRow}>
                <span className={styles.auditPriceNum}>{t.auditPrice}</span>
              </div>
              <div className={styles.auditPriceNote}>{t.auditPriceNote}</div>
            </div>

            <ul className={styles.auditList}>
              <li>✓ Kickoff de 45 min com arquiteto do time</li>
              <li>✓ Mapeamento das 5 transações mais críticas</li>
              <li>✓ 100.000+ escalonamentos de concorrência</li>
              <li>✓ Relatório final entregue e apresentado ao time</li>
              <li>✓ 3 meses inclusos de ChaosSQL Cloud Team</li>
            </ul>

            <button
              className={styles.auditCtaBtn}
              onClick={() => handleOpenModal('VIP Concurrency Safety Audit ($1,490)')}
            >
              {t.auditCta} →
            </button>
          </div>
        </div>
      </div>

      {/* Checkout / Contact Modal */}
      {modalOpen && (
        <div className={styles.modalBackdrop} onClick={() => setModalOpen(false)}>
          <div className={styles.modalCard} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div>
                <span className={styles.modalSelectedPlan}>{selectedPlan}</span>
                <h3 className={styles.modalTitle}>{t.modalTitle}</h3>
              </div>
              <button className={styles.modalClose} onClick={() => setModalOpen(false)}>
                <X size={20} />
              </button>
            </div>

            {submitted ? (
              <div className={styles.successState}>
                <div className={styles.successIcon}>✓</div>
                <h4>{t.successTitle}</h4>
                <p>{t.successDesc}</p>
                <button className={styles.closeBtnFinal} onClick={() => setModalOpen(false)}>
                  Fechar
                </button>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className={styles.formContent}>
                <div className={styles.inputGroup}>
                  <label>{t.nameLabel}</label>
                  <input
                    type="text"
                    required
                    placeholder="Ricardo Bregalda"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  />
                </div>

                <div className={styles.inputGroup}>
                  <label>{t.emailLabel}</label>
                  <input
                    type="email"
                    required
                    placeholder="ricardo@empresa.com"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  />
                </div>

                <div className={styles.formRow}>
                  <div className={styles.inputGroup}>
                    <label>{t.companyLabel}</label>
                    <input
                      type="text"
                      required
                      placeholder="Acme Fintech"
                      value={formData.company}
                      onChange={(e) => setFormData({ ...formData, company: e.target.value })}
                    />
                  </div>
                  <div className={styles.inputGroup}>
                    <label>{t.dbLabel}</label>
                    <select
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
                  <label>{t.notesLabel}</label>
                  <textarea
                    rows={3}
                    placeholder="Volume de transações, repositórios ou prazos de release..."
                    value={formData.notes}
                    onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                  />
                </div>

                <button type="submit" className={styles.modalSubmitBtn}>
                  {t.submitBtn} →
                </button>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
