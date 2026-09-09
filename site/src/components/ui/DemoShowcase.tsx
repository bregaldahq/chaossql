import { useState } from 'react';
import { Copy, Check, Play, AlertTriangle, GitBranch } from 'lucide-react';
import styles from './DemoShowcase.module.css';

export interface DemoShowcaseProps {
  lang?: 'pt' | 'en';
}

interface ScenarioData {
  id: 'banking' | 'hospital' | 'deadlock';
  title: { pt: string; en: string };
  badge: string;
  anomalyName: { pt: string; en: string };
  description: { pt: string; en: string };
  cliCmd: string;
  playgroundQuery: string;
  steps: Array<{ worker: 'T1' | 'T2'; sql: string }>;
  expected: { label: { pt: string; en: string }; value: string; sub: { pt: string; en: string } };
  actual: { label: { pt: string; en: string }; value: string; sub: { pt: string; en: string } };
}

const SCENARIOS: ScenarioData[] = [
  {
    id: 'banking',
    title: { pt: '🏦 Transferência Bancária', en: '🏦 Banking Transfer' },
    badge: 'P4 ANOMALY',
    anomalyName: { pt: 'Lost Update Silencioso', en: 'Silent Lost Update' },
    description: {
      pt: 'Duas transações leem o mesmo saldo (R$ 2.000) simultaneamente em READ COMMITTED. Ambas aplicam débitos válidos, mas a segunda sobrescreve o commit da primeira, destruindo R$ 50 sem aviso nos logs.',
      en: 'Two concurrent transactions read balance ($2,000) simultaneously under READ COMMITTED. Both apply valid debits, but T2 overwrites T1, silently dropping $50 from the financial balance.',
    },
    cliCmd: 'chaossql run examples/banking_lost_update/chaos.yaml',
    playgroundQuery: '#/playground',
    steps: [
      { worker: 'T1', sql: 'SELECT balance FROM accounts WHERE id = 1  -- reads 2000' },
      { worker: 'T2', sql: 'SELECT balance FROM accounts WHERE id = 1  -- reads 2000' },
      { worker: 'T1', sql: 'UPDATE accounts SET balance = 1950 WHERE id = 1; COMMIT;' },
      { worker: 'T2', sql: 'UPDATE accounts SET balance = 1900 WHERE id = 1; COMMIT; -- overwrites T1' },
    ],
    expected: {
      label: { pt: 'Saldo Correto Esperado', en: 'Expected Account Balance' },
      value: '$1,850.00',
      sub: { pt: 'Total = 2000 - 50 - 100', en: 'Total = 2000 - 50 - 100' },
    },
    actual: {
      label: { pt: 'Saldo Corrompido Detectado', en: 'Corrupted Balance Detected' },
      value: '$1,900.00',
      sub: { pt: 'Perda silenciosa de $50 (Race condition)', en: 'Silent $50 loss (Data race)' },
    },
  },
  {
    id: 'hospital',
    title: { pt: '🏥 Escala Médica', en: '🏥 Hospital Shift' },
    badge: 'A5B ANOMALY',
    anomalyName: { pt: 'Write Skew de Capacidade', en: 'Capacity Write Skew' },
    description: {
      pt: 'Dois médicos de plantão pedem licença concorrentemente sob REPEATABLE READ. Ambos consultam médicos ativos (count = 2, mínimo exigido >= 1). Ambos confirmam a saída, deixando o hospital sem nenhum médico.',
      en: 'Two on-call doctors simultaneously request leave under REPEATABLE READ. Both check active count (count = 2, min required >= 1). Both commit absence, leaving 0 doctors on duty.',
    },
    cliCmd: 'chaossql run examples/hospital_write_skew/chaos.yaml',
    playgroundQuery: '#/playground',
    steps: [
      { worker: 'T1', sql: 'SELECT COUNT(*) FROM on_call WHERE active = true  -- returns 2' },
      { worker: 'T2', sql: 'SELECT COUNT(*) FROM on_call WHERE active = true  -- returns 2' },
      { worker: 'T1', sql: 'UPDATE on_call SET active = false WHERE doctor_id = 1; COMMIT;' },
      { worker: 'T2', sql: 'UPDATE on_call SET active = false WHERE doctor_id = 2; COMMIT;' },
    ],
    expected: {
      label: { pt: 'Invariante de Plantão', en: 'Invariant Constraint' },
      value: 'count >= 1 doctor',
      sub: { pt: 'Mínimo de segurança hospitalar', en: 'Minimum staff safety requirement' },
    },
    actual: {
      label: { pt: 'Estado Crítico Detectado', en: 'Critical Violation Detected' },
      value: '0 doctors on duty',
      sub: { pt: 'Violação formal de Snapshot Isolation', en: 'Snapshot Isolation invariant broken' },
    },
  },
  {
    id: 'deadlock',
    title: { pt: '🔒 Travamento Cruzado', en: '🔒 Lock Inversion' },
    badge: 'G-DL ANOMALY',
    anomalyName: { pt: 'Ciclo de Deadlock Mútuo', en: 'Mutual Deadlock Cycle' },
    description: {
      pt: 'O Worker 1 adquire o lock no Registro A e solicita o Registro B. Ao mesmo tempo, o Worker 2 adquire B e solicita A. Ambas as conexões ficam travadas até timeout ou cancelamento abrupto.',
      en: 'Worker 1 locks Record A and requests Record B. Simultaneously, Worker 2 locks B and requests A. Both workers block indefinitely until query timeout or engine abort.',
    },
    cliCmd: 'chaossql run examples/deadlock_cycle/chaos.yaml',
    playgroundQuery: '#/playground',
    steps: [
      { worker: 'T1', sql: 'UPDATE accounts SET balance = balance - 10 WHERE id = 1;' },
      { worker: 'T2', sql: 'UPDATE accounts SET balance = balance - 20 WHERE id = 2;' },
      { worker: 'T1', sql: 'UPDATE accounts SET balance = balance + 10 WHERE id = 2; -- waits for T2' },
      { worker: 'T2', sql: 'UPDATE accounts SET balance = balance + 20 WHERE id = 1; -- waits for T1 (DEADLOCK)' },
    ],
    expected: {
      label: { pt: 'Ordem Canônica Esperada', en: 'Expected Lock Order' },
      value: 'ORDER BY id ASC',
      sub: { pt: 'Transações concluem com sucesso', en: 'Clean deterministic serialization' },
    },
    actual: {
      label: { pt: 'Deadlock Detectado', en: 'Deadlock Exception Detected' },
      value: '40P01 Deadlock Abort',
      sub: { pt: 'Ciclo de dependência causal detectado', en: 'Cyclic lock wait dependency cycle' },
    },
  },
];

export function DemoShowcase({ lang = 'pt' }: DemoShowcaseProps) {
  const [selectedId, setSelectedId] = useState<'banking' | 'hospital' | 'deadlock'>('banking');
  const [copied, setCopied] = useState(false);

  const scenario = SCENARIOS.find((s) => s.id === selectedId) ?? SCENARIOS[0];

  const handleCopy = () => {
    navigator.clipboard.writeText(scenario.cliCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section className={styles.section} id="scenarios">
      <div className={styles.header}>
        <div className={styles.eyebrow}>
          <AlertTriangle size={14} />
          {lang === 'pt' ? 'Evidência Determinística em Ação' : 'Deterministic Evidence in Action'}
        </div>
        <h2 className={styles.title}>
          {lang === 'pt' ? '3 Falhas Críticas de Concorrência' : '3 Critical Concurrency Failures'}
        </h2>
        <p className={styles.subtitle}>
          {lang === 'pt'
            ? 'Veja exatamente como race conditions que passam despercebidas em testes comuns são detectadas, isoladas e reproduzidas pelo ChaosSQL em milissegundos.'
            : 'See how race conditions that slip past normal unit tests are deterministically exposed, isolated, and synthesized by ChaosSQL in milliseconds.'}
        </p>
      </div>

      <div className={styles.tabsContainer}>
        {SCENARIOS.map((s) => (
          <button
            key={s.id}
            type="button"
            className={`${styles.tabButton} ${selectedId === s.id ? styles.tabButtonActive : ''}`}
            onClick={() => setSelectedId(s.id)}
          >
            <span>{s.title[lang]}</span>
            <span className={styles.anomalyBadge}>{s.badge}</span>
          </button>
        ))}
      </div>

      <div className={styles.card}>
        <div className={styles.cardTop}>
          <div className={styles.scenarioInfo}>
            <div className={styles.scenarioTitleRow}>
              <h3 className={styles.scenarioName}>{scenario.anomalyName[lang]}</h3>
              <span className={styles.anomalyBadge}>{scenario.badge}</span>
            </div>
            <p className={styles.scenarioDescription}>{scenario.description[lang]}</p>
          </div>
        </div>

        <div className={styles.gridComparison}>
          <div className={styles.traceBlock}>
            <div className={styles.traceTitle}>
              <GitBranch size={14} />
              {lang === 'pt' ? 'Intercalamento Causal Mínimo (ddmin sintetizado)' : 'Synthesized Causal Trace (ddmin)'}
            </div>
            <div className={styles.traceSteps}>
              {scenario.steps.map((step, idx) => (
                <div key={idx} className={styles.traceStep}>
                  <span className={`${styles.workerTag} ${step.worker === 'T1' ? styles.worker1 : styles.worker2}`}>
                    {step.worker}
                  </span>
                  <span className={styles.stepSql}>{step.sql}</span>
                </div>
              ))}
            </div>
          </div>

          <div className={styles.metricsPanel}>
            <div className={`${styles.metricBox} ${styles.metricExpected}`}>
              <div className={styles.metricLabel}>{scenario.expected.label[lang]}</div>
              <div className={`${styles.metricValue} ${styles.metricValueGreen}`}>{scenario.expected.value}</div>
              <div className={styles.metricSubtext}>{scenario.expected.sub[lang]}</div>
            </div>

            <div className={`${styles.metricBox} ${styles.metricActual}`}>
              <div className={styles.metricLabel}>{scenario.actual.label[lang]}</div>
              <div className={`${styles.metricValue} ${styles.metricValueRed}`}>{scenario.actual.value}</div>
              <div className={styles.metricSubtext}>{scenario.actual.sub[lang]}</div>
            </div>
          </div>
        </div>

        <div className={styles.cardFooter}>
          <div className={styles.cliSnippet}>
            <span className={styles.cliText}>{scenario.cliCmd}</span>
            <button
              type="button"
              className={styles.copyBtn}
              onClick={handleCopy}
              title={lang === 'pt' ? 'Copiar comando' : 'Copy command'}
            >
              {copied ? <Check size={16} color="#22c55e" /> : <Copy size={16} />}
            </button>
          </div>

          <a href={scenario.playgroundQuery} className={styles.actionBtn}>
            <Play size={15} fill="currentColor" />
            {lang === 'pt' ? 'Testar no Playground WASM' : 'Test in WASM Playground'}
          </a>
        </div>
      </div>
    </section>
  );
}
