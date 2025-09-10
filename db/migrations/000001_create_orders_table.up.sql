CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR PRIMARY KEY,
    track_number VARCHAR,
    entry VARCHAR,
    locale VARCHAR,
    internal_signature VARCHAR,
    customer_id VARCHAR,
    delivery_service VARCHAR,
    shardkey VARCHAR,
    sm_id INT,
    date_created TIMESTAMPTZ,
    oof_shard VARCHAR
);

CREATE TABLE IF NOT EXISTS delivery (
    order_uid VARCHAR PRIMARY KEY REFERENCES orders(order_uid),
    name VARCHAR,
    phone VARCHAR,
    zip VARCHAR,
    city VARCHAR,
    address VARCHAR,
    region VARCHAR,
    email VARCHAR
);

CREATE TABLE IF NOT EXISTS payment (
    transaction VARCHAR PRIMARY KEY,
    order_uid VARCHAR UNIQUE REFERENCES orders(order_uid),
    request_id VARCHAR,
    currency VARCHAR,
    provider VARCHAR,
    amount BIGINT,
    payment_dt BIGINT,
    bank VARCHAR,
    delivery_cost BIGINT,
    goods_total BIGINT,
    custom_fee BIGINT
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR REFERENCES orders(order_uid),
    chrt_id BIGINT,
    track_number VARCHAR,
    price BIGINT,
    rid VARCHAR,
    name VARCHAR,
    sale INT,
    size VARCHAR,
    total_price BIGINT,
    nm_id BIGINT,
    brand VARCHAR,
    status INT
);

CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);
