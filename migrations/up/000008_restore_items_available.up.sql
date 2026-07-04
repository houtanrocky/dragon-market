UPDATE items
SET status = CASE
                 WHEN available THEN 'free'
                 ELSE 'listed_in_order'
    END;

ALTER TABLE items DROP COLUMN available;