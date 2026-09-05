// tools/test_playground_ui.js
// Automated behavioral tests for ChaosSQL Playground UI Studio & Visualizers

const fs = require("fs");
const path = require("path");
const vm = require("vm");
const assert = require("assert");

const SITE_DIR = path.resolve(__dirname, "../site");
const appJsPath = path.join(SITE_DIR, "app.js");
const appJsCode = fs.readFileSync(appJsPath, "utf8");

function createPlaygroundDOM() {
  const elements = new Map();

  function createMockElement(tag, id = "") {
    const el = {
      tagName: tag.toUpperCase(),
      id,
      className: "",
      style: {},
      value: "",
      disabled: false,
      _listeners: {},
      classList: {
        _set: new Set(),
        add(c) { this._set.add(c); el.className = Array.from(this._set).join(" "); },
        remove(c) { this._set.delete(c); el.className = Array.from(this._set).join(" "); },
        toggle(c, force) {
          if (force === undefined) {
            if (this._set.has(c)) this.remove(c); else this.add(c);
          } else if (force) this.add(c);
          else this.remove(c);
        },
        contains(c) { return this._set.has(c); }
      },
      attributes: {},
      setAttribute(k, v) { this.attributes[k] = String(v); },
      getAttribute(k) { return this.attributes[k] !== undefined ? this.attributes[k] : null; },
      removeAttribute(k) { delete this.attributes[k]; },
      children: [],
      appendChild(child) {
        this.children.push(child);
        return child;
      },
      get textContent() {
        if (this._textContent !== undefined) return this._textContent;
        if (this.children.length > 0) return this.children.map(c => c.textContent).join("");
        return "";
      },
      set textContent(val) {
        this._textContent = String(val);
        this.children = [];
      },
      _innerHTML: "",
      get innerHTML() {
        return this._innerHTML;
      },
      set innerHTML(val) {
        this._innerHTML = String(val);
      },
      addEventListener(evt, fn) {
        if (!this._listeners[evt]) this._listeners[evt] = [];
        this._listeners[evt].push(fn);
      },
      removeEventListener(evt, fn) {
        if (!this._listeners[evt]) return;
        this._listeners[evt] = this._listeners[evt].filter(f => f !== fn);
      },
      trigger(evt, eventObj = {}) {
        if (this._listeners[evt]) {
          for (const fn of this._listeners[evt]) {
            fn.call(this, eventObj);
          }
        }
      }
    };
    if (id) elements.set(id, el);
    return el;
  }

  const knownIds = [
    "pgGanttContainer", "pgAdyaGraphContent", "pgAdyaSvg",
    "pgMetricOps", "pgMetricAnomaly", "pgMetricCycle", "pgMetricDuration",
    "pgWasmStatus", "pgWasmStatusText", "pgAlertBox", "pgConsoleOutput",
    "pgRunBtn", "pgCancelBtn", "pgValidateBtn", "pgPresetSelect",
    "pgWorkersRange", "pgWorkersVal", "pgIterationsRange", "pgIterationsVal",
    "pgJitterRange", "pgJitterVal", "pgSeedInput", "pgSeedVal",
    "pgYamlEditor", "pgResetYamlBtn", "pgTabAdya", "pgTabGantt", "pgTabLog",
    "pgPaneAdya", "pgPaneGantt", "pgPaneLog", "terminalOutput",
    "navMobileToggle", "mobileNavDrawer", "view-playground"
  ];

  for (const id of knownIds) {
    createMockElement("div", id);
  }

  const documentElement = createMockElement("html");
  documentElement.lang = "pt";

  const doc = {
    readyState: "complete",
    documentElement,
    getElementById(id) {
      if (!elements.has(id)) {
        createMockElement("div", id);
      }
      return elements.get(id);
    },
    createElement(tag) {
      return createMockElement(tag);
    },
    querySelectorAll(selector) {
      const results = [];
      for (const el of elements.values()) {
        if (selector.startsWith("#") && el.id === selector.slice(1)) results.push(el);
        if (selector.startsWith(".") && el.classList.contains(selector.slice(1))) results.push(el);
        if (selector.startsWith("[data-i18n]") && el.getAttribute("data-i18n")) results.push(el);
        if (selector.startsWith("[data-i18n-attr]") && el.getAttribute("data-i18n-attr")) results.push(el);
      }
      return results;
    },
    querySelector(selector) {
      const res = this.querySelectorAll(selector);
      return res.length > 0 ? res[0] : null;
    },
    addEventListener() {},
    removeEventListener() {}
  };

  class MockWorker {
    constructor() {
      this.onmessage = null;
      this.onerror = null;
    }
    postMessage() {}
    addEventListener() {}
  }

  const sandboxWindow = {
    addEventListener() {},
    location: { hash: "#/playground" },
    scrollTo: () => {},
    localStorage: {
      _store: {},
      getItem(k) { return this._store[k] || null; },
      setItem(k, v) { this._store[k] = String(v); }
    },
    navigator: { language: "pt-BR" },
    currentLang: "pt",
    setTimeout: (fn, delay) => { if (delay === 0) fn(); return 1; },
    clearTimeout: () => {},
    Worker: MockWorker
  };

  const sandbox = {
    window: sandboxWindow,
    document: doc,
    Worker: MockWorker,
    navigator: sandboxWindow.navigator,
    localStorage: sandboxWindow.localStorage,
    setTimeout: sandboxWindow.setTimeout,
    clearTimeout: sandboxWindow.clearTimeout,
    console: {
      log: () => {},
      warn: () => {},
      error: () => {}
    },
    Math,
    JSON,
    Array,
    Object,
    String,
    Number,
    Boolean,
    Set,
    Map,
    RegExp,
    parseInt,
    parseFloat,
    encodeURIComponent,
    decodeURIComponent
  };

  sandbox.window.window = sandbox.window;
  sandbox.window.document = doc;

  const context = vm.createContext(sandbox);
  vm.runInContext(appJsCode, context);

  return { context, doc, elements, window: sandbox.window };
}

async function runTests() {
  console.log("Running ChaosSQL Playground UI Studio & Visualizer test suite...\n");

  // =========================================================================
  // Test Suite 1: Gantt Visualizer (renderPlaygroundGantt)
  // =========================================================================
  console.log("--> Test Suite 1: Gantt Visualizer (renderPlaygroundGantt)");
  {
    const { context, doc } = createPlaygroundDOM();
    const ganttContainer = doc.getElementById("pgGanttContainer");

    const realGoTrace = [
      { worker_id: 1, type: "BEGIN", sql: "BEGIN;", start_time_ns: 1000, duration_ns: 200 },
      { worker_id: 1, type: "EXEC", sql: "SELECT balance FROM accounts WHERE id = 1;", start_time_ns: 1200, duration_ns: 500 },
      { worker_id: 2, type: "BEGIN", sql: "BEGIN;", start_time_ns: 1100, duration_ns: 200 },
      { worker_id: 2, type: "EXEC", sql: "UPDATE accounts SET balance = balance - 100 WHERE id = 1;", start_time_ns: 1700, duration_ns: 800 },
      { worker_id: 1, type: "COMMIT", sql: "COMMIT;", start_time_ns: 2500, duration_ns: 300 },
      { worker_id: 2, type: "ROLLBACK", sql: "ROLLBACK;", start_time_ns: 2800, duration_ns: 400 }
    ];

    vm.runInContext("renderPlaygroundGantt", context)(realGoTrace);
    const html = ganttContainer.innerHTML;

    assert(!html.includes("undefined"), 'Gantt HTML must not contain "undefined" from schema mismatch');
    assert(html.includes("Worker #1") || html.includes("Worker 1"), "Must render Worker #1 swimlane");
    assert(html.includes("Worker #2") || html.includes("Worker 2"), "Must render Worker #2 swimlane");

    assert(html.includes("SELECT balance FROM accounts"), "Must render actual SQL statement for Worker 1 EXEC");
    assert(html.includes("UPDATE accounts SET balance"), "Must render actual SQL statement for Worker 2 EXEC");
    assert(html.includes("BEGIN"), "Must render BEGIN event");
    assert(html.includes("COMMIT"), "Must render COMMIT event");
    assert(html.includes("ROLLBACK"), "Must render ROLLBACK event");

    assert(html.includes("ev-begin"), "Must have ev-begin class for BEGIN");
    assert(html.includes("ev-exec"), "Must have ev-exec class for EXEC");
    assert(html.includes("ev-commit"), "Must have ev-commit class for COMMIT");
    assert(html.includes("ev-rollback"), "Must have ev-rollback class for ROLLBACK");

    console.log("  ✔ 1.1: Go TraceEvent snake_case schema renders multiple swimlanes with actual SQL and badges");

    vm.runInContext("renderPlaygroundGantt", context)([]);
    const emptyHtml = ganttContainer.innerHTML;
    assert(!emptyHtml.includes("undefined"), "Empty state must not contain undefined");
    assert(emptyHtml.includes("pg-empty-state"), "Empty state must render .pg-empty-state");
    assert(emptyHtml.length > 20, "Empty state must contain message text");
    console.log("  ✔ 1.2: Empty trace renders localized empty state container");
  }

  // =========================================================================
  // Test Suite 2: Adya Graph Visualizer (renderPlaygroundAdya)
  // =========================================================================
  console.log("\n--> Test Suite 2: Adya Graph Visualizer (renderPlaygroundAdya)");
  {
    const { context, doc } = createPlaygroundDOM();
    const adyaContainer = doc.getElementById("pgAdyaGraphContent");

    const testEdges = [
      { from: "T1-0", to: "T2-0", type: "RW", item: "balance" },
      { from: "T2-0", to: "T1-0", type: "WW", item: "balance" }
    ];

    vm.runInContext("renderPlaygroundAdya", context)(testEdges, "A5B");
    const svgHtml = adyaContainer.innerHTML;

    assert(!svgHtml.includes("T1-0"), "Must normalize T1-0 to T1");
    assert(!svgHtml.includes("T2-0"), "Must normalize T2-0 to T2");
    assert(svgHtml.includes(">T1<") || svgHtml.includes("> T1 <") || svgHtml.includes("T1"), "Must render normalized node label T1");
    assert(svgHtml.includes(">T2<") || svgHtml.includes("> T2 <") || svgHtml.includes("T2"), "Must render normalized node label T2");

    assert(!svgHtml.includes("translate(200, 180)"), "Normalized nodes must map to valid coords, not fallback (200, 180)");
    console.log("  ✔ 2.1: Transaction IDs normalized from instance buckets to base nodes (T1-0 -> T1)");

    const paths = svgHtml.match(/<path[^>]+d="([^"]+)"[^>]*>/g) || [];
    assert.strictEqual(paths.length, 2, "Must render exactly 2 edge paths");

    const d1Match = paths[0].match(/d="([^"]+)"/);
    const d2Match = paths[1].match(/d="([^"]+)"/);
    assert(d1Match && d2Match, "Paths must have valid d attributes");
    assert.notStrictEqual(d1Match[1], d2Match[1], "Bidirectional paths must NOT have identical d strings (must separate curves)");

    const q1 = d1Match[1].match(/Q\s*([-\d.]+)\s+([-\d.]+)/);
    const q2 = d2Match[1].match(/Q\s*([-\d.]+)\s+([-\d.]+)/);
    assert(q1 && q2, "Paths must use quadratic bezier curves (Q)");
    const q1Y = parseFloat(q1[2]);
    const q2Y = parseFloat(q2[2]);
    assert(Math.abs(q1Y - q2Y) >= 20, `Bidirectional curve control points must have distinct Y coordinates (got ${q1Y} vs ${q2Y})`);
    console.log("  ✔ 2.2: Bidirectional edges separated with distinct quadratic bezier curves (no overlap)");

    const endPointMatch1 = d1Match[1].match(/Q\s*[\d.]+\s+[\d.]+\s+([\d.]+)\s+([\d.]+)/);
    assert(endPointMatch1, "Path 1 must have end coordinates");
    const endX1 = parseFloat(endPointMatch1[1]);
    assert(endX1 < 445, `Path ending at T2 (x=450) must terminate at node boundary (<= 445) to prevent marker occlusion, got ${endX1}`);
    console.log("  ✔ 2.3: Edge endpoints terminate at node boundary radius offset (no marker occlusion)");

    const cycleEdges = [
      { from: "T3-0", to: "T4-0", type: "RW", item: "counter" },
      { from: "T4-0", to: "T3-0", type: "RW", item: "counter" }
    ];
    vm.runInContext("renderPlaygroundAdya", context)(cycleEdges, "G2");
    const cycleSvg = adyaContainer.innerHTML;

    const t1Group = cycleSvg.match(/<g[^>]*transform="translate\(150,\s*100\)"[^>]*>([\s\S]*?)<\/g>/);
    const t3Group = cycleSvg.match(/<g[^>]*transform="translate\(450,\s*260\)"[^>]*>([\s\S]*?)<\/g>/);
    assert(t1Group && t3Group, "Must render T1 and T3 node groups");
    assert(!t1Group[1].includes("<animate"), "T1 must NOT have pulse animation when not in cycle");
    assert(t3Group[1].includes("<animate"), "T3 must have pulse animation when participating in cycle");
    console.log("  ✔ 2.4: Dynamic cycle highlighting identifies actual participating cycle nodes");

    const arbitraryEdges = [
      { from: "WorkerAlpha-0", to: "WorkerBeta-0", type: "RW", item: "state" }
    ];
    vm.runInContext("renderPlaygroundAdya", context)(arbitraryEdges, "ANOMALY");
    const arbSvg = adyaContainer.innerHTML;
    assert(!arbSvg.includes("NaN"), "Arbitrary node layout must not produce NaN coordinates");
    assert(arbSvg.includes("WorkerAlpha") || arbSvg.includes("WorkerA"), "Arbitrary nodes must be rendered");
    console.log("  ✔ 2.5: Arbitrary node keys placed dynamically with valid geometric coordinates");
  }

  // =========================================================================
  // Test Suite 3: Worker Message Processing (handlePlaygroundWorkerMessage)
  // =========================================================================
  console.log("\n--> Test Suite 3: Worker Message Handler (handlePlaygroundWorkerMessage)");
  {
    const { context, doc } = createPlaygroundDOM();
    const handleMsg = vm.runInContext("handlePlaygroundWorkerMessage", context);

    const runBtn = doc.getElementById("pgRunBtn");
    const cancelBtn = doc.getElementById("pgCancelBtn");
    runBtn.disabled = true;
    cancelBtn.style.display = "inline-flex";

    handleMsg({ type: "PROGRESS", status: "Cancelled" });
    assert.strictEqual(runBtn.disabled, false, "Run button must be re-enabled on Cancelled progress");
    assert.strictEqual(cancelBtn.style.display, "none", "Cancel button must be hidden on Cancelled progress");
    console.log("  ✔ 3.1: PROGRESS envelope (Cancelled) resets execution buttons and state");

    handleMsg({ type: "CYCLE_DETECTED", anomalyType: "A5B_WRITE_SKEW" });
    const anomalyMetric = doc.getElementById("pgMetricAnomaly");
    const cycleMetric = doc.getElementById("pgMetricCycle");
    assert.strictEqual(anomalyMetric.textContent, "A5B_WRITE_SKEW", "Must update anomaly metric on CYCLE_DETECTED");
    assert(cycleMetric.textContent.includes("CICLO") || cycleMetric.textContent.includes("CYCLE"), "Must update cycle metric on CYCLE_DETECTED");
    console.log("  ✔ 3.2: CYCLE_DETECTED envelope updates anomaly badge and cycle detection metrics");

    handleMsg({
      type: "COMPLETE",
      report: {
        totalOps: 42,
        reducedOps: 6,
        anomalyType: "G1c",
        durationMs: 128,
        adyaEdges: [{ from: "T1-0", to: "T2-0", type: "RW", item: "q" }],
        trace: [{ worker_id: 1, type: "EXEC", sql: "SELECT 1;" }]
      }
    });

    const opsMetric = doc.getElementById("pgMetricOps");
    const durMetric = doc.getElementById("pgMetricDuration");
    const gantt = doc.getElementById("pgGanttContainer");
    const adya = doc.getElementById("pgAdyaGraphContent");

    assert(opsMetric.textContent.includes("42 ops"), "Must update ops metric from report");
    assert(durMetric.textContent.includes("128ms"), "Must update duration metric from report");
    assert(gantt.innerHTML.includes("Worker #1") || gantt.innerHTML.includes("Worker 1"), "COMPLETE must trigger Gantt render");
    assert(adya.innerHTML.includes("T1") && adya.innerHTML.includes("T2"), "COMPLETE must trigger Adya render");
    console.log("  ✔ 3.3: COMPLETE envelope parses report and dispatches metrics and visualizers");

    handleMsg({ type: "ERROR", error: "WebAssembly execution aborted" });
    const alertBox = doc.getElementById("pgAlertBox");
    assert(alertBox.textContent.includes("WebAssembly execution aborted"), "ERROR must display message in alert box");
    assert.strictEqual(alertBox.style.display, "block", "ERROR must show alert box");
    console.log("  ✔ 3.4: ERROR envelope updates status badge and displays error alert box");
  }

  // =========================================================================
  // Test Suite 4: i18n Completeness & Language Switching
  // =========================================================================
  console.log("\n--> Test Suite 4: i18n Completeness & Dynamic Language Switching");
  {
    const { context, doc, window: win } = createPlaygroundDOM();
    const I18N = vm.runInContext("I18N", context);

    assert(I18N && I18N.pt && I18N.en, "I18N must define pt and en");
    assert(I18N.pt.playground && I18N.en.playground, "I18N must define playground in pt and en");

    const ptKeys = Object.keys(I18N.pt.playground).sort();
    const enKeys = Object.keys(I18N.en.playground).sort();
    assert.deepStrictEqual(ptKeys, enKeys, "I18N.pt.playground and I18N.en.playground must have identical keys");

    const requiredKeys = [
      "cycleDetected", "okSerializable", "adyaCycleClassified",
      "adyaPlaceholder", "emptyGantt", "legendRw", "legendWw", "legendWr"
    ];
    for (const key of requiredKeys) {
      assert(ptKeys.includes(key), `Missing required key in playground i18n: ${key}`);
      assert.strictEqual(typeof I18N.pt.playground[key], "string", `Value for ${key} in PT must be string`);
      assert.strictEqual(typeof I18N.en.playground[key], "string", `Value for ${key} in EN must be string`);
      assert(I18N.pt.playground[key].length > 0, `Value for ${key} in PT must not be empty`);
      assert(I18N.en.playground[key].length > 0, `Value for ${key} in EN must not be empty`);
    }

    const ptRemnants = ["ciclo", "detectado", "serializável", "executar", "raias", "concorrente"];
    for (const [k, v] of Object.entries(I18N.en.playground)) {
      const lower = v.toLowerCase();
      for (const word of ptRemnants) {
        assert(!lower.includes(word), `English playground string "${k}" must not contain Portuguese word "${word}": got "${v}"`);
      }
    }
    console.log("  ✔ 4.1: Symmetrical playground i18n dictionary with no Portuguese in English translations");

    vm.runInContext('isWasmReady = true', context);
    vm.runInContext('currentRoute = "playground"', context);

    vm.runInContext('setLanguage("en")', context);
    assert.strictEqual(win.currentLang, "en", 'Window currentLang must be "en"');

    const wasmStatusText = doc.getElementById("pgWasmStatusText");
    assert.strictEqual(wasmStatusText.textContent, I18N.en.playground.statusReady, "Status text must update to English");

    vm.runInContext('setLanguage("pt")', context);
    assert.strictEqual(win.currentLang, "pt", 'Window currentLang must be "pt"');
    assert.strictEqual(wasmStatusText.textContent, I18N.pt.playground.statusReady, "Status text must update to Portuguese");
    console.log("  ✔ 4.2: setLanguage() dynamically re-renders playground UI via updatePlaygroundTranslations()");
  }

  console.log("\n✔ All ChaosSQL Playground UI behavioral tests passed successfully.");
}

runTests().catch((err) => {
  console.error("\n❌ Playground UI test failure:", err.message);
  if (err.stack) console.error(err.stack);
  process.exit(1);
});
