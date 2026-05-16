ALTER TABLE device_tokens DROP CONSTRAINT IF EXISTS device_tokens_user_id_fkey;

ALTER TABLE device_tokens
    ALTER COLUMN user_id TYPE TEXT USING user_id::text;

DROP TABLE IF EXISTS users;
