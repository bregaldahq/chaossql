// site/assets/wasm-bench.js
// Client-side In-Browser Web Worker & WebAssembly Concurrency Stress Benchmark
// Dispatches consecutive scenario runs to the active Web Worker, measures frame timing via
// requestAnimationFrame, and computes ops/sec throughput and 60 FPS UI frame budget compliance.

(function (global) {
  "use strict";

  const BENCH_PRESETS = [
    {
      key: "banking",
      name: "banking_lost_update",
      yaml: `version: "1.0"
name: "banking_lost_update"
database:
  driver: "sqlite"
  schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);"
  seed: "INSERT INTO accounts VALUES (1, 1000);"
invariants:
  - name: "total_balance"
    query: "SELECT sum(balance) AS total FROM accounts;"
    assert: "total == 1000"
operations:
  - name: "withdraw_100"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1"
        capture: "cur_bal"
      - sql: "UPDATE accounts SET balance = {cur_bal - 100} WHERE id = 1"`
    },
    {
      key: "inventory",
      name: "inventory_oversell",
      yaml: `version: "1.0"
name: "inventory_oversell"
database:
  driver: "sqlite"
  schema: "CREATE TABLE inventory (item_id INT PRIMARY KEY, stock INT NOT NULL);"
  seed: "INSERT INTO inventory VALUES (1, 10);"
invariants:
  - name: "no_negative_stock"
    query: "SELECT stock FROM inventory WHERE item_id = 1;"
    assert: "stock >= 0"
operations:
  - name: "buy_item"
    steps:
      - sql: "SELECT stock FROM inventory WHERE item_id = 1"
        capture: "cur"
      - sql: "UPDATE inventory SET stock = {cur - 1} WHERE item_id = 1 AND {cur > 0}"`
    },
    {
      key: "hospital",
      name: "hospital_write_skew",
      yaml: `version: "1.0"
name: "hospital_write_skew"
database:
  driver: "sqlite"
  schema: "CREATE TABLE doctors (id INT PRIMARY KEY, name TEXT NOT NULL, on_duty BOOLEAN NOT NULL);"
  seed: "INSERT INTO doctors VALUES (1, 'Dr. Alice', 1), (2, 'Dr. Bob', 1);"
invariants:
  - name: "at_least_one_doctor"
    query: "SELECT count(*) AS active FROM doctors WHERE on_duty = 1;"
    assert: "active >= 1"
operations:
  - name: "sign_off_alice"
    steps:
      - sql: "SELECT count(*) AS active FROM doctors WHERE on_duty = 1"
        capture: "act"
      - sql: "UPDATE doctors SET on_duty = 0 WHERE id = 1 AND {act >= 2}"
  - name: "sign_off_bob"
    steps:
      - sql: "SELECT count(*) AS active FROM doctors WHERE on_duty = 1"
        capture: "act"
      - sql: "UPDATE doctors SET on_duty = 0 WHERE id = 2 AND {act >= 2}"`
    },
    {
      key: "deadlock",
      name: "deadlock_cycle",
      yaml: `version: "1.0"
name: "deadlock_cycle"
database:
  driver: "sqlite"
  schema: "CREATE TABLE locks (id INT PRIMARY KEY, v INT);"
  seed: "INSERT INTO locks VALUES (1, 10), (2, 20);"
invariants:
  - name: "locks_valid"
    query: "SELECT count(*) AS c FROM locks;"
    assert: "c == 2"
operations:
  - name: "tx_1"
    steps:
      - sql: "UPDATE locks SET v = 11 WHERE id = 1"
      - sql: "UPDATE locks SET v = 21 WHERE id = 2"
  - name: "tx_2"
    steps:
      - sql: "UPDATE locks SET v = 22 WHERE id = 2"
      - sql: "UPDATE locks SET v = 12 WHERE id = 1"`
    }
  ];

  function getPerformance() {
    if (typeof performance !== "undefined") return performance;
    return { now: () => Date.now() };
  }

  function percentile(arr, p) {
    if (arr.length === 0) return 0;
    const sorted = [...arr].sort((a, b) => a - b);
    const index = Math.min(Math.floor((p / 100) * sorted.length), sorted.length - 1);
    return sorted[index];
  }

  /**
   * Run client-side stress benchmark against active Web Worker.
   * @param {number} iterations - Number of consecutive scenario runs (default: 20)
   * @param {function} onProgress - Optional callback(progressData)
   * @param {object} options - Optional worker or timing overrides
   * @returns {Promise<object>} Benchmark report
   */
  async function ChaosSQL_RunStressTest(iterations = 20, onProgress = null, options = {}) {
    const perf = getPerformance();
    const targetIterations = Math.max(1, parseInt(iterations, 10) || 20);

    // 1. Resolve or spawn Worker
    let worker = options.worker || (typeof window !== "undefined" ? window.wasmWorker : null);
    let ownWorker = false;

    if (!worker) {
      if (typeof Worker !== "undefined") {
        worker = new Worker("assets/wasm-worker.js");
        ownWorker = true;
      } else {
        throw new Error("No active Web Worker available and Web Workers unsupported in this context");
      }
    }

    // Initialize worker if needed
    if (ownWorker) {
      await new Promise((resolve, reject) => {
        const onInitMsg = (e) => {
          const msg = e.data || {};
          if (msg.type === "READY") {
            worker.removeEventListener("message", onInitMsg);
            resolve();
          } else if (msg.type === "ERROR") {
            worker.removeEventListener("message", onInitMsg);
            reject(new Error("Worker initialization failed: " + msg.error));
          }
        };
        worker.addEventListener("message", onInitMsg);
        worker.postMessage({ action: "INIT", wasmUrl: "chaossql.wasm" });
      });
    }

    // 2. Start frame timing monitor
    const frameDeltas = [];
    let rafActive = true;
    let rafHandle = null;
    let lastFrameTimestamp = perf.now();

    const requestFrame = typeof requestAnimationFrame === "function"
      ? requestAnimationFrame
      : (fn) => setTimeout(() => fn(perf.now()), 16);
    const cancelFrame = typeof cancelAnimationFrame === "function"
      ? cancelAnimationFrame
      : (id) => clearTimeout(id);

    function frameLoop(now) {
      if (!rafActive) return;
      const delta = now - lastFrameTimestamp;
      lastFrameTimestamp = now;
      if (delta > 0) {
        frameDeltas.push(delta);
      }
      rafHandle = requestFrame(frameLoop);
    }
    rafHandle = requestFrame(frameLoop);

    // 3. Dispatch stress iterations
    const runLatencies = [];
    let totalOperations = 0;
    let anomaliesCount = 0;
    const benchStartTime = perf.now();

    try {
      for (let i = 0; i < targetIterations; i++) {
        const preset = BENCH_PRESETS[i % BENCH_PRESETS.length];
        const scenarioConfig = {
          yamlContent: preset.yaml,
          workers: 2,
          iterations: 5,
          jitterMs: 0,
          seed: 100 + i
        };

        const iterationStart = perf.now();

        const result = await new Promise((resolve, reject) => {
          let cycleAnomaly = null;
          let cycleEdges = [];
          const timeout = setTimeout(() => {
            cleanup();
            reject(new Error(`Iteration ${i + 1} timed out after 10000ms`));
          }, 10000);

          function onMessage(e) {
            const msg = e.data || {};
            if (msg.type === "CYCLE_DETECTED") {
              cycleAnomaly = msg.anomalyType;
              if (msg.edges) cycleEdges = msg.edges;
            } else if (msg.type === "COMPLETE") {
              cleanup();
              let parsed = {};
              try {
                parsed = typeof msg.report === "string" ? JSON.parse(msg.report) : (msg.report || {});
              } catch (_) {
                parsed = { raw: msg.report };
              }
              resolve({
                report: parsed,
                cycleAnomaly,
                cycleEdges: parsed.adyaEdges || cycleEdges
              });
            } else if (msg.type === "ERROR") {
              cleanup();
              reject(new Error(msg.error || "WASM scenario execution failed"));
            }
          }

          function cleanup() {
            clearTimeout(timeout);
            worker.removeEventListener("message", onMessage);
          }

          worker.addEventListener("message", onMessage);
          worker.postMessage({ action: "RUN", config: scenarioConfig });
        });

        const iterationDuration = perf.now() - iterationStart;
        runLatencies.push(iterationDuration);

        const rep = result.report || {};
        const ops = rep.totalOps || 10;
        totalOperations += ops;

        if (rep.violationFound || result.cycleAnomaly) {
          anomaliesCount++;
        }

        if (typeof onProgress === "function") {
          try {
            onProgress({
              currentIteration: i + 1,
              totalIterations: targetIterations,
              scenario: preset.name,
              durationMs: iterationDuration,
              ops,
              anomalyType: rep.anomalyType || result.cycleAnomaly || (rep.violationFound ? "VIOLATION" : null)
            });
          } catch (_) {}
        }
      }
    } finally {
      rafActive = false;
      if (rafHandle) cancelFrame(rafHandle);
      if (ownWorker && worker.terminate) {
        worker.terminate();
      }
    }

    const totalDuration = perf.now() - benchStartTime;
    const throughputOpsSec = totalOperations / (totalDuration / 1000);

    // 4. Compute frame timing analytics
    const jankFrames = frameDeltas.filter(d => d > 16.67).length;
    const maxFrameDelta = frameDeltas.length > 0 ? Math.max(...frameDeltas) : 0;
    const compliancePct = frameDeltas.length > 0
      ? ((frameDeltas.length - jankFrames) / frameDeltas.length) * 100
      : 100;

    const avgLatency = runLatencies.reduce((a, b) => a + b, 0) / runLatencies.length;
    const minLatency = Math.min(...runLatencies);
    const maxLatency = Math.max(...runLatencies);
    const p95Latency = percentile(runLatencies, 95);

    const report = {
      iterations: targetIterations,
      completedRuns: runLatencies.length,
      totalOperations,
      totalDurationMs: totalDuration,
      throughputOpsPerSec: throughputOpsSec,
      anomaliesDetected: anomaliesCount,
      latency: {
        avgMs: avgLatency,
        minMs: minLatency,
        maxMs: maxLatency,
        p95Ms: p95Latency
      },
      frameMetrics: {
        totalFrames: frameDeltas.length,
        jankFrames,
        maxFrameDeltaMs: maxFrameDelta,
        frameBudgetCompliancePct: compliancePct,
        fps60Compliant: compliancePct >= 90 || jankFrames === 0
      }
    };

    return report;
  }

  // Export to global window and module exports
  if (typeof window !== "undefined") {
    window.ChaosSQL_RunStressTest = ChaosSQL_RunStressTest;
    window.BENCH_PRESETS = BENCH_PRESETS;
  }
  if (typeof globalThis !== "undefined") {
    globalThis.ChaosSQL_RunStressTest = ChaosSQL_RunStressTest;
  }
  if (typeof module !== "undefined" && module.exports) {
    module.exports = {
      ChaosSQL_RunStressTest,
      BENCH_PRESETS
    };
  }
})(typeof globalThis !== "undefined" ? globalThis : typeof window !== "undefined" ? window : this);
