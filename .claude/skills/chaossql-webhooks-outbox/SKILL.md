---
name: chaossql-webhooks-outbox
description: Webhook alerts in the Cloud control plane — registration and SSRF validation, the safe HTTP client that pins validated IPs, event matching, the transactional webhook_outbox with retries and backoff, the background OutboxDispatcher, Discord/Slack/generic payloads, and test endpoints. Use when changing internal/server/webhooks.go or alert delivery.
---

# Webhooks and Alert Outbox

## When to use

- Changing webhook registration, dispatch formats, SSRF protection, or retries.
- Alerts are not delivered, duplicated, or blocked.

## Registration (`handleCreateWebhook`, admin)

Body `{target_type, url, events}`; `ValidateWebhookTarget(url)`:
https only, no userinfo, host required; literal IPs checked directly,
hostnames resolved via `lookupIP` (swappable in tests); rejects loopback,
private, link-local, multicast, unspecified, CGNAT `100.64.0.0/10` and other
internal ranges. `events` defaults to `"regression,all"`. Listing returns
redacted URLs.

## Delivery path

1. **Enqueue** inside the ingestion transaction (`chaossql-ingestion-baselines`):
   event `regression` when a regression was detected, else `failure` when the
   run did not pass. `GetActiveWebhooksForEvent` matches comma-separated
   events equal to the event or `all`. Each item stores a JSON
   `RegressionAlert` (repo, branch, PR, SHA, anomaly, driver, scenario, seed,
   duration, baseline status, run URL, regression flag).
2. **Dispatch** (`OutboxDispatcher`, started by `newRouter` whenever a store
   exists): wakes every 20 ms or on `Trigger()`; takes up to 20 `pending`
   items with `next_retry_at <= now`; loads the webhook (missing/inactive →
   failed attempt); `DispatchAlert` with a 5 s timeout.
3. **Result**: success → `delivered`; failure →
   `attempts+1`, `next_retry_at = now + 2 × 2^(attempts-1) s`, status `failed`
   after 3 attempts.

`DispatchAlert` formats by `target_type`: `discord` (embed), `slack`
(blocks), anything else → generic JSON of the alert. All POSTs go through
`NewSafeHTTPClient`: its dialer re-resolves and re-validates every address at
connect time and dials the validated IP directly (closing DNS rebinding), and
redirects to internal ranges are blocked.

## Test endpoints

`POST .../webhooks/test` (unsaved target, validated the same way) and
`POST .../webhooks/{wh_id}/test` (saved) dispatch a synthetic alert
synchronously.

## Gotchas

- The comment on `ValidateWebhookTarget` says DNS rebinding is not covered;
  `NewSafeHTTPClient` now covers it at dial time — keep both layers.
- Default events `regression,all` means failures alert too.
- The dispatcher polls every 20 ms for the life of the process (keep the
  machine running — see `deploy/fly/fly.toml`).
- Retention deletes non-pending outbox rows older than the plan window.

## Tests

- `internal/server/webhooks_test.go`, `internal/server/webhook_safety_regression_test.go`

## Source map

- `internal/server/webhooks.go`
- `internal/server/handlers.go`
- `internal/server/store.go`
- `internal/server/webhooks_test.go`
- `internal/server/webhook_safety_regression_test.go`
- `specs/20_saas_reliable_ingestion_and_baselines.md`

## Related skills

- `chaossql-ingestion-baselines`, `chaossql-control-plane-api`,
  `chaossql-plans-retention`, `chaossql-edge-worker` (separate portal webhook relay)
