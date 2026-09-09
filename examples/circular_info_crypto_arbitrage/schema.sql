CREATE TABLE dex_pools (
    id INT PRIMARY KEY,
    name TEXT NOT NULL,
    price_eth INT NOT NULL
);

CREATE TABLE trade_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dex_id INT NOT NULL,
    trader TEXT NOT NULL,
    amount INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
