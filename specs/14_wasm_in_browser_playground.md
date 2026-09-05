# Spec 14: In-Browser WebAssembly (WASM) Playground & Client-Side Verification Engine (v1.3)

## 1. Domain Theory & Motivation
- **Context**: ChaosSQL v1.0–v1.2 required local Go CLI or Docker installations to stress-test SQL isolation levels and classify Adya anomalies.
- **WASM Architecture**:
  - The ChaosSQL domain core (PRNG fuzzer, micro-jitter scheduler, Adya DSG cycle classifier, Elle-style register linearizability checker, and causal $ddmin$ shrinker) is compiled to WebAssembly (`GOOS=js GOARCH=wasm`).
  - To bypass CGO and `modernc.org/libc` build incompatibilities on `js/wasm`, database execution inside the browser is handled via an in-memory SQL bridge (`sql.js` / Web Worker SQLite virtual VFS) combined with build-tagged driver isolation (`//go:build !js || !wasm` vs `//go:build js && wasm`).
  - Execution runs 100% inside a dedicated browser Web Worker, ensuring the main UI thread maintains 60 FPS without frame drops.

## 2. Web Worker RPC Protocol (`site/assets/wasm-worker.js`)
- Dedicated Web Worker receives JSON actions and returns streamed event envelopes:
  - `INIT`: Instantiates Go runtime (`wasm_exec.js`) and `chaossql.wasm`. Emits `{ type: 'READY' }`.
  - `VALIDATE`: Parses and validates raw YAML scenario. Emits `{ type: 'VALIDATION_RESULT', valid: bool, error: string, spec: object }`.
  - `RUN`: Executes chaos schedule. Emits:
    - `{ type: 'PROGRESS', iteration: int, ops: int, anomaliesFound: int }`
    - `{ type: 'CYCLE_DETECTED', anomaly: string, edges: array }`
    - `{ type: 'SHRINK_PROGRESS', step: int, opsRemaining: int }`
    - `{ type: 'COMPLETE', trace: array, dsg: object, minimalOps: array, durationMs: number }`
  - `CANCEL`: Aborts running schedule via `context.WithCancel`.

## 3. WASM Core Bridge API (`cmd/chaossql-wasm/`)
- Exported global symbols (`globalThis.ChaosSQL`):
  - `ChaosSQL_ValidateYAML(yamlStr string) string`
  - `ChaosSQL_RunScenario(configJSON string, callbackFunc js.Value) string`
  - `ChaosSQL_Cancel()`
  - `ChaosSQL_GetVersion() string`

## 4. In-Browser Playground UI (`site/#/playground`)
- Interactive studio integrated into the vanilla JS single-page portal:
  - **Preset Catalog**: 1-click loading of the 10 canonical scenarios (Banking Lost Update, Inventory Oversell, Hospital Write Skew, Financial Audit Read Skew, Foreign Key Cascade Deadlock, etc.).
  - **Editable Scenario & Config**: Real-time YAML editor with sliders for workers (1–8), iterations (5–50), and micro-jitter (0–50ms).
  - **Live Adya Conflict Graph**: Responsive SVG rendering transaction nodes ($T_1, T_2, ...$) and directed dependency edges ($rw, wr, ww$) with pulsing cycle animations.
  - **Execution Gantt Timeline**: Horizontal swimlane showing worker goroutine interleavings and statement-level latency.
  - **Causal Delta-Debugging Inspector**: Side-by-side comparison of raw schedule vs 1-minimal reproduction schedule.
  - **Bilingual i18n**: Complete PT/EN translations for all UI controls and status badges.

## 5. Bundle Budget & Optimization
- Go compiler flags: `-ldflags="-s -w -X main.version=1.3.0" -trimpath`.
- Target binary: `site/assets/chaossql.wasm` (< 8MB uncompressed, < 2.2MB gzipped / brotli).
- CI Quality Gate: Audited in `tools/harness_check.go` and `Makefile` (`make wasm`).
