CREATE TABLE parent_orders (
    id INT PRIMARY KEY,
    customer_id INT NOT NULL,
    status TEXT NOT NULL,
    total_cents INT NOT NULL
);

CREATE TABLE child_items (
    id INT PRIMARY KEY,
    order_id INT NOT NULL,
    sku TEXT NOT NULL,
    price_cents INT NOT NULL,
    FOREIGN KEY (order_id) REFERENCES parent_orders(id) ON DELETE CASCADE
);
