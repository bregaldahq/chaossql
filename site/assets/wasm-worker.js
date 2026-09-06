// site/assets/wasm-worker.js
// Dedicated Web Worker running ChaosSQL WebAssembly Engine

/* global importScripts, Go, WebAssembly */
importScripts('wasm_exec.js');

let isWasmReady = false;
let goInstance = null;

self.onmessage = async function(e) {
  const data = e.data || {};
  const { action, wasmUrl, yamlContent, config } = data;

  switch (action) {
    case 'INIT': {
      try {
        if (isWasmReady) {
          self.postMessage({ type: 'READY' });
          return;
        }
        goInstance = new Go();
        const url = wasmUrl || 'chaossql.wasm';
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error('Failed to fetch WASM binary: ' + response.status + ' ' + response.statusText);
        }

        let result;
        if (WebAssembly.instantiateStreaming) {
          const fallbackResp = response.clone();
          try {
            result = await WebAssembly.instantiateStreaming(response, goInstance.importObject);
          } catch (streamErr) {
            // Fallback to arrayBuffer if instantiateStreaming fails (e.g. content-type issues)
            const buffer = await fallbackResp.arrayBuffer();
            result = await WebAssembly.instantiate(buffer, goInstance.importObject);
          }
        } else {
          const buffer = await response.arrayBuffer();
          result = await WebAssembly.instantiate(buffer, goInstance.importObject);
        }

        Promise.resolve(goInstance.run(result.instance)).catch((err) => {
          self.postMessage({
            type: 'ERROR',
            error: 'Go runtime exited: ' + (err && err.message ? err.message : err),
          });
        });

        isWasmReady = true;
        self.postMessage({ type: 'READY' });
      } catch (err) {
        self.postMessage({ type: 'ERROR', error: 'WASM initialization failed: ' + err.message });
      }
      break;
    }

    case 'VALIDATE': {
      if (!isWasmReady || !self.ChaosSQL_ValidateYAML) {
        self.postMessage({ type: 'VALIDATION_RESULT', valid: false, error: 'WASM engine not ready' });
        return;
      }
      try {
        const rawRes = self.ChaosSQL_ValidateYAML(yamlContent || '');
        const res = JSON.parse(rawRes);
        self.postMessage({ type: 'VALIDATION_RESULT', ...res });
      } catch (err) {
        self.postMessage({ type: 'VALIDATION_RESULT', valid: false, error: err.message });
      }
      break;
    }

    case 'RUN': {
      if (!isWasmReady || !self.ChaosSQL_RunScenario) {
        self.postMessage({ type: 'ERROR', error: 'WASM engine not ready' });
        return;
      }
      try {
        const cfgStr = JSON.stringify(config || {});
        self.ChaosSQL_RunScenario(cfgStr, (eventData) => {
          if (typeof eventData === 'string') {
            try {
              self.postMessage(JSON.parse(eventData));
            } catch (_) {
              self.postMessage({ type: 'PROGRESS', raw: eventData });
            }
          } else {
            self.postMessage(eventData);
          }
        });
      } catch (err) {
        self.postMessage({ type: 'ERROR', error: 'Execution failed: ' + err.message });
      }
      break;
    }

    case 'CANCEL': {
      if (self.ChaosSQL_Cancel) {
        self.ChaosSQL_Cancel();
      }
      self.postMessage({ type: 'PROGRESS', status: 'Cancelled' });
      break;
    }

    default:
      self.postMessage({ type: 'ERROR', error: 'Unknown action: ' + action });
  }
};
