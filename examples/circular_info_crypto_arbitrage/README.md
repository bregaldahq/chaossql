# Crypto Arbitrage Circular Information Flow ($G1c$)

## Business Context
In decentralized finance (DeFi), automated market maker (AMM) arbitrage bots monitor relative pool prices between Uniswap and Sushiswap. When pool $1$ updates its price, Bot 1 reads pool $2$'s price to execute a swap. Concurrently, Bot 2 updates pool $2$ and reads pool $1$'s price.

## Mathematical Formulation ($G1c$)
According to Adya's direct serialization graph $SG(S)$, a Circular Information Flow anomaly occurs when transactions observe each other's uncommitted or intermediate states through pure read-after-write ($wr$) dependencies:

$$T_1 \xrightarrow{wr} T_2 \xrightarrow{wr} T_1$$

## Remediation
Ensure both pools are locked atomically within a single transaction using `SERIALIZABLE` isolation or explicit lock ordering.
