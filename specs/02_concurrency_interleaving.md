# Spec 02: Concurrency Scheduler & Step Interleaving

## Objetivo
Orquestrar m�ltiplos workers assíncronos concorrentes e injetar jitter de latência entre passos (steps) de transações para provocar race conditions.

## Requisitos Verificáveis
1. **Distribuição em Filas:** Cada worker possui sua própria fila (`asyncio.Queue`) e conexão de banco isolada.
2. **Interleaving Intra-Transação:** Se uma operação tem $k > 1$ passos, o scheduler injeta `asyncio.sleep(delay_ms / 1000)` entre os passos, permitindo que outros workers intercalem operações no banco.
3. **Random Rollbacks:** Suporte a abortar transações aleatoriamente para testar resiliência a falhas parciais.
4. **Trace Append-Only:** Todo evento (execução, commit, rollback, erro) é registrado com timestamp, worker_id, op_id e SQL.
