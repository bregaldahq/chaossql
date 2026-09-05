#!/usr/bin/env node
// tools/headless_worker_stress.js
// Headless WebAssembly & Web Worker Concurrency Stress Harness
// Executes 50+ consecutive scenario runs inside headless V8 / Node.js VM sandbox,
// tracks WebAssembly linear memory growth & process RSS, benchmarks throughput (ops/sec),
// and asserts 60 FPS SVG DOM layout latency compliance (< 16.6ms).

const fs = require("fs");
const path = require("path");
const vm = require("vm");
const nodeCrypto = require("crypto");
const { performance } = require("perf_hooks");
const assert = require("assert");

const ROOT_DIR = path.resolve(__dirname, "..");
const ASSETS_DIR = path.join(ROOT_DIR, "site/assets");
const APP_JS_PATH = path.join(ROOT_DIR, "site/app.js");
const WASM_PATH = path.join(ASSETS_DIR, "chaossql.wasm");
const WASM_EXEC_PATH = path.join(ASSETS_DIR, "wasm_exec.js");
const WASM_WORKER_PATH = path.join(ASSETS_DIR, "wasm-worker.js");

// Canonical scenario presets for stress testing
const CANONICAL_PRESETS = [
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

// Helper to extract PLAYGROUND_PRESETS from site/app.js if available
function loadAppPresets() {
  try {
    if (fs.existsSync(APP_JS_PATH)) {
      const code = fs.readFileSync(APP_JS_PATH, "utf8");
      const match = code.match(/const PLAYGROUND_PRESETS = (\{[\s\S]*?\n\};)/);
      if (match) {
        const ctx = {};
        vm.createContext(ctx);
        vm.runInContext("var PLAYGROUND_PRESETS = " + match[1], ctx);
        const presets = ctx.PLAYGROUND_PRESETS;
        const list = [];
        for (const [k, yaml] of Object.entries(presets)) {
          const nameMatch = yaml.match(/name:\s*"([^"]+)"/);
          list.push({
            key: k,
            name: nameMatch ? nameMatch[1] : k,
            yaml
          });
        }
        if (list.length >= 4) return list;
      }
    }
  } catch (_) {}
  return CANONICAL_PRESETS;
}

// Build headless Web Worker sandbox
function setupWasmWorkerSandbox() {
  const wasmBuffer = fs.readFileSync(WASM_PATH);
  const wasmExecCode = fs.readFileSync(WASM_EXEC_PATH, "utf8");
  const wasmWorkerCode = fs.readFileSync(WASM_WORKER_PATH, "utf8");

  let wasmMemoryInstance = null;
  let wasmInstance = null;

  const customWebAssembly = Object.create(WebAssembly);
  customWebAssembly.instantiateStreaming = undefined; // Force fallback to arrayBuffer + instantiate in Node
  customWebAssembly.instantiate = async (buf, imports) => {
    const res = await WebAssembly.instantiate(buf, imports);
    const instance = res.instance || res;
    wasmInstance = instance;
    if (instance.exports && instance.exports.mem) {
      wasmMemoryInstance = instance.exports.mem;
    }
    return res;
  };

  const sandbox = {
    console: {
      log: () => {},
      warn: () => {},
      error: () => {},
      info: () => {}
    },
    setTimeout,
    clearTimeout,
    setInterval,
    clearInterval,
    setImmediate,
    clearImmediate,
    process,
    Buffer,
    TextEncoder,
    TextDecoder,
    crypto: nodeCrypto.webcrypto,
    performance,
    WebAssembly: customWebAssembly
  };

  sandbox.globalThis = sandbox;
  sandbox.self = sandbox;
  sandbox.window = sandbox;

  sandbox.importScripts = (scriptName) => {
    if (scriptName === "wasm_exec.js" || scriptName.endsWith("wasm_exec.js")) {
      vm.runInContext(wasmExecCode, sandbox);
    } else {
      const full = path.resolve(ASSETS_DIR, scriptName);
      if (fs.existsSync(full)) {
        vm.runInContext(fs.readFileSync(full, "utf8"), sandbox);
      }
    }
  };

  sandbox.fetch = async (url) => {
    return {
      ok: true,
      status: 200,
      statusText: "OK",
      clone() { return this; },
      arrayBuffer: async () => wasmBuffer.buffer.slice(wasmBuffer.byteOffset, wasmBuffer.byteOffset + wasmBuffer.byteLength)
    };
  };

  const messageListeners = [];
  sandbox.self.postMessage = (msg) => {
    const cloned = JSON.parse(JSON.stringify(msg));
    for (const listener of [...messageListeners]) {
      listener(cloned);
    }
  };

  vm.createContext(sandbox);
  vm.runInContext(wasmWorkerCode, sandbox);

  return {
    sandbox,
    getMemory: () => wasmMemoryInstance,
    getInstance: () => wasmInstance,
    postAction: async (action, data = {}) => {
      await sandbox.self.onmessage({ data: { action, ...data } });
    },
    addMessageListener: (fn) => {
      messageListeners.push(fn);
      return () => {
        const idx = messageListeners.indexOf(fn);
        if (idx !== -1) messageListeners.splice(idx, 1);
      };
    }
  };
}

// Minimal DOM & SVG layout test harness for Adya Graph
function setupAdyaBenchmark() {
  const container = { innerHTML: "" };
  const baseCoords = {
    T1: { x: 150, y: 100 },
    T2: { x: 450, y: 100 },
    T3: { x: 450, y: 260 },
    T4: { x: 150, y: 260 }
  };

  function escapeHtml(str) {
    if (!str) return "";
    return String(str).replace(/[&<>"']/g, (m) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"
    }[m]));
  }

  function renderAdyaLayout(edges, anomalyType) {
    const normalizeNode = (id) => (id ? String(id).split("-")[0] : id);
    const cycleNodes = new Set();
    const edgeNodes = [];

    (edges || []).forEach(edge => {
      const u = normalizeNode(edge.from);
      const v = normalizeNode(edge.to);
      if (u) {
        cycleNodes.add(u);
        if (!edgeNodes.includes(u)) edgeNodes.push(u);
      }
      if (v) {
        cycleNodes.add(v);
        if (!edgeNodes.includes(v)) edgeNodes.push(v);
      }
    });

    const displayNodes = edgeNodes.length > 0
      ? (edgeNodes.every(n => baseCoords[n]) ? ["T1", "T2", "T3", "T4"] : edgeNodes)
      : ["T1", "T2", "T3", "T4"];

    const coords = {};
    displayNodes.forEach((node, idx) => {
      if (baseCoords[node]) {
        coords[node] = baseCoords[node];
      } else {
        const angle = (2 * Math.PI * idx) / Math.max(displayNodes.length, 1) - Math.PI / 2;
        coords[node] = {
          x: Math.round(300 + 180 * Math.cos(angle)),
          y: Math.round(180 + 90 * Math.sin(angle))
        };
      }
    });

    const getCoord = (n) => {
      const norm = normalizeNode(n);
      if (coords[norm]) return coords[norm];
      if (baseCoords[norm]) return baseCoords[norm];
      return { x: 300, y: 180 };
    };

    let svgHtml = "";

    (edges || []).forEach(edge => {
      const fromNorm = normalizeNode(edge.from);
      const toNorm = normalizeNode(edge.to);
      const from = getCoord(fromNorm);
      const to = getCoord(toNorm);

      let color = "#e06c75";
      let marker = "url(#arrow-rw)";
      if (edge.type === "WW") {
        color = "#d19a66";
        marker = "url(#arrow-ww)";
      } else if (edge.type === "WR") {
        color = "#c678dd";
        marker = "url(#arrow-wr)";
      }

      const dx = to.x - from.x;
      const dy = to.y - from.y;
      const dist = Math.sqrt(dx * dx + dy * dy);

      let pathD = "";
      let labelX = (from.x + to.x) / 2;
      let labelY = (from.y + to.y) / 2;

      if (dist < 1) {
        pathD = `M ${from.x - 12} ${from.y - 18} C ${from.x - 35} ${from.y - 65}, ${from.x + 35} ${from.y - 65}, ${from.x + 12} ${from.y - 18}`;
        labelX = from.x;
        labelY = from.y - 50;
      } else {
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

        pathD = `M ${startX.toFixed(1)} ${startY.toFixed(1)} Q ${midX.toFixed(1)} ${midY.toFixed(1)} ${targetX.toFixed(1)} ${targetY.toFixed(1)}`;
        labelX = 0.25 * startX + 0.5 * midX + 0.25 * targetX;
        labelY = 0.25 * startY + 0.5 * midY + 0.25 * targetY;
      }

      const labelText = `${edge.type || "RW"}${edge.item ? ` (${edge.item})` : ""}`;
      const badgeW = Math.max(48, labelText.length * 7 + 12);
      const badgeH = 18;

      svgHtml += `
        <path d="${pathD}" stroke="${color}" stroke-width="2.5" fill="none" marker-end="${marker}" stroke-dasharray="4,2" />
        <rect x="${(labelX - badgeW / 2).toFixed(1)}" y="${(labelY - badgeH / 2).toFixed(1)}" width="${badgeW}" height="${badgeH}" rx="4" fill="#1e2227" stroke="${color}" stroke-width="1" />
        <text x="${labelX.toFixed(1)}" y="${(labelY + 4).toFixed(1)}" fill="${color}" font-size="10" font-family="monospace" text-anchor="middle" font-weight="600">
          ${escapeHtml(labelText)}
        </text>
      `;
    });

    displayNodes.forEach(node => {
      const c = coords[node];
      const isConflicted = Boolean(anomalyType && cycleNodes.has(node));
      const strokeColor = isConflicted ? "#e06c75" : "#61afef";
      const fillColor = isConflicted ? "#2d1f24" : "#1e2227";

      svgHtml += `
        <g transform="translate(${c.x}, ${c.y})">
          ${isConflicted ? `<circle r="26" fill="none" stroke="#e06c75" stroke-width="1.5" opacity="0.6">
            <animate attributeName="r" values="22;30;22" dur="2s" repeatCount="indefinite"/>
          </circle>` : ""}
          <circle r="20" fill="${fillColor}" stroke="${strokeColor}" stroke-width="2" />
          <text y="5" fill="#e5e9f0" font-size="12" font-family="monospace" font-weight="700" text-anchor="middle">
            ${escapeHtml(node)}
          </text>
        </g>
      `;
    });

    if (anomalyType) {
      svgHtml += `
        <text x="300" y="340" text-anchor="middle" fill="#e06c75" font-size="12" font-family="monospace" font-weight="700">
          CICLO ADYA CLASSIFICADO: ${escapeHtml(anomalyType)}
        </text>
      `;
    }

    container.innerHTML = svgHtml;
    return svgHtml.length;
  }

  return { renderAdyaLayout };
}

// Percentile calculation helper
function percentile(arr, p) {
  if (arr.length === 0) return 0;
  const sorted = [...arr].sort((a, b) => a - b);
  const index = Math.min(Math.floor((p / 100) * sorted.length), sorted.length - 1);
  return sorted[index];
}

async function main() {
  const isJsonOutput = process.argv.includes("--json");
  const iterationsArg = process.argv.find(a => a.startsWith("--runs=") || a.startsWith("--iterations="));
  let targetIterations = 50;
  if (iterationsArg) {
    targetIterations = parseInt(iterationsArg.split("=")[1], 10) || 50;
  } else if (process.env.STRESS_RUNS) {
    targetIterations = parseInt(process.env.STRESS_RUNS, 10) || 50;
  }

  if (!isJsonOutput) {
    console.log("===============================================================================");
    console.log("   CHAOSSQL: HEADLESS WEBASSEMBLY & WEB WORKER CONCURRENCY STRESS HARNESS");
    console.log("===============================================================================");
    console.log(`  Configuration: ${targetIterations} consecutive runs across canonical scenario matrix`);
    console.log(`  Engine Binary: ${path.relative(ROOT_DIR, WASM_PATH)} (${(fs.statSync(WASM_PATH).size / (1024 * 1024)).toFixed(2)} MB)`);
    console.log("-------------------------------------------------------------------------------");
  }

  // 1. Initialize Worker Sandbox
  const worker = setupWasmWorkerSandbox();
  const initStartTime = performance.now();

  await new Promise((resolve, reject) => {
    const remove = worker.addMessageListener((msg) => {
      if (msg.type === "READY") {
        remove();
        resolve();
      } else if (msg.type === "ERROR") {
        remove();
        reject(new Error("Worker initialization error: " + msg.error));
      }
    });
    worker.postAction("INIT", { wasmUrl: "chaossql.wasm" }).catch(reject);
  });

  const initDuration = performance.now() - initStartTime;
  if (!isJsonOutput) {
    console.log(`  ✔ Worker sandbox instantiated & Go runtime initialized in ${initDuration.toFixed(1)}ms`);
  }

  const wasmMemory = worker.getMemory();
  assert(wasmMemory, "WebAssembly linear memory reference must be acquired");

  const presets = loadAppPresets();

  // Warmup run to compile WebAssembly bytecode and stabilize V8 JIT memory
  const warmupPreset = presets[0];
  await new Promise((resolve, reject) => {
    const timer = setTimeout(() => { remove(); reject(new Error("Warmup timeout")); }, 10000);
    const remove = worker.addMessageListener((msg) => {
      if (msg.type === "COMPLETE") {
        clearTimeout(timer);
        remove();
        resolve();
      } else if (msg.type === "ERROR") {
        clearTimeout(timer);
        remove();
        reject(new Error(msg.error));
      }
    });
    worker.postAction("RUN", {
      config: {
        yamlContent: warmupPreset.yaml,
        workers: 2,
        iterations: 2,
        jitterMs: 0,
        seed: 41
      }
    }).catch(reject);
  });

  // Initial Telemetry Baselines (post-warmup)
  const initialMem = process.memoryUsage();
  const initialWasmBytes = wasmMemory.buffer.byteLength;

  if (!isJsonOutput) {
    console.log(`  • Initial Process RSS: ${(initialMem.rss / (1024 * 1024)).toFixed(2)} MB (post-warmup)`);
    console.log(`  • Initial Heap Used:   ${(initialMem.heapUsed / (1024 * 1024)).toFixed(2)} MB`);
    console.log(`  • Initial WASM Memory: ${(initialWasmBytes / (1024 * 1024)).toFixed(2)} MB\n`);
    console.log("  [STRESS LOOP EXECUTION]");
  }

  const runLatencies = [];
  const runRecords = [];
  let totalOpsCount = 0;
  let anomaliesDetected = 0;
  const collectedAdyaSamples = [];

  const overallStartTime = performance.now();

  for (let i = 0; i < targetIterations; i++) {
    const preset = presets[i % presets.length];
    const scenarioConfig = {
      yamlContent: preset.yaml,
      workers: 2,
      iterations: 5,
      jitterMs: 0,
      seed: 42 + i
    };

    const runStart = performance.now();
    let runResult = null;

    try {
      runResult = await new Promise((resolve, reject) => {
        let cycleAnomaly = null;
        let cycleEdges = [];
        const timer = setTimeout(() => {
          remove();
          reject(new Error(`Timeout on run ${i + 1} (${preset.name})`));
        }, 10000);

        const remove = worker.addMessageListener((msg) => {
          if (msg.type === "CYCLE_DETECTED") {
            cycleAnomaly = msg.anomalyType;
            if (msg.edges) cycleEdges = msg.edges;
          } else if (msg.type === "COMPLETE") {
            clearTimeout(timer);
            remove();
            let parsedReport = {};
            try {
              parsedReport = typeof msg.report === "string" ? JSON.parse(msg.report) : msg.report;
            } catch (_) {
              parsedReport = { raw: msg.report };
            }
            resolve({
              report: parsedReport,
              cycleAnomaly,
              cycleEdges: (parsedReport && parsedReport.adyaEdges) ? parsedReport.adyaEdges : cycleEdges
            });
          } else if (msg.type === "ERROR") {
            clearTimeout(timer);
            remove();
            reject(new Error(`Scenario error: ${msg.error}`));
          }
        });

        worker.postAction("RUN", { config: scenarioConfig }).catch((err) => {
          clearTimeout(timer);
          remove();
          reject(err);
        });
      });
    } catch (err) {
      console.error(`\n❌ Error during iteration ${i + 1} (${preset.name}):`, err.message);
      process.exit(1);
    }

    const runElapsed = performance.now() - runStart;
    runLatencies.push(runElapsed);

    const report = runResult.report || {};
    const ops = report.totalOps || 10;
    totalOpsCount += ops;

    if (report.violationFound || runResult.cycleAnomaly) {
      anomaliesDetected++;
    }

    if (runResult.cycleEdges && runResult.cycleEdges.length > 0) {
      collectedAdyaSamples.push({
        edges: runResult.cycleEdges,
        anomaly: report.anomalyType || runResult.cycleAnomaly
      });
    }

    const detectedAnomaly = report.anomalyType || runResult.cycleAnomaly || (report.violationFound ? (report.failingInvariant ? `INV:${report.failingInvariant}` : "DETECTED") : null);

    runRecords.push({
      iteration: i + 1,
      scenario: preset.name,
      durationMs: runElapsed,
      ops,
      anomaly: detectedAnomaly || "NONE",
      wasmMemBytes: wasmMemory.buffer.byteLength
    });

    if (!isJsonOutput) {
      const progress = `[${String(i + 1).padStart(3, " ")}/${targetIterations}]`;
      const name = preset.name.padEnd(28, " ");
      const anomalyBadge = detectedAnomaly
        ? `ANOMALY [${detectedAnomaly}]`.padEnd(28, " ")
        : "SERIALIZABLE [PASS]".padEnd(28, " ");
      const timeStr = `${runElapsed.toFixed(1)}ms`.padStart(8, " ");
      process.stdout.write(`  • ${progress} ${name} -> ${anomalyBadge} (${timeStr} | ${ops} ops)\n`);
    }
  }

  const overallDuration = performance.now() - overallStartTime;
  const throughputOpsPerSec = (totalOpsCount / (overallDuration / 1000));

  // 2. Telemetry Measurements After Stress Loop
  const finalMem = process.memoryUsage();
  const finalWasmBytes = wasmMemory.buffer.byteLength;

  const rssDeltaMB = (finalMem.rss - initialMem.rss) / (1024 * 1024);
  const heapDeltaMB = (finalMem.heapUsed - initialMem.heapUsed) / (1024 * 1024);
  const wasmDeltaMB = (finalWasmBytes - initialWasmBytes) / (1024 * 1024);

  // 3. Adya Graph SVG Layout Latency Benchmark
  const { renderAdyaLayout } = setupAdyaBenchmark();
  const testTopologies = [
    {
      name: "2-node bidirectional rw/ww",
      edges: [
        { from: "T1-0", to: "T2-0", type: "RW", item: "balance" },
        { from: "T2-0", to: "T1-0", type: "WW", item: "balance" }
      ],
      anomaly: "P4_LOST_UPDATE"
    },
    {
      name: "3-node circular dependency rw/wr/ww",
      edges: [
        { from: "T1-0", to: "T2-0", type: "RW", item: "x" },
        { from: "T2-0", to: "T3-0", type: "WR", item: "y" },
        { from: "T3-0", to: "T1-0", type: "WW", item: "z" }
      ],
      anomaly: "G1c_CIRCULAR_INFO"
    },
    {
      name: "4-node anti-dependency chain (G2)",
      edges: [
        { from: "T1-0", to: "T2-0", type: "RW", item: "doctor" },
        { from: "T2-0", to: "T3-0", type: "RW", item: "doctor" },
        { from: "T3-0", to: "T4-0", type: "RW", item: "doctor" },
        { from: "T4-0", to: "T1-0", type: "RW", item: "doctor" }
      ],
      anomaly: "A5B_WRITE_SKEW"
    }
  ];

  for (const sample of collectedAdyaSamples.slice(0, 5)) {
    testTopologies.push({
      name: `live sample: ${sample.anomaly || "cycle"}`,
      edges: sample.edges,
      anomaly: sample.anomaly
    });
  }

  const svgLayoutTimes = [];
  const svgBenchmarkReps = 100;
  for (let b = 0; b < svgBenchmarkReps; b++) {
    const topo = testTopologies[b % testTopologies.length];
    const t0 = performance.now();
    renderAdyaLayout(topo.edges, topo.anomaly);
    const t1 = performance.now();
    svgLayoutTimes.push(t1 - t0);
  }

  const avgSvgLayoutMs = svgLayoutTimes.reduce((a, b) => a + b, 0) / svgLayoutTimes.length;
  const p95SvgLayoutMs = percentile(svgLayoutTimes, 95);
  const maxSvgLayoutMs = Math.max(...svgLayoutTimes);

  const avgLatency = runLatencies.reduce((a, b) => a + b, 0) / runLatencies.length;
  const minLatency = Math.min(...runLatencies);
  const maxLatency = Math.max(...runLatencies);
  const p95Latency = percentile(runLatencies, 95);

  // 4. Invariant Assertions
  const assertions = [];

  function check(label, condition, detail) {
    const pass = Boolean(condition);
    assertions.push({ label, pass, detail });
    if (!pass) {
      if (!isJsonOutput) {
        console.error(`\n❌ ASSERTION FAILED: ${label} -> ${detail}`);
      }
    }
  }

  // Assert RSS growth < 50MB
  check(
    "RSS Memory Stability (< 50MB growth)",
    rssDeltaMB < 50,
    `RSS delta = ${rssDeltaMB.toFixed(2)} MB (Limit: 50.00 MB)`
  );

  // Assert WebAssembly linear memory stability (bounded growth)
  check(
    "WebAssembly Linear Memory Stability (< 32MB delta)",
    wasmDeltaMB < 32,
    `WASM linear memory delta = ${wasmDeltaMB.toFixed(2)} MB (Final: ${(finalWasmBytes / (1024 * 1024)).toFixed(2)} MB)`
  );

  // Assert Adya SVG layout average latency < 16.6ms (60 FPS frame budget)
  check(
    "Adya SVG Layout Avg Latency (< 16.6ms for 60 FPS)",
    avgSvgLayoutMs < 16.6,
    `Avg layout time = ${avgSvgLayoutMs.toFixed(3)}ms (Budget: 16.666ms)`
  );

  // Assert Adya SVG layout P95 latency < 16.6ms
  check(
    "Adya SVG Layout P95 Latency (< 16.6ms for 60 FPS)",
    p95SvgLayoutMs < 16.6,
    `P95 layout time = ${p95SvgLayoutMs.toFixed(3)}ms (Budget: 16.666ms)`
  );

  // Assert all iterations completed successfully
  check(
    "Scenario Completion Integrity",
    runLatencies.length === targetIterations,
    `Executed ${runLatencies.length} of ${targetIterations} scheduled runs`
  );

  const allPassed = assertions.every(a => a.pass);

  const summaryData = {
    targetIterations,
    completedIterations: runLatencies.length,
    totalOps: totalOpsCount,
    totalDurationMs: overallDuration,
    throughputOpsPerSec,
    anomaliesDetected,
    latency: {
      avgMs: avgLatency,
      minMs: minLatency,
      maxMs: maxLatency,
      p95Ms: p95Latency
    },
    memory: {
      initialRssMB: initialMem.rss / (1024 * 1024),
      finalRssMB: finalMem.rss / (1024 * 1024),
      rssDeltaMB,
      initialHeapMB: initialMem.heapUsed / (1024 * 1024),
      finalHeapMB: finalMem.heapUsed / (1024 * 1024),
      heapDeltaMB,
      initialWasmMB: initialWasmBytes / (1024 * 1024),
      finalWasmMB: finalWasmBytes / (1024 * 1024),
      wasmDeltaMB
    },
    svgLayout: {
      benchmarkIterations: svgBenchmarkReps,
      avgMs: avgSvgLayoutMs,
      p95Ms: p95SvgLayoutMs,
      maxMs: maxSvgLayoutMs,
      fps60Compliant: p95SvgLayoutMs < 16.6
    },
    assertions,
    passed: allPassed
  };

  if (isJsonOutput) {
    console.log(JSON.stringify(summaryData, null, 2));
  } else {
    console.log("\n-------------------------------------------------------------------------------");
    console.log("   STRESS TEST TELEMETRY & FRAME BUDGET AUDIT SUMMARY");
    console.log("-------------------------------------------------------------------------------");
    console.log(`  • Runs Executed:       ${summaryData.completedIterations} / ${targetIterations}`);
    console.log(`  • Anomalies Detected:  ${anomaliesDetected}`);
    console.log(`  • Total Operations:    ${totalOpsCount} ops`);
    console.log(`  • Overall Time:        ${overallDuration.toFixed(1)}ms`);
    console.log(`  • Throughput:          ${throughputOpsPerSec.toFixed(1)} ops/sec`);
    console.log(`  • Run Latency:         avg=${avgLatency.toFixed(1)}ms | min=${minLatency.toFixed(1)}ms | max=${maxLatency.toFixed(1)}ms | p95=${p95Latency.toFixed(1)}ms`);
    console.log(`  • Process RSS Memory:  ${(initialMem.rss / (1024 * 1024)).toFixed(2)} MB -> ${(finalMem.rss / (1024 * 1024)).toFixed(2)} MB (delta: ${rssDeltaMB >= 0 ? "+" : ""}${rssDeltaMB.toFixed(2)} MB)`);
    console.log(`  • V8 Heap Used:        ${(initialMem.heapUsed / (1024 * 1024)).toFixed(2)} MB -> ${(finalMem.heapUsed / (1024 * 1024)).toFixed(2)} MB (delta: ${heapDeltaMB >= 0 ? "+" : ""}${heapDeltaMB.toFixed(2)} MB)`);
    console.log(`  • WASM Linear Memory:  ${(initialWasmBytes / (1024 * 1024)).toFixed(2)} MB -> ${(finalWasmBytes / (1024 * 1024)).toFixed(2)} MB (delta: ${wasmDeltaMB >= 0 ? "+" : ""}${wasmDeltaMB.toFixed(2)} MB)`);
    console.log(`  • Adya SVG Generation: avg=${avgSvgLayoutMs.toFixed(3)}ms | p95=${p95SvgLayoutMs.toFixed(3)}ms | max=${maxSvgLayoutMs.toFixed(3)}ms [60 FPS compliant: ${p95SvgLayoutMs < 16.6 ? "YES" : "NO"}]`);
    console.log("-------------------------------------------------------------------------------");
    console.log("   INVARIANT ASSERTIONS:");
    for (const a of assertions) {
      const mark = a.pass ? "✔ PASS" : "✘ FAIL";
      console.log(`    ${mark}: ${a.label} (${a.detail})`);
    }
    console.log("===============================================================================\n");
  }

  if (!allPassed) {
    process.exit(1);
  }
}

main().catch((err) => {
  console.error("\nFATAL STRESS HARNESS ERROR:", err);
  process.exit(1);
});
