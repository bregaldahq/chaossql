INSERT INTO parent_orders (id, customer_id, status, total_cents) VALUES (1, 10, 'OPEN', 5000);
INSERT INTO parent_orders (id, customer_id, status, total_cents) VALUES (2, 20, 'OPEN', 4000);

INSERT INTO child_items (id, order_id, sku, price_cents) VALUES (1, 1, 'ITEM-101', 2500);
INSERT INTO child_items (id, order_id, sku, price_cents) VALUES (2, 1, 'ITEM-102', 2500);
INSERT INTO child_items (id, order_id, sku, price_cents) VALUES (3, 2, 'ITEM-201', 2000);
INSERT INTO child_items (id, order_id, sku, price_cents) VALUES (4, 2, 'ITEM-202', 2000);
