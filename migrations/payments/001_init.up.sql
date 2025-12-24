CREATE TABLE IF NOT EXISTS accounts (
    user_id    UUID PRIMARY KEY,
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS inbox (
    message_id  UUID PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
    order_id   UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    amount     BIGINT NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);

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
