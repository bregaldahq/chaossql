# Spec 03: Delta-Debugging Trace Shrinker

## Objective
Reduce a chaotic schedule of $N$ operations that violated an invariant into the **1-minimal subset** that still reproduces the failure.

## Verifiable Requirements
1. **$ddmin$ Algorithm:** Partitions the schedule into granular chunks (sizes $N/2, N/4, \dots, 1$) and tests the elimination of each chunk.
2. **Atomic Reset:** Every shrinker trial resets the database to its pristine initial state using `schema.sql` and `seed.sql`.
3. **1-Minimality Guarantee:** Upon completion, no single remaining operation can be eliminated from the reduced schedule without causing the invariant to pass.
4. **Reduction Metrics:** Quantifies the reduction ratio (typically $> 90\%$) and the total count of verification trials executed.
