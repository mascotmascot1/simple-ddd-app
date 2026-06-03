CREATE TABLE IF NOT EXISTS "customers"(
    customer_id UUID PRIMARY KEY 
);

CREATE TABLE IF NOT EXISTS "products"(
    product_id UUID PRIMARY KEY
);

----------------------------INSERT-BASE-VALUES----------------------------
INSERT INTO customers (customer_id) VALUES 
('11111111-1111-1111-1111-111111111111'),
('22222222-2222-2222-2222-222222222222')
ON CONFLICT DO NOTHING;

INSERT INTO products (product_id) VALUES 
('33333333-3333-3333-3333-333333333333'),
('44444444-4444-4444-4444-444444444444')
ON CONFLICT DO NOTHING;
----------------------------INSERT-BASE-VALUES----------------------------

CREATE TABLE IF NOT EXISTS "orders"(
    order_id UUID PRIMARY KEY,
    customer_id UUID REFERENCES customers (customer_id) NOT NULL, 
    total_amount BIGINT NOT NULL, -- Тип bigint занимает 8 байт и является точным аналогом int64 на стороне базы данных.
    currency VARCHAR(8) NOT NULL
);
CREATE INDEX orders_customer_id ON orders(customer_id);

CREATE TABLE IF NOT EXISTS "order_items"(
    order_id UUID REFERENCES orders (order_id) ON DELETE CASCADE, -- При удалении агрегата в таблице orders СУБД автоматически и атомарно очистит все дочерние позиции.
    product_id UUID REFERENCES products (product_id),
    amount BIGINT NOT NULL,
    currency VARCHAR(8) NOT NULL,
    quantity BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id) -- быстрый поиск будет или по ордер или по оредер+ продакт, а отдельно продакт не катит быстро 
);
CREATE INDEX order_items_product_id ON order_items(product_id);

-- В таблице order_items реализуется связь «один ко многим». Если бы ты оставил 
-- order_id единственным первичным ключом в этой таблице, база жестко ограничила 
-- бы заказ ровно одной позицией. Попытка добавить второй товар в тот же заказ вызвала 
-- бы ошибку нарушения уникальности.

-- Создавая составной ключ primary key (order_id, product_id), ты заставляешь базу данных
-- оценивать уникальность именно по комбинации этих двух полей.