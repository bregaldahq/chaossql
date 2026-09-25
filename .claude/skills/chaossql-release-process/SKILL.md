---
name: chaossql-release-process
description: How a ChaosSQL release is prepared — the single Go version source, every other place that carries a version string (Helm, README/action pins, portal snippets, worker footer, WASM rebuild, SDK packages), CHANGELOG and release notes conventions. Use when bumping the version or preparing a release pull request.
---

# Release Process

## When to use

- Preparing `chore(release): prepare vX.Y.Z`.
- Any change to version reporting.

## Version sources

`internal/version/version.go` (`const Version`) is the single source for Go:
`chaossql --version`, `chaossql-server --version`, the WASM
`ChaosSQL_GetVersion` (`-wasm` suffix), SARIF `ToolVersion`, and the Cloud
client User-Agent all read it.

Do not reintroduce `-ldflags -X main.version=...` in the `Dockerfile` or
`Makefile`; v1.6.0 removed them in favor of this package.

Strings that must be bumped by hand (as done for v1.6.0):
- `charts/chaossql-server/Chart.yaml` (`version`, `appVersion`) and
  `charts/chaossql-server/values.yaml` (`image.tag`);
- `README.md` release badge and `uses: bregaldahq/chaossql@vX.Y.Z` example;
- `site/src/pages/DashboardPage.tsx` action snippets (then rebuild the site);
- `worker.ts` and `site/_worker.js` footer text (`... • vX.Y.Z`);
- rebuild WASM (`make wasm`) so the embedded version matches, and rebuild
  the portal bundle (`chaossql-website-portal`).

SDK packages are versioned independently and currently lag (1.4.0):
`sdks/python/pyproject.toml`, `sdks/python/setup.py`,
`sdks/python/chaossql/__init__.py`, `sdks/typescript/package.json`,
`site/package.json`.

## Changelog and notes

- `CHANGELOG.md`: Keep a Changelog + SemVer; a dated `## [X.Y.Z] - YYYY-MM-DD`
  section with a short summary, then `### Added`, `### Changed`, `### Fixed`,
  `### Security` as needed, referencing PR numbers (`(#25)`); mention one-time
  migrations.
- `docs/releases/vX.Y.Z.md`: longer release notes.
- Tests: `internal/version/version_test.go`, `cmd/chaossql/version_test.go`.

## Checklist

1. Bump `internal/version/version.go`.
2. Update every hand-maintained string above (`grep -rn "X.Y.Z-previous"`
   outside `node_modules`, lockfiles and `docs/superpowers`).
3. `make wasm` and `cd site && npm run build`; commit new artifacts, remove
   stale hashed bundles.
4. CHANGELOG + release notes.
5. `make verify`.
6. Tag `vX.Y.Z` after merge (the action is consumed as `@vX.Y.Z`).
7. Commit messages and PR bodies carry no co-author or AI attribution.

## Source map

- `internal/version/version.go`
- `internal/version/version_test.go`
- `cmd/chaossql/version_test.go`
- `CHANGELOG.md`
- `docs/releases/v1.6.0.md`
- `charts/chaossql-server/Chart.yaml`
- `charts/chaossql-server/values.yaml`
- `README.md`
- `worker.ts`
- `site/_worker.js`
- `site/src/pages/DashboardPage.tsx`

## Related skills

- `chaossql-server-operations`, `chaossql-wasm-playground`,
  `chaossql-website-portal`, `chaossql-github-action`, `chaossql-docs-specs-adrs`
