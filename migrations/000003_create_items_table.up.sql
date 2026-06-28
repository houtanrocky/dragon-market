CREATE TABLE IF NOT EXISTS items (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    type       TEXT NOT NULL CHECK (type IN ('common', 'rare', 'legendary')),
    owner_id   TEXT NOT NULL REFERENCES guilds(id),
    available  BOOLEAN NOT NULL DEFAULT true,
    base_price NUMERIC(20,2) NOT NULL DEFAULT 0
);

