// tools/test_wasm_bench.js
// Automated verification for site/assets/wasm-bench.js

const assert = require("assert");
const bench = require("../site/assets/wasm-bench.js");

async function runTests() {
  console.log("Running wasm-bench.js test suite...");

  assert(typeof bench.ChaosSQL_RunStressTest === "function", "ChaosSQL_RunStressTest must be exported");
  assert(Array.isArray(bench.BENCH_PRESETS), "BENCH_PRESETS must be exported array");
  assert.strictEqual(bench.BENCH_PRESETS.length, 4, "Must define 4 canonical presets");

  // Test with Mock Worker
  class MockWorker {
    constructor() {
      this.listeners = [];
      this.posted = [];
    }
    addEventListener(evt, fn) {
      if (evt === "message") this.listeners.push(fn);
    }
    removeEventListener(evt, fn) {
      if (evt === "message") {
        this.listeners = this.listeners.filter(f => f !== fn);
      }
    }
    postMessage(data) {
      this.posted.push(data);
      if (data.action === "RUN") {
        setTimeout(() => {
          for (const fn of [...this.listeners]) {
            fn({
              data: {
                type: "COMPLETE",
                report: JSON.stringify({
                  violationFound: true,
                  anomalyType: "P4_LOST_UPDATE",
                  totalOps: 10,
                  adyaEdges: [{ from: "T1", to: "T2", type: "RW" }]
                })
              }
            });
          }
        }, 5);
      }
    }
    terminate() {}
  }

  const mockWorker = new MockWorker();
  const progressCalls = [];

  const report = await bench.ChaosSQL_RunStressTest(5, (prog) => {
    progressCalls.push(prog);
  }, { worker: mockWorker });

  assert.strictEqual(report.iterations, 5, "Must report 5 iterations");
  assert.strictEqual(report.completedRuns, 5, "Must complete 5 runs");
  assert.strictEqual(report.totalOperations, 50, "5 runs * 10 ops = 50 ops");
  assert.strictEqual(progressCalls.length, 5, "Must receive 5 progress callbacks");
  assert.strictEqual(progressCalls[0].currentIteration, 1);
  assert.strictEqual(progressCalls[4].currentIteration, 5);
  assert(report.throughputOpsPerSec > 0, "Throughput ops/sec must be positive");
  assert(report.frameMetrics.totalFrames >= 0, "Frame metrics must be computed");
  assert(typeof report.frameMetrics.fps60Compliant === "boolean", "fps60Compliant must be boolean");

  console.log("  ✔ wasm-bench.js: ChaosSQL_RunStressTest correctly coordinates runs, progress callbacks, and frame metrics");
  console.log("\n✔ All wasm-bench.js tests passed successfully.");
}

runTests().catch(err => {
  console.error("❌ Test failed:", err);
  process.exit(1);
});
