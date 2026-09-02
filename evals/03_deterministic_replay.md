# Eval 03: Reprodutibilidade e Convergência Determinística ($100\%$)

## Objetivo
Garantir que a mesma `seed` gere exatamente a mesma sequência de operações e resultado em 100% das execuções.

## Critérios de Aceite
1. **Identidade de Plano:** Duas execuções com a mesma seed devem gerar o mesmo slice de `ScheduledOp` (mesmos IDs, nomes e parâmetros).
2. **Determinismo do Shrinker:** O trace m�nimo resultante deve conter exatamente os mesmos IDs de operação em todas as rodadas.
