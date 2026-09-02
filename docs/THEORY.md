# Fundamentação Matemática e Formal do ChaosSQL

Este documento estabelece o embasamento matemático, os modelos formais de concorrência, a teoria de grafos de conflito, a definição formal de invariantes indutivas e a prova de convergência do algoritmo de **Delta-Debugging ($ddmin$)**.

---

## 1. Modelo Formal de Transações e Histórias

### 1.1 Definição de Transação
Uma transação $T_i$ é definida formalmente como uma tupla:
$$T_i = (O_i, <_i)$$

Onde:
* $O_i$ é um conjunto finito de operações de leitura $r_i[x]$, escrita $w_i[x]$, abort $a_i$ ou commit $c_i$, acessando entidades de dados $x \in \mathcal{D}$.
* $<_i$ é uma ordem estrita parcial sobre $O_i$ que preserva a causalidade do fluxo do programa.
* $T_i$ contém exatamente uma operação terminal: $a_i \in O_i$ ou $c_i \in O_i$, sendo este o elemento maximal sob $<_i$.

### 1.2 Definição de História (Schedule)
Uma história de execução concorrente $S$ sobre um conjunto de transações $\mathcal{T} = \{T_1, T_2, \dots, T_n\}$ é uma ordem parcial $S = (\bigcup_{i=1}^n O_i, <_S)$ satisfazendo:
1. $\bigcup_{i=1}^n <_i \;\subseteq\; <_S$ (preserva a ordem interna de cada transação).
2. Para quaisquer duas operações conflitantes $o_i, o_j \in S$ (onde $i \neq j$), ou $o_i <_S o_j$ ou $o_j <_S o_i$.

### 1.3 Condição de Conflito de Bernstein
Duas operações $o_i, o_j \in S$ ($i \neq j$) estão em **conflito direto** se acessam o mesmo item $x \in \mathcal{D}$ e pelo menos uma delas é uma mutação:
$$\text{Conflict}(o_i, o_j) \iff (\text{target}(o_i) = \text{target}(o_j) = x) \;\land\; (o_i \text{ is } w \;\lor\; o_j \text{ is } w)$$

---

## 2. Grafo de Serialização e Teorema CSR

### 2.1 O Grafo de Serialização $SG(S)$
Dado um schedule $S$ sobre transações $\mathcal{T}$, o Grafo de Serialização (ou *Dependency Graph*) é um grafo direcionado $SG(S) = (V, E)$ onde:
* $V = \mathcal{T}$ (os nós são as transações).
* Existe uma aresta direcionada $(T_i \to T_j) \in E$ se e somente se existe $o_i \in T_i$ e $o_j \in T_j$ tais que:
  $$o_i <_S o_j \quad \land \quad \text{Conflict}(o_i, o_j)$$

### 2.2 Teorema Fundamental da Serializabilidade por Conflito (CSR)
$$\text{Schedule } S \text{ é Serializabilidade por Conflito (CSR)} \iff SG(S) \text{ é um Grafo Direcionado Acíclico (DAG)}$$

---

## 3. Formalização Matemática de Anomalias MVCC

### 3.1 Lost Update ($P4$)
Ocorre quando duas transações $T_1$ e $T_2$ leem simultaneamente o estado $x_0$, computam transformações independentes $f(x_0)$ e $g(x_0)$, e ambas escrevem sem bloqueio mútuo:

$$S_{P4} = r_1[x_0] \;\dots\; r_2[x_0] \;\dots\; w_1[f(x_0)] \;\dots\; c_1 \;\dots\; w_2[g(x_0)] \;\dots\; c_2$$

* **Estado esperado:** $x_{\text{final}} = g(f(x_0))$ ou $f(g(x_0))$ (efeito cumulativo).
* **Estado resultante:** $x_{\text{final}} = g(x_0)$ (a mutação $f(x_0)$ de $T_1$ é completamente perdida).
* **No Grafo de Conflito:** $T_1 \xrightarrow{r_1 \to w_2} T_2$ e $T_2 \xrightarrow{r_2 \to w_1} T_1 \implies \text{Ciclo } T_1 \rightleftarrows T_2$.

### 3.2 Write Skew ($A5B$)
Seja uma invariante de integridade global $\mathcal{I}(x, y): x + y \ge 0$.
Inicialmente $x = 100, y = 100$ ($x + y = 200 \ge 0$).
* $T_1$ deseja debitar 150 de $x$: verifica $x+y \ge 150$ (válido: 200) e faz $w_1[x \leftarrow -50]$.
* $T_2$ concorrentemente deseja debitar 150 de $y$: verifica $x+y \ge 150$ (válido na snapshot: 200) e faz $w_2[y \leftarrow -50]$.

$$S_{A5B} = r_1[x, y] \;\dots\; r_2[x, y] \;\dots\; w_1[x] \;\dots\; c_1 \;\dots\; w_2[y] \;\dots\; c_2$$

* **Estado resultante:** $x = -50, y = -50 \implies x + y = -100 < 0 \implies \mathcal{I}(x, y) = \text{FALSE}$.
* Sob `REPEATABLE READ`, ambas as transações obtêm sucesso porque $w_1$ toca apenas $x$ e $w_2$ toca apenas $y$ (conjuntos de escrita disjuntos $\mathcal{W}_1 \cap \mathcal{W}_2 = \emptyset$), porém seus conjuntos de leitura intersectam as escritas da outra: $\mathcal{R}_1 \cap \mathcal{W}_2 \neq \emptyset$ e $\mathcal{R}_2 \cap \mathcal{W}_1 \neq \emptyset$.

---

## 4. Invariantes de Estado e Transições

O banco de dados é modelado como um sistema de transição de estados:
$$\mathcal{M} = (\Sigma, \sigma_0, \Delta, \mathcal{I})$$

Onde:
* $\Sigma$ é o universo de estados do banco de dados (tabelas, tuplas, índices).
* $\sigma_0 \in \Sigma$ é o estado inicial definido pelo DDL (`schema.sql`) e semente (`seed.sql`).
* $\Delta: \Sigma \times \mathcal{T} \to \Sigma$ é a função de transição que aplica uma transação ao estado atual.
* $\mathcal{I}: \Sigma \to \{0, 1\}$ é o predicado da invariante de negócio.

### 4.1 Invariante Indutiva
Um predicado $\mathcal{I}$ é indutivo para $\mathcal{M}$ se e somente se:
1. **Base:** $\mathcal{I}(\sigma_0) = 1$
2. **Passo Indutivo:** $\forall \sigma \in \Sigma, \forall T \in \mathcal{T}: (\mathcal{I}(\sigma) = 1) \implies (\mathcal{I}(\Delta(\sigma, T)) = 1)$

### 4.2 Falha por Intercalação não-atômica
Na execução caótica, a transição não ocorre em bloco atômico $\Delta(\sigma, T)$, mas sim em micro-passos $s_{i,k}$:
$$\sigma_{t+1} = \delta_{\text{step}}(\sigma_t, s_{i,k})$$

Se a história $S$ não for estritamente serializavel ($S \not\in \text{CSR}$), o estado final converge para:
$$\sigma_{\text{final}} = \delta_{\text{step}}(\dots \delta_{\text{step}}(\sigma_0, s_{1,1}) \dots s_{n,m}) \quad \text{tal que} \quad \mathcal{I}(\sigma_{\text{final}}) = 0$$

---

## 5. Teoria e Prova de Convergência do Delta-Debugging ($ddmin$)

O objetivo do módulo Shrinker do ChaosSQL é isolar a menor subsequência de transações que ainda falha a invariante.

### 5.1 Definição de $1$-Minimalidade
Seja $C = \langle op_1, op_2, \dots, op_N \rangle$ uma sequência ordenada de operações agendadas.
Seja a função orâculo $\text{test}: \mathcal{P}(C) \to \{\text{PASS}, \text{FAIL}\}$.
Sabemos que $\text{test}(C) = \text{FAIL}$ e $\text{test}(\emptyset) = \text{PASS}$.

Uma subsequência $C^* \subseteq C$ é dita **$1$-Minimal** se:
$$\text{test}(C^*) = \text{FAIL} \quad \land \quad \forall op \in C^*: \text{test}(C^* \setminus \{op\}) = \text{PASS}$$

Isto significa que nenhuma operação individual pode ser removida de $C^*$ sem que a falha desapareça.

### 5.2 Complexidade de $ddmin$
1. **Melhor Caso (Busca Binária Direta):**
  $$\text{Complexidade} = O(\log |C|)$$
2. **Pior Caso (Todas as operações interagem linearmente):**
  $$\text{Complexidade} = O(|C|^2)$$

---

## 6. Espaço Combinatório e Jitter

Para $W$ workers concorrentes executando $M$ passos por transação com $K$ transações no total, o número total de intercalações possíveis $\Omega$ é dado pelo coeficiente multinomial:

$$|\Omega| = \frac{(K \cdot M)!}{(M!)^K}$$

Para $K = 5$ transações de $M = 3$ passos:
$$|\Omega| = \frac{15!}{(3!)^5} = \frac{1.307.674.368.000}{7.776} \approx 168.168.000 \text{ intercalações possíveis}$$

Ao injetar jitter estocástico $d_i \sim \mathcal{U}(\delta_{\min}, \delta_{\max})$ com $\delta_{\max} \approx 20\text{ms}$:
$$\mathbb{P}(\text{Intercalação Crítica}) \to 1 - e^{-\lambda \cdot \tau_{\text{crit}}}$$

A probabilidade de interceptação da condição de corrida aumenta em várias ordens de grandeza em baterias curtas ($N \le 50$ operações).
