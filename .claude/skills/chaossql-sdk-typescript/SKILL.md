---
name: chaossql-sdk-typescript
description: The TypeScript/Node SDK (@chaossql/test in sdks/typescript) — ChaosHarness builder, binary discovery, executeIPC child-process client, seed safety checks, result mapping, build and tests. Use when changing sdks/typescript or the engine contracts it parses.
---

# TypeScript SDK (`sdks/typescript`)

## When to use

- Changing anything under `sdks/typescript`.
- Changing the `chaossql engine` response shape.

## Components

| File | Role |
| :--- | :--- |
| `src/engine.ts` | `findChaosSQLBinary`, `executeIPC(payload, binaryPath?, timeoutMs = 60000)`, `deriveAnomalyCode` |
| `src/harness.ts` | `ChaosHarness` fluent builder |
| `src/types.ts` | `ChaosResult`, `InvariantResult`, `ScheduledOp`, `SchedulePlan`, options |
| `src/index.ts` | public exports |

### Binary discovery

explicit path (exists) → `CHAOSSQL_BIN_PATH` → walk up 6 levels from
`__dirname` for `bin/chaossql[.exe]` → each `PATH` entry → error.

### `ChaosHarness`

```ts
const h = new ChaosHarness({ driver: 'sqlite', dsn: ':memory:', isolation: '' });
h.withSchema(sql).withSeed(sql).withInvariant(name, query, assertion)
 .addOperation('withdraw', ['SELECT balance FROM accounts WHERE id = 1 -> bal', 'UPDATE ...']);
const r = await h.run({ workers: 2, iterations: 10, seed: 42 });  // runAndShrink alias
await h.assertNoAnomalies();                 // throws with anomaly details
await h.exportStandaloneRepro('repro.test.js');
```

`buildPayload` sends the flat IPC form with defaults workers 2,
iterations 10, `seed_value` 42 and rejects seeds that are not non-negative
safe integers. `executeIPC` re-validates `seed_value` / `engine.seed`, spawns
`<bin> engine`, writes the JSON payload, kills the child on timeout, rejects on
non-zero exit with empty stdout or unparseable JSON, and maps snake_case keys
to camelCase (`anomalyType`, `anomalyCode`, `minimalOperations`, `reproGo`,
`reproPython`, `reproTypeScript`, ...), attaching `exportStandaloneRepro`.

## Gotchas

- Default `iterations` is 10 here but 20 in the Python harness.
- Package version (1.4.0) lags the Go version; there is no build-time check.
- The generated TypeScript repro's fidelity is limited
  (`chaossql-repro-synthesis`).

## Build and tests

- `make test-typescript` → `cd sdks/typescript && npm ci && npm run build && npm test`
  (`tsc` then `node --test tests/*.test.js`), after `make build`.
- Tests: `sdks/typescript/tests/harness.test.js`.

## Source map

- `sdks/typescript/src/engine.ts`
- `sdks/typescript/src/harness.ts`
- `sdks/typescript/src/types.ts`
- `sdks/typescript/src/index.ts`
- `sdks/typescript/tests/harness.test.js`
- `sdks/typescript/package.json`
- `sdks/typescript/tsconfig.json`

## Related skills

- `chaossql-engine-ipc`, `chaossql-sdk-python`, `chaossql-repro-synthesis`,
  `chaossql-release-process`
