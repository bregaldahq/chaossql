CREATE TABLE vault_positions (
    id INT PRIMARY KEY,
    owner TEXT NOT NULL,
    collateral_eth INT NOT NULL,
    debt_usd INT NOT NULL,
    is_liquidated INT DEFAULT 0
);

CREATE TABLE oracle_prices (
    id INT PRIMARY KEY,
    asset TEXT NOT NULL,
    price_usd INT NOT NULL
);
