CREATE TABLE idempotency_records
(
    key          TEXT PRIMARY KEY,
    operation    TEXT        NOT NULL,
    request_hash TEXT        NOT NULL,
    status_code  INTEGER,
    response     JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    CHECK (
        (completed_at IS NULL AND status_code IS NULL AND response IS NULL)
            OR
        (completed_at IS NOT NULL AND status_code IS NOT NULL)
        )
);

CREATE INDEX idx_idempotency_records_created_at
    ON idempotency_records (created_at);