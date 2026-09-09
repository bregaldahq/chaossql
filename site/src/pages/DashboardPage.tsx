import { useState } from 'react';
import {
  ShieldAlert,
  GitPullRequest,
  Activity,
  Layers,
  Download,
  Copy,
  Check,
  Search,
  RefreshCw,
  X,
  Play
} from 'lucide-react';
import styles from './DashboardPage.module.css';

interface InvariantViolation {
  name: string;
  query: string;
  assertion: string;
  actual: string;
}

interface TraceStep {
  worker: string;
  opType: string;
  table: string;
  sql: string;
}

interface RunItem {
  id: string;
  repo: string;
  branch: string;
  prNumber?: number;
  commitSHA: string;
  status: 'passed' | 'failed';
  anomalyType: string;
  anomalyName: string;
  isRegression: boolean;
  baselineStatus?: string;
  driver: string;
  isolation: string;
  scenario: string;
  schedulesCount: number;
  failedSchedules: number;
  seed: number;
  durationMS: number;
  timestamp: string;
  failingInvariant?: InvariantViolation;
  traceSteps?: TraceStep[];
  reproGoCode?: string;
}

const INITIAL_RUNS: RunItem[] = [
  {
    id: 'run_991823a',
    repo: 'acme/payments',
    branch: 'feat/wallet-withdraw',
    prNumber: 382,
    commitSHA: 'a8812ac',
    status: 'failed',
    anomalyType: 'P4',
    anomalyName: 'Lost Update',
    isRegression: true,
    baselineStatus: 'PASS',
    driver: 'PostgreSQL 16',
    isolation: 'READ COMMITTED',
    scenario: 'wallet_transfer',
    schedulesCount: 100,
    failedSchedules: 31,
    seed: 184729,
    durationMS: 340,
    timestamp: '12m atrás',
    failingInvariant: {
      name: 'total_wealth_conserved',
      query: 'SELECT sum(balance) FROM accounts;',
      assertion: 'total == 2000',
      actual: '1950',
    },
    traceSteps: [
      { worker: 'T1', opType: 'read', table: 'accounts', sql: 'SELECT balance FROM accounts WHERE id = 1;' },
      { worker: 'T2', opType: 'read', table: 'accounts', sql: 'SELECT balance FROM accounts WHERE id = 1;' },
      { worker: 'T1', opType: 'write', table: 'accounts', sql: 'UPDATE accounts SET balance = balance - 50 WHERE id = 1;' },
      { worker: 'T2', opType: 'write', table: 'accounts', sql: 'UPDATE accounts SET balance = balance - 50 WHERE id = 1;' },
    ],
    reproGoCode: `package repro_test

import (
	"context"
	"database/sql"
	"testing"
)

// Standalone reproduction synthesized by ChaosSQL
func TestReproduce_P4_LostUpdate(t *testing.T) {
	db, _ := sql.Open("postgres", "postgres://localhost:5432/test?sslmode=disable")
	defer db.Close()

	// Interleaved execution discovered under seed 184729
	// T1: SELECT -> T2: SELECT -> T1: UPDATE -> T2: UPDATE
}
`,
  },
  {
    id: 'run_881920b',
    repo: 'acme/payments',
    branch: 'main',
    commitSHA: '4f281e0',
    status: 'passed',
    anomalyType: 'NONE',
    anomalyName: 'Baseline Verificado',
    isRegression: false,
    driver: 'PostgreSQL 16',
    isolation: 'READ COMMITTED',
    scenario: 'wallet_transfer',
    schedulesCount: 100,
    failedSchedules: 0,
    seed: 42100,
    durationMS: 290,
    timestamp: '1h atrás',
  },
  {
    id: 'run_771822c',
    repo: 'acme/ledger',
    branch: 'fix/concurrent-settlement',
    prNumber: 104,
    commitSHA: 'c90181a',
    status: 'failed',
    anomalyType: 'DEADLOCK',
    anomalyName: 'Deadlock Cycle',
    isRegression: true,
    baselineStatus: 'PASS',
    driver: 'PostgreSQL 16',
    isolation: 'REPEATABLE READ',
    scenario: 'two_way_transfer',
    schedulesCount: 50,
    failedSchedules: 18,
    seed: 99182,
    durationMS: 510,
    timestamp: '3h atrás',
    failingInvariant: {
      name: 'no_deadlock_aborts',
      query: 'SELECT count(*) FROM pg_stat_activity WHERE state = \'active\';',
      assertion: 'aborted_tx == 0',
      actual: '2',
    },
    traceSteps: [
      { worker: 'T1', opType: 'write', table: 'accounts', sql: 'UPDATE accounts SET balance = balance - 100 WHERE id = 1;' },
      { worker: 'T2', opType: 'write', table: 'accounts', sql: 'UPDATE accounts SET balance = balance - 200 WHERE id = 2;' },
      { worker: 'T1', opType: 'write', table: 'accounts', sql: 'UPDATE accounts SET balance = balance + 100 WHERE id = 2;' },
      { worker: 'T2', opType: 'write', table: 'accounts', sql: 'UPDATE accounts SET balance = balance + 200 WHERE id = 1;' },
    ],
    reproGoCode: `package repro_test
// Deadlock cycle detected between T1 and T2 on cross-row locking order
`,
  },
  {
    id: 'run_661902d',
    repo: 'acme/checkout',
    branch: 'feat/idempotency-redis',
    prNumber: 77,
    commitSHA: '11e892b',
    status: 'passed',
    anomalyType: 'NONE',
    anomalyName: 'Livre de Anomalias',
    isRegression: false,
    driver: 'MySQL 8.0',
    isolation: 'READ COMMITTED',
    scenario: 'idempotent_order',
    schedulesCount: 150,
    failedSchedules: 0,
    seed: 77123,
    durationMS: 410,
    timestamp: '5h atrás',
  },
  {
    id: 'run_551934e',
    repo: 'acme/inventory',
    branch: 'feat/flash-sale-reserves',
    prNumber: 92,
    commitSHA: '33b190f',
    status: 'failed',
    anomalyType: 'A5A',
    anomalyName: 'Write Skew (A5A)',
    isRegression: true,
    baselineStatus: 'PASS',
    driver: 'PostgreSQL 16',
    isolation: 'SNAPSHOT',
    scenario: 'doctor_on_call',
    schedulesCount: 100,
    failedSchedules: 44,
    seed: 55410,
    durationMS: 380,
    timestamp: '1d atrás',
    failingInvariant: {
      name: 'at_least_one_active',
      query: 'SELECT count(*) FROM doctors WHERE on_call = true;',
      assertion: 'count >= 1',
      actual: '0',
    },
    traceSteps: [
      { worker: 'T1', opType: 'read', table: 'doctors', sql: 'SELECT count(*) FROM doctors WHERE on_call = true;' },
      { worker: 'T2', opType: 'read', table: 'doctors', sql: 'SELECT count(*) FROM doctors WHERE on_call = true;' },
      { worker: 'T1', opType: 'write', table: 'doctors', sql: 'UPDATE doctors SET on_call = false WHERE id = 1;' },
      { worker: 'T2', opType: 'write', table: 'doctors', sql: 'UPDATE doctors SET on_call = false WHERE id = 2;' },
    ],
    reproGoCode: `package repro_test
// Classic Write Skew (A5A): Both transactions read count=2, both deactivate
`,
  },
  {
    id: 'run_441890f',
    repo: 'acme/auth',
    branch: 'main',
    commitSHA: '9a012ff',
    status: 'passed',
    anomalyType: 'NONE',
    anomalyName: 'Baseline Verificado',
    isRegression: false,
    driver: 'SQLite 3.45',
    isolation: 'WAL Mode',
    scenario: 'session_rotation',
    schedulesCount: 80,
    failedSchedules: 0,
    seed: 12044,
    durationMS: 120,
    timestamp: '2d atrás',
  },
];

interface DashboardPageProps {
  lang: 'pt' | 'en';
}

export function DashboardPage({ lang }: DashboardPageProps) {
  const [runs, setRuns] = useState<RunItem[]>(INITIAL_RUNS);
  const [filterType, setFilterType] = useState<'all' | 'regressions' | 'prs' | 'passed'>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedRun, setSelectedRun] = useState<RunItem | null>(null);
  const [copiedCode, setCopiedCode] = useState(false);
  const [copiedCmd, setCopiedCmd] = useState(false);
  const [onboardingOpen, setOnboardingOpen] = useState(false);
  const [copiedWorkflow, setCopiedWorkflow] = useState(false);
  const [copiedToken, setCopiedToken] = useState(false);

  const t = {
    title: lang === 'pt' ? 'Dashboard de Concorrência' : 'Concurrency Health Dashboard',
    subtitle: lang === 'pt' 
      ? 'Observabilidade de isolamento transacional, baseline da branch principal e detecção de regressões de concorrência em tempo real.'
      : 'Transactional isolation observability, main branch baseline and real-time concurrency regression detection.',
    healthLabel: lang === 'pt' ? 'Índice de Saúde' : 'Concurrency Health',
    healthStatus: lang === 'pt' ? 'Estável & Protegido' : 'Stable & Protected',
    totalRuns: lang === 'pt' ? 'Execuções (30d)' : 'Runs (30d)',
    totalSchedules: lang === 'pt' ? 'Schedules Testados' : 'Schedules Tested',
    regressionsCaught: lang === 'pt' ? 'Regressões Detectadas' : 'Regressions Caught',
    openRegressions: lang === 'pt' ? 'Regressão em Aberto' : 'Open Regression',
    allRuns: lang === 'pt' ? 'Todos os Runs' : 'All Runs',
    regressionsOnly: lang === 'pt' ? 'Regressões Apenas 🚨' : 'Regressions Only 🚨',
    prsOnly: lang === 'pt' ? 'Pull Requests' : 'Pull Requests',
    passedOnly: lang === 'pt' ? 'Passados' : 'Passed',
    searchPlaceholder: lang === 'pt' ? 'Filtrar por repositório ou cenário...' : 'Filter by repository or scenario...',
    colRepo: lang === 'pt' ? 'Repositório' : 'Repository',
    colPR: lang === 'pt' ? 'Branch / PR' : 'Branch / PR',
    colStatus: lang === 'pt' ? 'Status' : 'Status',
    colFinding: lang === 'pt' ? 'Anomalia / Finding' : 'Finding / Anomaly',
    colEngine: lang === 'pt' ? 'Engine' : 'Engine',
    colDuration: lang === 'pt' ? 'Duração' : 'Duration',
    colTime: lang === 'pt' ? 'Momento' : 'Time',
    colAction: lang === 'pt' ? 'Ação' : 'Action',
    inspectBtn: lang === 'pt' ? 'Inspecionar' : 'Inspect',
    findingDetailTitle: lang === 'pt' ? 'Detalhe da Anomalia & Reprodução' : 'Finding Detail & Reproduction',
    invariantBoxTitle: lang === 'pt' ? 'Violação de Invariante de Negócio' : 'Business Invariant Violation',
    traceBoxTitle: lang === 'pt' ? 'Rastro Causal Mínimo (Delta-Debugging)' : 'Minimal Causal Trace (Delta-Debugging)',
    reproCodeTitle: lang === 'pt' ? 'Reprodutor Autônomo em Go' : 'Standalone Go Reproducer',
    openPlayground: lang === 'pt' ? 'Abrir no Playground WASM ↗' : 'Open in WASM Playground ↗',
    downloadRepro: lang === 'pt' ? 'Baixar repro_test.go' : 'Download repro_test.go',
    copyCmd: lang === 'pt' ? 'Copiar Comando CLI' : 'Copy CLI Command',
  };

  const filteredRuns = runs.filter((r) => {
    if (filterType === 'regressions' && !r.isRegression) return false;
    if (filterType === 'prs' && (!r.prNumber || r.prNumber <= 0)) return false;
    if (filterType === 'passed' && r.status !== 'passed') return false;

    if (searchQuery) {
      const q = searchQuery.toLowerCase();
      const matchRepo = r.repo.toLowerCase().includes(q);
      const matchScenario = r.scenario.toLowerCase().includes(q);
      const matchAnomaly = r.anomalyName.toLowerCase().includes(q);
      if (!matchRepo && !matchScenario && !matchAnomaly) return false;
    }
    return true;
  });

  const handleCopyCode = () => {
    if (!selectedRun?.reproGoCode) return;
    navigator.clipboard.writeText(selectedRun.reproGoCode);
    setCopiedCode(true);
    setTimeout(() => setCopiedCode(false), 2000);
  };

  const handleCopyCmd = () => {
    if (!selectedRun) return;
    const cmd = `chaossql run --seed ${selectedRun.seed} examples/banking_lost_update/chaos.yaml`;
    navigator.clipboard.writeText(cmd);
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 2000);
  };

  const handleCopyWorkflow = () => {
    const yaml = `name: Concurrency Verification (ChaosSQL)

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

permissions:
  contents: read
  pull-requests: write

jobs:
  concurrency-gate:
    name: Concurrency Invariant & Isolation Fuzzing
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run ChaosSQL
        uses: bregaldahq/chaossql@v1.5.0
        with:
          spec-path: 'chaos.yaml'
          cloud-token: \${{ secrets.CHAOSSQL_CLOUD_TOKEN }}
          github-token: \${{ secrets.GITHUB_TOKEN }}
          post-pr-comment: 'true'
`;
    navigator.clipboard.writeText(yaml);
    setCopiedWorkflow(true);
    setTimeout(() => setCopiedWorkflow(false), 2000);
  };

  const handleCopyToken = () => {
    navigator.clipboard.writeText("csql_live_demo_acme_8912b7fa");
    setCopiedToken(true);
    setTimeout(() => setCopiedToken(false), 2000);
  };

  const handleDownloadCode = () => {
    if (!selectedRun?.reproGoCode) return;
    const blob = new Blob([selectedRun.reproGoCode], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `repro_${selectedRun.anomalyType.toLowerCase()}_test.go`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className={styles.dashboardContainer} data-surface="light">
      <div className={styles.headerArea}>
        <div className={styles.headerLeft}>
          <div className={styles.orgBadge}>
            <span className={styles.orgDot}></span>
            <span>Organization: <strong>Acme Fintech</strong> (Team Pro)</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
            <h1 className={styles.mainTitle}>{t.title}</h1>
            <button
              className={styles.connectRepoBtn}
              onClick={() => setOnboardingOpen(true)}
            >
              + Conectar Novo Repositório
            </button>
          </div>
          <p className={styles.mainSubtitle}>{t.subtitle}</p>
        </div>

        <div className={styles.healthHeroBox}>
          <div className={styles.healthTop}>
            <span className={styles.healthIcon}>🛡️</span>
            <span className={styles.healthLabel}>{t.healthLabel}</span>
          </div>
          <div className={styles.healthScore}>98.7%</div>
          <div className={styles.healthSub}>{t.healthStatus}</div>
        </div>
      </div>

      {/* Metrics Row */}
      <div className={styles.metricsGrid}>
        <div className={styles.metricCard}>
          <div className={styles.metricLabel}>{t.totalRuns}</div>
          <div className={styles.metricValue}>184</div>
          <div className={styles.metricChange}>+24% vs mês anterior</div>
        </div>
        <div className={styles.metricCard}>
          <div className={styles.metricLabel}>{t.totalSchedules}</div>
          <div className={styles.metricValue}>284,128</div>
          <div className={styles.metricChange}>Escalonamentos determinísticos</div>
        </div>
        <div className={styles.metricCard}>
          <div className={styles.metricLabel}>{t.regressionsCaught}</div>
          <div className={styles.metricValue}>3</div>
          <div className={styles.metricChange}>2 resolvidas antes de produção</div>
        </div>
        <div className={`${styles.metricCard} ${styles.alertCard}`}>
          <div className={styles.metricLabel}>{t.openRegressions}</div>
          <div className={styles.metricValueAlert}>1</div>
          <div className={styles.metricChangeAlert}>🚨 PR #382 acme/payments</div>
        </div>
      </div>

      {/* Filter and Search Toolbar */}
      <div className={styles.toolbarArea}>
        <div className={styles.filterTabs}>
          <button
            className={`${styles.filterBtn} ${filterType === 'all' ? styles.filterBtnActive : ''}`}
            onClick={() => setFilterType('all')}
          >
            {t.allRuns} (6)
          </button>
          <button
            className={`${styles.filterBtn} ${styles.filterRegression} ${filterType === 'regressions' ? styles.filterBtnActive : ''}`}
            onClick={() => setFilterType('regressions')}
          >
            {t.regressionsOnly} (3)
          </button>
          <button
            className={`${styles.filterBtn} ${filterType === 'prs' ? styles.filterBtnActive : ''}`}
            onClick={() => setFilterType('prs')}
          >
            {t.prsOnly} (4)
          </button>
          <button
            className={`${styles.filterBtn} ${filterType === 'passed' ? styles.filterBtnActive : ''}`}
            onClick={() => setFilterType('passed')}
          >
            {t.passedOnly} (3)
          </button>
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
          <div className={styles.searchBox}>
            <Search size={16} className={styles.searchIcon} />
            <input
              type="text"
              className={styles.searchInput}
              placeholder={t.searchPlaceholder}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
          <button
            className={styles.refreshBtn}
            onClick={() => setRuns([...INITIAL_RUNS])}
            title={lang === 'pt' ? 'Atualizar execuções' : 'Refresh runs'}
          >
            <RefreshCw size={15} />
          </button>
        </div>
      </div>

      {/* Runs Table */}
      <div className={styles.tableCard}>
        <table className={styles.runsTable}>
          <thead>
            <tr>
              <th>{t.colRepo}</th>
              <th>{t.colPR}</th>
              <th>{t.colStatus}</th>
              <th>{t.colFinding}</th>
              <th>{t.colEngine}</th>
              <th>{t.colDuration}</th>
              <th>{t.colTime}</th>
              <th style={{ textAlign: 'right' }}>{t.colAction}</th>
            </tr>
          </thead>
          <tbody>
            {filteredRuns.map((r) => (
              <tr key={r.id} className={r.isRegression ? styles.regressionRow : undefined}>
                <td>
                  <div className={styles.repoName}>
                    <strong>{r.repo}</strong>
                    <span className={styles.scenarioTag}>{r.scenario}</span>
                  </div>
                </td>
                <td>
                  <div className={styles.branchInfo}>
                    {r.prNumber ? (
                      <span className={styles.prBadge}>
                        <GitPullRequest size={12} />
                        #{r.prNumber}
                      </span>
                    ) : (
                      <span className={styles.mainBadge}>main</span>
                    )}
                    <span className={styles.branchName}>{r.branch}</span>
                  </div>
                </td>
                <td>
                  {r.status === 'passed' ? (
                    <span className={styles.badgePassed}>
                      <Check size={12} /> PASS
                    </span>
                  ) : (
                    <span className={styles.badgeFailed}>
                      <ShieldAlert size={12} /> FAIL
                    </span>
                  )}
                </td>
                <td>
                  <div className={styles.findingCell}>
                    {r.isRegression ? (
                      <span className={styles.regressionAlertBadge}>🚨 REGRESSÃO</span>
                    ) : null}
                    <span className={styles.anomalyName}>{r.anomalyName}</span>
                    {r.anomalyType !== 'NONE' && (
                      <code className={styles.anomalyCode}>{r.anomalyType}</code>
                    )}
                  </div>
                </td>
                <td>
                  <div className={styles.driverCell}>
                    <span>{r.driver}</span>
                    <span className={styles.isolationText}>{r.isolation}</span>
                  </div>
                </td>
                <td>
                  <span className={styles.monoText}>{r.durationMS}ms</span>
                </td>
                <td>
                  <span className={styles.timeText}>{r.timestamp}</span>
                </td>
                <td style={{ textAlign: 'right' }}>
                  <button
                    className={styles.inspectBtn}
                    onClick={() => setSelectedRun(r)}
                  >
                    {t.inspectBtn} →
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Finding Detail Modal */}
      {selectedRun && (
        <div className={styles.modalBackdrop} onClick={() => setSelectedRun(null)}>
          <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div className={styles.modalHeaderLeft}>
                <div className={styles.modalPre}>
                  <span>{selectedRun.repo}</span> • <span>{selectedRun.scenario}</span>
                </div>
                <h2 className={styles.modalTitle}>
                  {selectedRun.status === 'failed' ? (
                    <span style={{ color: '#ef4444' }}>
                      {selectedRun.anomalyType} — {selectedRun.anomalyName}
                    </span>
                  ) : (
                    <span style={{ color: '#22c55e' }}>
                      ✅ Execução Aprovada (Sem Anomalias)
                    </span>
                  )}
                </h2>
              </div>
              <button className={styles.closeBtn} onClick={() => setSelectedRun(null)}>
                <X size={20} />
              </button>
            </div>

            <div className={styles.modalBody}>
              {/* Finding Metadata Grid */}
              <div className={styles.findingMetaGrid}>
                <div className={styles.findingMetaItem}>
                  <label>Repositório</label>
                  <div>{selectedRun.repo}</div>
                </div>
                <div className={styles.findingMetaItem}>
                  <label>Pull Request / Branch</label>
                  <div>
                    {selectedRun.prNumber ? `#${selectedRun.prNumber} (${selectedRun.branch})` : selectedRun.branch}
                  </div>
                </div>
                <div className={styles.findingMetaItem}>
                  <label>Commit SHA</label>
                  <code className={styles.monoText}>{selectedRun.commitSHA}</code>
                </div>
                <div className={styles.findingMetaItem}>
                  <label>Baseline (Branch Principal)</label>
                  <div className={styles.baselineCompare}>
                    <span className={styles.baselinePass}>PASS (main)</span> ➔{' '}
                    <span className={styles.prFail}>FAIL (PR)</span>
                  </div>
                </div>
                <div className={styles.findingMetaItem}>
                  <label>Banco de Dados & Driver</label>
                  <div>{selectedRun.driver}</div>
                </div>
                <div className={styles.findingMetaItem}>
                  <label>Nível de Isolamento</label>
                  <div>{selectedRun.isolation}</div>
                </div>
                <div className={styles.findingMetaItem}>
                  <label>Taxa de Falha</label>
                  <div>{selectedRun.failedSchedules} / {selectedRun.schedulesCount} escalonamentos</div>
                </div>
                <div className={styles.findingMetaItem}>
                  <label>PRNG Seed Determinístico</label>
                  <code className={styles.monoText}>{selectedRun.seed}</code>
                </div>
              </div>

              {/* Invariant Violation */}
              {selectedRun.failingInvariant && (
                <div className={styles.invariantBox}>
                  <div className={styles.sectionHeading}>
                    <ShieldAlert size={16} color="#ef4444" />
                    <span>{t.invariantBoxTitle}</span>
                  </div>
                  <div className={styles.invariantContent}>
                    <div className={styles.invariantQuery}>
                      <span className={styles.invLabel}>Query:</span>
                      <code>{selectedRun.failingInvariant.query}</code>
                    </div>
                    <div className={styles.invariantRow}>
                      <div>
                        <span className={styles.invLabel}>Esperado (Assertion):</span>
                        <code className={styles.passValue}>{selectedRun.failingInvariant.assertion}</code>
                      </div>
                      <div>
                        <span className={styles.invLabel}>Obtido no Banco (Actual):</span>
                        <code className={styles.failValue}>{selectedRun.failingInvariant.actual}</code>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Minimal Trace */}
              {selectedRun.traceSteps && selectedRun.traceSteps.length > 0 && (
                <div className={styles.traceBox}>
                  <div className={styles.sectionHeading}>
                    <Activity size={16} color="#f5c400" />
                    <span>{t.traceBoxTitle}</span>
                  </div>
                  <div className={styles.traceTimeline}>
                    {selectedRun.traceSteps.map((step, idx) => (
                      <div key={idx} className={styles.traceStepItem}>
                        <span className={styles.stepBadge}>{idx + 1}</span>
                        <span className={`${styles.workerBadge} ${step.worker === 'T1' ? styles.workerT1 : styles.workerT2}`}>
                          {step.worker}
                        </span>
                        <code className={styles.stepSql}>{step.sql}</code>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Reproducer Code */}
              {selectedRun.reproGoCode && (
                <div className={styles.reproBox}>
                  <div className={styles.reproHeader}>
                    <div className={styles.sectionHeading}>
                      <Layers size={16} />
                      <span>{t.reproCodeTitle}</span>
                    </div>
                    <div className={styles.reproActions}>
                      <button className={styles.actionBtnSmall} onClick={handleCopyCode}>
                        {copiedCode ? <Check size={14} /> : <Copy size={14} />}
                        {copiedCode ? 'Copiado!' : 'Copiar Go'}
                      </button>
                      <button className={styles.actionBtnSmall} onClick={handleDownloadCode}>
                        <Download size={14} />
                        {t.downloadRepro}
                      </button>
                    </div>
                  </div>
                  <pre className={styles.reproCodePre}>
                    <code>{selectedRun.reproGoCode}</code>
                  </pre>
                </div>
              )}
            </div>

            <div className={styles.modalFooter}>
              <button className={styles.copyCmdBtn} onClick={handleCopyCmd}>
                {copiedCmd ? <Check size={14} /> : <Copy size={14} />}
                {copiedCmd ? 'Comando Copiado!' : t.copyCmd}
              </button>

              <a
                href={`#/playground?scenario=${selectedRun.scenario}`}
                className={styles.playgroundBtn}
                onClick={() => setSelectedRun(null)}
              >
                <Play size={14} fill="currentColor" />
                {t.openPlayground}
              </a>
            </div>
          </div>
        </div>
      )}

      {/* Onboarding / Connect Repository Modal */}
      {onboardingOpen && (
        <div className={styles.modalBackdrop} onClick={() => setOnboardingOpen(false)}>
          <div className={styles.onboardingCard} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div>
                <span className={styles.modalSelectedPlan}>Setup em 60 Segundos</span>
                <h3 className={styles.modalTitle}>Conectar Repositório ao ChaosSQL Cloud</h3>
              </div>
              <button className={styles.closeBtn} onClick={() => setOnboardingOpen(false)}>
                <X size={20} />
              </button>
            </div>

            <div className={styles.onboardingBody}>
              <div className={styles.onboardingStep}>
                <div className={styles.stepNum}>1</div>
                <div className={styles.stepContent}>
                  <h4>Copie seu Token de Autenticação da Organização</h4>
                  <p>Adicione este token seguro como Secret no seu repositório GitHub para autenticar runners.</p>
                  <div className={styles.tokenBox}>
                    <code>csql_live_demo_acme_8912b7fa</code>
                    <button className={styles.actionBtnSmall} onClick={handleCopyToken}>
                      {copiedToken ? <Check size={14} /> : <Copy size={14} />}
                      {copiedToken ? 'Copiado!' : 'Copiar Token'}
                    </button>
                  </div>
                </div>
              </div>

              <div className={styles.onboardingStep}>
                <div className={styles.stepNum}>2</div>
                <div className={styles.stepContent}>
                  <h4>Configure o Secret no GitHub</h4>
                  <p>
                    No seu repositório no GitHub, acesse <strong>Settings → Secrets and variables → Actions → New repository secret</strong>.
                    Defina o nome como <code className={styles.secretName}>CHAOSSQL_CLOUD_TOKEN</code>.
                  </p>
                </div>
              </div>

              <div className={styles.onboardingStep}>
                <div className={styles.stepNum}>3</div>
                <div className={styles.stepContent}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <h4>Adicione o Workflow no Repositório</h4>
                    <button className={styles.actionBtnSmall} onClick={handleCopyWorkflow}>
                      {copiedWorkflow ? <Check size={14} /> : <Copy size={14} />}
                      {copiedWorkflow ? 'YAML Copiado!' : 'Copiar Workflow YAML'}
                    </button>
                  </div>
                  <p>Crie o arquivo <code>.github/workflows/concurrency.yml</code> com o conteúdo abaixo:</p>
                  <pre className={styles.yamlPre}>
                    <code>{`name: Concurrency Verification (ChaosSQL)

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

permissions:
  contents: read
  pull-requests: write

jobs:
  concurrency-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: bregaldahq/chaossql@v1.5.0
        with:
          spec-path: 'chaos.yaml'
          cloud-token: \${{ secrets.CHAOSSQL_CLOUD_TOKEN }}
          github-token: \${{ secrets.GITHUB_TOKEN }}
          post-pr-comment: 'true'`}</code>
                  </pre>
                </div>
              </div>

              <div className={styles.onboardingStep}>
                <div className={styles.stepNum}>4</div>
                <div className={styles.stepContent}>
                  <h4>Abra um Pull Request de Teste</h4>
                  <p>
                    O ChaosSQL executará os testes de concorrência em paralelo, comentará o resultado no PR e sincronizará o histórico aqui no Dashboard automaticamente!
                  </p>
                </div>
              </div>
            </div>

            <div className={styles.modalFooter}>
              <button className={styles.copyCmdBtn} onClick={() => setOnboardingOpen(false)}>
                Concluir & Voltar ao Dashboard
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}
