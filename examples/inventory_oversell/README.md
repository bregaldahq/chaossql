# Cenário 02: E-Commerce Inventory Oversell

## Contexto de Negócio
Uma promoção relâmpago disponibiliza apenas **10 unidades** de uma Super GPU. Dezenas de usuários tentam comprar simultaneamente.

## O Bug (Condição de Corrida de Estoque)
1. **Workers leem o estoque** (ex: `stock = 1`) e consideram que a compra é válida.
2. M workers criam suas ordens (insert em `orders`) e decrementam o estoque.
3. **Resultado Caótico:** Foram vendidas 14 unidades, mas havia apenas 10 em estoque!

## Invariante de Negócio
$$\text{Estoque Restante} + \text{Total Vendido} == 10 \quad \land \quad \text{Total Vendido} \le 10$$

## Como Corrigir
* **Decremento Atômico com Guard:**
  ```sql
  UPDATE products SET stock = stock - 1 WHERE id = 1 AND stock >= 1;
  ```
  Se nenhuma linha for afetada, a compra é rejeitada imediatamente.
