# Flash Crash Liquidation Dirty Read ($G1a$)

## Business Context
In decentralized lending and collateral protocols (such as MakerDAO or Aave), a market maker submits a batch of price updates to an on-chain oracle. During extreme volatility, a transaction updates the price of ETH to an erroneous flash crash value ($1500 instead of $3000) before aborting or reverting due to a failed slippage check.

Concurrently, an autonomous liquidation bot executes under `READ UNCOMMITTED` or weak snapshot isolation, reads the uncommitted flash crash price ($1500), and liquidates a healthy loan vault ($10\text{ ETH} \times \$3000 = \$30,000 > \$24,000 \text{ threshold}$).

## Mathematical Formulation ($G1a$)
According to Atul Adya (1999) and Berenson et al. (1995), Phenomenon $G1a$ (Aborted Read / Dirty Read) is defined as:

$$w_1(x) \dots r_2(x) \dots (a_1 \text{ and } c_2 \text{ in any order})$$

Where:
- $w_1(x)$ updates oracle price to $\$1500$.
- $r_2(x)$ reads dirty uncommitted price $\$1500$.
- $a_1$ transaction 1 aborts / rolls back.
- $c_2$ transaction 2 commits an irreversible state change (vault liquidation).

## Remediation
Ensure the database operates at `READ COMMITTED` or higher, preventing readers from acquiring dirty pages or uncommitted undo records.
