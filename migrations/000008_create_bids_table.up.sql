CREATE TABLE IF NOT EXISTS bids (
    id         TEXT PRIMARY KEY,
    auction_id TEXT NOT NULL REFERENCES auctions(id),
    bidder_id  TEXT NOT NULL REFERENCES guilds(id),
    amount     NUMERIC(20,2) NOT NULL,
    placed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- used to quickly find the top bid for an auction
CREATE INDEX IF NOT EXISTS idx_bids_auction_amount ON bids (auction_id, amount DESC);
