---
name: chaossql-wasm-playground
description: The in-browser engine — cmd/chaossql-wasm (js/wasm build, exported ChaosSQL_* globals, ExecuteWasmScenario on the mock driver), the Web Worker message protocol, the TypeScript bridge used by the portal playground, the three committed copies of the WASM artifacts, and the headless Node stress/protocol tests. Use when changing the WASM engine, the worker, the playground bridge, or make wasm / make stress-wasm.
---

# WASM Playground Engine

The portal's playground runs the Go engine as WebAssembly inside a Web Worker,
entirely client-side (spec 14).

## When to use

- Changing `cmd/chaossql-wasm`, `site/assets/wasm-worker.js`,
  `site/src/lib/wasm-bridge.ts`, `site/assets/wasm-bench.js`.
- Rebuilding or shipping `chaossql.wasm`.
- `make stress-wasm` or the Node WASM tests fail.

## Go side (`cmd/chaossql-wasm`)

`main.go` has the `js && wasm` build tag; `bridge.go` has none, so its logic
compiles and is tested natively (`bridge_test.go` runs in `make test`).

`main` registers globals and blocks forever:
- `ChaosSQL_ValidateYAML(yaml)` → JSON `{valid, error, name, numOperations, numInvariants}`.
- `ChaosSQL_RunScenario(configJSON, callback)` → returns `"OK"` immediately;
  cancels any previous run; runs `ExecuteWasmScenario` in a goroutine and calls
  `callback` with JSON strings: progress events (`CYCLE_DETECTED`, ...),
  `{"type":"COMPLETE","report":"<json>"}` or `{"type":"ERROR","error":...}`;
  panics are recovered into `ERROR`.
- `ChaosSQL_Cancel()` → `"true"`/`"false"`.
- `ChaosSQL_GetVersion()` → `version.Version + "-wasm"`.

`ExecuteWasmScenario` (`bridge.go`): config
`{yamlContent, workers, iterations, jitterMs, seed}`; seeds must be JavaScript
safe integers (≤ 2^53−1); `jitterMs` sets `jitter_ms = [0, jitterMs]`;
`ParseSpecBytes` (no `.sql` file resolution); **always `drivers.NewMockDriver()`**;
`runner.Run`; classification uses `cycles[0]`; all graph edges are reported;
shrink + verify on violation. Report: `success, violationFound, seed, schedule,
failingInvariant, anomalyType, totalOps, reducedOps, trace, reducedTrace,
adyaEdges, durationMs`.

## Worker protocol (`site/assets/wasm-worker.js`)

Messages `{action, ...}` in, `{type, ...}` out:
- `INIT {wasmUrl}` → loads `wasm_exec.js` via `importScripts`, fetches and
  instantiates the module, runs the Go program → `READY` or `ERROR`.
- `VALIDATE {yamlContent}` → `VALIDATION_RESULT`.
- `RUN {yamlContent, config}` → `PROGRESS {raw}` events, then the completion.
- `CANCEL` → `PROGRESS {status: 'Cancelled'}`.
- Unknown action → `ERROR`.

The portal bridge (`ChaosSqlWasmBridge`, singleton `getWasmBridge()`) creates
`new Worker('/wasm/wasm-worker.js')` and initializes with
`wasmUrl: '/wasm/chaossql.wasm'`, then maps reports into UI types
(`buildExecutionReport`).

## Artifacts (committed, three copies)

`make wasm` builds with `CGO_ENABLED=0 GOOS=js GOARCH=wasm -ldflags="-s -w" -trimpath`
into `site/assets/chaossql.wasm` and copies it to `site/public/wasm/` (Vite
public dir → `/wasm/`) and `site/wasm/` (served as-is because Pages deploys
the `site/` directory). The worker and `wasm_exec.js` also exist in all three
places and must stay identical.

## Gotchas

- Because of the mock driver every invariant sees zeros; playground results
  reflect the assertion's behavior on zeros plus the trace's Adya cycles, not a
  real isolation level (`chaossql-database-drivers`).
- `make wasm` does not refresh `wasm_exec.js`; the committed copy predates the
  Go 1.25 toolchain version. When rebuilding for a new Go version copy
  `$(go env GOROOT)/lib/wasm/wasm_exec.js` into all three locations.
- `make stress-wasm` builds a separate, unstripped `bin/chaossql-test.wasm`
  and never touches the committed artifacts.
- Only one run at a time: a new `RunScenario` cancels the previous one.

## Tests

- `make stress-wasm` → `CHAOSSQL_WASM_PATH=bin/chaossql-test.wasm node tools/headless_worker_stress.js`
  (50+ runs in a Node VM sandbox: memory growth, throughput, protocol).
- `node tools/test_wasm_worker.js` (worker protocol envelopes, fetch fallback).
- `node tools/test_wasm_bench.js` (`site/assets/wasm-bench.js`).
- `node tools/test_playground_ui.js` (playground behaviors).
- Go: `cmd/chaossql-wasm/bridge_test.go` (native).

## Source map

- `cmd/chaossql-wasm/main.go`
- `cmd/chaossql-wasm/bridge.go`
- `cmd/chaossql-wasm/bridge_test.go`
- `site/assets/wasm-worker.js`
- `site/assets/wasm_exec.js`
- `site/assets/wasm-bench.js`
- `site/assets/chaossql.wasm`
- `site/public/wasm/wasm-worker.js`
- `site/wasm/wasm-worker.js`
- `site/src/lib/wasm-bridge.ts`
- `tools/headless_worker_stress.js`
- `tools/test_wasm_worker.js`
- `tools/test_wasm_bench.js`
- `specs/14_wasm_in_browser_playground.md`

## Related skills

- `chaossql-website-portal`, `chaossql-database-drivers`, `chaossql-quality-gate`
