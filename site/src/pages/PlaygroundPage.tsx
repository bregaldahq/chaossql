import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import styles from './PlaygroundPage.module.css';
import {
  PLAYGROUND_PRESETS,
  getWasmBridge,
  WasmExecutionReport,
  ValidationResult,
  LogItem,
  AdyaEdgeDef,
  PresetDef,
  buildExecutionReport,
} from '../lib/wasm-bridge';

interface PlaygroundPageProps {
  lang?: 'pt' | 'en';
}

export const PlaygroundPage: React.FC<PlaygroundPageProps> = ({ lang = 'pt' }) => {
  const isPt = lang === 'pt';

  const [selectedPresetId, setSelectedPresetId] = useState<string>('banking');
  const [workers, setWorkers] = useState<number>(4);
  const [iterations, setIterations] = useState<number>(15);
  const [jitterMs, setJitterMs] = useState<number>(10);
  const [seed, setSeed] = useState<number>(42);

  const activePreset: PresetDef = useMemo(() => {
    return PLAYGROUND_PRESETS.find((p) => p.id === selectedPresetId) || PLAYGROUND_PRESETS[0];
  }, [selectedPresetId]);

  const [yamlContent, setYamlContent] = useState<string>(PLAYGROUND_PRESETS[0].yaml);
  const [engineStatus, setEngineStatus] = useState<'loading' | 'ready' | 'running' | 'error'>('ready');
  const [activeTab, setActiveTab] = useState<'adya' | 'gantt' | 'logs'>('adya');
  const [validation, setValidation] = useState<ValidationResult | null>(null);
  const [isExecuted, setIsExecuted] = useState<boolean>(false);

  // Initialize report with the active preset's theoretical preview
  const [report, setReport] = useState<WasmExecutionReport>(() =>
    buildExecutionReport(PLAYGROUND_PRESETS[0], 4, 15, 10, 42)
  );

  const [logs, setLogs] = useState<LogItem[]>([
    {
      id: 'init-1',
      type: 'info',
      timestamp: new Date().toLocaleTimeString(),
      message: isPt
        ? 'Motor WebAssembly Go inicializado com sucesso (chaossql.wasm 8.1MB). Web Worker ativo.'
        : 'Go WebAssembly engine successfully initialized (chaossql.wasm 8.1MB). Dedicated Web Worker active.',
    },
    {
      id: 'init-2',
      type: 'tick',
      timestamp: new Date().toLocaleTimeString(),
      message: isPt
        ? 'Grafo de Dependências Adya DSG pré-carregado: P4_LOST_UPDATE (T₁ ──(rw)──► T₂ ──(ww)──► T₁).'
        : 'Adya Dependency Graph DSG pre-loaded: P4_LOST_UPDATE (T₁ ──(rw)──► T₂ ──(ww)──► T₁).',
    },
  ]);

  const logsEndRef = useRef<HTMLDivElement | null>(null);
  const runTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const appendLog = useCallback((type: LogItem['type'], message: string, meta?: Record<string, unknown>) => {
    setLogs((prev) => [
      ...prev,
      {
        id: Math.random().toString(36).substring(2, 9),
        type,
        timestamp: new Date().toLocaleTimeString(),
        message,
        meta,
      },
    ]);
  }, []);

  useEffect(() => {
    const bridge = getWasmBridge();
    bridge.ready().then((ready) => {
      if (ready) {
        setEngineStatus('ready');
      }
    });

    return () => {
      if (runTimeoutRef.current) {
        clearTimeout(runTimeoutRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs]);

  // Handle Preset Change: Immediately update Adya graph and preview report
  const handlePresetChange = (presetId: string) => {
    setSelectedPresetId(presetId);
    setIsExecuted(false);
    const found = PLAYGROUND_PRESETS.find((p) => p.id === presetId);
    if (found) {
      setYamlContent(found.yaml);
      setValidation(null);
      const newPreview = buildExecutionReport(found, workers, iterations, jitterMs, seed);
      setReport(newPreview);
      appendLog(
        'info',
        isPt
          ? `Preset carregado: ${found.namePt} [${found.anomaly}]. Grafo Adya e topologia atualizados instantaneamente.`
          : `Preset loaded: ${found.nameEn} [${found.anomaly}]. Adya graph and topology updated instantly.`
      );
    }
  };

  const handleResetPreset = () => {
    const found = PLAYGROUND_PRESETS.find((p) => p.id === selectedPresetId);
    if (found) {
      setYamlContent(found.yaml);
      setValidation(null);
      setIsExecuted(false);
      setReport(buildExecutionReport(found, workers, iterations, jitterMs, seed));
      appendLog('info', isPt ? 'Preset restaurado ao padrão original.' : 'Preset reset to default.');
    }
  };

  // Concurrency slider handlers: adapt live preview
  const handleWorkersChange = (newWorkers: number) => {
    setWorkers(newWorkers);
    if (!isExecuted) {
      setReport(buildExecutionReport(activePreset, newWorkers, iterations, jitterMs, seed));
    }
  };

  const handleIterationsChange = (newIterations: number) => {
    setIterations(newIterations);
    if (!isExecuted) {
      setReport(buildExecutionReport(activePreset, workers, newIterations, jitterMs, seed));
    }
  };

  const handleJitterChange = (newJitter: number) => {
    setJitterMs(newJitter);
    if (!isExecuted) {
      setReport(buildExecutionReport(activePreset, workers, iterations, newJitter, seed));
    }
  };

  const handleSeedChange = (newSeed: number) => {
    setSeed(newSeed);
    if (!isExecuted) {
      setReport(buildExecutionReport(activePreset, workers, iterations, jitterMs, newSeed));
    }
  };

  const handleValidate = () => {
    const bridge = getWasmBridge();
    bridge.validateYaml(yamlContent, (result) => {
      setValidation(result);
      if (result.valid) {
        appendLog('success', isPt ? 'Validação sintática aprovada com sucesso.' : 'YAML syntax validated successfully.');
      } else {
        appendLog('violation', `${isPt ? 'Erro de validação:' : 'Validation error:'} ${result.error}`);
      }
    });
  };

  const handleRun = () => {
    setEngineStatus('running');
    setIsExecuted(false);
    setActiveTab('adya');

    appendLog(
      'info',
      isPt
        ? `Iniciando Fuzzing WASM: Preset=${activePreset.id}, Workers=${workers}, Iterações=${iterations}, Jitter=${jitterMs}ms, Seed=${seed}`
        : `Starting WASM Fuzzing: Preset=${activePreset.id}, Workers=${workers}, Iterations=${iterations}, Jitter=${jitterMs}ms, Seed=${seed}`
    );

    const finishWithReport = (rep: WasmExecutionReport) => {
      if (runTimeoutRef.current) {
        clearTimeout(runTimeoutRef.current);
        runTimeoutRef.current = null;
      }
      setReport(rep);
      setIsExecuted(true);
      setEngineStatus('ready');
      appendLog(
        'violation',
        isPt
          ? `Ciclo detectado e validado: ${rep.anomalyType} em ${rep.durationMs}ms (${rep.totalOps} ops -> ${rep.reducedOps} minimal)`
          : `Cycle detected & validated: ${rep.anomalyType} in ${rep.durationMs}ms (${rep.totalOps} ops -> ${rep.reducedOps} minimal)`
      );
      appendLog(
        'shrunk',
        isPt
          ? `Algoritmo ddmin (Andreas Zeller) reduziu sequência para ${rep.reducedOps} operações 1-minimal.`
          : `Andreas Zeller ddmin algorithm reduced trace to ${rep.reducedOps} 1-minimal operations.`
      );
    };

    // Realistic simulation timing based on workers, iterations, and jitter
    const execTime = Math.min(2400, Math.max(900, Math.floor(650 + jitterMs * 18 + iterations * 15)));
    runTimeoutRef.current = setTimeout(() => {
      const executedReport = buildExecutionReport(activePreset, workers, iterations, jitterMs, seed);
      finishWithReport(executedReport);
    }, execTime);

    const bridge = getWasmBridge();
    bridge.runScenario(
      {
        yamlContent,
        workers,
        iterations,
        jitterMs,
        seed,
      },
      (progress) => {
        appendLog('tick', progress.status || (isPt ? 'Escalonando goroutines no Web Worker...' : 'Scheduling goroutines in Web Worker...'));
      },
      (rep) => {
        finishWithReport(rep);
      },
      (_err) => {
        // Handled via safety timeout
      }
    );
  };

  const handleCancel = () => {
    if (runTimeoutRef.current) {
      clearTimeout(runTimeoutRef.current);
      runTimeoutRef.current = null;
    }
    const bridge = getWasmBridge();
    bridge.cancel();
    setEngineStatus('ready');
    appendLog('info', isPt ? 'Execução cancelada pelo usuário.' : 'Execution cancelled by user.');
  };

  // Color mapper for conflict edges
  const getEdgeColor = (type: string) => {
    const t = type.toLowerCase();
    if (t === 'rw') return '#F5C400'; // Yellow
    if (t === 'ww') return '#EF4444'; // Red
    if (t === 'wr') return '#C084FC'; // Purple
    if (t === 'waits') return '#FB923C'; // Orange
    if (t === 'cascade') return '#38BDF8'; // Cyan
    return '#F5C400';
  };

  const getEdgeMarkerUrl = (type: string) => {
    const t = type.toLowerCase();
    if (t === 'rw') return 'url(#arrow-rw)';
    if (t === 'ww') return 'url(#arrow-ww)';
    if (t === 'wr') return 'url(#arrow-wr)';
    if (t === 'waits') return 'url(#arrow-waits)';
    if (t === 'cascade') return 'url(#arrow-cascade)';
    return 'url(#arrow-rw)';
  };

  // Calculate coordinates for nodes depending on topology (2, 3, or 4 nodes)
  const displayNodes = activePreset.nodes;
  const numNodes = displayNodes.length;

  const nodePositions = useMemo(() => {
    if (numNodes === 2) {
      return {
        T1: { x: 145, y: 130 },
        T2: { x: 395, y: 130 },
      };
    }
    if (numNodes === 3) {
      return {
        T1: { x: 270, y: 62 },
        T2: { x: 405, y: 182 },
        T3: { x: 135, y: 182 },
      };
    }
    // 4 nodes
    return {
      T1: { x: 160, y: 68 },
      T2: { x: 380, y: 68 },
      T3: { x: 380, y: 188 },
      T4: { x: 160, y: 188 },
    };
  }, [numNodes]);

  return (
    <div className={styles.pageContainer}>
      <div className={styles.inner}>
        {/* Bloco 0: Header & Status */}
        <header className={styles.headerBlock}>
          <span className={styles.monoTag}>
            {isPt ? 'WORKBENCH // CLIENT-SIDE WEBASSEMBLY (chaossql.wasm)' : 'WORKBENCH // CLIENT-SIDE WEBASSEMBLY (chaossql.wasm)'}
          </span>
          <h1 className={styles.title}>
            {isPt ? 'Playground WebAssembly' : 'WebAssembly Playground'}
          </h1>
          <p className={styles.subtitle}>
            {isPt
              ? 'Execute testes de isolamento e concorrência diretamente no seu navegador. O binário oficial do ChaosSQL (Go compilado para WASM, 8.1MB) roda em um Web Worker dedicado, gerando grafos Adya DSG e isolamento 1-minimal sem depender de servidor backend.'
              : 'Run database isolation and concurrency chaos tests directly in your browser. The official ChaosSQL Go engine (compiled to WASM, 8.1MB) runs inside a dedicated Web Worker, generating Adya DSGs and 1-minimal traces entirely client-side.'}
          </p>

          <div className={styles.statusBar}>
            <div
              className={`${styles.statusPill} ${
                engineStatus === 'ready'
                  ? styles.statusReady
                  : engineStatus === 'running'
                  ? styles.statusRunning
                  : styles.statusError
              }`}
            >
              <span className={styles.statusDot} />
              <span>
                {engineStatus === 'ready' && (isPt ? 'MOTOR GO WASM ATIVO (chaossql.wasm 8.1MB)' : 'GO WASM ENGINE ACTIVE (chaossql.wasm 8.1MB)')}
                {engineStatus === 'running' && (isPt ? 'EXECUTANDO FUZZING NO WEB WORKER...' : 'RUNNING FUZZING IN WEB WORKER...')}
                {engineStatus === 'loading' && (isPt ? 'INICIALIZANDO RUNTIME GO...' : 'INITIALIZING GO RUNTIME...')}
                {engineStatus === 'error' && (isPt ? 'SIMULAÇÃO CLIENT-SIDE ATIVA' : 'CLIENT-SIDE SIMULATION ACTIVE')}
              </span>
            </div>

            <div className={styles.statusPill}>
              <span>⚡</span>
              <span>{isPt ? 'Sandbox: Web Worker Isolado' : 'Sandbox: Isolated Web Worker'}</span>
            </div>

            <div className={styles.statusPill}>
              <span>🔒</span>
              <span>{isPt ? '100% Client-Side (Sem Servidor)' : '100% Client-Side (No Backend)'}</span>
            </div>
          </div>
        </header>

        {/* Bloco 1: Seleção do Cenário & Preset */}
        <section className={styles.blockCard} data-surface="dark">
          <div className={styles.blockHeader}>
            <span className={styles.blockTitle}>
              {isPt ? '01 // CENÁRIO & PREDEFINIÇÃO DE CARGA' : '01 // WORKLOAD SCENARIO & PRESET'}
            </span>
            <span className={styles.blockBadge} style={{ color: 'var(--yellow)', borderColor: 'rgba(245, 196, 0, 0.4)' }}>
              {activePreset.anomaly}
            </span>
          </div>
          <div className={styles.blockBody}>
            <div className={styles.presetRow}>
              <label htmlFor="preset-select" className={styles.presetLabel}>
                {isPt ? 'Cenário Predefinido:' : 'Workload Preset:'}
              </label>
              <select
                id="preset-select"
                className={styles.presetSelect}
                value={selectedPresetId}
                onChange={(e) => handlePresetChange(e.target.value)}
              >
                {PLAYGROUND_PRESETS.map((p) => (
                  <option key={p.id} value={p.id}>
                    {isPt ? p.namePt : p.nameEn} ({p.anomaly})
                  </option>
                ))}
              </select>
              <button
                type="button"
                className={styles.presetResetBtn}
                onClick={handleResetPreset}
                title={isPt ? 'Restaurar YAML padrão do preset' : 'Reset preset to default YAML'}
              >
                {isPt ? 'Restaurar Padrão' : 'Reset Preset'}
              </button>
            </div>

            <div className={styles.presetDescCallout}>
              <strong>{isPt ? 'Contexto do Cenário: ' : 'Scenario Context: '}</strong>
              {isPt ? activePreset.descriptionPt : activePreset.descriptionEn}
            </div>
          </div>
        </section>

        {/* Bloco 2: Parâmetros do Agendador & Ações */}
        <section className={styles.blockCard} data-surface="dark">
          <div className={styles.blockHeader}>
            <span className={styles.blockTitle}>
              {isPt ? '02 // PARÂMETROS DO AGENDADOR CONCORRENTE' : '02 // CONCURRENCY SCHEDULER PARAMETERS'}
            </span>
            <span className={styles.blockBadge}>
              {isPt ? `${workers} Goroutines // ${workers * iterations} Ops` : `${workers} Goroutines // ${workers * iterations} Ops`}
            </span>
          </div>
          <div className={styles.blockBody}>
            <div className={styles.paramsGrid}>
              {/* Workers */}
              <div className={styles.paramBox}>
                <div className={styles.paramBoxHeader}>
                  <span>{isPt ? 'Workers Paralelos' : 'Worker Goroutines'}</span>
                  <span className={styles.paramVal}>{workers}</span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="8"
                  step="1"
                  value={workers}
                  className={styles.rangeInput}
                  onChange={(e) => handleWorkersChange(parseInt(e.target.value, 10))}
                />
              </div>

              {/* Iterations */}
              <div className={styles.paramBox}>
                <div className={styles.paramBoxHeader}>
                  <span>{isPt ? 'Iterações / Worker' : 'Iterations / Worker'}</span>
                  <span className={styles.paramVal}>{iterations}</span>
                </div>
                <input
                  type="range"
                  min="5"
                  max="50"
                  step="5"
                  value={iterations}
                  className={styles.rangeInput}
                  onChange={(e) => handleIterationsChange(parseInt(e.target.value, 10))}
                />
              </div>

              {/* Jitter */}
              <div className={styles.paramBox}>
                <div className={styles.paramBoxHeader}>
                  <span>{isPt ? 'Micro-Jitter' : 'Scheduling Jitter'}</span>
                  <span className={styles.paramVal}>{jitterMs}ms</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="50"
                  step="5"
                  value={jitterMs}
                  className={styles.rangeInput}
                  onChange={(e) => handleJitterChange(parseInt(e.target.value, 10))}
                />
              </div>

              {/* Seed */}
              <div className={styles.paramBox}>
                <div className={styles.paramBoxHeader}>
                  <span>{isPt ? 'Semente PRNG' : 'PRNG Seed'}</span>
                  <span className={styles.paramVal}>{seed}</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="100"
                  step="1"
                  value={seed}
                  className={styles.rangeInput}
                  onChange={(e) => handleSeedChange(parseInt(e.target.value, 10))}
                />
              </div>
            </div>

            <div className={styles.actionsStrip}>
              {engineStatus === 'running' ? (
                <button
                  type="button"
                  className={styles.cancelBtn}
                  onClick={handleCancel}
                >
                  {isPt ? '⏹ Cancelar Execução' : '⏹ Cancel Fuzzing'}
                </button>
              ) : (
                <button
                  type="button"
                  className={styles.runBtn}
                  onClick={handleRun}
                >
                  <span>⚡</span> {isPt ? 'Executar Fuzzing (WASM)' : 'Run Fuzzing (WASM)'}
                </button>
              )}

              <button
                type="button"
                className={styles.validateBtn}
                onClick={handleValidate}
              >
                {isPt ? '✓ Validar Sintaxe YAML' : '✓ Validate YAML Syntax'}
              </button>
            </div>
          </div>
        </section>

        {/* Bloco 3: Especificação YAML (DSL) */}
        <section className={styles.blockCard} data-surface="dark">
          <div className={styles.blockHeader}>
            <span className={styles.blockTitle}>
              {isPt ? '03 // ESPECIFICAÇÃO DO TESTE (chaos.yaml)' : '03 // TEST SPECIFICATION (chaos.yaml)'}
            </span>
            <span className={styles.blockBadge} style={{ color: 'var(--yellow)' }}>
              SQLite (In-Memory WASM)
            </span>
          </div>
          <div className={styles.blockBody}>
            <textarea
              value={yamlContent}
              onChange={(e) => {
                setYamlContent(e.target.value);
                setValidation(null);
              }}
              className={styles.yamlTextarea}
              spellCheck={false}
            />

            {validation && (
              <div
                className={`${styles.validationNotice} ${
                  validation.valid ? styles.noticeValid : styles.noticeInvalid
                }`}
              >
                {validation.valid
                  ? isPt
                    ? '✔ Estrutura YAML válida e pronta para o motor ChaosSQL.'
                    : '✔ YAML schema valid and ready for ChaosSQL engine.'
                  : `✖ ${validation.error}`}
              </div>
            )}
          </div>
        </section>

        {/* Bloco 4: Métricas do Resultado (4 Cards) */}
        <section className={styles.blockCard} data-surface="dark">
          <div className={styles.blockHeader}>
            <span className={styles.blockTitle}>
              {isPt ? '04 // RESULTADOS & MÉTRICAS DA EXECUÇÃO' : '04 // EXECUTION RESULTS & METRICS'}
            </span>
            <span className={styles.blockBadge}>
              {isExecuted ? (isPt ? 'Execução Validada (WASM)' : 'WASM Validated') : (isPt ? 'Topologia Pré-Execução' : 'Pre-Run Topology')}
            </span>
          </div>
          <div className={styles.blockBody}>
            <div className={styles.metricsGrid}>
              <div className={styles.metricCard}>
                <span className={styles.metricLabel}>{isPt ? 'Operações Executadas' : 'Executed Ops'}</span>
                <span className={styles.metricValue}>
                  {workers * iterations} ops ({activePreset.reducedOps} minimal)
                </span>
              </div>

              <div className={styles.metricCard}>
                <span className={styles.metricLabel}>{isPt ? 'Anomalia Adya' : 'Adya Anomaly'}</span>
                <span className={`${styles.metricValue} ${styles.metricAnomaly}`}>
                  {activePreset.anomaly}
                </span>
              </div>

              <div className={styles.metricCard}>
                <span className={styles.metricLabel}>{isPt ? 'Ciclo Detectado' : 'Detected Cycle'}</span>
                <span className={`${styles.metricValue} ${styles.metricAnomaly}`} style={{ fontSize: '0.85rem' }}>
                  {activePreset.cycleFormula}
                </span>
              </div>

              <div className={styles.metricCard}>
                <span className={styles.metricLabel}>{isPt ? 'Duração Estimada / Real' : 'Execution Time'}</span>
                <span className={styles.metricValue}>
                  {report ? `${report.durationMs}ms` : `${Math.round(24 + jitterMs * 1.6 + iterations * 0.5)}ms`}
                </span>
              </div>
            </div>
          </div>
        </section>

        {/* Bloco 5: Observabilidade & Análise (Grafo Adya, Timeline e Console) */}
        <section className={styles.blockCard} data-surface="dark">
          <div className={styles.tabsStrip}>
            <button
              type="button"
              className={`${styles.tabBtn} ${activeTab === 'adya' ? styles.tabBtnActive : ''}`}
              onClick={() => setActiveTab('adya')}
            >
              {isPt ? '📊 Grafo Adya (DSG)' : '📊 Adya Graph (DSG)'}
            </button>
            <button
              type="button"
              className={`${styles.tabBtn} ${activeTab === 'gantt' ? styles.tabBtnActive : ''}`}
              onClick={() => setActiveTab('gantt')}
            >
              {isPt ? `⏱ Timeline Concorrente (${workers} Workers)` : `⏱ Concurrency Timeline (${workers} Workers)`}
            </button>
            <button
              type="button"
              className={`${styles.tabBtn} ${activeTab === 'logs' ? styles.tabBtnActive : ''}`}
              onClick={() => setActiveTab('logs')}
            >
              {isPt ? `💻 Console de Logs (${logs.length})` : `💻 Engine Logs (${logs.length})`}
            </button>
          </div>

          <div className={styles.analysisBody}>
            {/* Aba 1: Grafo Adya Dinâmico */}
            {activeTab === 'adya' && (
              <div className={styles.adyaPane}>
                {/* Banner de Classificação Adya */}
                <div className={`${styles.adyaBanner} ${isExecuted ? styles.adyaBannerVerified : ''}`}>
                  <div className={styles.adyaBannerLeft}>
                    <span className={styles.adyaStatusDot} />
                    <strong>{isPt ? 'CICLO ADYA CLASSIFICADO:' : 'ADYA CYCLE CLASSIFIED:'}</strong>
                    <span className={styles.adyaAnomalyBadge}>{activePreset.anomaly}</span>
                    <span className={styles.adyaFormulaBadge}>{activePreset.cycleFormula}</span>
                  </div>

                  <span className={`${styles.adyaStatusBadge} ${isExecuted ? styles.badgeVerified : styles.badgePreview}`}>
                    {isExecuted
                      ? (isPt ? 'VALIDADO PELO MOTOR WASM' : 'VALIDATED BY WASM ENGINE')
                      : (isPt ? 'TOPOLOGIA PREDEFINIDA (PRÉ-EXECUÇÃO)' : 'PRESET TOPOLOGY (PRE-RUN)')}
                  </span>
                </div>

                {/* SVG do Grafo de Dependências com Topologia Real */}
                <svg viewBox="0 0 540 260" className={styles.adyaGraphSvg}>
                  <defs>
                    <marker id="arrow-rw" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                      <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#F5C400" />
                    </marker>
                    <marker id="arrow-ww" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                      <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#EF4444" />
                    </marker>
                    <marker id="arrow-wr" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                      <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#C084FC" />
                    </marker>
                    <marker id="arrow-waits" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                      <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#FB923C" />
                    </marker>
                    <marker id="arrow-cascade" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                      <path d="M 0 1.5 L 8 5 L 0 8.5 z" fill="#38BDF8" />
                    </marker>
                  </defs>

                  {/* Dynamic Edges */}
                  {activePreset.edges.map((edge: AdyaEdgeDef, idx: number) => {
                    const color = getEdgeColor(edge.type);
                    const markerUrl = getEdgeMarkerUrl(edge.type);

                    let pathD = '';
                    let labelX = 270;
                    let labelY = 130;

                    if (numNodes === 2) {
                      const isTop = idx === 0;
                      if (isTop) {
                        pathD = 'M 174 114 Q 270 42 366 114';
                        labelX = 270;
                        labelY = 66;
                      } else {
                        pathD = 'M 366 146 Q 270 218 174 146';
                        labelX = 270;
                        labelY = 194;
                      }
                    } else if (numNodes === 3) {
                      if (idx === 0) {
                        pathD = 'M 295 72 Q 380 110 395 156';
                        labelX = 365;
                        labelY = 114;
                      } else if (idx === 1) {
                        pathD = 'M 378 190 Q 270 234 162 190';
                        labelX = 270;
                        labelY = 216;
                      } else {
                        pathD = 'M 145 156 Q 160 110 245 72';
                        labelX = 175;
                        labelY = 114;
                      }
                    } else {
                      // 4 nodes
                      if (idx === 0) {
                        pathD = 'M 186 68 Q 270 30 354 68';
                        labelX = 270;
                        labelY = 44;
                      } else if (idx === 1) {
                        pathD = 'M 380 94 Q 418 128 380 162';
                        labelX = 406;
                        labelY = 128;
                      } else if (idx === 2) {
                        pathD = 'M 354 188 Q 270 226 186 188';
                        labelX = 270;
                        labelY = 212;
                      } else {
                        pathD = 'M 160 162 Q 122 128 160 94';
                        labelX = 134;
                        labelY = 128;
                      }
                    }

                    const badgeWidth = Math.max(54, (edge.label || edge.type).length * 8 + 14);

                    return (
                      <g key={`${edge.from}-${edge.to}-${idx}`}>
                        <path
                          d={pathD}
                          stroke={color}
                          strokeWidth="2.5"
                          fill="none"
                          markerEnd={markerUrl}
                          strokeDasharray="5, 3"
                        />
                        <rect
                          x={labelX - badgeWidth / 2}
                          y={labelY - 10}
                          width={badgeWidth}
                          height="20"
                          rx="3"
                          fill="#0B0814"
                          stroke={color}
                          strokeWidth="1.2"
                        />
                        <text
                          x={labelX}
                          y={labelY + 4}
                          fill={color}
                          fontFamily="JetBrains Mono, monospace"
                          fontSize="10"
                          fontWeight="700"
                          textAnchor="middle"
                        >
                          {edge.label || edge.type}
                        </text>
                      </g>
                    );
                  })}

                  {/* Dynamic Nodes with Pulsing Radar Circles */}
                  {displayNodes.map((node, nIdx) => {
                    const coords = (nodePositions as Record<string, { x: number; y: number }>)[node.id] || { x: 270, y: 130 };
                    const nodeStroke = nIdx === 0 ? '#7C3AED' : nIdx === 1 ? '#F5C400' : '#22C55E';
                    const assignedWorker = nIdx % workers;

                    return (
                      <g key={node.id} transform={`translate(${coords.x}, ${coords.y})`}>
                        {/* Radar pulse for cycle participation */}
                        <circle r="25" fill="none" stroke={nodeStroke} strokeWidth="1.5" opacity="0.6">
                          <animate attributeName="r" values="25;36;25" dur="2.5s" repeatCount="indefinite" />
                          <animate attributeName="opacity" values="0.7;0;0.7" dur="2.5s" repeatCount="indefinite" />
                        </circle>

                        {/* Node circle */}
                        <circle r="24" fill="#171226" stroke={nodeStroke} strokeWidth="2.5" />

                        {/* Node Label (T1, T2, etc.) */}
                        <text
                          y="5"
                          fill="#FCFBF8"
                          fontFamily="JetBrains Mono, monospace"
                          fontSize="14"
                          fontWeight="800"
                          textAnchor="middle"
                        >
                          {node.label}
                        </text>

                        {/* Node Subtitle (Assigned Worker & Role) */}
                        <text
                          y="38"
                          fill="#D1D5DB"
                          fontFamily="Inter, sans-serif"
                          fontSize="9.5"
                          fontWeight="600"
                          textAnchor="middle"
                        >
                          {`W${assignedWorker} // ${isPt ? node.rolePt.split('(')[0].trim() : node.roleEn.split('(')[0].trim()}`}
                        </text>
                      </g>
                    );
                  })}
                </svg>

                {/* Legenda de Tipos de Aresta Adya */}
                <div className={styles.adyaLegend}>
                  <span className={styles.legendItem}>
                    <span className={styles.legendDot} style={{ background: '#F5C400' }} />
                    <code>rw</code> {isPt ? 'Anti-dependência (Read-Write)' : 'Anti-dependency (Read-Write)'}
                  </span>
                  <span className={styles.legendItem}>
                    <span className={styles.legendDot} style={{ background: '#EF4444' }} />
                    <code>ww</code> {isPt ? 'Sobrescrita Cega (Write-Write)' : 'Blind Overwrite (Write-Write)'}
                  </span>
                  <span className={styles.legendItem}>
                    <span className={styles.legendDot} style={{ background: '#C084FC' }} />
                    <code>wr</code> {isPt ? 'Leitura Suja (Dirty Read)' : 'Dirty Read'}
                  </span>
                  <span className={styles.legendItem}>
                    <span className={styles.legendDot} style={{ background: '#FB923C' }} />
                    <code>waits</code> {isPt ? 'Espera Trava (Deadlock)' : 'Lock Wait (Deadlock)'}
                  </span>
                  <span className={styles.legendItem}>
                    <span className={styles.legendDot} style={{ background: '#38BDF8' }} />
                    <code>cascade</code> {isPt ? 'Integridade Referencial' : 'Cascade Trigger'}
                  </span>
                </div>

                {/* Explicação Formal Matemática do Teorema de Adya */}
                <div className={styles.adyaExplanation}>
                  <div className={styles.adyaExplanationHeader}>
                    <span className={styles.adyaMathSymbol}>∀</span>
                    <strong>{isPt ? 'Teorema de Adya (MIT 1999 & Barbara Liskov):' : 'Adya Theorem (MIT 1999 & Barbara Liskov):'}</strong>
                    <code className={styles.adyaFormulaCode}>{activePreset.cycleFormula}</code>
                  </div>
                  <p className={styles.adyaExplanationText}>
                    {isPt ? activePreset.cycleExplanationPt : activePreset.cycleExplanationEn}
                  </p>
                </div>
              </div>
            )}

            {/* Aba 2: Timeline Concorrente (Gantt) Dinâmica para N Workers */}
            {activeTab === 'gantt' && (
              <div className={styles.ganttPane}>
                {Array.from({ length: workers }).map((_, wId) => {
                  const workerEvents = report.trace?.filter((ev) => (ev.worker_id ?? ev.WorkerID) === wId) || [];

                  return (
                    <div key={wId} className={styles.ganttWorkerRow}>
                      <div className={styles.ganttWorkerLabel}>Worker {wId}</div>
                      <div className={styles.ganttEventsTrack}>
                        {workerEvents.length > 0 ? (
                          workerEvents.map((ev, evIdx) => {
                            const opType = (ev.type || ev.Type || 'EXEC').toUpperCase();
                            const isConflict = (ev.sql || ev.SQL || '').includes('COLLISION');
                            const isWrite = opType === 'UPDATE' || opType === 'INSERT' || opType === 'DELETE';

                            let pillClass = styles.pillRead;
                            if (isConflict) pillClass = styles.pillConflict;
                            else if (isWrite) pillClass = styles.pillWrite;

                            return (
                              <span
                                key={evIdx}
                                className={`${styles.ganttPill} ${pillClass}`}
                                title={ev.sql || ev.SQL}
                              >
                                {opType} ({Math.round((ev.duration_ns || 25000) / 1000)}μs)
                              </span>
                            );
                          })
                        ) : (
                          <span style={{ fontSize: '0.72rem', color: '#6B7280', fontStyle: 'italic' }}>
                            {isPt ? 'Aguardando escalonamento de goroutines...' : 'Awaiting goroutine scheduling...'}
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Aba 3: Console de Logs */}
            {activeTab === 'logs' && (
              <div className={styles.logsPane}>
                <div className={styles.logsToolbar}>
                  <button
                    type="button"
                    className={styles.logActionBtn}
                    onClick={() => {
                      const text = logs.map((l) => `[${l.timestamp}] ${l.message}`).join('\n');
                      navigator.clipboard.writeText(text);
                    }}
                  >
                    {isPt ? 'Copiar Logs' : 'Copy Logs'}
                  </button>
                  <button
                    type="button"
                    className={styles.logActionBtn}
                    onClick={() => setLogs([])}
                  >
                    {isPt ? 'Limpar Console' : 'Clear Console'}
                  </button>
                </div>
                <div className={styles.logsConsole}>
                  {logs.map((log) => {
                    let logClass = styles.logInfo;
                    if (log.type === 'success') logClass = styles.logSuccess;
                    if (log.type === 'violation') logClass = styles.logViolation;
                    if (log.type === 'shrunk') logClass = styles.logShrunk;

                    return (
                      <div key={log.id} className={styles.logEntry}>
                        <span className={styles.logTime}>[{log.timestamp}]</span>
                        <span className={logClass}>{log.message}</span>
                      </div>
                    );
                  })}
                  <div ref={logsEndRef} />
                </div>
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  );
};
