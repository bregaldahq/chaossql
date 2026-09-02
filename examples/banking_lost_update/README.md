# Cenário 01: Fintech Lost Update (Anomalia P4)

## Contexto de Negócio
Em um sistema bancário, dois saques simultâneos ocorrem na conta de *Alice* (saldo inicial: R$ 1.000).

## O Bug (Read-Modify-Write Sem Lock)
1. **Worker 1** le o saldo (R$ 1.000) e prepara saque de R$ 100 (saldo final deveria ser R$ 900).
2. **Worker 2** le o mesmo saldo (R$ 1.000) antes do Worker 1 comitar e prepara saque de R$ 150 (saldo final deveria ser R$ 850).
3. Ambos escrevem os saldos calculados em memória.

* **Resultado Real:** O saldo vira R$ 850, mas o extrato (ledger) registrou R$ 250 de débitos! O banco perdeu R$ 100.

## Invariante de Consistência
$$\text{Saldo Atual} == 1000 - \sum(\text{Débitos do Extrato})$$

## Como Corrigir
* **Correção Atômica no SQL:**
  ```sql
  UPDATE accounts SET balance = balance - :amount WHERE id = 1 AND balance >= :amount;
  ```
