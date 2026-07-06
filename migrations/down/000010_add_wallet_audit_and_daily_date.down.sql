DROP TABLE IF EXISTS wallet_transactions;
ALTER TABLE guilds
    DROP CONSTRAINT IF EXISTS guilds_daily_values_nonnegative,
    DROP CONSTRAINT IF EXISTS guilds_reserved_valid,
    DROP CONSTRAINT IF EXISTS guilds_gold_nonnegative,
    DROP COLUMN IF EXISTS daily_spent_on;
