CREATE TABLE IF NOT EXISTS limit_orders (
    id        TEXT PRIMARY KEY,
    item_id   TEXT NOT NULL REFERENCES items(id),
    seller_id TEXT NOT NULL REFERENCES guilds(id),
    buyer_id  TEXT REFERENCES guilds(id),
    price     BIGINT NOT NULL,
    status    TEXT NOT NULL CHECK (status IN ('listed', 'sold', 'canceled')) DEFAULT 'listed',
    listed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
