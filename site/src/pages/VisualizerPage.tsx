import React, { useState, useMemo } from 'react';
import styles from './VisualizerPage.module.css';
import { RAW_TRACE_OPS, SHRUNK_TRACE_OPS, TraceOp } from '../lib/wasm-bridge';
import { CodeBlock } from '../components/docs/CodeBlock';

interface VisualizerPageProps {
  lang?: 'pt' | 'en';
}

export const VisualizerPage: React.FC<VisualizerPageProps> = ({ lang = 'pt' }) => {
  const [mode, setMode] = useState<'raw' | 'shrunk'>('raw');
  const [selectedWorker, setSelectedWorker] = useState<string>('all');
  const [activeOpId, setActiveOpId] = useState<string>('op_13');
  const [animatingIndex, setAnimatingIndex] = useState<number | null>(null);

  const isPt = lang === 'pt';

  const opsList = mode === 'raw' ? RAW_TRACE_OPS : SHRUNK_TRACE_OPS;

  const currentOp = useMemo(() => {
    const found = opsList.find((op) => op.id === activeOpId);
    if (found) return found;
    return mode === 'raw' ? RAW_TRACE_OPS[13] : SHRUNK_TRACE_OPS[3];
  }, [opsList, activeOpId, mode]);

  const maxTime = 250; // μs

  const filteredOps = useMemo(() => {
    if (selectedWorker === 'all') return opsList;
    return opsList.filter((op) => op.worker.toString() === selectedWorker);
  }, [opsList, selectedWorker]);

  const handleModeChange = (newMode: 'raw' | 'shrunk') => {
    setMode(newMode);
    if (newMode === 'raw') {
      setActiveOpId('op_13');
    } else {
      setActiveOpId('op_s3');
    }
  };

  const handleSelectOp = (op: TraceOp) => {
    setActiveOpId(op.id);
  };

  const handleAnimate = () => {
    let idx = 0;
    const interval = setInterval(() => {
      if (idx >= opsList.length) {
        clearInterval(interval);
        setAnimatingIndex(null);
        return;
      }
      setAnimatingIndex(idx);
      setActiveOpId(opsList[idx].id);
      idx++;
    }, 280);
  };

  // Collision marker position
  const collisionUs = mode === 'raw' ? 184 : 155;
  const collisionPct = (collisionUs / maxTime) * 100;

  const workers = [0, 1, 2, 3];

  return (
    <div className={styles.pageContainer}>
      <div className={styles.inner}>
        {/* Header */}
        <header className={styles.header}>
          <span className={styles.monoTag}>
            {isPt ? 'TERMINAL // CONCURRENCY TRACE (chaossql ui)' : 'TERMINAL // CONCURRENCY TRACE (chaossql ui)'}
          </span>
          <h1 className={styles.title}>
            {isPt ? 'Trace Visualizer' : 'Trace Visualizer'}
          </h1>
          <p className={styles.subtitle}>
            {isPt
              ? 'Simulação interativa da suíte chaossql ui. Inspecione o entrelaçamento de goroutines em escala de microssegundos, investigue o Grafo de Dependências Adya (DSG) e compare o trace original de 20 operações com a síntese 1-minimal isolada por delta-debugging (ddmin).'
              : 'Interactive high-fidelity simulation of the chaossql ui suite. Inspect microsecond-scale goroutine interleaving, explore the cyclic Adya Dependency Graph (DSG), and compare raw 20-operation executions with 1-minimal ddmin shrunk traces.'}
          </p>
        </header>

        {/* Controls Bar */}
        <div className={styles.controlsBar}>
          <div className={styles.modeSegmentedGroup}>
            <button
              type="button"
              className={`${styles.modeBtn} ${mode === 'raw' ? styles.modeBtnActive : ''}`}
              onClick={() => handleModeChange('raw')}
            >
              {isPt ? 'Raw Trace (20 ops)' : 'Raw Trace (20 ops)'}
            </button>
            <button
              type="button"
              className={`${styles.modeBtn} ${mode === 'shrunk' ? styles.modeBtnActive : ''}`}
              onClick={() => handleModeChange('shrunk')}
            >
              {isPt ? '1-Minimal Shrunk (4 ops / ddmin)' : '1-Minimal Shrunk (4 ops / ddmin)'}
            </button>
          </div>

          <div className={styles.filterGroup}>
            <span className={styles.filterLabel}>
              {isPt ? 'Filtrar Worker:' : 'Filter:'}
            </span>
            <button
              type="button"
              className={`${styles.filterBtn} ${selectedWorker === 'all' ? styles.filterBtnActive : ''}`}
              onClick={() => setSelectedWorker('all')}
            >
              {isPt ? 'Todos' : 'All'}
            </button>
            {workers.map((w) => (
              <button
                key={w}
                type="button"
                className={`${styles.filterBtn} ${selectedWorker === w.toString() ? styles.filterBtnActive : ''}`}
                onClick={() => setSelectedWorker(w.toString())}
              >
                W{w}
              </button>
            ))}
          </div>

          <button
            type="button"
            className={styles.animateBtn}
            onClick={handleAnimate}
            title={isPt ? 'Reproduzir sequência de execução' : 'Replay execution sequence'}
          >
            <span>▶</span> {isPt ? 'Simular Replay' : 'Replay Trace'}
          </button>

          <div className={styles.statusPill}>
            <span>⚡</span>
            <span>
              {isPt
                ? `P4_LOST_UPDATE detectado em t=${collisionUs}μs`
                : `P4_LOST_UPDATE detected at t=${collisionUs}μs`}
            </span>
          </div>
        </div>

        {/* Gantt Chart Terminal */}
        <div className={styles.ganttCard} data-surface="dark">
          <div className={styles.ganttHeader}>
            <span className={styles.ganttTitle}>
              {isPt ? 'Linha do Tempo de Concorrência (Gantt μs)' : 'Concurrency Timeline (Gantt μs)'}
            </span>
            <div className={styles.ganttLegend}>
              <div className={styles.legendItem}>
                <span className={`${styles.legendColor} ${styles.legendRead}`} />
                <span>{isPt ? 'Leitura (Read)' : 'Read'}</span>
              </div>
              <div className={styles.legendItem}>
                <span className={`${styles.legendColor} ${styles.legendWrite}`} />
                <span>{isPt ? 'Escrita / Commit' : 'Write / Commit'}</span>
              </div>
              <div className={styles.legendItem}>
                <span className={`${styles.legendColor} ${styles.legendConflict}`} />
                <span>{isPt ? 'Colisão / Invariante' : 'Collision / Invariant'}</span>
              </div>
            </div>
          </div>

          <div className={styles.ganttInner}>
            {/* Axis */}
            <div className={styles.ganttAxis}>
              <div className={styles.axisTick}>0μs</div>
              <div className={styles.axisTick}>50μs</div>
              <div className={styles.axisTick}>100μs</div>
              <div className={styles.axisTick}>150μs</div>
              <div className={styles.axisTick}>200μs</div>
              <div className={styles.axisTick}>250μs</div>
            </div>

            {/* Collision Marker Line */}
            <div
              className={styles.collisionMarker}
              style={{ left: `calc(90px + (100% - 90px) * (${collisionPct} / 100))` }}
            >
              <div className={styles.collisionLabel}>
                {isPt ? `P4 Colisão (${collisionUs}μs)` : `P4 Collision (${collisionUs}μs)`}
              </div>
            </div>

            {/* Worker Lanes */}
            <div className={styles.ganttLanes}>
              {workers.map((w) => {
                if (selectedWorker !== 'all' && selectedWorker !== w.toString()) {
                  return null;
                }
                const workerOps = filteredOps.filter((op) => op.worker === w);

                return (
                  <div key={w} className={styles.laneRow}>
                    <div className={styles.laneLabel}>
                      Worker {w}
                    </div>
                    <div className={styles.laneTrack}>
                      {workerOps.map((op) => {
                        const leftPct = (op.startUs / maxTime) * 100;
                        const widthPct = Math.max((op.durationUs / maxTime) * 100, 7.5);
                        const isSelected = op.id === activeOpId;
                        const isPulsing = animatingIndex !== null && opsList[animatingIndex]?.id === op.id;

                        let blockClass = styles.opRead;
                        if (op.type === 'write') blockClass = styles.opWrite;
                        if (op.type === 'conflict') blockClass = styles.opConflict;

                        return (
                          <div
                            key={op.id}
                            className={`${styles.ganttBlock} ${blockClass} ${isSelected ? styles.ganttBlockActive : ''}`}
                            style={{
                              left: `${leftPct}%`,
                              width: `${widthPct}%`,
                              transform: isPulsing ? 'scale(1.12)' : undefined,
                            }}
                            onClick={() => handleSelectOp(op)}
                            title={`${op.tx}: ${op.name}`}
                          >
                            {op.tx}: {op.type.toUpperCase()} ({op.durationUs}μs)
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Detail Split Grid: Adya Graph & Query Inspector */}
        <div className={styles.splitGrid}>
          {/* Adya Dependency Graph (DSG) */}
          <div className={styles.detailCard} data-surface="dark">
            <div className={styles.cardToolbar}>
              <span className={styles.cardTitle}>
                {isPt ? 'Grafo de Dependências Adya (DSG)' : 'Adya Dependency Graph (DSG)'}
              </span>
              <span className={`${styles.cardBadge} ${styles.badgeConflict}`}>
                {isPt ? 'Ciclo Anômalo: rw ∘ ww' : 'Anomaly Cycle: rw ∘ ww'}
              </span>
            </div>

            <div className={styles.adyaWrapper}>
              <svg viewBox="0 0 420 190" className={styles.adyaSvg}>
                <defs>
                  <marker
                    id="viz-arrow-yellow"
                    viewBox="0 0 10 10"
                    refX="7"
                    refY="5"
                    markerWidth="7"
                    markerHeight="7"
                    orient="auto-start-reverse"
                  >
                    <path d="M 0 1 L 10 5 L 0 9 z" fill="#F5C400" />
                  </marker>
                  <marker
                    id="viz-arrow-red"
                    viewBox="0 0 10 10"
                    refX="7"
                    refY="5"
                    markerWidth="7"
                    markerHeight="7"
                    orient="auto-start-reverse"
                  >
                    <path d="M 0 1 L 10 5 L 0 9 z" fill="#EF4444" />
                  </marker>
                </defs>

                {/* Path T1 -> T2 (rw anti-dependency) */}
                <g
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    const op = opsList.find((o) => o.tx === 'T1' && o.type === 'read') || opsList[0];
                    setActiveOpId(op.id);
                  }}
                >
                  <path
                    d="M 110 80 C 160 20, 260 20, 310 80"
                    stroke="#F5C400"
                    strokeWidth="2.5"
                    fill="none"
                    markerEnd="url(#viz-arrow-yellow)"
                    strokeDasharray="5, 3"
                  />
                  <rect x="185" y="24" width="50" height="22" rx="3" fill="#0D0A17" stroke="#F5C400" strokeWidth="1.2" />
                  <text x="210" y="39" fill="#F5C400" fontFamily="JetBrains Mono" fontSize="11" fontWeight="700" textAnchor="middle">
                    rw
                  </text>
                </g>

                {/* Path T2 -> T1 (ww write-write conflict) */}
                <g
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    const op = opsList.find((o) => o.type === 'conflict') || opsList[opsList.length - 1];
                    setActiveOpId(op.id);
                  }}
                >
                  <path
                    d="M 310 110 C 260 170, 160 170, 110 110"
                    stroke="#EF4444"
                    strokeWidth="2.5"
                    fill="none"
                    markerEnd="url(#viz-arrow-red)"
                    strokeDasharray="5, 3"
                  />
                  <rect x="185" y="142" width="50" height="22" rx="3" fill="#0D0A17" stroke="#EF4444" strokeWidth="1.2" />
                  <text x="210" y="157" fill="#EF4444" fontFamily="JetBrains Mono" fontSize="11" fontWeight="700" textAnchor="middle">
                    ww
                  </text>
                </g>

                {/* Node T1 */}
                <g
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    const op = opsList.find((o) => o.tx === 'T1' && o.type === 'write') || opsList[0];
                    setActiveOpId(op.id);
                  }}
                >
                  <circle cx="90" cy="95" r="30" fill="#1F1934" stroke="#7C3AED" strokeWidth="2.5" />
                  <text x="90" y="101" fill="#FCFBF8" fontFamily="Inter" fontSize="16" fontWeight="700" textAnchor="middle">
                    T₁
                  </text>
                </g>

                {/* Node T2 */}
                <g
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    const op = opsList.find((o) => o.tx === 'T2') || opsList[1];
                    setActiveOpId(op.id);
                  }}
                >
                  <circle cx="330" cy="95" r="30" fill="#1F1934" stroke="#F5C400" strokeWidth="2.5" />
                  <text x="330" y="101" fill="#FCFBF8" fontFamily="Inter" fontSize="16" fontWeight="700" textAnchor="middle">
                    T₂
                  </text>
                </g>
              </svg>

              <div className={styles.adyaFootnote}>
                <strong>{isPt ? 'Teorema de Adya (MIT 1999):' : 'Adya Theorem (MIT 1999):'}</strong>{' '}
                {isPt
                  ? 'A presença do ciclo direcionado T1 ──(rw)──► T2 ──(ww)──► T1 prova matematicamente a quebra de serializabilidade (Lost Update P4). Clique nos nós ou arestas para sincronizar com o inspetor.'
                  : 'The directed cycle T1 ──(rw)──► T2 ──(ww)──► T1 formally proves violation of serializability (P4 Lost Update). Click nodes or edges to sync with the inspector.'}
              </div>
            </div>
          </div>

          {/* Operation & Query Inspector */}
          <div className={styles.detailCard} data-surface="dark">
            <div className={styles.cardToolbar}>
              <span className={styles.cardTitle}>
                {isPt ? 'Inspetor de Operações & Queries' : 'Operation & Query Inspector'}
              </span>
              <span
                className={`${styles.cardBadge} ${
                  currentOp.type === 'conflict' ? styles.badgeConflict : styles.badgeSuccess
                }`}
              >
                {currentOp.status}
              </span>
            </div>

            <div className={styles.inspectorGrid}>
              <span className={styles.inspectorKey}>
                {isPt ? 'Transação / Worker:' : 'Transaction / Worker:'}
              </span>
              <span className={styles.inspectorVal}>
                <strong>{currentOp.tx}</strong> (Goroutine Worker {currentOp.worker})
              </span>

              <span className={styles.inspectorKey}>
                {isPt ? 'Timestamp / Duração:' : 'Timestamp / Duration:'}
              </span>
              <span className={styles.inspectorVal}>
                {currentOp.startUs}μs (+{currentOp.durationUs}μs {isPt ? 'de execução' : 'latency'})
              </span>

              <span className={styles.inspectorKey}>
                {isPt ? 'Variáveis / Estado:' : 'Variables / State:'}
              </span>
              <span className={`${styles.inspectorVal} ${styles.inspectorValHighlight}`}>
                {currentOp.vars}
              </span>

              <span className={styles.inspectorKey}>
                {isPt ? 'Grafo de Conflito:' : 'Conflict Graph:'}
              </span>
              <span
                className={`${styles.inspectorVal} ${
                  currentOp.type === 'conflict' ? styles.inspectorValConflict : ''
                }`}
              >
                {currentOp.type === 'conflict'
                  ? 'T1 ──(rw)──► T2 ──(ww)──► T1 [CICLO ANÔMALO]'
                  : isPt
                  ? 'Passo serializável (Sem conflito direto)'
                  : 'Serializable step (No direct conflict)'}
              </span>
            </div>

            <div className={styles.sqlBox}>
              <CodeBlock code={currentOp.name} language="sql" filename={`op_${currentOp.id}.sql`} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
