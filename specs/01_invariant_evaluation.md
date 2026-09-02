# Spec 01: Invariant Evaluation Engine

## Objetivo
Avaliar determinística e seguramente as asserções de invariantes expressas em SQL.

## Requisitos Verificáveis
1. **Execução SQL:** A query da invariante deve retornar uma única linha com colunas nomeadas (ex: `SELECT SUM(balance) AS total FROM accounts;`).
2. **Avaliação Safe Eval:** Os nomes das colunas são injetados como variáveis locais na expressão booleana (ex: `total == 10000`).
3. **Isolamento de Segurança:** Nenhuma função de SO ou builtin perigoso (ex: `__import__`, `os`) é permitido no ambiente de evaluation.
4. **Relatório Estruturado:** Se a asserção falhar, o resultado deve conter: nome da invariante, expressão, valores reais encontrados no banco e mensagem de erro explicável.
