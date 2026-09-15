# ENG-03 Replay and Shrinker Design

## Problem

ChaosSQL currently minimizes any invariant violation, even when a candidate reproduces a different invariant from the original failure. The shrink result can therefore be small without being causally related to the reported finding. The `replay` command only renders a stored trace; it cannot reset the database and execute the recorded operations again.

For a commercial service, every minimized finding needs a portable evidence artifact that identifies the exact failure, preserves the effective schedule inputs, and can verify whether the same failure occurs again.

## Failure identity

A violation is identified by a stable `FailureSignature` containing the execution status and failing invariant name. For invariant violations, a candidate reproduces the target only when it finishes with `violation` and fails the same named invariant. Infrastructure errors, cancellations, inconclusive evaluations, and different invariant failures are non-reproductions.

The signature deliberately excludes trace timestamps and physical anomaly classification. Those values can change with database lock and runtime timing even when the harness-controlled logical schedule is identical.

## Shrinking contract

The shrinker keeps the existing oracle convention: `false` means that the target failure reproduced. It checks cancellation before every oracle call, memoizes subsets by their ordered operation IDs, and reports the number of actual oracle trials. It executes the empty subset through the same oracle and returns an explicit baseline failure when the target reproduces without operations. After ddmin converges, an explicit single-removal audit verifies 1-minimality and continues reducing if any operation can still be removed.

Every engine entry point that shrinks a violation derives one target signature from the original result and tests candidates against that signature. A different failure can never satisfy the oracle.

## Replay artifact

The version 1 replay artifact contains:

- the complete normalized specification, including schema and seed SQL;
- effective seed and versioned logical schedule;
- exact scheduled operations and their bound parameters;
- observed trace for inspection;
- expected execution status, failure signature, and anomaly metadata.

`chaossql run --export-result <path>` writes this JSON artifact after execution. The export contains the minimal operations when shrinking succeeds, together with the logical schedule rebuilt for that minimal set. It contains the original operations when no shrink result exists.

Existing trace-only replay JSON remains readable.

## Executable replay

`chaossql replay --verify <artifact.json>` validates the artifact, opens its configured database driver, resets the database through `RunSchedule`, and executes the stored operations with the stored seed. Verification succeeds only when:

1. the regenerated logical schedule equals the stored schedule; and
2. the result reproduces the stored failure signature, or reproduces the stored passing status for a non-failure artifact.

The command prints a concise verification result and returns a nonzero error when schedule identity or failure identity differs. Plain `chaossql replay` remains a read-only trace renderer.

## Compatibility and safety

Replay artifact versioning allows future schema changes. Version `0` is accepted only for legacy trace inspection and cannot be executed. Verification rejects a missing specification, empty operation set, unsupported artifact or schedule versions, and inconsistent seeds before opening a driver.

Artifacts may contain database credentials and SQL data inherited from the specification. They are written only to the explicit local path requested by the caller and are not uploaded automatically. Publication uses a temporary owner-only file and an atomic replacement. Unix uses mode `0600`; Windows creates the temporary file with a protected ACL granting full access only to the current account.

## Acceptance

- A candidate that violates a different invariant is rejected by the shrink oracle.
- Synthetic and engine-backed shrink tests prove 1-minimality and deterministic operation IDs.
- Trial metrics count actual oracle executions.
- An exported SQLite artifact replays the same invariant failure from a fresh database reset.
- A tampered seed, schedule, or failure signature fails verification clearly.
- Legacy trace-only files continue to render.
- `make verify`, race tests, and zero-CGO native and WASM builds pass.
