CREATE TABLE IF NOT EXISTS users
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username     VARCHAR(64) UNIQUE,
    email        VARCHAR(255) UNIQUE,
    password     VARCHAR(255)                   NOT NULL,
    phone_number VARCHAR(64) UNIQUE             NOT NULL,
    role         INTEGER          DEFAULT 1     NOT NULL,
    created_at   TIMESTAMP        DEFAULT NOW() NOT NULL,
    updated_at   TIMESTAMP
);

ALTER TABLE device_tokens
    ALTER COLUMN user_id TYPE UUID USING user_id::uuid;

ALTER TABLE device_tokens
    ADD CONSTRAINT device_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
