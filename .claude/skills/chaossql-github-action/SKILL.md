---
name: chaossql-github-action
description: The composite GitHub Action in action.yml (inputs, outputs, how it builds and invokes chaossql run, Cloud and PR comment wiring, exit behavior) and the repository workflows that dogfood ChaosSQL — concurrency-ci.yml and the multi-engine swarm.yml with mutation, SARIF upload and WASM stress. Use when changing action.yml, its inputs/outputs, or these workflows.
---

# GitHub Action and Dogfooding Workflows

## When to use

- Adding/changing an action input or output.
- Changing how CI users run ChaosSQL.
- Editing `.github/workflows/concurrency-ci.yml` or `swarm.yml`.

## `action.yml` (composite)

Steps:
1. `actions/setup-go@v5` with `go-version: stable`.
2. Build: `cd $ACTION_PATH && CGO_ENABLED=0 go build -o $RUNNER_TEMP/chaossql ./cmd/chaossql`.
3. Execute `chaossql run <spec-path>` with arguments assembled from inputs
   (values passed through `env`, never interpolated into the script).

| Input | Maps to |
| :--- | :--- |
| `spec-path` | positional spec |
| `workers`, `iterations`, `seed` | `--workers`, `--iterations`, `--seed` (only when non-empty) |
| `export-html`, `export-junit` | `--export-html`, `--export-junit` |
| `export-summary` | `--export-summary`; defaults to `$GITHUB_STEP_SUMMARY` when empty |
| `export-repro`, `export-mermaid` | `--export-repro`, `--export-mermaid` when `true` |
| `cloud-token`, `cloud-url` | `CHAOSSQL_CLOUD_TOKEN`, `CHAOSSQL_CLOUD_URL` env (default URL `https://api.chaossql.bregalda.com`) |
| `cloud-fail-fast` | `--cloud-fail-fast` when `true` |
| `github-token` (default `github.token`) | `GITHUB_TOKEN` env for PR comments |
| `post-pr-comment` | `--pr-comment=false` when `false` |

Outputs `cloud-run-id`, `cloud-run-url`, `is-regression` come from the CLI
appending to `$GITHUB_OUTPUT` after a successful Cloud publish
(`chaossql-cloud-publishing`).

### Exit behavior (verified)

The step fails only when `chaossql run` exits non-zero: execution errors,
inconclusive or canceled runs, setup failures, or Cloud failures with
`cloud-fail-fast`. **A detected violation exits 0**, so the step stays green;
gate on `is-regression`, the JUnit report, or parse `--json` if you need a red
build on anomalies.

## `concurrency-ci.yml`

Runs the action from the repository itself (`uses: ./`) on push/PR to `main`.

## `swarm.yml`

Job 1 (with PostgreSQL/MySQL services, user `chaossql`): `make build`;
`chaossql mutate` every `examples/*/chaos.yaml` into `examples/*/mutated`
(3 variants); `chaossql swarm diff --drivers sqlite,mock,postgres,mysql
--markdown-summary $GITHUB_STEP_SUMMARY`; `run banking --export-sarif
findings.sarif --export-repro || true`; upload SARIF to code scanning
(`continue-on-error`) and archive `repro_test.go` + SARIF.
Job 2: `make wasm` and `make stress-wasm`.

## Change checklist

- New CLI flag worth exposing → input + env + argument assembly + README
  example (`uses: bregaldahq/chaossql@vX.Y.Z`) + portal snippets
  (`site/src/pages/DashboardPage.tsx`).
- Test: `cmd/chaossql/action_integration_test.go`.

## Source map

- `action.yml`
- `.github/workflows/concurrency-ci.yml`
- `.github/workflows/swarm.yml`
- `cmd/chaossql/action_integration_test.go`
- `internal/cloud/action.go`
- `README.md`

## Related skills

- `chaossql-cli-run-pipeline`, `chaossql-cloud-publishing`,
  `chaossql-differential-fuzzing`, `chaossql-quality-gate`, `chaossql-release-process`
