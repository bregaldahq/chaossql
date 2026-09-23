import { useState, useRef } from 'react';
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
  Play,
  ExternalLink,
  Server,
  Wifi,
  WifiOff,
  AlertTriangle,
  Bell,
  Trash2,
  Send,
  CheckCircle2
} from 'lucide-react';
import styles from './DashboardPage.module.css';
import { CloudAPI, type WebhookItem } from '../lib/cloud-api';

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
  status: string;
  anomalyType: string;
  anomalyName: string;
  isRegression: boolean | null;
  baselineStatus?: string;
  driver: string;
  isolation: string;
  scenario: string;
  schedulesCount: number | null;
  failedSchedules: number | null;
  seed: number | null;
  durationMS: number | null;
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
      query: "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';",
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

  // Credentials only live in this mounted component; never persist them.
  const [isLiveMode, setIsLiveMode] = useState(false);
  const [apiUrl, setApiUrl] = useState('');
  const [apiToken, setApiToken] = useState('');
  const [apiStatus, setApiStatus] = useState<'idle' | 'checking' | 'online' | 'offline'>('idle');
  const [liveLoading, setLiveLoading] = useState(false);
  const [liveError, setLiveError] = useState<string | null>(null);
  const [webhooksModalOpen, setWebhooksModalOpen] = useState(false);
  const [webhooks, setWebhooks] = useState<WebhookItem[]>([]);
  const [newWhTarget, setNewWhTarget] = useState<WebhookItem['target']>('discord');
  const [newWhUrl, setNewWhUrl] = useState('');
  const [testingWebhookId, setTestingWebhookId] = useState<string | null>(null);
  const [testSuccessId, setTestSuccessId] = useState<string | null>(null);
  const [webhookError, setWebhookError] = useState<string | null>(null);
  const [webhookBusy, setWebhookBusy] = useState(false);
  const [issuedToken, setIssuedToken] = useState('');
  const [tokenError, setTokenError] = useState<string | null>(null);
  const [tokenBusy, setTokenBusy] = useState(false);
  const connectionVersion = useRef(0);
  const api = () => new CloudAPI(apiUrl, apiToken);
  const errorMessage = (error: unknown) => error instanceof Error ? error.message : 'Request failed';

  const openWebhooks = async () => {
    setWebhooksModalOpen(true);
    setWebhookError(null);
    if (!isLiveMode || apiStatus !== 'online') return;
    const version = connectionVersion.current;
    setWebhookBusy(true);
    try { const hooks = await api().webhooks(); if (version === connectionVersion.current) setWebhooks(hooks); }
    catch (error) { if (version === connectionVersion.current) setWebhookError(errorMessage(error)); }
    finally { if (version === connectionVersion.current) setWebhookBusy(false); }
  };
  const handleAddWebhook = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isLiveMode || apiStatus !== 'online') return;
    const version = connectionVersion.current;
    setWebhookBusy(true); setWebhookError(null);
    try {
      const hook = await api().createWebhook(newWhTarget, newWhUrl.trim());
      if (version === connectionVersion.current) { setWebhooks(current => [...current, hook]); setNewWhUrl(''); }
    } catch (error) { if (version === connectionVersion.current) setWebhookError(errorMessage(error)); }
    finally { if (version === connectionVersion.current) setWebhookBusy(false); }
  };
  const handleDeleteWebhook = async (id: string) => {
    const version = connectionVersion.current;
    setWebhookBusy(true); setWebhookError(null);
    try { await api().deleteWebhook(id); if (version === connectionVersion.current) setWebhooks(current => current.filter(hook => hook.id !== id)); }
    catch (error) { if (version === connectionVersion.current) setWebhookError(errorMessage(error)); }
    finally { if (version === connectionVersion.current) setWebhookBusy(false); }
  };
  const handleTestWebhook = async (hook: WebhookItem) => {
    const version = connectionVersion.current;
    setTestingWebhookId(hook.id); setTestSuccessId(null); setWebhookError(null);
    try { await api().testWebhook(hook.id); if (version === connectionVersion.current) setTestSuccessId(hook.id); }
    catch (error) { if (version === connectionVersion.current) setWebhookError(errorMessage(error)); }
    finally { if (version === connectionVersion.current) setTestingWebhookId(null); }
  };
  const handleCreateToken = async () => {
    const version = connectionVersion.current;
    setTokenBusy(true); setTokenError(null);
    try { const result = await api().createToken(); if (version === connectionVersion.current) setIssuedToken(result.token); }
    catch (error) { if (version === connectionVersion.current) setTokenError(errorMessage(error)); }
    finally { if (version === connectionVersion.current) setTokenBusy(false); }
  };
  const resetConnection = () => {
    connectionVersion.current += 1;
    setRuns([]); setSelectedRun(null); setWebhooks([]); setIssuedToken('');
    setNewWhUrl(''); setWebhookError(null); setTokenError(null);
    setApiStatus('idle'); setLiveError(null); setLiveLoading(false);
    setWebhookBusy(false); setTokenBusy(false); setTestingWebhookId(null); setTestSuccessId(null);
  };

  const t = {
    title: lang === 'pt' ? 'Dashboard de Concorrência' : 'Concurrency Health Dashboard',
    subtitle: lang === 'pt'
      ? 'Observabilidade de isolamento transacional, baseline da branch principal e detecção de regressões de concorrência em tempo real.'
      : 'Transactional isolation observability, main branch baseline and real-time concurrency regression detection.',
    healthLabel: lang === 'pt' ? 'Taxa de Aprovação' : 'Pass Rate',
    healthStatus: lang === 'pt' ? 'Nas execuções exibidas' : 'Across displayed runs',
    totalRuns: lang === 'pt' ? 'Execuções recentes (até 50)' : 'Recent runs (up to 50)',
    totalSchedules: lang === 'pt' ? 'Schedules Testados' : 'Schedules Tested',
    regressionsCaught: lang === 'pt' ? 'Regressões Detectadas' : 'Regressions Caught',
    openRegressions: lang === 'pt' ? 'nas execuções exibidas' : 'in displayed runs',
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
    openVisualizer: lang === 'pt' ? 'Abrir no Visualizador de Intercalamento ➔' : 'Open in Interleaving Visualizer ➔',
    downloadRepro: lang === 'pt' ? 'Baixar repro_test.go' : 'Download repro_test.go',
    copyCmd: lang === 'pt' ? 'Copiar Comando CLI' : 'Copy CLI Command',
    liveMode: lang === 'pt' ? 'Live Cloud' : 'Live Cloud',
    demoMode: lang === 'pt' ? 'Modo Demo' : 'Demo Mode',
    connected: lang === 'pt' ? 'Conectado à API' : 'Connected to API',
    disconnected: lang === 'pt' ? 'API Offline' : 'API Offline',
    checking: lang === 'pt' ? 'Verificando conexão...' : 'Checking connection...',
    refresh: lang === 'pt' ? 'Atualizar' : 'Refresh',
  };

  const fetchLiveCloudData = async () => {
    const version = ++connectionVersion.current;
    setLiveLoading(true); setLiveError(null); setApiStatus('checking'); setRuns([]); setSelectedRun(null);
    try {
      const result = await api().runs();
      if (version === connectionVersion.current) { setRuns(result); setApiStatus('online'); }
      const linkedRun = new URLSearchParams(window.location.search).get('run');
      if (linkedRun && version === connectionVersion.current) {
        const detail = result.find(run => run.id === linkedRun) || await api().run(linkedRun);
        if (version === connectionVersion.current) setSelectedRun(detail);
      }
    } catch (error) {
      if (version === connectionVersion.current) { setApiStatus('offline'); setLiveError(errorMessage(error)); }
    } finally { if (version === connectionVersion.current) setLiveLoading(false); }
  };
  const handleToggleMode = (mode: boolean) => {
    resetConnection(); setIsLiveMode(mode);
    if (!mode) { setRuns(INITIAL_RUNS); setApiToken(''); }
  };
  const handleInspectRun = (run: RunItem) => setSelectedRun(run);

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

  const totalRunsCount = runs.length;
  const passedRunsCount = runs.filter((r) => r.status === 'passed').length;
  const healthScore = totalRunsCount > 0 ? ((passedRunsCount / totalRunsCount) * 100).toFixed(1) : '—';
  const totalSchedulesSum = runs.every(r => r.schedulesCount !== null) ? runs.reduce((acc, r) => acc + (r.schedulesCount ?? 0), 0) : null;
  const regressionsCount = runs.filter((r) => r.isRegression).length;

  const handleCopyCode = () => {
    if (!selectedRun?.reproGoCode) return;
    navigator.clipboard.writeText(selectedRun.reproGoCode);
    setCopiedCode(true);
    setTimeout(() => setCopiedCode(false), 2000);
  };

  const handleCopyCmd = () => {
    if (!selectedRun) return;
    const cmd = `chaossql run --scenario=${selectedRun.scenario} --seed=${selectedRun.seed} --driver=${selectedRun.driver.toLowerCase().includes('postgres') ? 'postgres' : 'sqlite'}`;
    navigator.clipboard.writeText(cmd);
    setCopiedCmd(true);
    setTimeout(() => setCopiedCmd(false), 2000);
  };

  const handleDownloadRepro = () => {
    if (!selectedRun) return;
    const code = selectedRun.reproGoCode || `package repro_test\n\n// Reproduction test for scenario: ${selectedRun.scenario}\n// Deterministic seed: ${selectedRun.seed}\n`;
    const blob = new Blob([code], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `repro_${selectedRun.scenario}_seed${selectedRun.seed}_test.go`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
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
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: bregaldahq/chaossql@v1.5.0
        with:
          spec-path: 'chaos.yaml'
          cloud-url: \${{ vars.CHAOSSQL_CLOUD_URL }}
          cloud-token: \${{ secrets.CHAOSSQL_CLOUD_TOKEN }}
          github-token: \${{ secrets.GITHUB_TOKEN }}
          post-pr-comment: 'true'`;
    navigator.clipboard.writeText(yaml);
    setCopiedWorkflow(true);
    setTimeout(() => setCopiedWorkflow(false), 2000);
  };

  const handleCopyToken = () => {
    if (!issuedToken) return;
    navigator.clipboard.writeText(issuedToken);
    setCopiedToken(true);
    setTimeout(() => setCopiedToken(false), 2000);
  };

  return (
    <div className={styles.pageContainer} data-surface="light">
      {/* Header matching MatrixPage standard */}
      <div className={styles.header}>
        <p className="technical-label" style={{ color: 'var(--purple)' }}>
          {lang === 'pt' ? 'Observabilidade de Isolamento Concorrente // Control Plane' : 'Concurrency Isolation Observability // Control Plane'}
        </p>
        <div className={styles.headerTitleRow}>
          <div>
            <h1 className={styles.title}>{t.title}</h1>
            <p className={styles.subtitle}>{t.subtitle}</p>
          </div>

          <div className={styles.headerActions}>
            {/* Live vs Demo Segmented Control */}
            <div className={styles.modeSegmentedGroup}>
              <button
                type="button"
                className={`${styles.modeBtn} ${!isLiveMode ? styles.modeBtnActive : ''}`}
                onClick={() => handleToggleMode(false)}
              >
                {t.demoMode}
              </button>
              <button
                type="button"
                className={`${styles.modeBtn} ${isLiveMode ? styles.modeBtnActive : ''}`}
                onClick={() => handleToggleMode(true)}
              >
                <Server size={13} style={{ marginRight: 5 }} />
                {t.liveMode}
              </button>
            </div>

            <button
              type="button"
              className={styles.webhookModalBtn}
              onClick={openWebhooks}
            >
              <Bell size={14} />
              {lang === 'pt' ? 'Alertas & Webhooks' : 'Alerts & Webhooks'}
              {webhooks.length > 0 && (
                <span
                  style={{
                    background: 'var(--purple)',
                    color: 'var(--cream)',
                    borderRadius: '10px',
                    padding: '1px 6px',
                    fontSize: '0.68rem',
                    fontWeight: 700,
                    marginLeft: 4,
                  }}
                >
                  {webhooks.length}
                </span>
              )}
            </button>

            <button
              type="button"
              className={styles.connectRepoBtn}
              onClick={() => setOnboardingOpen(true)}
            >
              + {lang === 'pt' ? 'Conectar Repositório' : 'Connect Repository'}
            </button>
          </div>
        </div>
      </div>

      {/* Live Mode Configuration & Connection Banner */}
      {isLiveMode && (
        <div className={styles.liveConfigStrip}>
          <div className={styles.liveEndpointBox}>
            <span style={{ fontSize: 'var(--type-meta)', fontFamily: 'var(--font-jetbrains-mono)', color: 'var(--text-secondary)' }}>
              {lang === 'pt' ? 'Servidor ChaosSQL:' : 'ChaosSQL Server:'}
            </span>
            {apiStatus === 'online' && (
              <span className={styles.badgePrevented}>
                <Wifi size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} /> {t.connected}
              </span>
            )}
            {apiStatus === 'offline' && (
              <span className={styles.badgeRegression}>
                <WifiOff size={12} style={{ marginRight: 4, verticalAlign: 'middle' }} /> {t.disconnected}
              </span>
            )}
            {apiStatus === 'checking' && (
              <span className={styles.badgePermitted}>
                <RefreshCw size={12} className={styles.spin} style={{ marginRight: 4, verticalAlign: 'middle' }} /> {t.checking}
              </span>
            )}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
            <input
              type="text"
              value={apiUrl}
              aria-label="API URL"
              onChange={(e) => { resetConnection(); setApiUrl(e.target.value); }}
              className={styles.apiUrlInput}
              placeholder={lang === 'pt' ? 'Origem atual (padrão)' : 'Same origin (default)'}
            />
            <input type="password" aria-label="API token" placeholder="API token" autoComplete="off"
              value={apiToken} onChange={e => { resetConnection(); setApiToken(e.target.value); }} className={styles.apiUrlInput} />
            <button
              type="button"
              className={styles.refreshBtn}
              onClick={fetchLiveCloudData}
              disabled={liveLoading || !apiToken.trim()}
            >
              <RefreshCw size={13} className={liveLoading ? styles.spin : ''} />
              {t.refresh}
            </button>
          </div>

          <p>{lang === 'pt' ? 'Token mantido apenas na memória desta página.' : 'Token kept only in this page’s memory.'}</p>
          {liveError && <p role="alert"><AlertTriangle size={14} /> {liveError}</p>}
        </div>
      )}

      {/* 4 Metric Cards Grid */}
      <div className={styles.metricsGrid}>
        <div className={styles.metricCard}>
          <div className={styles.metricHeader}>
            <span className={styles.metricTitle}>{t.healthLabel}</span>
            <Activity size={16} color="var(--green)" />
          </div>
          <div className={styles.metricValue} style={{ color: 'var(--green)' }}>{healthScore === '—' ? '—' : `${healthScore}%`}</div>
          <div className={styles.metricSub}>
            <span style={{ color: 'var(--green)' }}>●</span> {t.healthStatus}
          </div>
        </div>

        <div className={styles.metricCard}>
          <div className={styles.metricHeader}>
            <span className={styles.metricTitle}>{t.totalRuns}</span>
            <Layers size={16} color="var(--purple)" />
          </div>
          <div className={styles.metricValue}>{totalRunsCount}</div>
          <div className={styles.metricSub}>
            <span>{passedRunsCount} passados • {runs.length - passedRunsCount} com anomalias</span>
          </div>
        </div>

        <div className={styles.metricCard}>
          <div className={styles.metricHeader}>
            <span className={styles.metricTitle}>{t.totalSchedules}</span>
            <Activity size={16} color="var(--purple)" />
          </div>
          <div className={styles.metricValue}>{totalSchedulesSum === null || !runs.length ? '—' : totalSchedulesSum.toLocaleString()}</div>
          <div className={styles.metricSub}>
            <span>{lang === 'pt' ? 'Nas execuções exibidas; — indica dados ausentes' : 'Across displayed runs; — means unavailable'}</span>
          </div>
        </div>

        <div className={styles.metricCard}>
          <div className={styles.metricHeader}>
            <span className={styles.metricTitle}>{t.regressionsCaught}</span>
            <ShieldAlert size={16} color={regressionsCount > 0 ? '#EF4444' : 'var(--text-secondary)'} />
          </div>
          <div className={regressionsCount > 0 ? styles.metricValueAlert : styles.metricValue}>{regressionsCount}</div>
          <div className={styles.metricSub}>
            <span style={{ color: regressionsCount > 0 ? '#DC2626' : 'inherit' }}>
              {regressionsCount} {t.openRegressions}{runs.some(r => r.isRegression === null) ? (lang === 'pt' ? ' (dados incompletos)' : ' (incomplete metadata)') : ''}
            </span>
          </div>
        </div>
      </div>

      {/* Legend Bar (Mirrors MatrixPage) */}
      <div className={styles.legendBar}>
        <div className={styles.legendItemsGroup}>
          <div className={styles.legendItem}>
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: 'var(--green)' }} />
            <span>PASS: Sem anomalias</span>
          </div>
          <div className={styles.legendItem}>
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: 'var(--yellow)' }} />
            <span>ANOMALY: Risco transacional detectado</span>
          </div>
          <div className={styles.legendItem}>
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: '#EF4444' }} />
            <span>REGRESSION: Quebra vs Branch Base</span>
          </div>
          <div className={styles.legendItem}>
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: 'var(--purple)' }} />
            <span>CICLO: Deadlock / Serializability Cycle</span>
          </div>
        </div>
      </div>

      {/* Filter and Search Controls */}
      <div className={styles.tableControls}>
        <div className={styles.tabsGroup}>
          <button
            type="button"
            className={`${styles.tabBtn} ${filterType === 'all' ? styles.tabBtnActive : ''}`}
            onClick={() => setFilterType('all')}
          >
            {t.allRuns}
          </button>
          <button
            type="button"
            className={`${styles.tabBtn} ${filterType === 'regressions' ? styles.tabBtnActiveAlert : ''}`}
            onClick={() => setFilterType('regressions')}
          >
            {t.regressionsOnly} {regressionsCount > 0 ? `(${regressionsCount})` : ''}
          </button>
          <button
            type="button"
            className={`${styles.tabBtn} ${filterType === 'prs' ? styles.tabBtnActive : ''}`}
            onClick={() => setFilterType('prs')}
          >
            {t.prsOnly}
          </button>
          <button
            type="button"
            className={`${styles.tabBtn} ${filterType === 'passed' ? styles.tabBtnActive : ''}`}
            onClick={() => setFilterType('passed')}
          >
            {t.passedOnly}
          </button>
        </div>

        <div className={styles.searchBox}>
          <Search size={15} className={styles.searchIcon} />
          <input
            type="text"
            className={styles.searchInput}
            placeholder={t.searchPlaceholder}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Runs Table Card (Matching MatrixTable pattern) */}
      <div className={styles.tableCard}>
        {filteredRuns.length === 0 ? (
          <div className={styles.emptyState}>
            <Layers size={36} color="var(--purple)" style={{ opacity: 0.5, marginBottom: 8 }} />
            <h3>{lang === 'pt' ? 'Nenhuma execução encontrada' : 'No executions found'}</h3>
            <p>
              {isLiveMode
                ? (apiStatus === 'online' ? (lang === 'pt' ? 'Nenhuma execução corresponde a esta seleção.' : 'No runs match this selection.') : (lang === 'pt' ? 'Conecte à API com seu token para carregar execuções.' : 'Connect to the API with your token to load runs.'))
                : (lang === 'pt' ? 'Nenhum resultado para o filtro informado.' : 'No results matching your filters.')}
            </p>
          </div>
        ) : (
          <table className={styles.table}>
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
                <tr
                  key={r.id}
                  className={r.isRegression ? styles.rowRegression : styles.row}
                  onClick={() => handleInspectRun(r)}
                >
                  <td>
                    <div className={styles.repoCell}>
                      <span className={styles.repoName}>{r.repo}</span>
                      <span className={styles.scenarioName}>{r.scenario}</span>
                    </div>
                  </td>
                  <td>
                    <div className={styles.branchCell}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <code className={styles.branchCode}>{r.branch}</code>
                        {r.commitSHA && (
                          <a
                            href={`https://github.com/${r.repo}/commit/${r.commitSHA}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className={styles.prGithubBadge}
                            onClick={(e) => e.stopPropagation()}
                            title={lang === 'pt' ? 'Ver commit no GitHub' : 'View commit on GitHub'}
                          >
                            <code>{r.commitSHA.slice(0, 7)}</code>
                          </a>
                        )}
                      </div>
                      {r.prNumber && r.prNumber > 0 ? (
                        <a
                          href={`https://github.com/${r.repo}/pull/${r.prNumber}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={styles.prGithubBadge}
                          onClick={(e) => e.stopPropagation()}
                          title={lang === 'pt' ? 'Abrir Pull Request no GitHub' : 'Open Pull Request on GitHub'}
                        >
                          <GitPullRequest size={11} /> #{r.prNumber}
                          <ExternalLink size={9} style={{ marginLeft: 3 }} />
                        </a>
                      ) : null}
                    </div>
                  </td>
                  <td>
                    {r.status === 'passed' ? (
                      <span className={styles.badgePrevented}>PASS</span>
                    ) : (
                      <span className={styles.badgeRegression}>{r.status.toUpperCase()}</span>
                    )}
                  </td>
                  <td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                      <span
                        className={
                          r.anomalyType === 'NONE'
                            ? styles.badgePrevented
                            : r.anomalyType === 'DEADLOCK'
                            ? styles.badgeCycle
                            : styles.badgePermitted
                        }
                      >
                        {r.anomalyType}
                      </span>
                      <span style={{ fontWeight: 600, fontSize: '0.85rem' }}>{r.anomalyName}</span>
                    </div>
                    {r.isRegression && (
                      <div>
                        <span className={styles.badgeRegression}>
                          REGRESSION
                        </span>
                      </div>
                    )}
                  </td>
                  <td>
                    <div className={styles.engineCell}>
                      <span className={styles.driverText}>{r.driver}</span>
                      <span className={styles.isolationText}>{r.isolation}</span>
                    </div>
                  </td>
                  <td>
                    <div className={styles.durationCell}>
                      <span className={styles.durationMono}>{r.durationMS === null ? '—' : `${r.durationMS}ms`}</span>
                      <span className={styles.schedulesMono}>{r.schedulesCount ?? '—'} sched</span>
                    </div>
                  </td>
                  <td>
                    <span style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', fontFamily: 'var(--font-jetbrains-mono)' }}>
                      {r.timestamp}
                    </span>
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    <button
                      type="button"
                      className={styles.inspectBtn}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleInspectRun(r);
                      }}
                    >
                      {t.inspectBtn}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Finding Detail Modal */}
      {selectedRun && (
        <div className={styles.modalBackdrop} onClick={() => setSelectedRun(null)}>
          <div className={styles.modalCard} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div>
                <span className={styles.modalTag}>RUN FINDING EXPLORER // {selectedRun.id}</span>
                <h2 className={styles.modalTitle}>
                  {selectedRun.anomalyName} ({selectedRun.anomalyType})
                </h2>
                <div className={styles.modalMetaRow}>
                  <span>{selectedRun.repo}</span>
                  <span>•</span>
                  <span>{selectedRun.branch}</span>
                  {selectedRun.prNumber && selectedRun.prNumber > 0 && (
                    <>
                      <span>•</span>
                      <a
                        href={`https://github.com/${selectedRun.repo}/pull/${selectedRun.prNumber}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={styles.prGithubBadge}
                      >
                        <GitPullRequest size={11} /> PR #{selectedRun.prNumber}
                        <ExternalLink size={9} style={{ marginLeft: 3 }} />
                      </a>
                    </>
                  )}
                  <span>•</span>
                  <span>Seed: <code>{selectedRun.seed ?? '—'}</code></span>
                </div>
              </div>
              <button type="button" className={styles.closeBtn} onClick={() => setSelectedRun(null)}>
                <X size={20} />
              </button>
            </div>

            <div className={styles.modalBody}>
              {isLiveMode && <p>{lang === 'pt' ? 'Consulte os artefatos locais do CI para SQL, traces e reprodutores. A nuvem armazena apenas metadados.' : 'Use your local CI artifacts for SQL, traces and reproducers. The cloud stores metadata only.'}</p>}
              {/* Invariant Failure */}
              {selectedRun.failingInvariant && (
                <div className={styles.sectionBox}>
                  <h4 className={styles.sectionTitle}>{t.invariantBoxTitle}</h4>
                  <div className={styles.invariantGrid}>
                    <div>
                      <span className={styles.labelMuted}>Invariante:</span>
                      <code>{selectedRun.failingInvariant.name}</code>
                    </div>
                    <div>
                      <span className={styles.labelMuted}>Query de Verificação:</span>
                      <code>{selectedRun.failingInvariant.query}</code>
                    </div>
                    <div>
                      <span className={styles.labelMuted}>Condição Esperada:</span>
                      <span style={{ color: 'var(--green)', fontWeight: 600 }}>{selectedRun.failingInvariant.assertion}</span>
                    </div>
                    <div>
                      <span className={styles.labelMuted}>Valor Real Violado:</span>
                      <span style={{ color: '#DC2626', fontWeight: 600 }}>{selectedRun.failingInvariant.actual}</span>
                    </div>
                  </div>
                </div>
              )}

              {/* Minimal Causal Trace Steps */}
              {selectedRun.traceSteps && selectedRun.traceSteps.length > 0 && (
                <div className={styles.sectionBox}>
                  <h4 className={styles.sectionTitle}>{t.traceBoxTitle}</h4>
                  <div className={styles.traceTimeline}>
                    {selectedRun.traceSteps.map((s, idx) => (
                      <div key={idx} className={styles.traceRow}>
                        <span className={styles.workerPill}>{s.worker}</span>
                        <span className={s.opType === 'write' ? styles.opWrite : styles.opRead}>
                          {s.opType.toUpperCase()}
                        </span>
                        <code className={styles.sqlSnippet}>{s.sql}</code>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Reproducer Code in Go */}
              {selectedRun.reproGoCode && (
                <div className={styles.sectionBox}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                    <h4 className={styles.sectionTitle}>{t.reproCodeTitle}</h4>
                    <button type="button" className={styles.copyBtn} onClick={handleCopyCode}>
                      {copiedCode ? <Check size={14} /> : <Copy size={14} />}
                      {copiedCode ? 'Copiado!' : 'Copiar Go Code'}
                    </button>
                  </div>
                  <pre className={styles.codePre}>
                    <code>{selectedRun.reproGoCode}</code>
                  </pre>
                </div>
              )}
            </div>

            {!isLiveMode && <div className={styles.modalFooter}>
              {/* Deep-link direct to VisualizerPage */}
              <a
                href="/visualizer"
                className={styles.openVisualizerBtn}
                onClick={() => setSelectedRun(null)}
              >
                <Layers size={14} />
                {t.openVisualizer}
              </a>

              {/* Download standalone repro_test.go */}
              <button type="button" className={styles.downloadBtn} onClick={handleDownloadRepro}>
                <Download size={14} />
                {t.downloadRepro}
              </button>

              {/* Copy CLI command */}
              <button type="button" className={styles.copyCmdBtn} onClick={handleCopyCmd}>
                {copiedCmd ? <Check size={14} /> : <Copy size={14} />}
                {copiedCmd ? 'Comando Copiado!' : t.copyCmd}
              </button>

              <a
                href={`/playground?scenario=${selectedRun.scenario}`}
                className={styles.playgroundBtn}
                onClick={() => setSelectedRun(null)}
              >
                <Play size={14} fill="currentColor" />
                {t.openPlayground}
              </a>
            </div>}
          </div>
        </div>
      )}

      {/* Onboarding / Connect Repository Modal */}
      {onboardingOpen && (
        <div className={styles.modalBackdrop} onClick={() => (setOnboardingOpen(false), setIssuedToken(''), setTokenError(null))}>
          <div className={styles.onboardingCard} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div>
                <span className={styles.modalTag}>SETUP EM 60 SEGUNDOS</span>
                <h3 className={styles.modalTitle}>Conectar Repositório ao ChaosSQL Cloud</h3>
              </div>
              <button type="button" className={styles.closeBtn} onClick={() => (setOnboardingOpen(false), setIssuedToken(''), setTokenError(null))}>
                <X size={20} />
              </button>
            </div>

            <div className={styles.onboardingBody}>
              <div className={styles.onboardingStep}>
                <div className={styles.stepNum}>1</div>
                <div className={styles.stepContent}>
                  <h4>{lang === 'pt' ? 'Crie um token de CI' : 'Create a CI token'}</h4>
                  <p>{lang === 'pt' ? 'Conecte à API com um token de administrador. O novo token de membro será mostrado uma vez; salve-o como secret do GitHub.' : 'Connect with an administrator API token. The new member token is shown once; save it as a GitHub secret.'}</p>
                  <p>{lang === 'pt' ? 'Sem acesso de administrador? Peça ao operador do servidor para executar:' : 'Without administrator access, ask your server operator to run:'}</p>
                  <code>chaossql server create-token --org YOUR_ORG_ID --name "CI Token"</code>
                  {tokenError && <p role="alert">{tokenError}</p>}
                  {issuedToken ? <div className={styles.tokenBox}>
                    <code>{issuedToken}</code>
                    <button type="button" className={styles.actionBtnSmall} onClick={handleCopyToken}>
                      {copiedToken ? <Check size={14} /> : <Copy size={14} />}
                      {copiedToken ? 'Copied' : 'Copy Token'}
                    </button>
                  </div> : <button type="button" className={styles.actionBtnSmall} disabled={!isLiveMode || apiStatus !== 'online' || tokenBusy} onClick={handleCreateToken}>
                    {tokenBusy ? 'Creating…' : (lang === 'pt' ? 'Criar token de CI' : 'Create CI Token')}
                  </button>}
                </div>
              </div>

              <div className={styles.onboardingStep}>
                <div className={styles.stepNum}>2</div>
                <div className={styles.stepContent}>
                  <h4>Configure o Secret no GitHub</h4>
                  <p>{lang === 'pt' ? 'Defina também a variável CHAOSSQL_CLOUD_URL com a URL da API acessível pelo runner.' : 'Also set the CHAOSSQL_CLOUD_URL repository variable to the API URL reachable from the runner.'}</p>
                  <p>
                    No seu repositório no GitHub, acesse <strong>Settings → Secrets and variables → Actions → New repository secret</strong>.
                    Defina o nome como <code style={{ color: 'var(--purple)', background: 'color-mix(in srgb, var(--purple) 8%, var(--cream))', padding: '2px 6px', borderRadius: 3 }}>CHAOSSQL_CLOUD_TOKEN</code>.
                  </p>
                </div>
              </div>

              <div className={styles.onboardingStep}>
                <div className={styles.stepNum}>3</div>
                <div className={styles.stepContent}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 }}>
                    <h4>Adicione o Workflow no Repositório</h4>
                    <button type="button" className={styles.actionBtnSmall} onClick={handleCopyWorkflow}>
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
          cloud-url: \${{ vars.CHAOSSQL_CLOUD_URL }}
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
              <button type="button" className={styles.copyCmdBtn} onClick={() => (setOnboardingOpen(false), setIssuedToken(''), setTokenError(null))}>
                Concluir & Voltar ao Dashboard
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Webhooks & Alerts Modal */}
      {webhooksModalOpen && (
        <div className={styles.modalOverlay} onClick={() => setWebhooksModalOpen(false)}>
          <div className={styles.webhookModalCard} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <div>
                <h3 className={styles.modalTitle}>
                  <Bell size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom', color: 'var(--purple)' }} />
                  {lang === 'pt' ? 'Alertas & Webhooks em Tempo Real' : 'Real-Time Alert Webhooks'}
                </h3>
                <p className={styles.modalSubtitle}>
                  {lang === 'pt'
                    ? 'Receba alertas instantâneos no Discord, Slack ou SIEM corporativo quando uma regressão de concorrência for detectada no CI.'
                    : 'Get instant alerts in Discord, Slack or SIEM whenever a concurrency regression is caught in CI.'}
                </p>
              </div>
              <button
                type="button"
                className={styles.modalCloseBtn}
                onClick={() => setWebhooksModalOpen(false)}
              >
                <X size={18} />
              </button>
            </div>

            <div className={styles.modalBody}>
              {!isLiveMode || apiStatus !== 'online' ? <p>{lang === 'pt' ? 'Conecte à API para gerenciar webhooks.' : 'Connect to the API to manage webhooks.'}</p> : null}
              {webhookBusy && <p>{lang === 'pt' ? 'Carregando…' : 'Loading…'}</p>}
              {webhookError && <p role="alert">{webhookError}</p>}
              {/* Existing Webhooks List */}
              <div style={{ marginBottom: 'var(--space-3)' }}>
                <h4
                  style={{
                    fontFamily: 'var(--font-jetbrains-mono)',
                    fontSize: '0.82rem',
                    color: 'var(--ink)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    marginBottom: 'var(--space-2)',
                  }}
                >
                  {lang === 'pt' ? 'Canais Conectados' : 'Connected Channels'} ({webhooks.length})
                </h4>

                {webhooks.length === 0 ? (
                  <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', padding: '16px 0' }}>
                    {lang === 'pt' ? 'Nenhum webhook configurado ainda.' : 'No webhooks configured yet.'}
                  </p>
                ) : (
                  <div className={styles.webhookList}>
                    {webhooks.map((wh) => (
                      <div key={wh.id} className={styles.webhookItem}>
                        <div className={styles.webhookItemHeader}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            {wh.target === 'discord' && (
                              <span className={styles.webhookTargetBadgeDiscord}>
                                Discord
                              </span>
                            )}
                            {wh.target === 'slack' && (
                              <span className={styles.webhookTargetBadgeSlack}>
                                Slack
                              </span>
                            )}
                            {wh.target === 'generic' && (
                              <span className={styles.webhookTargetBadgeGeneric}>
                                Generic JSON
                              </span>
                            )}
                            <strong style={{ fontFamily: 'var(--font-inter)', fontSize: '0.88rem', color: 'var(--ink)' }}>
                              {wh.id}
                            </strong>
                          </div>

                          <div className={styles.webhookActions}>
                            <button
                              type="button"
                              className={styles.webhookTestBtn}
                              disabled={testingWebhookId !== null || webhookBusy}
                              onClick={() => handleTestWebhook(wh)}
                            >
                              {testingWebhookId === wh.id ? (
                                <>
                                  <RefreshCw size={12} className="spin" />
                                  {lang === 'pt' ? 'Disparando...' : 'Sending...'}
                                </>
                              ) : testSuccessId === wh.id ? (
                                <>
                                  <CheckCircle2 size={12} style={{ color: '#059669' }} />
                                  <span style={{ color: '#059669' }}>
                                    {lang === 'pt' ? 'Enviado!' : 'Delivered!'}
                                  </span>
                                </>
                              ) : (
                                <>
                                  <Send size={12} />
                                  {lang === 'pt' ? 'Testar Alerta' : 'Test Alert'}
                                </>
                              )}
                            </button>

                            <button
                              type="button"
                              className={styles.webhookDeleteBtn}
                              disabled={webhookBusy || testingWebhookId !== null}
                              onClick={() => handleDeleteWebhook(wh.id)}
                              title={lang === 'pt' ? 'Remover webhook' : 'Delete webhook'}
                            >
                              <Trash2 size={13} />
                              {lang === 'pt' ? 'Remover' : 'Remove'}
                            </button>
                          </div>
                        </div>

                        <div className={styles.webhookUrlCode}>
                          {wh.url.length > 70 ? `${wh.url.slice(0, 40)}...${wh.url.slice(-25)}` : wh.url}
                        </div>

                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 12,
                            fontSize: '0.75rem',
                            color: 'var(--text-secondary)',
                            fontFamily: 'var(--font-jetbrains-mono)',
                          }}
                        >
                          <span>
                            {lang === 'pt' ? 'Eventos: ' : 'Events: '}
                            <strong>{wh.events.join(', ')}</strong>
                          </span>
                          <span>•</span>
                          <span>
                            {lang === 'pt' ? 'Status: ' : 'Status: '}
                            <span style={{ color: wh.active ? '#059669' : '#9CA3AF', fontWeight: 600 }}>
                              {wh.active ? (lang === 'pt' ? 'Ativo' : 'Active') : (lang === 'pt' ? 'Inativo' : 'Disabled')}
                            </span>
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Add Webhook Form */}
              <div className={styles.webhookFormCard}>
                <h4
                  style={{
                    fontFamily: 'var(--font-jetbrains-mono)',
                    fontSize: '0.82rem',
                    color: 'var(--ink)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    margin: 0,
                  }}
                >
                  + {lang === 'pt' ? 'Adicionar Novo Destino de Alerta' : 'Add New Alert Destination'}
                </h4>
                <form onSubmit={handleAddWebhook} className={styles.webhookFormGrid}>
                  <select
                    className={styles.webhookSelect}
                    value={newWhTarget}
                    onChange={(e) => setNewWhTarget(e.target.value as WebhookItem['target'])}
                  >
                    <option value="discord">Discord Webhook</option>
                    <option value="slack">Slack Incoming</option>
                    <option value="generic">Custom JSON POST</option>
                  </select>

                  <input
                    type="url"
                    required
                    placeholder={
                      newWhTarget === 'discord'
                        ? 'https://discord.com/api/webhooks/...'
                        : newWhTarget === 'slack'
                        ? 'https://hooks.slack.com/services/...'
                        : 'https://api.empresa.com/webhooks/concurrency'
                    }
                    className={styles.webhookInput}
                    value={newWhUrl}
                    onChange={(e) => setNewWhUrl(e.target.value)}
                  />

                  <button type="submit" className={styles.webhookSaveBtn} disabled={!isLiveMode || apiStatus !== 'online' || webhookBusy}>
                    + {lang === 'pt' ? 'Salvar Webhook' : 'Save Webhook'}
                  </button>
                </form>


              </div>
            </div>

            <div className={styles.modalFooter}>
              <button
                type="button"
                className={styles.copyCmdBtn}
                onClick={() => setWebhooksModalOpen(false)}
              >
                {lang === 'pt' ? 'Fechar' : 'Close'}
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}
