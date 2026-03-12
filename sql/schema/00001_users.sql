-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email           citext UNIQUE NOT NULL,
  password_hash   TEXT NOT NULL,
  name            TEXT NOT NULL,
  activated       BOOLEAN DEFAULT false,
  notify_email    BOOLEAN DEFAULT true,
  notify_push     BOOLEAN DEFAULT false,
  tier            TEXT DEFAULT 'free',   -- 'free' | 'pro'
  created_at      TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS users;
