# Spec 03: Delta-Debugging Trace Shrinker

## Objetivo
Reduzir uma sequência caótica de $N$ operações que quebrou uma invariante para o **subconjunto mínimo (1-minimal)** que ainda reproduz a falha.

## Requisitos Verificáveis
1. **Algoritmo ddmin:** Divide o plano em blocos (chunks de tamanho $N/2, N/4, \dots, 1$) e testa a remoção de cada bloco.
2. **Reset Atômico:** Cada teste do shrinker reseta o banco para o estado inicial com `schema.sql` e `seed.sql`.
3. **Garantia de 1-Minimalidade:** Ao concluir, nenhuma única operação pode ser removida do plano resultante sem que a invariante volte a passar.
4. **Metricas de Redução:** Calcula a porcentagem de redução (geralmente $> 90%$) e o número de execuções de verificação.
