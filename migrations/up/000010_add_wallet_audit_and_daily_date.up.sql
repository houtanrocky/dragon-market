ALTER TABLE guilds
    ADD COLUMN daily_spent_on DATE NOT NULL DEFAULT CURRENT_DATE,
    ADD CONSTRAINT guilds_gold_nonnegative CHECK (gold >= 0),
    ADD CONSTRAINT guilds_reserved_valid CHECK (reserved >= 0 AND reserved <= gold),
    ADD CONSTRAINT guilds_daily_values_nonnegative CHECK (daily_limit >= 0 AND daily_spent >= 0);

CREATE TABLE wallet_transactions (
    id           BIGSERIAL PRIMARY KEY,
    guild_id     TEXT NOT NULL REFERENCES guilds(id),
    operation    TEXT NOT NULL CHECK (operation IN ('reserve', 'release', 'spend', 'deduct', 'earn')),
    amount       BIGINT NOT NULL CHECK (amount > 0),
    gold_after   BIGINT NOT NULL CHECK (gold_after >= 0),
    reserved_after BIGINT NOT NULL CHECK (reserved_after >= 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallet_transactions_guild_created
    ON wallet_transactions (guild_id, created_at DESC);
