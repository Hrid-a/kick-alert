-- +goose Up
CREATE TABLE IF NOT EXISTS price_history (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  price       NUMERIC NOT NULL,
  in_stock    BOOLEAN NOT NULL,
  scraped_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE notification_type AS ENUM ('PRICE_DROP', 'RESTOCK');

CREATE TABLE IF NOT EXISTS notifications (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  product_id       UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  watchlist_id     UUID NOT NULL REFERENCES watchlist(id) ON DELETE CASCADE,
  type             notification_type NOT NULL,
  old_price        NUMERIC,
  new_price        NUMERIC,
  read             BOOLEAN NOT NULL DEFAULT false,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;
DROP TABLE IF EXISTS price_history;
