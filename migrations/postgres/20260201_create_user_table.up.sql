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

CREATE INDEX name_index ON users (username);
CREATE INDEX email_index ON users (email);
CREATE INDEX phone_number_index ON users (phone_number);
CREATE INDEX created_at_index ON users (created_at);