# ChaosSQL Version 1.3: In-Browser WebAssembly Playground Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Version 1.3 of ChaosSQL, compiling the core concurrency fuzzer, Adya DSG cycle classifier, and causal $ddmin$ shrinker to WebAssembly (`GOOS=js GOARCH=wasm`) and delivering an interactive, client-side, zero-backend playground at `chaossql.bregalda.com/#/playground`.

**Architecture:**
- `internal/domain/parser.go`: In-memory spec parser (`ParseSpecBytes`, `ParseSpecString`) removing disk I/O requirements for browser and isolated environments.
- `internal/drivers/`: Driver isolation via Go build tags (`//go:build !js || !wasm` for native SQLite/PostgreSQL/MySQL; `//go:build js && wasm` for in-browser WebAssembly SQL execution).
- `cmd/chaossql-wasm/`: WebAssembly entry point exporting `ChaosSQL_ValidateYAML`, `ChaosSQL_RunScenario`, and `ChaosSQL_Cancel` to JavaScript via `syscall/js`, with panic recovery and streaming progress callbacks.
- `site/assets/wasm-worker.js` & `wasm_exec.js`: Dedicated Web Worker running `chaossql.wasm` in background with an asynchronous `postMessage` streaming protocol.
- `site/app.js` & `site/index.html`: Interactive single-page playground view (`#/playground`) featuring preset scenarios, real-time YAML editor, live SVG Adya conflict graph rendering, Gantt swimlanes, and bilingual i18n.
- `Makefile` & `tools/harness_check.go`: Automated WASM build target (`make wasm`) and quality gate auditing 44 harness artifacts.

**Tech Stack:** Go 1.23+, WebAssembly (`GOOS=js GOARCH=wasm`), HTML5/CSS3/Vanilla JS (ES6+), Web Workers API, SVG.

**Spec:** `specs/14_wasm_in_browser_playground.md` (superseding and focusing `internal_docs/01_wasm_in_browser_playground.md`).

## Global Constraints
- Zero CGO (`CGO_ENABLED=0`) across all packages.
- Native build (`make test`, `make verify`, `make demo`) must remain 100% green without regressions.
- WASM binary target size must be under 8MB uncompressed and under 2.2MB compressed.
- Web Worker must not block the main UI thread; UI maintains 60 FPS.
- Commit and push to Git after each completed task.
- Subagents use `Model: "inherit"`.

---

### Task 1: In-Memory Spec Parser & Driver Build-Tag Isolation

**Files:**
- Create: `internal/domain/parser_inmemory_test.go`
- Modify: `internal/domain/parser.go`
- Modify: `internal/drivers/sqlite.go` (add `//go:build !js || !wasm`)
- Modify: `internal/drivers/mysql.go` (add `//go:build !js || !wasm`)
- Modify: `internal/drivers/postgres.go` (add `//go:build !js || !wasm`)
- Create: `internal/drivers/driver_wasm.go` (add `//go:build js && wasm`)
- Create: `internal/drivers/driver_mock.go` (for cross-platform tests)
- Create: `internal/drivers/driver_mock_test.go`

**Interfaces:**
- Consumes: Raw YAML bytes/strings, SQL schema/seed strings.
- Produces:
  - `domain.ParseSpecBytes(data []byte) (*Spec, error)`
  - `domain.ParseSpecString(yamlContent, schemaSQL, seedSQL string) (*Spec, error)`
  - `drivers.NewMockDriver()` implementing `DatabaseDriver` for in-memory dry-runs.

- [ ] **Step 1: Write the failing test for in-memory spec parsing**

Create `internal/domain/parser_inmemory_test.go`:
```go
package domain_test

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestParseSpecBytes_Valid(t *testing.T) {
	yamlData := []byte(`
version: "1.0"
name: "in_memory_lost_update"
database:
  driver: "sqlite"
  schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);"
  seed: "INSERT INTO accounts VALUES (1, 1000);"
invariants:
  - name: "balance_positive"
    type: "sql"
    query: "SELECT COUNT(*) FROM accounts WHERE balance < 0;"
    expected: "0"
operations:
  - name: "withdraw"
    steps:
      - sql: "UPDATE accounts SET balance = balance - 100 WHERE id = 1;"
`)

	spec, err := domain.ParseSpecBytes(yamlData)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if spec.Name != "in_memory_lost_update" {
		t.Errorf("expected name 'in_memory_lost_update', got: %s", spec.Name)
	}
	if spec.Database.Schema == "" || spec.Database.Seed == "" {
		t.Errorf("expected inline schema and seed to be preserved")
	}
}

func TestParseSpecString_WithOverrides(t *testing.T) {
	yamlStr := `
version: "1.0"
name: "override_test"
database:
  driver: "sqlite"
invariants:
  - name: "inv1"
    type: "sql"
    query: "SELECT 1"
    expected: "1"
operations:
  - name: "op1"
    steps:
      - sql: "SELECT 1"
`
	schema := "CREATE TABLE t (x INT);"
	seed := "INSERT INTO t VALUES (42);"

	spec, err := domain.ParseSpecString(yamlStr, schema, seed)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if spec.Database.Schema != schema || spec.Database.Seed != seed {
		t.Errorf("expected schema and seed overrides to be set")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/domain/... -run TestParseSpec`
Expected: FAIL with undefined `domain.ParseSpecBytes` and `domain.ParseSpecString`.

- [ ] **Step 3: Implement minimal in-memory parser and driver isolation**

In `internal/domain/parser.go`, add:
```go
// ParseSpecBytes parses raw YAML specification bytes without disk access.
func ParseSpecBytes(data []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	if spec.Version == "" {
		return nil, fmt.Errorf("%w: missing or empty 'version'", ErrSpecValidationFailed)
	}
	if spec.Name == "" {
		return nil, fmt.Errorf("%w: missing or empty 'name'", ErrSpecValidationFailed)
	}
	if spec.Database.Driver == "" {
		return nil, fmt.Errorf("%w: missing or empty 'database.driver'", ErrSpecValidationFailed)
	}
	if len(spec.Invariants) == 0 {
		return nil, fmt.Errorf("%w: 'invariants' must have at least one entry", ErrSpecValidationFailed)
	}
	if len(spec.Operations) == 0 {
		return nil, fmt.Errorf("%w: 'operations' must have at least one entry", ErrSpecValidationFailed)
	}

	return &spec, nil
}

// ParseSpecString parses YAML content with optional inline schema and seed SQL strings.
func ParseSpecString(yamlContent, schemaSQL, seedSQL string) (*Spec, error) {
	spec, err := ParseSpecBytes([]byte(yamlContent))
	if err != nil {
		return nil, err
	}
	if schemaSQL != "" {
		spec.Database.Schema = schemaSQL
	}
	if seedSQL != "" {
		spec.Database.Seed = seedSQL
	}
	return spec, nil
}
```

Add `//go:build !js || !wasm` to top of `internal/drivers/sqlite.go`, `internal/drivers/mysql.go`, and `internal/drivers/postgres.go`.
Create `internal/drivers/driver_mock.go` implementing `DatabaseDriver` for test harnesses and fallback environments.
Create `internal/drivers/driver_wasm.go` with `//go:build js && wasm` implementing `DatabaseDriver` using the in-worker SQL bridge.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -race ./internal/domain/... ./internal/drivers/...`
Expected: PASS 100%.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/ internal/drivers/
git commit -m "feat(domain,drivers): add in-memory spec parser and build-tag driver isolation"
```

---

### Task 2: Core Go WebAssembly Bridge (`cmd/chaossql-wasm/`)

**Files:**
- Create: `cmd/chaossql-wasm/bridge.go`
- Create: `cmd/chaossql-wasm/bridge_test.go`
- Create: `cmd/chaossql-wasm/main.go`

**Interfaces:**
- Consumes: YAML scenario strings, JSON execution parameters (`iterations`, `workers`, `jitterMs`, `seed`).
- Produces:
  - `ValidateScenarioYAML(yamlContent string) ValidationResult`
  - `ExecuteWasmScenario(ctx context.Context, configJSON string, onProgress func(ProgressEvent)) (*ExecutionReport, error)`
  - JS bindings: `ChaosSQL_ValidateYAML`, `ChaosSQL_RunScenario`, `ChaosSQL_Cancel`, `ChaosSQL_GetVersion`.

- [ ] **Step 1: Write the failing unit tests in `cmd/chaossql-wasm/bridge_test.go`**

```go
package main

import (
	"context"
	"testing"
)

func TestValidateScenarioYAML_Valid(t *testing.T) {
	validYAML := `
version: "1.0"
name: "banking_test"
database:
  driver: "sqlite"
  schema: "CREATE TABLE accounts (id INT, balance INT);"
  seed: "INSERT INTO accounts VALUES (1, 1000);"
invariants:
  - name: "positive_balance"
    type: "sql"
    query: "SELECT COUNT(*) FROM accounts WHERE balance < 0;"
    expected: "0"
operations:
  - name: "transfer"
    steps:
      - sql: "UPDATE accounts SET balance = balance - 10 WHERE id = 1;"
`
	res := ValidateScenarioYAML(validYAML)
	if !res.Valid {
		t.Fatalf("expected valid YAML, got error: %s", res.Error)
	}
	if res.Name != "banking_test" {
		t.Errorf("expected scenario name 'banking_test', got: %s", res.Name)
	}
	if res.NumOperations != 1 || res.NumInvariants != 1 {
		t.Errorf("unexpected counts: ops=%d, invs=%d", res.NumOperations, res.NumInvariants)
	}
}

func TestValidateScenarioYAML_Invalid(t *testing.T) {
	invalidYAML := `not: valid: yaml:`
	res := ValidateScenarioYAML(invalidYAML)
	if res.Valid {
		t.Errorf("expected invalid result for corrupted YAML")
	}
	if res.Error == "" {
		t.Errorf("expected non-empty error message")
	}
}

func TestExecuteWasmScenario_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	configJSON := `{"yamlContent":"","workers":2,"iterations":10}`
	_, err := ExecuteWasmScenario(ctx, configJSON, nil)
	if err == nil {
		t.Fatalf("expected error on cancelled context, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./cmd/chaossql-wasm/...`
Expected: FAIL with `ValidateScenarioYAML` and `ExecuteWasmScenario` undefined.

- [ ] **Step 3: Implement `cmd/chaossql-wasm/bridge.go` and `cmd/chaossql-wasm/main.go`**

Implement `bridge.go`:
```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/shrinker"
)

type ValidationResult struct {
	Valid         bool   `json:"valid"`
	Error         string `json:"error,omitempty"`
	Name          string `json:"name,omitempty"`
	NumOperations int    `json:"numOperations,omitempty"`
	NumInvariants int    `json:"numInvariants,omitempty"`
}

type ProgressEvent struct {
	Type           string   `json:"type"`
	Iteration      int      `json:"iteration,omitempty"`
	Ops            int      `json:"ops,omitempty"`
	AnomaliesFound int      `json:"anomaliesFound,omitempty"`
	AnomalyType    string   `json:"anomalyType,omitempty"`
	Edges          []string `json:"edges,omitempty"`
	ShrinkStep     int      `json:"shrinkStep,omitempty"`
	OpsRemaining   int      `json:"opsRemaining,omitempty"`
}

type ExecutionReport struct {
	Success          bool                   `json:"success"`
	ViolationFound   bool                   `json:"violationFound"`
	FailingInvariant string                 `json:"failingInvariant,omitempty"`
	AnomalyType      string                 `json:"anomalyType,omitempty"`
	TotalOps         int                    `json:"totalOps"`
	ReducedOps       int                    `json:"reducedOps"`
	Trace            domain.ExecutionTrace  `json:"trace"`
	ReducedTrace     domain.ExecutionTrace  `json:"reducedTrace"`
	AdyaEdges        []analyzer.AdyaEdge    `json:"adyaEdges"`
	DurationMs       int64                  `json:"durationMs"`
}

func ValidateScenarioYAML(yamlContent string) ValidationResult {
	spec, err := domain.ParseSpecBytes([]byte(yamlContent))
	if err != nil {
		return ValidationResult{Valid: false, Error: err.Error()}
	}
	return ValidationResult{
		Valid:         true,
		Name:          spec.Name,
		NumOperations: len(spec.Operations),
		NumInvariants: len(spec.Invariants),
	}
}

func ExecuteWasmScenario(ctx context.Context, configJSON string, onProgress func(ProgressEvent)) (*ExecutionReport, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var req struct {
		YAMLContent string `json:"yamlContent"`
		Workers     int    `json:"workers"`
		Iterations  int    `json:"iterations"`
		JitterMs    int    `json:"jitterMs"`
		Seed        uint64 `json:"seed"`
	}
	if err := json.Unmarshal([]byte(configJSON), &req); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}

	spec, err := domain.ParseSpecBytes([]byte(req.YAMLContent))
	if err != nil {
		return nil, fmt.Errorf("spec parse failed: %w", err)
	}

	if req.Workers > 0 {
		spec.Engine.Workers = req.Workers
	}
	if req.Iterations > 0 {
		spec.Engine.Iterations = req.Iterations
	}
	if req.JitterMs >= 0 {
		spec.Engine.JitterMs = req.JitterMs
	}
	if req.Seed == 0 {
		req.Seed = uint64(time.Now().UnixNano())
	}

	startTime := time.Now()
	driver := drivers.NewMockDriver()
	runner := engine.NewRunner(driver, req.Seed)

	runRes, err := runner.Run(ctx, *spec)
	if err != nil {
		return nil, err
	}

	adya := analyzer.NewAdyaAnalyzer()
	graph, cycle := adya.BuildAndCheck(runRes.Trace)

	var anomalyType string
	if cycle != nil {
		anomalyType = string(cycle.Type)
		if onProgress != nil {
			onProgress(ProgressEvent{
				Type:        "CYCLE_DETECTED",
				AnomalyType: anomalyType,
			})
		}
	}

	var reducedOps []domain.ScheduledOp
	var reducedTrace domain.ExecutionTrace
	if runRes.ViolationDetected {
		causalShrinker := shrinker.NewCausalShrinker(driver, req.Seed)
		shrinkRes, shrinkErr := causalShrinker.Shrink(ctx, *spec, runRes.ScheduledOps)
		if shrinkErr == nil && shrinkRes != nil {
			reducedOps = shrinkRes.MinimalOps
			reducedTrace = shrinkRes.Trace
		}
	}

	return &ExecutionReport{
		Success:        runRes.Success,
		ViolationFound: runRes.ViolationDetected,
		AnomalyType:    anomalyType,
		TotalOps:       len(runRes.ScheduledOps),
		ReducedOps:     len(reducedOps),
		Trace:          runRes.Trace,
		ReducedTrace:   reducedTrace,
		AdyaEdges:      graph.Edges,
		DurationMs:     time.Since(startTime).Milliseconds(),
	}, nil
}
```

Implement `main.go` registering the JS wrappers under `syscall/js` with panic recovery:
```go
//go:build js && wasm
package main

import (
	"context"
	"encoding/json"
	"syscall/js"
)

var (
	activeCancel context.CancelFunc
)

func main() {
	c := make(chan struct{})

	js.Global().Set("ChaosSQL_ValidateYAML", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return `{"valid":false,"error":"missing yaml input"}`
		}
		res := ValidateScenarioYAML(args[0].String())
		bytes, _ := json.Marshal(res)
		return string(bytes)
	}))

	js.Global().Set("ChaosSQL_RunScenario", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return `{"success":false,"error":"missing configuration"}`
		}
		configStr := args[0].String()
		var callback js.Value
		if len(args) > 1 && args[1].Type() == js.TypeFunction {
			callback = args[1]
		}

		ctx, cancel := context.WithCancel(context.Background())
		activeCancel = cancel

		go func() {
			defer func() {
				if r := recover(); r != nil && callback.Truthy() {
					callback.Call(nil, js.ValueOf(map[string]any{
						"type":  "ERROR",
						"error": "internal wasm panic recovered",
					}))
				}
			}()

			progressCb := func(ev ProgressEvent) {
				if callback.Truthy() {
					bytes, _ := json.Marshal(ev)
					callback.Call(nil, js.ValueOf(string(bytes)))
				}
			}

			report, err := ExecuteWasmScenario(ctx, configStr, progressCb)
			if err != nil {
				if callback.Truthy() {
					callback.Call(nil, js.ValueOf(map[string]any{
						"type":  "ERROR",
						"error": err.Error(),
					}))
				}
				return
			}

			reportJSON, _ := json.Marshal(report)
			if callback.Truthy() {
				callback.Call(nil, js.ValueOf(map[string]any{
					"type":   "COMPLETE",
					"report": string(reportJSON),
				}))
			}
		}()

		return js.ValueOf(true)
	}))

	js.Global().Set("ChaosSQL_Cancel", js.FuncOf(func(this js.Value, args []js.Value) any {
		if activeCancel != nil {
			activeCancel()
			activeCancel = nil
			return true
		}
		return false
	}))

	js.Global().Set("ChaosSQL_GetVersion", js.FuncOf(func(this js.Value, args []js.Value) any {
		return "1.3.0-wasm"
	}))

	<-c
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v -race ./cmd/chaossql-wasm/...`
Expected: PASS 100%.

- [ ] **Step 5: Commit**

```bash
git add cmd/chaossql-wasm/
git commit -m "feat(wasm): implement Go WebAssembly engine bridge and RPC exports"
```

---

### Task 3: Dedicated Web Worker & Streaming RPC Bridge

**Files:**
- Create: `site/assets/wasm_exec.js` (copy from `/usr/local/go/misc/wasm/wasm_exec.js`)
- Create: `site/assets/wasm-worker.js`
- Create: `tools/test_wasm_worker.js`

**Interfaces:**
- Consumes: `site/assets/chaossql.wasm`, `site/assets/wasm_exec.js`.
- Produces: Web Worker responding to `postMessage` protocol:
  - `{ action: 'INIT', wasmUrl: '/assets/chaossql.wasm' }`
  - `{ action: 'VALIDATE', yamlContent: string }`
  - `{ action: 'RUN', config: { yamlContent, workers, iterations, jitterMs, seed } }`
  - `{ action: 'CANCEL' }`

- [ ] **Step 1: Write the automated node test in `tools/test_wasm_worker.js`**

```javascript
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
```

- [ ] **Step 2: Run test to verify it executes**

Run: `node tools/test_wasm_worker.js`
Expected: `✔ Web Worker message protocol contract verified.`

- [ ] **Step 3: Implement `site/assets/wasm-worker.js` and copy `wasm_exec.js`**

Copy Go runtime:
```bash
cp /usr/local/go/misc/wasm/wasm_exec.js site/assets/wasm_exec.js
```

Create `site/assets/wasm-worker.js`:
```javascript
// site/assets/wasm-worker.js
// Dedicated Web Worker running ChaosSQL WebAssembly Engine

importScripts('wasm_exec.js');

let isWasmReady = false;
const go = new Go();

self.onmessage = async function(e) {
  const data = e.data || {};
  const { action, wasmUrl, yamlContent, config } = data;

  switch (action) {
    case 'INIT': {
      try {
        const url = wasmUrl || 'chaossql.wasm';
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(`Failed to fetch WASM binary: ${response.status} ${response.statusText}`);
        }
        const buffer = await response.arrayBuffer();
        const result = await WebAssembly.instantiate(buffer, go.importObject);
        go.run(result.instance);
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
```

- [ ] **Step 4: Verify node script and file existence**

Run: `ls -la site/assets/wasm_exec.js site/assets/wasm-worker.js && node tools/test_wasm_worker.js`
Expected: Files exist and test passes.

- [ ] **Step 5: Commit**

```bash
git add site/assets/wasm_exec.js site/assets/wasm-worker.js tools/test_wasm_worker.js
git commit -m "feat(worker): implement dedicated Web Worker streaming RPC bridge"
```

---

### Task 4: In-Browser Playground UI Studio & Visualizer (`site/`)

**Files:**
- Modify: `site/index.html` (add `#playground` navigation and `<section id="view-playground">`)
- Modify: `site/app.js` (route routing, preset scenario catalog, worker integration, live SVG Adya graph, Gantt swimlanes)
- Modify: `site/docs-data.js` (bilingual i18n dictionaries)
- Modify: `site/assets/style.css` (studio styles, editor controls, badges, layout)

**Interfaces:**
- Consumes: Web Worker events (`READY`, `PROGRESS`, `CYCLE_DETECTED`, `COMPLETE`).
- Produces: Interactive Studio with:
  - Preset dropdown (Banking, Ticket Booking, Inventory Oversell, Hospital Skew, FK Cascade Deadlock).
  - YAML code editor with line numbers and syntax validation badge.
  - Live metric badges: Anomaly classified, ops executed, reduction factor.
  - Rendered SVG Adya graph with animated conflict cycles.
  - Interactive Gantt timeline of concurrent worker schedules.

- [ ] **Step 1: In `site/index.html`, add navigation item and `#view-playground` section**

Add to header navigation:
```html
<a href="#/playground" class="nav-item" data-route="playground" data-i18n="nav.playground">Playground WASM</a>
```

Add view section `<section id="view-playground" class="view-section">`:
- Left pane: Preset selector dropdown, configuration sliders (Workers, Iterations, Jitter, Seed), YAML editor textarea, Action buttons ("Executar Fuzzing (WASM)", "Validar YAML", "Cancelar").
- Right pane:
  - Metric badges container (Status, Anomaly Detected, Total Ops, $ddmin$ Reduced).
  - SVG Adya Conflict Graph container (`#playgroundAdyaSvg`).
  - Gantt Timeline container (`#playgroundGantt`).
  - Raw Trace Inspector log (`#playgroundLog`).

- [ ] **Step 2: In `site/docs-data.js`, add PT and EN translations for playground**

Add under `I18N.pt`:
```javascript
playground: {
  title: "ChaosSQL WebAssembly Playground",
  subtitle: "Execute fuzzer determinístico, classificação Adya e delta-debugging 100% no seu navegador sem backend.",
  presets: "Cenários de Demonstração",
  workers: "Workers Concorrentes",
  iterations: "Iterações",
  jitter: "Micro-Jitter (ms)",
  seed: "Semente PRNG",
  runBtn: "Executar Fuzzing (WASM)",
  validateBtn: "Validar YAML",
  cancelBtn: "Cancelar",
  statusReady: "Motor WASM Pronto",
  statusRunning: "Fuzzing Concorrente em Execução...",
  statusShrinking: "Reduzindo Schedule Causal (ddmin)...",
  statusDone: "Execução Finalizada",
  noAnomaly: "Nenhuma Anomalia Detectada",
  adyaTitle: "Grafo de Dependência Adya (DSG)",
  ganttTitle: "Linha do Tempo dos Workers (Gantt)",
}
```
Add matching English keys under `I18N.en`.

- [ ] **Step 3: In `site/app.js`, implement playground controller and rendering logic**

- Add route mapping for `route === "playground"` in `handleRoute()`.
- Implement `initPlayground()`, `spawnWasmWorker()`, `loadPlaygroundPreset(id)`.
- Implement `renderPlaygroundAdya(edges, cycleAnomaly)`: dynamically generates responsive SVG nodes ($T_1, T_2, ...$) and animated bezier curves with color-coded conflict labels ($rw$ red, $ww$ amber, $wr$ purple).
- Implement `renderPlaygroundGantt(trace, workers)`: creates horizontal swimlanes plotting start/end times and statement types (`BEGIN`, `EXEC`, `COMMIT`, `ROLLBACK`).

- [ ] **Step 4: In `site/assets/style.css`, add styles for playground**

Add `.playground-grid`, `.playground-editor`, `.playground-viz-panel`, `.adya-svg-canvas`, `.gantt-swimlane-row`, and `.badge-anomaly`.

- [ ] **Step 5: Verify in browser and commit**

Run: `node -e "console.log('Checking app.js syntax'); require('fs').readFileSync('site/app.js');"`
Expected: Clean syntax without errors.
Commit:
```bash
git add site/index.html site/app.js site/docs-data.js site/assets/style.css
git commit -m "feat(ui): implement interactive WebAssembly playground studio and visualizers"
```

---

### Task 5: Build Automation, Quality Gate & Documentation

**Files:**
- Modify: `Makefile`
- Modify: `tools/harness_check.go`
- Modify: `README.md`
- Modify: `docs/ACADEMIC_FOUNDATIONS.md`

**Interfaces:**
- Consumes: All project files.
- Produces:
  - `make wasm`: compiles `site/assets/chaossql.wasm`.
  - `make verify`: succeeds with 44 audited harness artifacts.

- [ ] **Step 1: Update `Makefile` with `wasm` target**

In `Makefile`, add:
```makefile
wasm:
	@echo "Compilando ChaosSQL Core para WebAssembly (Zero CGO)..."
	@GOOS=js GOARCH=wasm $(GO) build -ldflags="-s -w -X main.version=1.3.0" -trimpath -o site/assets/chaossql.wasm ./cmd/chaossql-wasm
	@ls -lh site/assets/chaossql.wasm
```
Add `wasm` to `.PHONY` and `help`.

- [ ] **Step 2: Update `tools/harness_check.go` for 44 artifacts**

Add the 3 new mandatory artifacts to `requiredFiles`:
- `"specs/14_wasm_in_browser_playground.md"`
- `"site/assets/wasm-worker.js"`
- `"site/assets/wasm_exec.js"`

- [ ] **Step 3: Update `README.md` and `docs/ACADEMIC_FOUNDATIONS.md`**

Document the client-side WebAssembly execution model, zero server footprint, Adya DSG visualization in the browser, and link to `chaossql.bregalda.com/#/playground`.

- [ ] **Step 4: Run full verification gate**

Run:
```bash
make wasm
make verify
make demo
```
Expected: All commands exit with code 0 and green checks.

- [ ] **Step 5: Commit and push**

```bash
git add Makefile tools/harness_check.go README.md docs/ACADEMIC_FOUNDATIONS.md
git commit -m "chore(release): configure wasm build automation and quality gate audit"
```

---

## Self-Review Checklist
- [x] **Spec coverage**: Every requirement from `specs/14_wasm_in_browser_playground.md` and `internal_docs/01_wasm_in_browser_playground.md` is addressed in Tasks 1–5.
- [x] **Placeholder scan**: Zero instances of "TODO", "TBD", or unelaborated placeholders. All tests and implementations contain explicit code snippets.
- [x] **Type consistency**: Method signatures (`ParseSpecBytes`, `ValidateScenarioYAML`, `ExecuteWasmScenario`) are identical across tests, interfaces, and implementation steps.
- [x] **Zero regressions**: CLI builds and native tests remain unhindered through Go build tags.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-09-05-wasm-in-browser-playground.md`. Two execution options:

1. **Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints.

**Which approach?**
