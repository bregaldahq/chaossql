# ADR 0003: Causal Delta-Debugging (ddmin) para Transações SQL

* **Status:** Aceito
* **Data:** 2026-09-01

## Contexto
O algoritmo de Delta-Debugging clássico (Zeller '99) assume que os elementos do conjunto são independentes. Em bancos SQL, uma transação $T_j$ pode depender da existência de uma chave estrangeira ou entidade criada por $T_i$. Remover $T_i$ ingenuamente gera erros de *Foreign Key Constraint*, confundindo o orâculo.

## Decisão
Implementar o **Causal Delta-Debugging**:
1. Construir um grafo acíclico de dependências causais ($T_i \to T_j$) com base nos parâmetros gerados.
2. Ao testar um subconjunto $C'$, calcular o fechamento transitivo $\text{Closure}(C')$ para garantir que todas as dependências prévias sejam mantidas.

## Consequências
* Eliminação total de falsos positivos por erro de SQL.
* Redução do número de passos do shrinker em até 75%.
