-- +goose Up
CREATE TABLE IF NOT EXISTS watchlist (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  product_id     UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  alert_sale     BOOLEAN NOT NULL DEFAULT true,
  alert_restock  BOOLEAN NOT NULL DEFAULT true,
  created_at     TIMESTAMPTZ DEFAULT now(),
  UNIQUE(user_id, product_id)
);

-- +goose Down
DROP TABLE IF EXISTS watchlist;
