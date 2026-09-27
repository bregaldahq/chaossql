---
name: chaossql-sdk-python
description: The Python SDK (chaossql-py in sdks/python) — ChaosHarness builder, binary discovery, the IPC subprocess client, ChaosResult and AnomalyType, the pytest plugin (chaossql_runner fixture and chaossql marker) that shells out to chaossql run --json, packaging and tests. Use when changing sdks/python or the contracts it depends on.
---

# Python SDK (`sdks/python`)

## When to use

- Changing anything under `sdks/python`.
- Changing `chaossql engine` or `chaossql run --json` output (both are parsed here).

## Components

| Module | Role |
| :--- | :--- |
| `chaossql/engine.py` | `find_chaossql_binary`, `execute_ipc(payload, bin_path, timeout_sec=60)` |
| `chaossql/harness.py` | `ChaosHarness` fluent builder |
| `chaossql/types.py` | `ChaosResult` dataclass, `AnomalyType` constants, `anomaly_code` |
| `chaossql/pytest_plugin.py` | `ChaosSQLRunner`, `chaossql_runner` / `chaossql_fixture` fixtures, `chaossql` marker |

### Binary discovery (`find_chaossql_binary`)

1. explicit path (must be an executable file) → 2. `CHAOSSQL_BIN_PATH` →
3. walk up to 5 parents from the package looking for `bin/chaossql` →
4. `PATH`. Otherwise `FileNotFoundError` telling you to `make build`.

### `ChaosHarness`

```python
h = ChaosHarness(driver="sqlite", dsn=":memory:", bin_path=None, isolation="")
h.with_schema(sql).with_seed(sql).with_invariant(name, query, assertion)
h.add_operation("withdraw", ["SELECT balance FROM accounts WHERE id = 1 -> bal",
                             {"sql": "UPDATE ...", "capture": ""}], weight=1.0, params={...})
result = h.run(workers=2, iterations=20, seed=42)   # run_and_shrink is an alias
h.assert_no_anomalies(...)                         # AssertionError with minimal ops + Mermaid
h.export_standalone_repro("repro.py")              # runs first if needed
```

`_build_payload` sends the flat IPC form (`driver`, `dsn`, `isolation`,
`schema`, `seed`, `invariants`, `operations`, `workers`, `iterations`,
`seed_value`) — see `chaossql-engine-ipc`. `execute_ipc` runs
`<bin> engine` with `subprocess.Popen`, kills on timeout, raises on non-zero
exit with empty stdout or on invalid JSON, and maps keys with defaults into
`ChaosResult`.

`assert_no_anomalies` raises `RuntimeError` when `result.error` is set or the
run neither succeeded nor violated (execution errors), and `AssertionError`
on violations.

### pytest plugin

`ChaosSQLRunner.run_scenario(path, workers=None, iterations=None, seed=None)`
runs `chaossql run <abs path> --json [--workers] [--iterations] [--seed]`
(defaults can come from `@pytest.mark.chaossql(workers=..., seed=...)`), and
maps the CLI JSON (`shrink.minimal_ops` → `minimal_operations`). The binary
comes from `CHAOSSQL_BIN_PATH`. The marker is registered in `pytest_configure`.

## Gotchas

- `run --json` output has no `repro_python`/`repro_typescript` keys, so plugin
  results only carry `repro_go`.
- `ChaosSQLRunner` never checks the CLI exit code; it relies on parsing stdout.
- The assertion message reads `failing_invariant["observed"]`, a key the engine
  never sends (it sends `actual_values`).
- Versions are hard-coded in three places (`pyproject.toml`, `setup.py`,
  `chaossql/__init__.py`) and currently lag the Go version (1.4.0 vs 1.6.0).

## Tests

- `make test-python` (builds the binary, then
  `CHAOSSQL_BIN_PATH=bin/chaossql PYTHONPATH=sdks/python PYTEST_DISABLE_PLUGIN_AUTOLOAD=1 python3 -m pytest -p chaossql.pytest_plugin sdks/python/tests -v`).
- CI pins dependencies with `PIP_CONSTRAINT=sdks/python/test-constraints.txt`.
- Files: `sdks/python/tests/test_harness.py`, `sdks/python/tests/test_pytest_plugin.py`,
  `sdks/python/tests/conftest.py`.

## Source map

- `sdks/python/chaossql/__init__.py`
- `sdks/python/chaossql/engine.py`
- `sdks/python/chaossql/harness.py`
- `sdks/python/chaossql/types.py`
- `sdks/python/chaossql/pytest_plugin.py`
- `sdks/python/pyproject.toml`
- `sdks/python/setup.py`
- `sdks/python/test-constraints.txt`
- `sdks/python/tests/test_harness.py`
- `sdks/python/README.md`

## Related skills

- `chaossql-engine-ipc`, `chaossql-cli-run-pipeline`, `chaossql-sdk-typescript`,
  `chaossql-release-process`
