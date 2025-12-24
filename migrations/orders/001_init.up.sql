CREATE TABLE IF NOT EXISTS orders (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL,
    amount      BIGINT NOT NULL CHECK (amount > 0),
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL CHECK (status IN ('NEW','FINISHED','FAILED')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

CREATE TABLE IF NOT EXISTS outbox (
    id            BIGSERIAL PRIMARY KEY,
    topic         TEXT NOT NULL,
    key           TEXT NOT NULL,
    payload       JSONB NOT NULL,
    processed_at  TIMESTAMPTZ NULL,
    attempts      INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_until  TIMESTAMPTZ NULL,
    last_error    TEXT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_outbox_ready
    ON outbox(processed_at, next_retry_at)
    WHERE processed_at IS NULL;
