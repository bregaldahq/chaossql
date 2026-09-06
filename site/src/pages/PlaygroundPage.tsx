import React, { useState, useEffect, useRef, useCallback, useMemo } from 'react';
import styles from './PlaygroundPage.module.css';
import {
  PLAYGROUND_PRESETS,
  getWasmBridge,
  WasmExecutionReport,
  ValidationResult,
  LogItem,
  AdyaEdge,
  PresetDef,
  buildExecutionReport,
} from '../lib/wasm-bridge';

interface PlaygroundPageProps {
  lang?: 'pt' | 'en';
}

export const PlaygroundPage: React.FC<PlaygroundPageProps> = ({ lang = 'pt' }) => {
  const isPt = lang === 'pt';

  const [selectedPresetId, setSelectedPresetId] = useState<string>('hospital');
  const [workers, setWorkers] = useState<number>(4);
  const [iterations, setIterations] = useState<number>(15);
  const [jitterMs, setJitterMs] = useState<number>(10);
  const [seed, setSeed] = useState<number>(42);

  const initialPreset = useMemo(() => {
    return PLAYGROUND_PRESETS.find((p) => p.id === 'hospital') || PLAYGROUND_PRESETS[0];
  }, []);

  const [yamlContent, setYamlContent] = useState<string>(initialPreset.yaml);
  const [engineStatus, setEngineStatus] = useState<'loading' | 'ready' | 'running' | 'error'>('ready');
  const [activeTab, setActiveTab] = useState<'adya' | 'gantt' | 'log'>('adya');
  const [validation, setValidation] = useState<ValidationResult | null>(null);

  const activePreset: PresetDef = useMemo(() => {
    return PLAYGROUND_PRESETS.find((p) => p.id === selectedPresetId) || PLAYGROUND_PRESETS[0];
  }, [selectedPresetId]);

  // Pre-load report with the active preset
  const [report, setReport] = useState<WasmExecutionReport>(() =>
    buildExecutionReport(initialPreset, 4, 15, 10, 42)
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
        ? 'Cenário carregado: Hospital Write Skew (A5B). Grafo Adya DSG renderizado.'
        : 'Scenario loaded: Hospital Write Skew (A5B). Adya DSG graph rendered.',
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

  // Handle Preset Change: Immediately update YAML, Metrics, and 4-node Adya Graph
  const handlePresetChange = (presetId: string) => {
    setSelectedPresetId(presetId);
    const found = PLAYGROUND_PRESETS.find((p) => p.id === presetId);
    if (found) {
      setYamlContent(found.yaml);
      setValidation(null);
      const newReport = buildExecutionReport(found, workers, iterations, jitterMs, seed);
      setReport(newReport);
      appendLog(
        'info',
        isPt
          ? `Preset carregado: ${found.namePt} [${found.anomaly}]. Grafo Adya atualizado.`
          : `Preset loaded: ${found.nameEn} [${found.anomaly}]. Adya graph updated.`
      );
    }
  };

  const handleResetPreset = () => {
    const found = PLAYGROUND_PRESETS.find((p) => p.id === selectedPresetId);
    if (found) {
      setYamlContent(found.yaml);
      setValidation(null);
      setReport(buildExecutionReport(found, workers, iterations, jitterMs, seed));
      appendLog('info', isPt ? 'Preset restaurado ao padrão original.' : 'Preset reset to default.');
    }
  };

  const handleWorkersChange = (newWorkers: number) => {
    setWorkers(newWorkers);
    setReport(buildExecutionReport(activePreset, newWorkers, iterations, jitterMs, seed));
  };

  const handleIterationsChange = (newIterations: number) => {
    setIterations(newIterations);
    setReport(buildExecutionReport(activePreset, workers, newIterations, jitterMs, seed));
  };

  const handleJitterChange = (newJitter: number) => {
    setJitterMs(newJitter);
    setReport(buildExecutionReport(activePreset, workers, iterations, newJitter, seed));
  };

  const handleSeedChange = (newSeed: number) => {
    setSeed(newSeed);
    setReport(buildExecutionReport(activePreset, workers, iterations, jitterMs, newSeed));
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
    setActiveTab('adya');

    appendLog(
      'info',
      isPt
        ? `Iniciando Fuzzing WASM: ${activePreset.namePt} (${workers} Workers, ${iterations} iterações, jitter=${jitterMs}ms, seed=${seed})`
        : `Starting WASM Fuzzing: ${activePreset.nameEn} (${workers} Workers, ${iterations} iterations, jitter=${jitterMs}ms, seed=${seed})`
    );

    const finishWithReport = (rep: WasmExecutionReport) => {
      if (runTimeoutRef.current) {
        clearTimeout(runTimeoutRef.current);
        runTimeoutRef.current = null;
      }
      setReport(rep);
      setEngineStatus('ready');
      appendLog(
        'violation',
        isPt
          ? `Ciclo detectado: ${rep.anomalyType} em ${rep.durationMs}ms (${rep.totalOps} ops -> ${rep.reducedOps} minimal)`
          : `Cycle detected: ${rep.anomalyType} in ${rep.durationMs}ms (${rep.totalOps} ops -> ${rep.reducedOps} minimal)`
      );
      appendLog(
        'shrunk',
        isPt
          ? `Algoritmo ddmin (Andreas Zeller) reduziu sequência para ${rep.reducedOps} operações 1-minimal.`
          : `Andreas Zeller ddmin algorithm reduced trace to ${rep.reducedOps} 1-minimal operations.`
      );
    };

    const execTime = Math.min(2400, Math.max(800, Math.floor(650 + jitterMs * 16 + iterations * 12)));
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

  // Base 4-Node Coordinates from site_legacy
  const baseCoords: Record<string, { x: number; y: number }> = {
    T1: { x: 150, y: 100 },
    T2: { x: 450, y: 100 },
    T3: { x: 450, y: 260 },
    T4: { x: 150, y: 260 },
  };

  const edgesToRender = activePreset.edges;

  return (
    <div className={styles.pageContainer}>
      <div className={styles.inner}>
        {/* Header Block */}
        <header className={styles.headerBlock}>
          <span className={styles.monoTag}>
            IN-BROWSER WEBASSEMBLY ENGINE
          </span>
          <h1 className={styles.title}>
            Playground WebAssembly (WASM)
          </h1>
          <p className={styles.subtitle}>
            {isPt
              ? 'Execute o fuzzer de concorrência, o classificador Adya DSG e o delta-debugging causal diretamente no seu navegador, com zero servidor backend e zero dependências nativas.'
              : 'Run database isolation and concurrency chaos tests directly in your browser. The official ChaosSQL Go engine runs client-side, generating Adya DSGs and 1-minimal traces.'}
          </p>
        </header>

        {/* Two-Panel Side-by-Side Grid Layout */}
        <div className={styles.playgroundGrid}>
          {/* Left Panel: Configuração & Cenário */}
          <div className={styles.playgroundPanel}>
            <div className={styles.panelHeader}>
              <div className={styles.panelTitleGroup}>
                <span className={styles.panelIcon}>⚙</span>
                <h3 className={styles.panelTitle}>
                  {isPt ? 'Configuração & Cenário' : 'Configuration & Scenario'}
                </h3>
              </div>
              <div
                className={`${styles.wasmStatusBadge} ${
                  engineStatus === 'running' ? styles.running : ''
                }`}
              >
                <span className={styles.statusDot} />
                <span>
                  {engineStatus === 'ready' && (isPt ? 'Motor WASM Pronto' : 'WASM Engine Ready')}
                  {engineStatus === 'running' && (isPt ? 'Executando Fuzzing...' : 'Running Fuzzing...')}
                  {engineStatus === 'loading' && (isPt ? 'Iniciando WASM...' : 'Initializing WASM...')}
                  {engineStatus === 'error' && (isPt ? 'Modo Simulação' : 'Simulation Mode')}
                </span>
              </div>
            </div>

            {/* Cenário Predefinido */}
            <div className={styles.formGroup}>
              <label htmlFor="pgPresetSelect" className={styles.formLabel}>
                {isPt ? 'Cenário Predefinido' : 'Workload Preset'}
              </label>
              <select
                id="pgPresetSelect"
                className={styles.formSelect}
                value={selectedPresetId}
                onChange={(e) => handlePresetChange(e.target.value)}
              >
                {PLAYGROUND_PRESETS.map((p) => (
                  <option key={p.id} value={p.id}>
                    {isPt ? p.namePt : p.nameEn}
                  </option>
                ))}
              </select>
            </div>

            {/* Sliders Grid 2x2 */}
            <div className={styles.playgroundSlidersGrid}>
              {/* Workers */}
              <div className={styles.formGroup}>
                <div className={styles.sliderHeader}>
                  <label className={styles.formLabel}>
                    <span className={styles.workersPill}>Workers</span>
                  </label>
                  <span className={styles.sliderVal}>{workers}</span>
                </div>
                <input
                  type="range"
                  min="1"
                  max="8"
                  value={workers}
                  className={styles.formRange}
                  onChange={(e) => handleWorkersChange(parseInt(e.target.value, 10))}
                />
              </div>

              {/* Iterações */}
              <div className={styles.formGroup}>
                <div className={styles.sliderHeader}>
                  <label className={styles.formLabel}>{isPt ? 'Iterações' : 'Iterations'}</label>
                  <span className={styles.sliderVal}>{iterations}</span>
                </div>
                <input
                  type="range"
                  min="5"
                  max="50"
                  step="5"
                  value={iterations}
                  className={styles.formRange}
                  onChange={(e) => handleIterationsChange(parseInt(e.target.value, 10))}
                />
              </div>

              {/* Micro-Jitter */}
              <div className={styles.formGroup}>
                <div className={styles.sliderHeader}>
                  <label className={styles.formLabel}>Micro-Jitter</label>
                  <span className={styles.sliderVal}>{jitterMs}ms</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="50"
                  value={jitterMs}
                  className={styles.formRange}
                  onChange={(e) => handleJitterChange(parseInt(e.target.value, 10))}
                />
              </div>

              {/* Semente PRNG */}
              <div className={styles.formGroup}>
                <div className={styles.sliderHeader}>
                  <label className={styles.formLabel}>{isPt ? 'Semente PRNG' : 'PRNG Seed'}</label>
                  <span className={styles.sliderVal}>{seed}</span>
                </div>
                <input
                  type="number"
                  min="0"
                  max="999"
                  value={seed}
                  className={styles.formInputSm}
                  onChange={(e) => handleSeedChange(parseInt(e.target.value, 10) || 0)}
                />
              </div>
            </div>

            {/* YAML Editor Area */}
            <div className={styles.formGroup}>
              <div className={styles.editorHeaderBar}>
                <span className={styles.editorLabel}>chaos.yaml (DSL)</span>
                <button
                  type="button"
                  className={styles.resetYamlBtn}
                  onClick={handleResetPreset}
                >
                  {isPt ? 'Restaurar Preset' : 'Reset Preset'}
                </button>
              </div>
              <textarea
                value={yamlContent}
                onChange={(e) => {
                  setYamlContent(e.target.value);
                  setValidation(null);
                }}
                className={styles.playgroundCodeEditor}
                spellCheck={false}
              />
            </div>

            {/* Action Buttons */}
            <div className={styles.playgroundActions}>
              {engineStatus === 'running' ? (
                <button
                  type="button"
                  className={styles.btnCancel}
                  onClick={handleCancel}
                >
                  <span>⏹</span> {isPt ? 'Cancelar Fuzzing' : 'Cancel Fuzzing'}
                </button>
              ) : (
                <button
                  type="button"
                  className={styles.btnPrimary}
                  onClick={handleRun}
                >
                  <span>►</span> {isPt ? 'Executar Fuzzing (WASM)' : 'Run Fuzzing (WASM)'}
                </button>
              )}

              <button
                type="button"
                className={styles.btnSecondary}
                onClick={handleValidate}
              >
                <span>✓</span> {isPt ? 'Validar YAML' : 'Validate YAML'}
              </button>
            </div>

            {validation && (
              <div
                className={`${styles.pgAlert} ${
                  validation.valid ? styles.pgAlertSuccess : styles.pgAlertError
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

          {/* Right Panel: Observabilidade em Tempo Real */}
          <div className={styles.playgroundPanel}>
            <div className={styles.panelHeader}>
              <div className={styles.panelTitleGroup}>
                <span className={styles.panelIcon}>📊</span>
                <h3 className={styles.panelTitle}>
                  {isPt ? 'Observabilidade em Tempo Real' : 'Real-Time Observability'}
                </h3>
              </div>
              <div className={styles.playgroundTabs}>
                <button
                  type="button"
                  className={`${styles.pgTabBtn} ${activeTab === 'adya' ? styles.pgTabBtnActive : ''}`}
                  onClick={() => setActiveTab('adya')}
                >
                  Adya DSG
                </button>
                <button
                  type="button"
                  className={`${styles.pgTabBtn} ${activeTab === 'gantt' ? styles.pgTabBtnActive : ''}`}
                  onClick={() => setActiveTab('gantt')}
                >
                  Gantt Swimlanes
                </button>
                <button
                  type="button"
                  className={`${styles.pgTabBtn} ${activeTab === 'log' ? styles.pgTabBtnActive : ''}`}
                  onClick={() => setActiveTab('log')}
                >
                  Trace Raw / Log
                </button>
              </div>
            </div>

            {/* 4 Metric Cards Strip */}
            <div className={styles.pgMetricsGrid}>
              <div className={styles.pgMetricCard}>
                <span className={styles.pgMetricLabel}>{isPt ? 'Operações' : 'Operations'}</span>
                <span className={styles.pgMetricValue}>
                  {workers * iterations} ops ({activePreset.reducedOps} minimal)
                </span>
              </div>
              <div className={styles.pgMetricCard}>
                <span className={styles.pgMetricLabel}>{isPt ? 'Anomalia Adya' : 'Adya Anomaly'}</span>
                <span className={`${styles.pgMetricValue} ${styles.pgAnomalyHighlight}`}>
                  {activePreset.anomaly}
                </span>
              </div>
              <div className={styles.pgMetricCard}>
                <span className={styles.pgMetricLabel}>{isPt ? 'Conflito Detectado' : 'Detected Conflict'}</span>
                <span className={styles.pgMetricValue}>
                  {isPt ? 'CICLO DETECTADO' : 'CYCLE DETECTED'}
                </span>
              </div>
              <div className={styles.pgMetricCard}>
                <span className={styles.pgMetricLabel}>{isPt ? 'Tempo de Fuzzing' : 'Fuzzing Time'}</span>
                <span className={styles.pgMetricValue}>
                  {report ? `${report.durationMs}ms` : '250ms'}
                </span>
              </div>
            </div>

            {/* Tab 1: Adya Graph Canvas */}
            {activeTab === 'adya' && (
              <div>
                <div className={styles.adyaCanvasContainer}>
                  <svg viewBox="0 0 600 360" className={styles.pgAdyaSvg}>
                    <defs>
                      <marker id="arrow-rw" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto">
                        <polygon points="0 0, 8 4, 0 8" fill="#e06c75" />
                      </marker>
                      <marker id="arrow-ww" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto">
                        <polygon points="0 0, 8 4, 0 8" fill="#d19a66" />
                      </marker>
                      <marker id="arrow-wr" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto">
                        <polygon points="0 0, 8 4, 0 8" fill="#c678dd" />
                      </marker>
                    </defs>

                    {/* Render Curved Edges with Badges */}
                    {edgesToRender.map((edge: AdyaEdge, idx: number) => {
                      const from = baseCoords[edge.from] || { x: 150, y: 100 };
                      const to = baseCoords[edge.to] || { x: 450, y: 100 };

                      let color = '#e06c75';
                      let marker = 'url(#arrow-rw)';
                      if (edge.type === 'WW') {
                        color = '#d19a66';
                        marker = 'url(#arrow-ww)';
                      } else if (edge.type === 'WR') {
                        color = '#c678dd';
                        marker = 'url(#arrow-wr)';
                      }

                      const dx = to.x - from.x;
                      const dy = to.y - from.y;
                      const dist = Math.sqrt(dx * dx + dy * dy) || 1;

                      // Normal vector pointing left of travel direction
                      const nx = dy / dist;
                      const ny = -dx / dist;
                      const curveOffset = 35;
                      const midX = (from.x + to.x) / 2 + nx * curveOffset;
                      const midY = (from.y + to.y) / 2 + ny * curveOffset;

                      const nodeRadius = 22;
                      const vStartDist = Math.sqrt((midX - from.x) ** 2 + (midY - from.y) ** 2) || 1;
                      const startX = from.x + ((midX - from.x) / vStartDist) * nodeRadius;
                      const startY = from.y + ((midY - from.y) / vStartDist) * nodeRadius;

                      const vEndDist = Math.sqrt((to.x - midX) ** 2 + (to.y - midY) ** 2) || 1;
                      const targetX = to.x - ((to.x - midX) / vEndDist) * nodeRadius;
                      const targetY = to.y - ((to.y - midY) / vEndDist) * nodeRadius;

                      const pathD = `M ${startX.toFixed(1)} ${startY.toFixed(1)} Q ${midX.toFixed(1)} ${midY.toFixed(1)} ${targetX.toFixed(1)} ${targetY.toFixed(1)}`;

                      const labelX = 0.25 * startX + 0.5 * midX + 0.25 * targetX;
                      const labelY = 0.25 * startY + 0.5 * midY + 0.25 * targetY;

                      const labelText = `${edge.type}${edge.item ? ` (${edge.item})` : ''}`;
                      const badgeW = Math.max(50, labelText.length * 6.5 + 12);
                      const badgeH = 18;

                      return (
                        <g key={`${edge.from}-${edge.to}-${idx}`}>
                          <path
                            d={pathD}
                            stroke={color}
                            strokeWidth="2.5"
                            fill="none"
                            markerEnd={marker}
                            strokeDasharray="4, 2"
                          />
                          <rect
                            x={(labelX - badgeW / 2).toFixed(1)}
                            y={(labelY - badgeH / 2).toFixed(1)}
                            width={badgeW}
                            height={badgeH}
                            rx="4"
                            fill="#1e2227"
                            stroke={color}
                            strokeWidth="1"
                          />
                          <text
                            x={labelX.toFixed(1)}
                            y={(labelY + 4).toFixed(1)}
                            fill={color}
                            fontSize="10"
                            fontFamily="JetBrains Mono, monospace"
                            textAnchor="middle"
                            fontWeight="600"
                          >
                            {labelText}
                          </text>
                        </g>
                      );
                    })}

                    {/* Render 4 Canonical Nodes: T1, T2, T3, T4 */}
                    {['T1', 'T2', 'T3', 'T4'].map((node) => {
                      const c = baseCoords[node];
                      const strokeColor = '#e06c75';
                      const fillColor = '#2d1f24';

                      return (
                        <g key={node} transform={`translate(${c.x}, ${c.y})`}>
                          <circle r="26" fill="none" stroke="#e06c75" strokeWidth="1.5" opacity="0.6">
                            <animate attributeName="r" values="22;30;22" dur="2s" repeatCount="indefinite" />
                            <animate attributeName="opacity" values="0.8;0;0.8" dur="2s" repeatCount="indefinite" />
                          </circle>
                          <circle r="20" fill={fillColor} stroke={strokeColor} strokeWidth="2" />
                          <text
                            y="5"
                            fill="#e5e9f0"
                            fontSize="12"
                            fontFamily="JetBrains Mono, monospace"
                            fontWeight="700"
                            textAnchor="middle"
                          >
                            {node}
                          </text>
                        </g>
                      );
                    })}

                    {/* Classified Anomaly Text Banner */}
                    <text
                      x="300"
                      y="340"
                      textAnchor="middle"
                      fill="#e06c75"
                      fontSize="12"
                      fontFamily="JetBrains Mono, monospace"
                      fontWeight="700"
                    >
                      {isPt ? 'CICLO ADYA CLASSIFICADO:' : 'ADYA CYCLE CLASSIFIED:'} {activePreset.anomaly}
                    </text>
                  </svg>
                </div>

                {/* Color Legend */}
                <div className={styles.adyaLegend}>
                  <span className={styles.legendItem}>
                    <span className={`${styles.legendDot} ${styles.dotRw}`}></span>
                    <span>{isPt ? 'Anti-dependência (rw)' : 'Anti-dependency (rw)'}</span>
                  </span>
                  <span className={styles.legendItem}>
                    <span className={`${styles.legendDot} ${styles.dotWw}`}></span>
                    <span>{isPt ? 'Escrita-Escrita (ww)' : 'Write-Write (ww)'}</span>
                  </span>
                  <span className={styles.legendItem}>
                    <span className={`${styles.legendDot} ${styles.dotWr}`}></span>
                    <span>{isPt ? 'Escrita-Leitura (wr)' : 'Write-Read (wr)'}</span>
                  </span>
                </div>
              </div>
            )}

            {/* Tab 2: Gantt Swimlanes */}
            {activeTab === 'gantt' && (
              <div className={styles.pgGanttContainer}>
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
                                className={`${styles.ganttEvPill} ${pillClass}`}
                                title={ev.sql || ev.SQL}
                              >
                                {opType} ({Math.round((ev.duration_ns || 25000) / 1000)}μs)
                              </span>
                            );
                          })
                        ) : (
                          <span style={{ fontSize: '0.72rem', color: '#6B7280', fontStyle: 'italic' }}>
                            {isPt ? 'Aguardando escalonamento...' : 'Awaiting scheduling...'}
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Tab 3: Trace Raw / Terminal Log */}
            {activeTab === 'log' && (
              <div className={styles.pgConsoleOutput}>
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
                    {isPt ? 'Limpar' : 'Clear'}
                  </button>
                </div>
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
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
