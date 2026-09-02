# Eval 02: Taxa de Falsos Positivos ($0\%$)

## Objetivo
Garantir que o ChaosSQL nunca acuse uma violação de invariante em um sistema que implementa corretamente bloqueio pessimista (`FOR UPDATE`) ou `SERIALIZABLE`.

## Critérios de Aceite
1. **Falsos Positivos:** Em 100 baterias de teste com o SQL corrigido, 0 violações devem ser reportadas.
2. **Tratamento de Erros do Driver:** Erros de serialização (`40001`) e deadlocks (`40P01`) devem ser abortados e contabilizados sem quebrar a invariante.
