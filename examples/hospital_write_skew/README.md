# Cenário 03: Hospital Write Skew (Anomalia A5B)

## Contexto de Negócio
Regra hospitalar: **Pelo menos um médico deve estar de plantão ativo a qualquer momento**.
Inicialmente, o Dr. Alice e o Dr. Bob estão de plantão (`is_on_call = 1`).

## O Bug (Write Skew sob Snapshot Isolation)
1. **Dr. Alice** tenta sair do plantão: consulta (retorna 2 ativos). Como $2 >= 2$, ela atualiza seu status para 0 e comita.
2. **Dr. Bob** simultaneamente tenta sair do plantão: consulta (retorna 2 ativos na sua snapshot). Como $2 >= 2$, ele atualiza seu status para 0 e comita.
3. **Resultado Catastrófico:** **Zero médicos de plantão!** A invariante hospitalar foi violada.

## Invariante Hospitalar
$$\sum(\text{is_on_call}) \ge 1$$

## Como Corrigir
* **Isolamento SERIALIZABLE (SSI):** O banco (ex: PostgreSQL) detecta o ciclo $T_1 \rightleftarrows T_2$ e aborta uma das transações com erro `40001 serialization_failure`.
* **Bloqueio Pessimista:** Fazer `SELECT ... FOR UPDATE` nas linhas lidas.
