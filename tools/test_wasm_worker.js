// tools/test_wasm_worker.js
// Comprehensive Node.js vm sandbox tests for site/assets/wasm-worker.js
// Tests protocol envelopes, HTTP response status verification, single-fetch fallback, and Go runtime errors.

const fs = require('fs');
const path = require('path');
const vm = require('vm');
const assert = require('assert');

const workerPath = path.resolve(__dirname, '../site/assets/wasm-worker.js');
const workerCode = fs.readFileSync(workerPath, 'utf8');

function createWorkerContext(options = {}) {
  const postedMessages = [];
  const fetchCalls = [];
  let cloned = false;

  const mockResponse = {
    ok: options.fetchOk !== undefined ? options.fetchOk : true,
    status: options.fetchStatus !== undefined ? options.fetchStatus : 200,
    statusText: options.fetchStatusText !== undefined ? options.fetchStatusText : 'OK',
    clone() {
      cloned = true;
      return {
        ok: this.ok,
        status: this.status,
        statusText: this.statusText,
        arrayBuffer: async () => new ArrayBuffer(16),
      };
    },
    arrayBuffer: async () => new ArrayBuffer(16),
  };

  const mockFetch = async (url) => {
    fetchCalls.push(url);
    if (options.fetchThrows) {
      throw options.fetchThrows;
    }
    return mockResponse;
  };

  class MockGo {
    constructor() {
      this.importObject = { env: {} };
    }
    run(instance) {
      if (options.goRunReject) {
        return Promise.reject(options.goRunReject);
      }
      return new Promise(() => {});
    }
  }

  let mockWebAssembly = options.WebAssembly;
  if (!mockWebAssembly) {
    mockWebAssembly = {
      instantiateStreaming: options.streamUnsupported
        ? undefined
        : async (resp, imports) => {
            if (options.streamThrows) {
              throw options.streamThrows;
            }
            return { instance: { exports: {} } };
          },
      instantiate: async (buf, imports) => {
        if (options.instantiateThrows) {
          throw options.instantiateThrows;
        }
        return { instance: { exports: {} } };
      },
    };
  }

  const sandbox = {
    console,
    Promise,
    JSON,
    Error,
    TypeError,
    setTimeout,
    clearTimeout,
    importScripts: options.importScripts || ((...args) => {}),
    fetch: options.fetch || mockFetch,
    WebAssembly: mockWebAssembly,
    Go: options.Go || MockGo,
  };
  sandbox.self = sandbox;
  // Structured clone semantics across worker boundary
  sandbox.self.postMessage = (msg) => {
    postedMessages.push(JSON.parse(JSON.stringify(msg)));
  };

  vm.createContext(sandbox);
  vm.runInContext(workerCode, sandbox);

  return {
    sandbox,
    postedMessages,
    fetchCalls,
    getCloned: () => cloned,
    sendMessage: async (data) => {
      await sandbox.self.onmessage({ data });
    },
    sendRawEvent: async (event) => {
      await sandbox.self.onmessage(event);
    },
  };
}

// Protocol helper verification
function validateMessageProtocol(msg) {
  const validActions = ['INIT', 'VALIDATE', 'RUN', 'CANCEL'];
  if (!validActions.includes(msg.action)) {
    throw new Error(`Invalid action: ${msg.action}`);
  }
  return true;
}

function parseWorkerEnvelope(eventData) {
  assert(eventData && typeof eventData.type === 'string', 'Message must contain a string type');
  const validTypes = [
    'READY',
    'VALIDATION_RESULT',
    'PROGRESS',
    'CYCLE_DETECTED',
    'SHRINK_PROGRESS',
    'COMPLETE',
    'ERROR',
  ];
  assert(validTypes.includes(eventData.type), `Unexpected event type: ${eventData.type}`);
  return true;
}

async function runTests() {
  console.log('Running wasm-worker.js VM sandbox test suite...');

  // Baseline protocol asserts
  assert(validateMessageProtocol({ action: 'INIT', wasmUrl: '/assets/chaossql.wasm' }));
  assert(validateMessageProtocol({ action: 'VALIDATE', yamlContent: 'version: "1.0"' }));
  assert(validateMessageProtocol({ action: 'RUN', config: { iterations: 10 } }));
  assert(validateMessageProtocol({ action: 'CANCEL' }));

  assert(parseWorkerEnvelope({ type: 'READY' }));
  assert(parseWorkerEnvelope({ type: 'PROGRESS', iteration: 1, ops: 4 }));
  assert(parseWorkerEnvelope({ type: 'CYCLE_DETECTED', anomalyType: 'P4' }));
  assert(parseWorkerEnvelope({ type: 'COMPLETE', report: '{}' }));

  // Test 1: INIT - successful instantiateStreaming
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'INIT', wasmUrl: 'chaossql.wasm' });
    assert.strictEqual(ctx.fetchCalls.length, 1);
    assert.strictEqual(ctx.fetchCalls[0], 'chaossql.wasm');
    assert.deepStrictEqual(ctx.postedMessages, [{ type: 'READY' }]);
    parseWorkerEnvelope(ctx.postedMessages[0]);
    console.log('  ✔ Test 1: INIT with instantiateStreaming succeeded');
  }

  // Test 2: INIT - idempotent when already ready
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'INIT' });
    assert.deepStrictEqual(ctx.postedMessages, [{ type: 'READY' }]);
    await ctx.sendMessage({ action: 'INIT' });
    assert.strictEqual(ctx.fetchCalls.length, 1, 'Should not re-fetch WASM on subsequent INIT');
    assert.strictEqual(ctx.postedMessages.length, 2);
    assert.deepStrictEqual(ctx.postedMessages[1], { type: 'READY' });
    console.log('  ✔ Test 2: INIT is idempotent when already ready');
  }

  // Test 3: INIT - HTTP status failure throws descriptive error
  {
    const ctx = createWorkerContext({
      fetchOk: false,
      fetchStatus: 404,
      fetchStatusText: 'Not Found',
    });
    await ctx.sendMessage({ action: 'INIT' });
    assert.strictEqual(ctx.postedMessages.length, 1);
    assert.strictEqual(ctx.postedMessages[0].type, 'ERROR');
    assert.strictEqual(
      ctx.postedMessages[0].error,
      'WASM initialization failed: Failed to fetch WASM binary: 404 Not Found'
    );
    parseWorkerEnvelope(ctx.postedMessages[0]);
    console.log('  ✔ Test 3: INIT rejects non-ok HTTP responses');
  }

  // Test 4: INIT - streaming failure falls back to arrayBuffer using response.clone() (single fetch)
  {
    const ctx = createWorkerContext({
      streamThrows: new Error('instantiateStreaming failed'),
    });
    await ctx.sendMessage({ action: 'INIT' });
    assert.strictEqual(ctx.fetchCalls.length, 1, 'Must avoid double fetch by cloning response');
    assert.strictEqual(ctx.getCloned(), true, 'Must clone response before instantiateStreaming');
    assert.deepStrictEqual(ctx.postedMessages, [{ type: 'READY' }]);
    console.log('  ✔ Test 4: INIT falls back to cloned arrayBuffer on streaming error');
  }

  // Test 5: INIT - fallback when instantiateStreaming is unsupported
  {
    const ctx = createWorkerContext({ streamUnsupported: true });
    await ctx.sendMessage({ action: 'INIT' });
    assert.strictEqual(ctx.fetchCalls.length, 1);
    assert.deepStrictEqual(ctx.postedMessages, [{ type: 'READY' }]);
    console.log('  ✔ Test 5: INIT handles environments without instantiateStreaming');
  }

  // Test 6: Go runtime rejection posts ERROR envelope (Error instance & raw string)
  {
    const ctx = createWorkerContext({
      goRunReject: new Error('runtime exit code 1'),
    });
    await ctx.sendMessage({ action: 'INIT' });
    await new Promise((r) => setTimeout(r, 20));
    const errorMsg = ctx.postedMessages.find((m) => m.type === 'ERROR');
    assert(errorMsg, 'Expected ERROR envelope on Go runtime rejection');
    assert.strictEqual(errorMsg.error, 'Go runtime exited: runtime exit code 1');
    parseWorkerEnvelope(errorMsg);

    // Also test string rejection
    const ctx2 = createWorkerContext({
      goRunReject: 'abnormal termination',
    });
    await ctx2.sendMessage({ action: 'INIT' });
    await new Promise((r) => setTimeout(r, 20));
    const errorMsg2 = ctx2.postedMessages.find((m) => m.type === 'ERROR');
    assert(errorMsg2);
    assert.strictEqual(errorMsg2.error, 'Go runtime exited: abnormal termination');
    console.log('  ✔ Test 6: Go runtime rejection handled and posted to postMessage');
  }

  // Test 7: VALIDATE - when engine not ready
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'VALIDATE', yamlContent: 'version: "1.0"' });
    assert.deepStrictEqual(ctx.postedMessages, [
      { type: 'VALIDATION_RESULT', valid: false, error: 'WASM engine not ready' },
    ]);
    parseWorkerEnvelope(ctx.postedMessages[0]);
    console.log('  ✔ Test 7: VALIDATE before INIT returns error');
  }

  // Test 8: VALIDATE - when ready, handles valid and invalid YAML
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'INIT' });

    // Mock validate function returning success JSON
    ctx.sandbox.self.ChaosSQL_ValidateYAML = (yaml) => {
      assert.strictEqual(yaml, 'valid: yaml');
      return JSON.stringify({ valid: true, summary: 'Passed 4 invariants' });
    };
    await ctx.sendMessage({ action: 'VALIDATE', yamlContent: 'valid: yaml' });
    assert.deepStrictEqual(ctx.postedMessages[1], {
      type: 'VALIDATION_RESULT',
      valid: true,
      summary: 'Passed 4 invariants',
    });
    parseWorkerEnvelope(ctx.postedMessages[1]);

    // Mock validate function throwing exception
    ctx.sandbox.self.ChaosSQL_ValidateYAML = () => {
      throw new Error('YAML syntax error');
    };
    await ctx.sendMessage({ action: 'VALIDATE', yamlContent: 'bad: yaml' });
    assert.deepStrictEqual(ctx.postedMessages[2], {
      type: 'VALIDATION_RESULT',
      valid: false,
      error: 'YAML syntax error',
    });
    parseWorkerEnvelope(ctx.postedMessages[2]);
    console.log('  ✔ Test 8: VALIDATE parses and returns validation envelope');
  }

  // Test 9: RUN - when engine not ready
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'RUN', config: { iterations: 5 } });
    assert.deepStrictEqual(ctx.postedMessages, [
      { type: 'ERROR', error: 'WASM engine not ready' },
    ]);
    parseWorkerEnvelope(ctx.postedMessages[0]);
    console.log('  ✔ Test 9: RUN before INIT returns error');
  }

  // Test 10: RUN - when ready, streams events and handles exceptions
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'INIT' });

    ctx.sandbox.self.ChaosSQL_RunScenario = (cfgStr, cb) => {
      const cfg = JSON.parse(cfgStr);
      assert.strictEqual(cfg.iterations, 10);
      cb(JSON.stringify({ type: 'PROGRESS', iteration: 1, ops: 4 }));
      cb('unformatted progress text');
      cb({ type: 'COMPLETE', violations: 0 });
    };

    await ctx.sendMessage({ action: 'RUN', config: { iterations: 10 } });
    assert.deepStrictEqual(ctx.postedMessages.slice(1), [
      { type: 'PROGRESS', iteration: 1, ops: 4 },
      { type: 'PROGRESS', raw: 'unformatted progress text' },
      { type: 'COMPLETE', violations: 0 },
    ]);
    for (const msg of ctx.postedMessages.slice(1)) {
      parseWorkerEnvelope(msg);
    }

    // Test RUN throwing error
    ctx.sandbox.self.ChaosSQL_RunScenario = () => {
      throw new Error('Scheduler deadlock');
    };
    await ctx.sendMessage({ action: 'RUN', config: {} });
    const lastMsg = ctx.postedMessages[ctx.postedMessages.length - 1];
    assert.deepStrictEqual(lastMsg, {
      type: 'ERROR',
      error: 'Execution failed: Scheduler deadlock',
    });
    parseWorkerEnvelope(lastMsg);
    console.log('  ✔ Test 10: RUN streams parsed events and handles errors');
  }

  // Test 11: CANCEL - triggers ChaosSQL_Cancel and emits cancellation progress
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'INIT' });
    let cancelCalled = false;
    ctx.sandbox.self.ChaosSQL_Cancel = () => {
      cancelCalled = true;
    };
    await ctx.sendMessage({ action: 'CANCEL' });
    assert.strictEqual(cancelCalled, true, 'ChaosSQL_Cancel must be called');
    assert.deepStrictEqual(ctx.postedMessages[1], {
      type: 'PROGRESS',
      status: 'Cancelled',
    });
    parseWorkerEnvelope(ctx.postedMessages[1]);
    console.log('  ✔ Test 11: CANCEL invokes cancel handler and emits PROGRESS envelope');
  }

  // Test 12: Unknown action returns ERROR envelope
  {
    const ctx = createWorkerContext();
    await ctx.sendMessage({ action: 'NON_EXISTENT_ACTION' });
    assert.deepStrictEqual(ctx.postedMessages, [
      { type: 'ERROR', error: 'Unknown action: NON_EXISTENT_ACTION' },
    ]);
    parseWorkerEnvelope(ctx.postedMessages[0]);
    console.log('  ✔ Test 12: Unknown action returns ERROR envelope');
  }

  // Test 13: Null or missing event data handled gracefully
  {
    const ctx = createWorkerContext();
    await ctx.sendRawEvent({});
    assert.deepStrictEqual(ctx.postedMessages, [
      { type: 'ERROR', error: 'Unknown action: undefined' },
    ]);

    const ctx2 = createWorkerContext();
    await ctx2.sendRawEvent({ data: null });
    assert.deepStrictEqual(ctx2.postedMessages, [
      { type: 'ERROR', error: 'Unknown action: undefined' },
    ]);
    console.log('  ✔ Test 13: Graceful handling of empty or null event payloads');
  }

  console.log('\n✔ All wasm-worker.js VM sandbox tests passed successfully.');
}

runTests().catch((err) => {
  console.error('\n❌ Test failure:', err);
  process.exit(1);
});
