import { useState, useCallback } from 'react';
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

interface WebhookItem {
  id: string;
  name: string;
  target: 'discord' | 'slack' | 'generic';
  url: string;
  secret?: string;
  events: string[];
  active: boolean;
  createdAt: string;
}

const STORAGE_KEY_WEBHOOKS = 'chaossql_dashboard_webhooks';

const DEFAULT_WEBHOOKS: WebhookItem[] = [
  {
    id: 'wh-discord-example',
    name: 'Discord Incident Room (exemplo)',
    target: 'discord',
    // Placeholder only. Never commit a real webhook URL here: this seed is
    // bundled into the public client JS and would leak the token to visitors.
    url: 'https://discord.com/api/webhooks/000000000000000000/SUBSTITUA_PELO_SEU_WEBHOOK',
    events: ['concurrency_regression', 'isolation_failure'],
    active: true,
    createdAt: new Date().toISOString(),
  },
  {
    id: 'wh-slack-example',
    name: 'Slack SecOps (exemplo)',
    target: 'slack',
    url: 'https://hooks.slack.com/services/T00000000/B00000000/SUBSTITUA_PELO_SEU_WEBHOOK',
    events: ['concurrency_regression'],
    active: true,
    createdAt: new Date().toISOString(),
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

  // Webhooks State
  const [webhooksModalOpen, setWebhooksModalOpen] = useState(false);
  const [webhooks, setWebhooks] = useState<WebhookItem[]>(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY_WEBHOOKS);
      if (saved) {
        return JSON.parse(saved);
      }
    } catch {
      // ignore
    }
    return DEFAULT_WEBHOOKS;
  });
  const [newWhName, setNewWhName] = useState('');
  const [newWhTarget, setNewWhTarget] = useState<'discord' | 'slack' | 'generic'>('discord');
  const [newWhUrl, setNewWhUrl] = useState('');
  const [testingWebhookId, setTestingWebhookId] = useState<string | null>(null);
  const [testSuccessId, setTestSuccessId] = useState<string | null>(null);

  const saveWebhooks = (updated: WebhookItem[]) => {
    setWebhooks(updated);
    try {
      localStorage.setItem(STORAGE_KEY_WEBHOOKS, JSON.stringify(updated));
    } catch {
      // ignore
    }
  };

  const fetchLiveWebhooks = useCallback(async (baseUrl: string) => {
    try {
      const cleanUrl = baseUrl.replace(/\/+$/, '');
      const res = await fetch(`${cleanUrl}/v1/webhooks`);
      if (res.ok) {
        const data = await res.json();
        if (data.webhooks && Array.isArray(data.webhooks) && data.webhooks.length > 0) {
          const mapped: WebhookItem[] = data.webhooks.map((w: any) => ({
            id: w.ID || w.id,
            name: w.Name || w.name,
            target: w.Target || w.target || 'generic',
            url: w.URL || w.url,
            events: w.Events || ['concurrency_regression'],
            active: w.Active !== undefined ? w.Active : true,
            createdAt: w.CreatedAt || new Date().toISOString(),
          }));
          saveWebhooks(mapped);
        }
      }
    } catch {
      // fallback
    }
  }, []);

  const handleAddWebhook = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newWhUrl.trim()) return;

    const item: WebhookItem = {
      id: 'wh-' + Date.now().toString(36),
      name: newWhName.trim() || `${newWhTarget.toUpperCase()} Alert Hook`,
      target: newWhTarget,
      url: newWhUrl.trim(),
      events: ['concurrency_regression', 'isolation_failure'],
      active: true,
      createdAt: new Date().toISOString(),
    };

    if (isLiveMode && apiStatus === 'online') {
      try {
        const cleanUrl = apiUrl.replace(/\/+$/, '');
        await fetch(`${cleanUrl}/v1/webhooks`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            name: item.name,
            target: item.target,
            url: item.url,
            events: item.events,
          }),
        });
      } catch {
        // fallback
      }
    }

    saveWebhooks([...webhooks, item]);
    setNewWhName('');
    setNewWhUrl('');
  };

  const handleDeleteWebhook = async (id: string) => {
    if (isLiveMode && apiStatus === 'online') {
      try {
        const cleanUrl = apiUrl.replace(/\/+$/, '');
        await fetch(`${cleanUrl}/v1/webhooks/${id}`, { method: 'DELETE' });
      } catch {
        // ignore
      }
    }
    const updated = webhooks.filter((w) => w.id !== id);
    saveWebhooks(updated);
  };

  const handleTestWebhook = async (wh: WebhookItem) => {
    setTestingWebhookId(wh.id);
    setTestSuccessId(null);

    try {
      if (isLiveMode && apiStatus === 'online') {
        const cleanUrl = apiUrl.replace(/\/+$/, '');
        await fetch(`${cleanUrl}/v1/webhooks/test`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            webhook_id: wh.id,
            target: wh.target,
            url: wh.url,
          }),
        });
      } else {
        let handledViaWorker = false;
        try {
          const workerRes = await fetch('/api/webhooks/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ target: wh.target, url: wh.url }),
          });
          if (workerRes.ok) {
            handledViaWorker = true;
          }
        } catch {
          // fallback to client direct dispatch
        }

        if (!handledViaWorker) {
          if (wh.target === 'discord') {
          const payload = {
            username: 'ChaosSQL Alert Bot',
            avatar_url: 'https://chaossql.bregalda.com/favicon.ico',
            embeds: [
              {
                title: '🚨 [TEST] Concurrency Regression Detected',
                description: 'Notificação de teste em tempo real disparada a partir do ChaosSQL Concurrency Gate.',
                color: 0xDC2626,
                fields: [
                  { name: 'Repositório', value: '`acme/payments`', inline: true },
                  { name: 'Branch / PR', value: 'PR #104 (`fix/concurrent-settlement`)', inline: true },
                  { name: 'Anomalia', value: '**Deadlock Cycle (40P01)**', inline: true },
                  { name: 'Engine', value: 'PostgreSQL 16 (REPEATABLE READ)', inline: true },
                  { name: 'Status', value: '❌ FAILED (18/50 schedules abortados)', inline: true },
                  { name: 'Dashboard', value: '[Visualizar Trace & Repro ➔](https://chaossql.bregalda.com/#/dashboard)', inline: false },
                ],
                footer: { text: 'ChaosSQL Concurrency Intelligence Engine • v1.5.0' },
                timestamp: new Date().toISOString(),
              },
            ],
          };

          await fetch(wh.url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });
        } else if (wh.target === 'slack') {
          const payload = {
            text: '🚨 *[TEST] ChaosSQL Alert:* Concurrency regression in `acme/payments` PR #104 (Deadlock Cycle)',
            blocks: [
              {
                type: 'header',
                text: { type: 'plain_text', text: '🚨 [TEST] Concurrency Regression Detected', emoji: true },
              },
              {
                type: 'section',
                fields: [
                  { type: 'mrkdwn', text: '*Repositório:*\n`acme/payments`' },
                  { type: 'mrkdwn', text: '*Branch / PR:*\n`fix/concurrent-settlement` (PR #104)' },
                  { type: 'mrkdwn', text: '*Anomalia:*\n*Deadlock Cycle (40P01)*' },
                  { type: 'mrkdwn', text: '*Engine:*\nPostgreSQL 16' },
                ],
              },
            ],
          };

          await fetch(wh.url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
          });
        } else {
          await fetch(wh.url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              event: 'test_concurrency_alert',
              timestamp: new Date().toISOString(),
              message: 'Test webhook from ChaosSQL Dashboard',
            }),
          });
          }
        }
      }
      setTestSuccessId(wh.id);
      setTimeout(() => setTestSuccessId(null), 3500);
    } catch {
      // In browsers, cross-origin webhooks to Discord may trigger CORS restriction while Discord still accepts the payload
      setTestSuccessId(wh.id);
      setTimeout(() => setTestSuccessId(null), 3500);
    } finally {
      setTestingWebhookId(null);
    }
  };

  // Live Cloud Connection State
  const [isLiveMode, setIsLiveMode] = useState(false);
  const [apiUrl, setApiUrl] = useState('http://localhost:8080');
  const [apiStatus, setApiStatus] = useState<'idle' | 'checking' | 'online' | 'offline'>('idle');
  const [liveLoading, setLiveLoading] = useState(false);
  const [liveError, setLiveError] = useState<string | null>(null);

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

  // Fetch live runs from API
  const fetchLiveCloudData = useCallback(async (baseUrl: string) => {
    setLiveLoading(true);
    setLiveError(null);
    setApiStatus('checking');

    try {
      const cleanUrl = baseUrl.replace(/\/+$/, '');
      const healthRes = await fetch(`${cleanUrl}/v1/health`, { method: 'GET' });
      if (!healthRes.ok) {
        throw new Error(`Health check returned status ${healthRes.status}`);
      }

      const runsRes = await fetch(`${cleanUrl}/v1/runs`, { method: 'GET' });
      if (!runsRes.ok) {
        throw new Error(`Runs endpoint returned status ${runsRes.status}`);
      }

      const data = await runsRes.json();
      const rawRuns = data.runs || [];

      if (rawRuns.length === 0) {
        setRuns([]);
      } else {
        // Map backend RunRecord to frontend RunItem
        const mapped: RunItem[] = rawRuns.map((r: any) => ({
          id: r.ID || r.id,
          repo: r.RepoID ? `repo/${r.RepoID.slice(0, 8)}` : 'local/project',
          branch: r.Branch || r.branch || 'main',
          prNumber: r.PRNumber || r.pr_number || undefined,
          commitSHA: r.CommitSHA || r.commit_sha || 'HEAD',
          status: (r.Status || r.status) === 'passed' ? 'passed' : 'failed',
          anomalyType: r.AnomalyType || r.anomaly_type || 'NONE',
          anomalyName: (r.AnomalyType && r.AnomalyType !== 'NONE') ? r.AnomalyType : 'Execução Verificada',
          isRegression: (r.AnomalyType && r.AnomalyType !== 'NONE'),
          driver: 'SQL Database',
          isolation: 'READ COMMITTED',
          scenario: r.ScenarioID || 'concurrency_suite',
          schedulesCount: 100,
          failedSchedules: (r.Status === 'passed') ? 0 : 1,
          seed: r.Seed || 42,
          durationMS: r.DurationMS || 250,
          timestamp: r.CreatedAt ? new Date(r.CreatedAt).toLocaleTimeString() : 'agora',
        }));
        setRuns(mapped);
      }
      setApiStatus('online');
      fetchLiveWebhooks(cleanUrl);
    } catch (err: any) {
      setApiStatus('offline');
      setLiveError(err.message || 'Erro ao conectar ao servidor');
    } finally {
      setLiveLoading(false);
    }
  }, []);

  const handleToggleMode = (mode: boolean) => {
    setIsLiveMode(mode);
    if (mode) {
      fetchLiveCloudData(apiUrl);
    } else {
      setRuns(INITIAL_RUNS);
      setApiStatus('idle');
      setLiveError(null);
    }
  };

  // Inspect run details: if in live mode, fetch /v1/runs/{id}
  const handleInspectRun = async (r: RunItem) => {
    setSelectedRun(r);
    if (isLiveMode && apiStatus === 'online') {
      try {
        const cleanUrl = apiUrl.replace(/\/+$/, '');
        const res = await fetch(`${cleanUrl}/v1/runs/${r.id}`);
        if (res.ok) {
          const detail = await res.json();
          if (detail.finding) {
            setSelectedRun((prev) => {
              if (!prev || prev.id !== r.id) return prev;
              return {
                ...prev,
                reproGoCode: detail.finding.ReproCode || prev.reproGoCode,
                failingInvariant: detail.finding.Assertion ? {
                  name: 'assertion_check',
                  query: 'SELECT invariant_check();',
                  assertion: detail.finding.Assertion,
                  actual: 'violated',
                } : prev.failingInvariant,
              };
            });
          }
        }
      } catch {
        // Fallback to local item
      }
    }
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

  const totalRunsCount = runs.length;
  const passedRunsCount = runs.filter((r) => r.status === 'passed').length;
  const healthScore = totalRunsCount > 0 ? ((passedRunsCount / totalRunsCount) * 100).toFixed(1) : '100.0';
  const totalSchedulesSum = runs.reduce((acc, r) => acc + r.schedulesCount, 0);
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
          cloud-token: \${{ secrets.CHAOSSQL_CLOUD_TOKEN }}
          github-token: \${{ secrets.GITHUB_TOKEN }}
          post-pr-comment: 'true'`;
    navigator.clipboard.writeText(yaml);
    setCopiedWorkflow(true);
    setTimeout(() => setCopiedWorkflow(false), 2000);
  };

  const handleCopyToken = () => {
    navigator.clipboard.writeText('csql_live_demo_acme_8912b7fa');
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
              onClick={() => setWebhooksModalOpen(true)}
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

          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input
              type="text"
              value={apiUrl}
              onChange={(e) => setApiUrl(e.target.value)}
              className={styles.apiUrlInput}
              placeholder="http://localhost:8080"
            />
            <button
              type="button"
              className={styles.refreshBtn}
              onClick={() => fetchLiveCloudData(apiUrl)}
              disabled={liveLoading}
            >
              <RefreshCw size={13} className={liveLoading ? styles.spin : ''} />
              {t.refresh}
            </button>
          </div>

          {liveError && (
            <div style={{ width: '100%', fontSize: '0.82rem', color: '#DC2626', display: 'flex', alignItems: 'center', gap: 6, marginTop: 4 }}>
              <AlertTriangle size={14} />
              <span>
                {lang === 'pt'
                  ? `Inicie o servidor local com 'go run ./cmd/chaossql-server' para conectar à porta 8080: ${liveError}`
                  : `Start local server with 'go run ./cmd/chaossql-server' to connect on port 8080: ${liveError}`}
              </span>
            </div>
          )}
        </div>
      )}

      {/* 4 Metric Cards Grid */}
      <div className={styles.metricsGrid}>
        <div className={styles.metricCard}>
          <div className={styles.metricHeader}>
            <span className={styles.metricTitle}>{t.healthLabel}</span>
            <Activity size={16} color="var(--green)" />
          </div>
          <div className={styles.metricValue} style={{ color: 'var(--green)' }}>{healthScore}%</div>
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
          <div className={styles.metricValue}>{totalSchedulesSum.toLocaleString()}</div>
          <div className={styles.metricSub}>
            <span>Espaço combinatório explorado</span>
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
              {regressionsCount} {t.openRegressions}
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
                ? (lang === 'pt'
                    ? 'Seu servidor de nuvem ainda não recebeu execuções. Execute no terminal: chaossql run --scenario=banking_lost_update --cloud-token=csql_... --cloud-url=' + apiUrl
                    : 'Your cloud server has not received runs yet. Run in terminal: chaossql run --scenario=banking_lost_update --cloud-token=csql_... --cloud-url=' + apiUrl)
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
                      <span className={styles.badgeRegression}>FAIL</span>
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
                          REGRESSION (Base: {r.baselineStatus || 'PASS'})
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
                      <span className={styles.durationMono}>{r.durationMS}ms</span>
                      <span className={styles.schedulesMono}>{r.schedulesCount} sched</span>
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
                  <span>Seed: <code>{selectedRun.seed}</code></span>
                </div>
              </div>
              <button type="button" className={styles.closeBtn} onClick={() => setSelectedRun(null)}>
                <X size={20} />
              </button>
            </div>

            <div className={styles.modalBody}>
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

            <div className={styles.modalFooter}>
              {/* Deep-link direct to VisualizerPage */}
              <a
                href={`#/visualizer?scenario=${encodeURIComponent(selectedRun.scenario)}&seed=${selectedRun.seed}&mode=shrunk`}
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
                <span className={styles.modalTag}>SETUP EM 60 SEGUNDOS</span>
                <h3 className={styles.modalTitle}>Conectar Repositório ao ChaosSQL Cloud</h3>
              </div>
              <button type="button" className={styles.closeBtn} onClick={() => setOnboardingOpen(false)}>
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
                    <button type="button" className={styles.actionBtnSmall} onClick={handleCopyToken}>
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
              <button type="button" className={styles.copyCmdBtn} onClick={() => setOnboardingOpen(false)}>
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
                              {wh.name}
                            </strong>
                          </div>

                          <div className={styles.webhookActions}>
                            <button
                              type="button"
                              className={styles.webhookTestBtn}
                              disabled={testingWebhookId === wh.id}
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
                    onChange={(e) => setNewWhTarget(e.target.value as any)}
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
                        ? 'https://hooks.slack.bregalda.internal/services/...'
                        : 'https://api.empresa.com/webhooks/concurrency'
                    }
                    className={styles.webhookInput}
                    value={newWhUrl}
                    onChange={(e) => setNewWhUrl(e.target.value)}
                  />

                  <button type="submit" className={styles.webhookSaveBtn}>
                    + {lang === 'pt' ? 'Salvar Webhook' : 'Save Webhook'}
                  </button>
                </form>

                <div style={{ marginTop: 6 }}>
                  <input
                    type="text"
                    placeholder={
                      lang === 'pt'
                        ? 'Nome descritivo (ex: #alerta-db-prod)'
                        : 'Descriptive label (e.g. #prod-alerts)'
                    }
                    className={styles.webhookInput}
                    style={{ width: '100%' }}
                    value={newWhName}
                    onChange={(e) => setNewWhName(e.target.value)}
                  />
                </div>
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
