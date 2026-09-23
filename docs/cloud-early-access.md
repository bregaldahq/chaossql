# ChaosSQL Cloud early access

This guide connects one repository to ChaosSQL Cloud in about fifteen minutes.
Afterwards every pull request is checked for concurrency bugs, and a failure
that `main` did not have is flagged as a **regression**.

## What you need

- An organization and an **owner token**, sent to you privately by the
  ChaosSQL team. Keep it in a password manager; it is shown only once.
- A GitHub repository where you can add Actions secrets.
- A ChaosSQL scenario (`chaos.yaml`). If you do not have one yet, start from
  a folder in [`examples/`](../examples), such as `examples/banking_lost_update`.

## What leaves your CI

Only execution metadata:

- **CI:** repository, commit SHA and commit time, branch, base branch, pull
  request number, and workflow run ID.
- **Scenario:** name, a fingerprint computed locally, driver and version,
  workers, iterations, and seed.
- **Result:** status, anomaly type, duration, total and failed schedule counts,
  the *name* of the failing invariant, the size of the minimal reproduction,
  and how long shrinking took.

**No SQL, parameter values, schema, table names, invariant queries or values,
traces, or reproduction code is uploaded.** Detailed evidence stays in your
local artifacts.

## 1. Create a CI token

1. Open <https://api.chaossql.bregalda.com/#/dashboard>.
2. Paste the owner token and connect. The token stays in the browser tab's
   memory only.
3. Issue a **member** token named after the repository, for example
   `acme/payments CI`. Copy it; it is shown once.

Use the member token in CI, never the owner token. Members can publish and read
runs but cannot manage webhooks or tokens.

## 2. Add the secret

In the repository: **Settings → Secrets and variables → Actions → New
repository secret**. Name it `CHAOSSQL_CLOUD_TOKEN` and paste the member token.

## 3. Add the workflow

Create `.github/workflows/chaossql.yml`:

```yaml
name: ChaosSQL
on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read
  pull-requests: write

jobs:
  concurrency:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: bregaldahq/chaossql@v1.6.0
        with:
          spec-path: path/to/chaos.yaml
          cloud-token: ${{ secrets.CHAOSSQL_CLOUD_TOKEN }}
```

Both triggers matter. Runs on `main` establish the **baseline**; pull request
runs are compared against it. Without a baseline, a failing pull request is
reported as a failure but cannot be classified as a regression.

The scenario above uses SQLite and needs no services. For PostgreSQL or MySQL,
add a [service container](https://docs.github.com/actions/using-containerized-services/about-service-containers)
and point the scenario's DSN at it.

## 4. See the result

1. Push the workflow to `main`. The run appears in the dashboard and becomes
   the baseline for that scenario.
2. Open a pull request. The job summary and the pull request comment show the
   result, and the dashboard lists the run.
3. Any failed scenario fails the job, as it would without the Cloud. The
   Cloud adds the comparison: if `main` passed the same scenario, the run is
   marked as a **regression**. Only passing `main` runs become baselines, so a
   failure that `main` already has is not flagged as a regression.

The Action also exposes outputs for later steps:

| Output | Meaning |
| :--- | :--- |
| `cloud-run-id` | ID of the published run |
| `cloud-run-url` | Dashboard link to the run |
| `is-regression` | `true` when the run is a regression against the baseline |

To read them, give the step an `id` (for example `id: chaossql`) and use
`${{ steps.chaossql.outputs.cloud-run-url }}`.

## Options

| Input | Default | Use |
| :--- | :--- | :--- |
| `cloud-fail-fast` | `false` | Fail the job when publishing to the Cloud fails. By default a Cloud outage never blocks your CI. |
| `post-pr-comment` | `true` | Set to `false` to rely on the job summary only. |
| `export-repro` | `false` | Also write a local Go reproduction test as an artifact. |
| `seed` | from the scenario | Fix the schedule seed for a deterministic rerun. |

Changing what a scenario tests (schema, seed SQL, operations, invariants)
changes its fingerprint; runs with different fingerprints are never compared.
If you rewrite a scenario, give it a new name so it starts a fresh baseline.

## Limits during early access

- Plans limit the number of repositories and how long run history is kept
  (`team`: 10 repositories, 90 days). Ask us before you need more.
- Tokens cannot be revoked yet. Keep them in repository secrets only, and
  tell us immediately if one is exposed.
- There is no GitHub login yet; access is by token only.

## Getting help

Reply to the onboarding message or open an issue at
<https://github.com/bregaldahq/chaossql/issues> without including secrets,
SQL, or production data. We will ask for a thirty-minute call after your first
week; your feedback decides what we build next.
