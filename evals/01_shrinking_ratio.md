# Eval 01: Shrinker Reduction Ratio ($> 85\%$)

## Objective
Ensure that the Causal Delta-Debugging ($ddmin$) algorithm reduces chaotic traces of $N \ge 50$ operations to a 1-minimal counterexample subset (typically $\le 3$ operations).

## Acceptance Criteria
1. **Reduction Ratio Formula:** $\frac{|C_{\text{original}}| - |C_{\text{minimal}}|}{|C_{\text{original}}|} \times 100\% \ge 85\%$.
2. **1-Minimality:** For every $op \in C_{\text{minimal}}$, eliminating $op$ causes the invariant test to pass ($\text{test}(C \setminus \{op\}) = \text{PASS}$).
3. **Time Limit:** Shrinking convergence must complete in under 2 seconds in a local execution environment.
