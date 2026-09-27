import { useState } from 'react';
import { Database, FileCode2, MoveRight, SlidersHorizontal, Check } from 'lucide-react';
import { ArtifactCard } from './ArtifactCard';
import styles from './ChaosSqlArtifact.module.css';

const STEPS_SHRUNK = [
  { step: '01', tx: 'T1', op: 'SELECT balance (1000)', balance: '$1000' },
  { step: '02', tx: 'T2', op: 'SELECT balance (1000)', balance: '$1000' },
  { step: '03', tx: 'T1', op: 'UPDATE balance = 950', balance: '$950' },
  { step: '04', tx: 'T2', op: 'UPDATE balance = 900', balance: '$900' },
];

const STEPS_RAW = [
  { step: '01', tx: 'T1', op: 'SELECT balance (1000)', balance: '$1000' },
  { step: '02', tx: 'T3', op: 'INSERT INTO audit_log (ping)', balance: '—' },
  { step: '03', tx: 'T2', op: 'SELECT balance (1000)', balance: '$1000' },
  { step: '04', tx: 'T4', op: 'SELECT count(*) FROM accounts', balance: '1' },
  { step: '05', tx: 'T1', op: 'UPDATE balance = 950', balance: '$950' },
  { step: '06', tx: 'T2', op: 'UPDATE balance = 900', balance: '$900' },
];

export interface ChaosSqlArtifactProps {
  lang?: 'pt' | 'en';
}

export function ChaosSqlArtifact({ lang = 'en' }: ChaosSqlArtifactProps) {
  const pt = lang === 'pt';
  const [activeMode, setActiveMode] = useState<'shrunk' | 'raw'>('shrunk');

  const steps = activeMode === 'shrunk' ? STEPS_SHRUNK : STEPS_RAW;

  return (
    <ArtifactCard
      title="chaossql"
      icon={<Database size={16} />}
      tag="banking_lost_update"
      disclaimer="Synthetic illustration · architecture grounded in repository · no live database connected"
    >
      <div className={styles.chaosContainer}>
        <div className={styles.decision}>
          <div>
            <span className={styles.eyebrow}>Concurrent withdrawals · </span>
            <span className={styles.anomalyBadge}>Lost Update / P4</span>
          </div>

          <div style={{ display: 'flex', gap: '0.4rem' }}>
            <button
              type="button"
              onClick={() => setActiveMode('shrunk')}
              style={{
                background: activeMode === 'shrunk' ? 'var(--purple)' : 'transparent',
                color: 'var(--cream)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-control)',
                padding: '0.2rem 0.5rem',
                fontSize: '0.72rem',
                fontFamily: 'var(--font-jetbrains-mono), monospace',
                cursor: 'pointer',
              }}
            >
              {activeMode === 'shrunk' && <Check size={11} style={{ marginRight: 3, verticalAlign: 'middle' }} />}
              1-Minimal Shrunk
            </button>
            <button
              type="button"
              onClick={() => setActiveMode('raw')}
              style={{
                background: activeMode === 'raw' ? 'var(--purple)' : 'transparent',
                color: 'var(--cream)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-control)',
                padding: '0.2rem 0.5rem',
                fontSize: '0.72rem',
                fontFamily: 'var(--font-jetbrains-mono), monospace',
                cursor: 'pointer',
              }}
            >
              {activeMode === 'raw' && <SlidersHorizontal size={11} style={{ marginRight: 3, verticalAlign: 'middle' }} />}
              Raw Trace
            </button>
          </div>
        </div>

        <table
          className={styles.traceTable}
          aria-label={pt ? 'Tabela de intercalação de concorrência' : 'Concurrency interleaving table'}
        >
          <thead>
            <tr>
              <th scope="col">{pt ? 'Passo' : 'Step'}</th>
              <th scope="col">Tx</th>
              <th scope="col">{pt ? 'Operação' : 'Operation'}</th>
              <th scope="col">{pt ? 'Saldo' : 'Balance'}</th>
            </tr>
          </thead>
          <tbody>
            {steps.map((row, idx) => {
              const isViolated = idx === steps.length - 1;
              return (
                <tr
                  key={row.step}
                  className={isViolated ? styles.rowViolated : undefined}
                >
                  <td>{row.step}</td>
                  <td>
                    <span className={row.tx === 'T1' ? styles.txT1 : styles.txT2}>
                      {row.tx}
                    </span>
                  </td>
                  <td>{row.op}</td>
                  <td>{row.balance}</td>
                </tr>
              );
            })}
          </tbody>
        </table>

        <div className={styles.invariant}>
          <p className={styles.eyebrow}>ledger_balance_consistency</p>
          <p className={styles.violation}>Invariant Violated: A3/P4</p>
          <code>actual_balance == expected_balance</code>
          <dl className={styles.metricGrid}>
            <div className={styles.metricItem}>
              <dt>{pt ? 'Saldo Atual' : 'Actual Balance'}</dt>
              <dd style={{ color: 'var(--yellow)' }}>$900</dd>
            </div>
            <div className={styles.metricItem}>
              <dt>{pt ? 'Esperado pelo Ledger' : 'Ledger Expected'}</dt>
              <dd style={{ color: 'var(--green)' }}>$850</dd>
            </div>
          </dl>
          <p className={styles.caption}>
            {activeMode === 'shrunk'
              ? pt
                ? 'O algoritmo ddmin reduziu o rastro para as 4 operações estritamente necessárias para reproduzir o bug.'
                : 'ddmin reduced the trace to the 4 operations strictly required to reproduce the bug.'
              : pt
                ? 'Rastro bruto com ruído de transações paralelas antes da redução causal delta-debugging.'
                : 'Raw trace with noise from parallel transactions, before causal delta-debugging.'}
          </p>
        </div>

        <ol className={styles.reproduction} aria-label={pt ? 'Sequência de redução de falha' : 'Failure reduction sequence'}>
          <li>
            <span>Trace ({activeMode === 'raw' ? '20 ops' : '6 ops'})</span>
            <MoveRight size={14} aria-hidden="true" />
          </li>
          <li>
            <span>ddmin Shrink</span>
            <MoveRight size={14} aria-hidden="true" />
          </li>
          <li>
            <FileCode2 size={14} aria-hidden="true" />
            <code>repro_test.go (4 ops)</code>
          </li>
        </ol>
      </div>
    </ArtifactCard>
  );
}
