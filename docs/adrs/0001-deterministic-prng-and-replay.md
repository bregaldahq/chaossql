# ADR 0001: Geração Pseudo-Aleatória Determinística e Seeding

* **Status:** Aceito
* **Data:** 2026-09-01

## Contexto
Para que um fuzzer de concorrência seja útil em engenharia, qualquer falha encontrada DEVE ser 100% reproduz�ivel em outras máquinas e no CI.

## Decisão
Adotar uma instância dedicada de `random.Random(seed)` por execução. Todos os parâmetros, delays e ordem de operações são gerados a partir dessa seed, garantindo que a mesma seed sempre produza o mesmo plano.

## Consequências
* Developers podem compartilhar apenas a seed (ex: `--seed 42`) para reproduzir o bug.
* O algoritmo de shrinking pode re-executar subconjuntos sabendo que os parâmetros nenhuma vez mudarão.
