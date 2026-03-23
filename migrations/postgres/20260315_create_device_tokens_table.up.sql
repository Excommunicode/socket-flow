CREATE TABLE IF NOT EXISTS device_tokens
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token      VARCHAR(512) NOT NULL,
    platform   VARCHAR(32)  NOT NULL,
    created_at TIMESTAMP    DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP    DEFAULT NOW() NOT NULL
);

CREATE UNIQUE INDEX device_tokens_token_key ON device_tokens (token);
CREATE INDEX device_tokens_user_id_idx ON device_tokens (user_id);
