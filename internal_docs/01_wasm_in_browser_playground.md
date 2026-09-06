# Technical Architecture: In-Browser WebAssembly (WASM) Playground

- **Project**: ChaosSQL v1.3 Roadmap
- **Module**: In-Browser Client-Side Engine (`chaossql-wasm`)
- **Status**: Internal Technical Specification
- **Target**: `chaossql.bregalda.com/#/playground`

---

## 1. Overview & Motivation

Previously, running concurrency test scenarios with ChaosSQL required a local environment with Go installed (`go install`) or Docker.

Compiling ChaosSQL to **WebAssembly (WASM)** leverages one of the core architectural decisions of the project:
**Zero CGO dependencies and exclusive use of pure Go libraries (notably `modernc.org/sqlite`).**

This allows ChaosSQL to compile directly to the `js/wasm` architecture (`GOOS=js GOARCH=wasm`), enabling **the entire concurrency pipeline, micro-jitter fuzzer, Adya DSG cycle analyzer, and $ddmin$ shrinker to execute 100% within the client browser**, inside a serverless sandbox.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ CLIENT BROWSER (chaossql.bregalda.com)                                          │
│                                                                                 │
│  ┌───────────────────────┐              ┌────────────────────────────────────┐  │
│  │ Main UI Thread        │              │ Dedicated Web Worker               │  │
│  │ (React / Vanilla DOM) │ postMessage  │ (chaossql.wasm ~2.1MB gzipped)     │  │
│  │                       │ ───────────► │                                    │  │
│  │ • CodeMirror / Monaco │              │ • Go Runtime (wasm_exec.js)        │  │
│  │ • SVG Adya Graph      │ ◄─────────── │ • In-Memory SQLite (modernc)       │  │
│  │ • Gantt Swimlanes     │ Stream Evts  │ • Micro-Jitter Fuzzer Scheduler    │  │
│  │ • Metric Counters     │              │ • Adya Cycle Classifier            │  │
│  └───────────────────────┘              │ • Causal ddmin Shrinker            │  │
│                                         └────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. WASM Subsystem Architecture

### 2.1 Go Entry Point (`cmd/chaossql-wasm/main.go`)
The WASM binary exports functions to the global JavaScript scope (`globalThis.ChaosSQL`):

```go
// cmd/chaossql-wasm/main.go
package main

import (
	"context"
	"encoding/json"
	"syscall/js"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/shrinker"
)

func main() {
	c := make(chan struct{})

	// Register API on JavaScript global scope
	js.Global().Set("ChaosSQL_RunScenario", js.FuncOf(runScenarioWrapper))
	js.Global().Set("ChaosSQL_ValidateYAML", js.FuncOf(validateYAMLWrapper))
	js.Global().Set("ChaosSQL_GetVersion", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "1.2.0-wasm"
	}))

	// Keep main goroutine alive
	<-c
}
```

### 2.2 Web Worker Isolation
Concurrent database execution and compute-intensive $ddmin$ reduction loops must never block the main UI thread (which must sustain 60 FPS with smooth rendering).

1. The script `site/assets/wasm-worker.js` instantiates the Go runtime (`wasm_exec.js`).
2. The worker loads and compiles the streamed `chaossql.wasm` binary via `WebAssembly.instantiateStreaming`.
3. The worker communicates with the DOM through a typed JSON messaging protocol via `postMessage`.

#### Messaging Protocol (`postMessage`)

| Action (Request) | Payload | Worker Response (Stream Event) |
| :--- | :--- | :--- |
| `INIT` | `{ wasmUrl: '/chaossql.wasm' }` | `{ type: 'READY' }` |
| `VALIDATE` | `{ yamlContent: string }` | `{ type: 'VALIDATION_RESULT', valid: bool, errors: string[] }` |
| `RUN` | `{ yamlContent: string, workers: int, durationMs: int, seed: int }` | 1. `{ type: 'PROGRESS', iteration: int, ops: int, anomaliesFound: int }`<br>2. `{ type: 'CYCLE_DETECTED', anomaly: 'P4', edges: [...] }`<br>3. `{ type: 'SHRINK_PROGRESS', step: int, opsRemaining: int }`<br>4. `{ type: 'COMPLETE', trace: JSON, dsg: SVG/JSON, minimalOps: [...] }` |
| `TERMINATE` | `{}` | Cancels context via `context.WithCancel` |

---

## 3. Memory Management & Virtual SQLite Storage

The pure-Go SQLite driver (`modernc.org/sqlite`) operates seamlessly on top of the in-memory VFS (Virtual File System) within the Go WASM runtime.

- **In-Memory DSN**: Each execution uses `file::memory:?cache=shared&mode=memory` or a unique virtual database identifier `file:chaos_test_uuid.db?mode=memory`.
- **In-Browser Garbage Collection**: At the conclusion of each iteration, `driver.Reset()` and `PRAGMA drop_tables` deallocate all database B-Trees from WASM memory.
- **Memory Ceiling**: The Go WASM module is configured with an initial heap of 16MB expandable up to 128MB, preventing browser memory leaks.

---

## 4. Binary Size Optimization (Bundle Budget)

Go binaries compiled to WASM can range between 8MB and 15MB if unoptimized. For production web delivery, the following compression pipeline is applied:

1. **Go Compiler Flags**:
   ```bash
   GOOS=js GOARCH=wasm go build -ldflags="-s -w -X main.version=1.2.0" -trimpath -o site/chaossql.wasm ./cmd/chaossql-wasm
   ```
   *Effect*: Reduces uncompressed size from ~18MB to ~7.2MB.
2. **WebAssembly Optimizer (`wasm-opt`)**:
   ```bash
   wasm-opt -Oz --enable-bulk-memory --enable-mutable-globals site/chaossql.wasm -o site/chaossql.opt.wasm
   ```
   *Effect*: Reduces binary size to ~5.4MB.
3. **Brotli / Gzip Compression on Cloudflare Pages**:
   - Brotli level 11: **~1.7 MB** transferred across the wire.
   - The browser downloads the binary once and caches it in `CacheStorage` / IndexedDB via a ServiceWorker.

---

## 5. Three-Phase Implementation Roadmap

- **Phase 1 (Core Bridge)**:
  - Create package `cmd/chaossql-wasm/main.go`.
  - Integrate `modernc.org/sqlite` with verified `GOOS=js GOARCH=wasm` support.
  - Export synchronous validation and execution handlers.
- **Phase 2 (Web Worker & Streaming UI)**:
  - Implement `site/assets/wasm-worker.js` and `site/assets/wasm_exec.js`.
  - Implement asynchronous streaming via progress events.
  - Connect to the `#/playground` view on the web portal.
- **Phase 3 (Editor & Interactive Visualization)**:
  - Monaco / CodeMirror editor with syntax highlighting for YAML and SQL.
  - Pre-load interactive presets with all 10 standard race scenarios.
  - Render instant SVG Adya Graph visualizations directly on screen.
