-- Development-only data for manual API testing.
-- Kept outside migrations so deployments do not receive fixture guilds.
INSERT INTO guilds (id, gold, reserved, daily_limit, daily_spent, daily_spent_on)
VALUES
    ('guild-seller', 10000, 0, 5000, 0, CURRENT_DATE),
    ('guild-buyer', 10000, 0, 5000, 0, CURRENT_DATE),
    ('guild-bidder', 10000, 0, 5000, 0, CURRENT_DATE)
ON CONFLICT (id) DO NOTHING;
