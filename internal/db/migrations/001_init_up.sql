CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    endpoint_url TEXT NOT NULL,
    UNIQUE (event_type, endpoint_url)
);

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL
);

CREATE TYPE delivery_status AS ENUM (
    'pending',
    'processing',
    'delivered',
    'failed'
);

CREATE TABLE IF NOT EXISTS deliveries (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL,
    endpoint_url TEXT NOT NULL,
    status delivery_status NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ,
    last_error TEXT
);