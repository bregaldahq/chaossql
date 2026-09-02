# Cenário 04: Financial Audit Read Skew (Anomalia A5A)

## Contexto de Negócio
Em uma instituição financeira, um cliente possui duas contas vinculadas: **Checking** (Corrente, saldo inicial: R$ 500) e **Savings** (Poupança, saldo inicial: R$ 500). O patrimônio total do cliente é de R$ 1.000.
Um processo de auditoria ou extrato consolidado ($T_1$) calcula o patrimônio total lendo o saldo de ambas as contas. Concomitantemente, transações de transferência ($T_2$) movimentam fundos entre as contas.

## O Bug (Read Skew / A5A sob Read Committed)
A anomalia **A5A (Read Skew)** ocorre quando uma transação de leitura inconsistente observa um estado parcial de outra transação concorrente:
1. $T_1$ (Auditoria) lê a conta **Checking** ($x = 500$).
2. $T_2$ (Transferência) transfere R$ 100 de Checking para Savings:
   - Decrementa Checking: $x \leftarrow 400$
   - Incrementa Savings: $y \leftarrow 600$
   - Registra auditoria em `transfers`.
   - $T_2$ comita com sucesso.
3. $T_1$ lê a conta **Savings** ($y = 600$), observando a escrita feita por $T_2$.
4. $T_1$ calcula a riqueza total: $500 + 600 = 1100 \ne 1000$.

### Definição Matemática Formal (Adya / Berenson et al.)
No grafo de dependências diretas de Adya (DSG), o Read Skew é caracterizado pelo ciclo contendo uma anti-dependência de leitura ($rw$) e uma dependência de escrita-leitura ($wr$):
$$T_1 \xrightarrow{rw} T_2 \xrightarrow{wr} T_1$$

Onde:
- $T_1 \xrightarrow{rw} T_2$ no item $x$ (`accounts:1` / Checking): $T_1$ leu a versão de $x$ anterior à modificação feita por $T_2$.
- $T_2 \xrightarrow{wr} T_1$ no item $y$ (`accounts:2` / Savings): $T_1$ leu a versão de $y$ escrita e comitada por $T_2$.

## Invariante de Preservação de Riqueza
$$\text{total\_balance} == 1000$$

## Como Corrigir
* **Nível de Isolamento SERIALIZABLE ou REPEATABLE READ / SNAPSHOT ISOLATION:**
  Garante que $T_1$ leia uma visão consistente e congelada no tempo de todo o banco de dados (snapshot no início de $T_1$), lendo $x = 500$ e $y = 500$ (total = 1000).
* **Transações Atômicas com Lock Pessimista (`SELECT ... FOR UPDATE`):**
  Bloqueia ambos os registros durante a leitura da auditoria para impedir modificações concorrentes até o término de $T_1$.
