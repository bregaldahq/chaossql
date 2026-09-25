---
name: chaossql-quality-gate
description: The unified quality gate (make verify) and every stage behind it — harness check, go vet, race tests with coverage, Python/TypeScript SDK tests, frontend verify, headless WASM stress, waitlist and English-purity tests, WASM worker/playground/bench tests — plus the CI workflow, database services, and the fastest command to run for each kind of change. Use before committing, when CI is red, or when adding a new check.
---

# Quality Gate (`make verify`)

`AGENTS.md` rule #1: never break `make verify`.

## Stages (in order)

| Stage | Command | Needs |
| :--- | :--- | :--- |
| `check-harness` | `go run tools/harness_check.go` | required docs/files + agent skill harness (`chaossql-skill-maintenance`) |
| `lint` | `go vet ./...` | — |
| `test` | `go test -v -race ./internal/... ./cmd/... ./pkg/...` (+ `-covermode=atomic -coverprofile=$COVERPROFILE` when set) | optional PostgreSQL/MySQL |
| `test-python` | `make build`, then pytest with `CHAOSSQL_BIN_PATH`, `PYTHONPATH=sdks/python`, `PYTEST_DISABLE_PLUGIN_AUTOLOAD=1 -p chaossql.pytest_plugin` | Python 3 + pytest |
| `test-typescript` | `make build`, `cd sdks/typescript && npm ci && npm run build && npm test` | Node |
| `test-frontend` | `cd site && npm ci && npm run verify` | Node |
| `stress-wasm` | build `bin/chaossql-test.wasm`, `node tools/headless_worker_stress.js` | Node |
| node tests | `node --test tools/test_waitlist.mjs tools/test_english_purity.test.cjs` | `site/node_modules` (esbuild) from `test-frontend` |
| node scripts | `node tools/test_english_purity.js && node tools/test_wasm_worker.js && node tools/test_playground_ui.js && node tools/test_wasm_bench.js` | — |

`tools/harness_check.py` is a legacy subset and is not used by the Makefile.

## English purity (`tools/test_english_purity.js`)

Scans `docs specs evals examples cmd internal pkg tools .claude/skills` and
root files (`Makefile ARCHITECTURE.md AGENTS.md CLAUDE.md CONTRIBUTING.md
README.md SECURITY.md action.yml CHANGELOG.md`), excluding `site/`,
`docs/superpowers/`, `.superpowers/`, `node_modules/` and a few tools. Fails on
Portuguese accented characters (except allowed citations such as Poincaré) or
common Portuguese keywords.

## Fast feedback by change type

| You changed | Run first |
| :--- | :--- |
| `internal/engine`, `internal/shrinker`, `internal/analyzer` | `go test -race ./internal/engine/... ./internal/shrinker/... ./internal/analyzer/...` |
| `internal/drivers` | `go test ./internal/drivers/...` (+ databases, below) |
| `cmd/chaossql` | `go test ./cmd/chaossql/...` |
| `internal/server`, `internal/cloud` | `go test -race ./internal/server/... ./internal/cloud/...` |
| `sdks/*` or IPC | `make test-sdks` |
| `site/` | `make test-frontend` and the node scripts |
| `cmd/chaossql-wasm`, worker | `make stress-wasm && node tools/test_wasm_worker.js` |
| skills, docs | `make check-harness && node tools/test_english_purity.js` |

## Databases for integration tests

Driver integration tests skip without databases unless
`CHAOSSQL_REQUIRE_DATABASES=1`. CI provides PostgreSQL 16.4
(`DATABASE_URL=postgres://postgres:postgres@127.0.0.1:5432/chaossql_test?sslmode=disable`)
and MySQL 8.4.2 (`MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/chaossql_test?parseTime=true`).

## CI (`.github/workflows/ci.yml`)

On push/PR to `main`: Go 1.25.0, Node 20.19.5, Python 3.12.11 (pip pinned by
`PIP_CONSTRAINT=sdks/python/test-constraints.txt`); `make verify
COVERPROFILE=coverage.out`; Codecov upload via OIDC (`codecov.yml`); then
`make demo` (each demo ends with `|| true`, so demos never fail CI).
Other workflows: `concurrency-ci.yml` and `swarm.yml`
(`chaossql-github-action`), `deploy-pages.yml` and `static-pages.yml`
(`chaossql-edge-worker`).

## Gotchas

- `swarm.yml` pins Go 1.23 while `go.mod` requires 1.25 (setup-go may
  auto-download the toolchain).
- The Zero-CGO rule is enforced by building with `CGO_ENABLED=0`
  (`make build`, `make wasm`, Dockerfiles), not by a dedicated check.
- Adding a required artifact means editing the list in `tools/harness_check.go`.

## Source map

- `Makefile`
- `tools/harness_check.go`
- `tools/test_english_purity.js`
- `tools/test_english_purity.test.cjs`
- `tools/test_waitlist.mjs`
- `tools/headless_worker_stress.js`
- `tools/test_playground_ui.js`
- `tools/harness_check.py`
- `.github/workflows/ci.yml`
- `codecov.yml`
- `CONTRIBUTING.md`

## Related skills

- `chaossql-skill-maintenance`, `chaossql-database-drivers`,
  `chaossql-github-action`, `chaossql-wasm-playground`, `chaossql-website-portal`
