# Eval 01: Taxa de Redução do Shrinker ($> 85%$)

## Objetivo
Garantir que o algoritmo de Causal Delta-Debugging ($ddmin$) seja capaz de reduzir traces caóticos de $N \ge 50$ operações para um subconjunto $1$-minimal (geralmente $\le 3$ operações).

## Critérios de Aceite
1. **Taxa de Redução:** $\frac{|C_{\text{original}}| - |C_{\text{minimal}}|}{|C_{\text{original}}|} \times 100\% \ge 85\%$.
2. **1-Minimalidade:** Para todo $op \in C_{\text{minimal}}$, a remoção de $op$ faz a invariante passar ($\text{test}(C \setminus \{op\}) = \text{PASS}$).
3. **Tempo Limite:** O shrinking deve convergir em menos de 2 segundos em ambiente local.
