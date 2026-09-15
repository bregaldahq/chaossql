# Spec 03: Delta-Debugging Trace Shrinker

## Objective
Reduce a chaotic schedule of $N$ operations that violated an invariant into the **1-minimal subset** that still reproduces the failure.

## Verifiable Requirements
1. **$ddmin$ Algorithm:** Partitions the schedule into granular chunks (sizes $N/2, N/4, \dots, 1$) and tests the elimination of each chunk.
2. **Atomic Reset:** Every shrinker trial resets the database to its pristine initial state using `schema.sql` and `seed.sql`.
3. **1-Minimality Guarantee:** Upon completion, no single remaining operation can be eliminated from the reduced schedule without causing the invariant to pass.
4. **Reduction Metrics:** Quantifies the reduction ratio (typically $> 90\%$) and the total count of verification trials executed.
5. **Failure Identity:** A candidate reproduces an invariant violation only when it finishes with `violation` and fails the same named invariant as the original execution. A different invariant failure, execution error, cancellation, or inconclusive result is a non-reproduction.
6. **Explicit Minimality Audit:** After ddmin converges, the shrinker tests every single-operation removal and continues reducing until none reproduces the target failure.
7. **Cancellation:** A canceled context prevents new oracle trials. Cancellation detected during a trial stops minimization without publishing a partial result.
8. **Trial Accounting:** `trials` counts actual oracle executions, including the explicit empty baseline check. Memoized lookups do not increment this metric; `iterations` remains the number of ddmin partition rounds for compatibility.
9. **Baseline Attribution:** The empty operation set is executed through the same oracle. If it reproduces the target failure, shrinking stops with an explicit baseline failure because no operation-level reproducer can be attributed.
