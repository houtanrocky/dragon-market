-- 1. Add back the available column
ALTER TABLE items ADD COLUMN available BOOLEAN NOT NULL DEFAULT true;

-- 2. Backfill available from status
UPDATE items SET available = true WHERE status = 'free';
UPDATE items SET available = false WHERE status IN ('listed_in_order', 'listed_in_auction');

-- 3. Drop the status column
ALTER TABLE items DROP COLUMN status;