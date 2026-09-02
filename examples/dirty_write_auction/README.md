# Cenário 05: Auction Bidding Dirty Write (Anomalia G0)

## Contexto de Negócio
Em uma plataforma de leilões virtuais em tempo real, múltiplos participantes submetem lances concorrentes para o mesmo item de colecionador (`auction_items:1`).
Para cada lance submetido, o sistema atualiza o maior valor ofertado (`highest_bid`), registra o histórico em `bids_log` e atribui o identificador do vencedor provisório (`winner_id`).

## O Bug (Dirty Write / G0 sob Concorrência Desprotegida)
A anomalia **G0 (Dirty Write)** ocorre quando duas transações modificam simultaneamente os mesmos itens ou colunas sem controle de concorrência ou serialização atômica, resultando em escritas cruzadas e sobrescrita de dados não comitados (ou intermediários):
1. **Licitante 1 ($T_1$)** submete um lance de R$ 200:
   - Atualiza `highest_bid = 200` na tabela `auction_items`.
2. **Licitante 2 ($T_2$)** concorrentemente submete um lance de R$ 350:
   - Atualiza `highest_bid = 350`.
   - Insere o registro em `bids_log (item_id=1, bidder_id=202, bid_amount=350)`.
   - Atualiza `winner_id = 202`.
3. **Licitante 1 ($T_1$)** finaliza sua sequência desprotegida:
   - Insere o registro em `bids_log (item_id=1, bidder_id=101, bid_amount=200)`.
   - Atualiza `winner_id = 101`.
4. **Estado Final Inconsistente:** O item registra `highest_bid = 350`, porém com `winner_id = 101`!
   O Licitante 1 (que ofertou R$ 200) é declarado vencedor provisório do leilão com o valor de R$ 350 ofertado pelo Licitante 2.

### Definição Matemática Formal (Adya / Berenson et al.)
No grafo de dependências diretas de Adya (DSG), a anomalia $G0$ (Dirty Write) é caracterizada por um ciclo composto puramente por dependências diretas de escrita-escrita ($ww$):
$$T_1 \xrightarrow{ww} T_2 \xrightarrow{ww} T_1$$

Onde:
- $T_1 \xrightarrow{ww} T_2$ em `auction_items:1`: $T_2$ sobrescreveu a coluna `highest_bid` previamente escrita por $T_1$.
- $T_2 \xrightarrow{ww} T_1$ em `auction_items:1`: $T_1$ sobrescreveu a coluna `winner_id` previamente escrita por $T_2$.

## Invariante de Consistência do Leilão
O vencedor registrado em `auction_items` deve coincidir estritamente com o participante que ofertou o `highest_bid` registrado:
$$\text{is\_consistent} == 1 \lor \text{is\_consistent} == \text{true}$$

## Como Corrigir
* **Atualização Atômica com Validação Condicional:**
  Executar o update de `highest_bid` e `winner_id` em uma única instrução SQL atômica com predicado de validação:
  ```sql
  UPDATE auction_items 
  SET highest_bid = :new_bid, winner_id = :bidder_id 
  WHERE id = 1 AND :new_bid > highest_bid;
  ```
* **Nível de Isolamento SERIALIZABLE ou Locks Pessimistas (`SELECT ... FOR UPDATE`):**
  Bloquear o registro do item durante a validação do lance e posterior escrita, impedindo que transações concorrentes intercalem escritas parciais.
