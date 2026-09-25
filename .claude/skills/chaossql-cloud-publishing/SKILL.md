---
name: chaossql-cloud-publishing
description: CLI-side ChaosSQL Cloud integration in internal/cloud — when publishing happens, CI context detection (GitHub Actions, GitLab, local git), the local scenario fingerprint, the metadata-only projection and identifier validation, idempotency keys, retries and error types, GitHub Action outputs, Step Summary, and PR comments. Use when changing internal/cloud, publishToCloud in cmd/chaossql/main.go, or anything about what data leaves the machine.
---

# Cloud Publishing (client side)

`chaossql run/demo --cloud-token ...` publishes a **metadata-only** summary of
the run to `POST /v1/runs`. Privacy is a hard contract (spec 19): no SQL,
parameters, schema, traces, repro code, or connection strings ever leave the
machine.

## When to use

- Changing `internal/cloud/*` or `publishToCloud` (`cmd/chaossql/main.go`).
- Adding a field to the hosted payload.
- Debugging "unsafe identifier", 400/401/409 responses, or missing PR comments.

## Flow (`publishToCloud`)

1. `cloud.ScenarioFingerprint(spec)`: SHA-256 of the spec JSON after clearing
   `DSN`, `Description` and `Engine.Seed` — computed locally; SQL never leaves.
2. Build `RunIngestRequest`: status `passed` iff `Success`, otherwise `failed`
   (execution errors included); `execution_status`, anomaly type, duration,
   `TotalSchedules = 1`, `FailedSchedules = 1` on violation, failing invariant,
   reproduction data (Go repro, Mermaid, sanitized trace — **dropped** by the
   projection below), `CI = DetectCIContext()`, schedule plan (also dropped).
3. `NewClient(Config{BaseURL, Token, FailFast})`: defaults base URL
   `https://api.chaossql.bregalda.com`, 10 s timeout, 3 retries.
4. `PublishRun`:
   - `projectMetadataPayload` keeps only: idempotency key, version, timestamp,
     CI (provider, repository, commit SHA, commit timestamp, branch, base
     branch, PR number, run ID — **not** actor), scenario (name, fingerprint,
     driver, driver version, workers, iterations, seed), result summary,
     failing invariant **name**, and reproduction counts
     (`minimal_operations_count`, `shrink_duration_ms`);
   - generates a random 16-byte hex idempotency key if none;
   - rejects bodies over `MaxPayloadBytes` (64 KiB) → `ErrPayloadTooLarge`;
   - `validateMetadataPayload`: every identifier must match
     `^[A-Za-z0-9][A-Za-z0-9._/+:-]*$`, contain no `://`, and respect a length
     limit; provider ∈ {github-actions, gitlab-ci, local}; driver, statuses and
     anomaly type must be on allow-lists → `ErrUnsafeMetadata`;
   - POST with `Authorization: Bearer`, `Idempotency-Key` header, UA
     `ChaosSQL-CLI/v<version> (pure-go)`; retries network errors, 5xx, 408, 429
     with backoff `2^attempt × 50 ms`; 401/403 → `ErrUnauthorized`; 400 →
     `ErrBadRequest`; other 4xx → error; exhausted → `ErrCloudUnavailable`.
5. On success: print the run URL / regression banner, append
   `cloud-run-id`, `cloud-run-url`, `is-regression` to `$GITHUB_OUTPUT`
   (rejects CR/LF/NUL), append the Markdown report to `$GITHUB_STEP_SUMMARY`,
   and post a PR comment when `--pr-comment`, a GitHub token, a repository and
   a PR number are present.

Without `--cloud-fail-fast`, publish errors are printed as warnings (or
attached as `cloud_error` in `--json`) and the command result is unchanged.

## CI detection (`ci.go`)

- GitHub Actions (`GITHUB_ACTIONS=true`): repo, SHA, branch
  (`GITHUB_HEAD_REF` or `GITHUB_REF_NAME`), base `GITHUB_BASE_REF`, PR number
  from `refs/pull/N/merge` or the event file, run ID, actor.
- GitLab (`GITLAB_CI=true`): project path, SHA, ref, MR IID and target branch.
- Otherwise local git (`git rev-parse`), sanitized remote as repository.
- Commit timestamp: `git show -s --format=%cI <sha>^{commit}` only when the SHA
  looks valid; never upload time.

## PR comment (`pr_reporter.go`)

`FormatPRMarkdown` renders only safe identifiers (fallbacks otherwise).
`PostPRComment` POSTs a **new** issue comment to
`https://api.github.com/repos/<repo>/issues/<n>/comments` on every run (no
update-in-place), sent by the CLI itself — the server never talks to GitHub.

## Gotchas

- Scenario names with spaces or other characters outside the identifier
  pattern make publishing fail with `ErrUnsafeMetadata`.
- The server decodes with `DisallowUnknownFields`: a new client field breaks
  ingestion against older servers — ship server support first.
- `SanitizedTraceEvent.Table/SQL` and `ReproductionData` code fields exist
  only for decoding legacy payloads; hosted ingestion rejects them.
- Plan flags such as `has_pr_comments` are not consulted by the CLI.

## Change checklist (new hosted field)

1. Add to `metadataPayload` + `projectMetadataPayload` + `validateMetadataPayload`
   + `DecodeMetadataOnlyRequest` (same struct is the server contract).
2. Server storage and projections (`chaossql-ingestion-baselines`,
   `chaossql-control-plane-api`).
3. Privacy review against `specs/19_saas_cloud_payload_privacy.md`.
4. Dashboard client (`site/src/lib/cloud-api.ts`) if displayed.

## Tests

- `internal/cloud/*_test.go` (client, reliability, sanitizer, fingerprint, CI,
  action outputs, PR reporter), `cmd/chaossql/cloud_integration_test.go`

## Source map

- `internal/cloud/client.go`
- `internal/cloud/payload.go`
- `internal/cloud/types.go`
- `internal/cloud/fingerprint.go`
- `internal/cloud/ci.go`
- `internal/cloud/sanitizer.go`
- `internal/cloud/action.go`
- `internal/cloud/pr_reporter.go`
- `cmd/chaossql/main.go`
- `specs/19_saas_cloud_payload_privacy.md`
- `docs/cloud-early-access.md`

## Related skills

- `chaossql-ingestion-baselines`, `chaossql-control-plane-api`,
  `chaossql-github-action`, `chaossql-cli-run-pipeline`
