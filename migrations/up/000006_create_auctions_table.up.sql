CREATE TABLE IF NOT EXISTS auctions (
    id        TEXT PRIMARY KEY,
    item_id   TEXT NOT NULL REFERENCES items(id),
    seller_id TEXT NOT NULL REFERENCES guilds(id),
    ends_at   TIMESTAMPTZ NOT NULL,
    status    TEXT NOT NULL CHECK (status IN ('active', 'ended')) DEFAULT 'active'
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_one_active_auction_per_item
    ON auctions (item_id) WHERE status = 'active';
