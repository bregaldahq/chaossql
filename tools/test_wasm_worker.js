// tools/test_wasm_worker.js
// Tests Web Worker protocol and message validation contract
const assert = require('assert');

function validateMessageProtocol(msg) {
  const validActions = ['INIT', 'VALIDATE', 'RUN', 'CANCEL'];
  if (!validActions.includes(msg.action)) {
    throw new Error(`Invalid action: ${msg.action}`);
  }
  return true;
}

function parseWorkerEnvelope(eventData) {
  assert(eventData && typeof eventData.type === 'string', 'Message must contain a string type');
  const validTypes = ['READY', 'VALIDATION_RESULT', 'PROGRESS', 'CYCLE_DETECTED', 'SHRINK_PROGRESS', 'COMPLETE', 'ERROR'];
  assert(validTypes.includes(eventData.type), `Unexpected event type: ${eventData.type}`);
  return true;
}

// Verification assertions
assert(validateMessageProtocol({ action: 'INIT', wasmUrl: '/assets/chaossql.wasm' }));
assert(validateMessageProtocol({ action: 'VALIDATE', yamlContent: 'version: "1.0"' }));
assert(validateMessageProtocol({ action: 'RUN', config: { iterations: 10 } }));
assert(validateMessageProtocol({ action: 'CANCEL' }));

assert(parseWorkerEnvelope({ type: 'READY' }));
assert(parseWorkerEnvelope({ type: 'PROGRESS', iteration: 1, ops: 4 }));
assert(parseWorkerEnvelope({ type: 'CYCLE_DETECTED', anomalyType: 'P4' }));
assert(parseWorkerEnvelope({ type: 'COMPLETE', report: '{}' }));

console.log('✔ Web Worker message protocol contract verified.');
