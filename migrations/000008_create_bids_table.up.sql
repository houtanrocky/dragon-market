CREATE TABLE IF NOT EXISTS bids (
    id         TEXT PRIMARY KEY,
    auction_id TEXT NOT NULL REFERENCES auctions(id),
    bidder_id  TEXT NOT NULL REFERENCES guilds(id),
    amount     BIGINT NOT NULL CHECK (amount > 0),
    placed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status     TEXT NOT NULL CHECK (status IN ('active', 'outbid', 'cancelled', 'winning')) DEFAULT 'active'
);

-- used to quickly find the top bid for an auction
CREATE INDEX IF NOT EXISTS idx_bids_auction_amount ON bids (auction_id, amount DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_one_active_bid_per_auction
    ON bids (auction_id) WHERE status = 'active';
