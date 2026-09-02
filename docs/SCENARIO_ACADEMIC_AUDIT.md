# Auditoria e Revisão Acadêmica dos Cenários do ChaosSQL

Este documento fornece a auditoria epistemólogica e formal dos 3 cenários de demonstração, vinculando cada um às publicações seminais de concorrência (SIGMOD, VLDB, ACM TODS, CACM).

---

## 1. Quadro Comparativo de Alinhamento Acadêmico

| Cenário | Anomalia (Berenson '95) | Classificação (Adya '99) | Teorema & Paper Seminal | Solução Formal |
| :--- | :--- | :--- | :--- | :--- |
| **Banking** | **P4** (Lost Update) | **G-single** (Ciclo $rw + ww$) | Berenson et al. (*SIGMOD 1995*) | Atomic Update ou Pessimistic Locking ($2PL$) |
| **Inventory** | **A3 / P4** (Predicate Depletion) | **G-phantom** (Conflito de Predicado) | Eswaran et al. (*CACM 1976*) | Guarded Decrement ($WHERE stock \ge 1$) |
| **Hospital** | **A5B** (Write Skew) | **G2-item** (Ciclo $rw + rw$) | Fekete et al. (*TODS 2005*) & Ports (*VLDB 2012*) | Serializable Snapshot Isolation (SSI) |

---

## 2. Auditoria Detalhada por Cenário

### Cenário 01: Banking Lost Update ($P4$)
* **Fonte Seminal:** Berenson, Bernstein, Gray, Melton, O'Neil, O'Neil (*A Critique of ANSI SQL Isolation Levels*, SIGMOD 1995).
* **Formalismo de Adya (1999):** Anomalia $G-\text{single}$. O histórico contém um ciclo de dependência direcionada:
  $$T_1 \xrightarrow{rw} T_2 \xrightarrow{ww} T_1$$
* **Diagnóstico Empírico:** O nível `READ COMMITTED` (padrão do PostgreSQL e Oracle) **não protege** contra essa anomalia se a aplicação fizer leitura em memória (Read-Modify-Write).

---

### Cenário 02: Inventory Oversell ($A3 / Predicate Depletion$)
* **Fonte Seminal:** Eswaran, Gray, Lorie, Traiger (*The Notions of Consistency and Predicate Locks in a Database System*, CACM 1976).
* **Formalismo de Predicado:** A asserção de negócio envolve um predicado agregado $\sum \text{quantity} \le \text{initial_stock}$. Leituras de predicado ($r_i[P]$) são invalidadas por inserções concorrentes ($w_j[y \in P]$).
* **Diagnóstico Empírico:** Simples (`stock >= 0`) na tabela de produtos não impede que a tabela de ordens (`orders`) acumule mais compras do que o estoque permitia.

---

### Cenário 03: Hospital Write Skew ($A5B$)
* **Fontes Seminais:**
  1. Berenson et al. (*SIGMOD 1995*) — Definição do fenômeno $A5B$.
  2. Fekete, Li, O'Neil, O'Neil (*Making Snapshot Isolation Serializable*, ACM TODS 2005) — **Teorema da Estrutura Perigosa (Dangerous Structure)**.
  3. Ports & Grittner (*Serializable Snapshot Isolation in PostgreSQL*, VLDB 2012).
* **Teorema de Fekete (TODS 2005):**
  Toda execução não-serializavel sob Snapshot Isolation OBRIGATORIAMENTE contém duas arestas consecutivas de anti-dependência ($rw$):
  $$T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_1$$
* **Diagnóstico Empírico:** Como $\mathcal{W}_1 \cap \mathcal{W}_2 = \emptyset$ (Dr. Alice escreve apenas na tupla 1 e Dr. Bob apenas na tupla 2), o motor MVCC **não detecta nenhum conflito de escrita** sob `REPEATABLE READ`, gerando a catastrófica violação da invariante hospitalar.
