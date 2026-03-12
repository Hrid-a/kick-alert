-- +goose Up
CREATE TABLE IF NOT EXISTS products (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug            TEXT UNIQUE NOT NULL,   -- 'air-max-ishod-iron-grey-and-photon-dust-1'
  name            TEXT NOT NULL,
  sku             TEXT NOT NULL, 
  external_id     TEXT UNIQUE NOT NULL,   -- Nike cloudProductId
  category        TEXT NOT NULL DEFAULT 'FOOTWEAR',  -- 'FOOTWEAR' | 'APPAREL' | 'EQUIPMENT'
  url             TEXT NOT NULL,
  image_url       TEXT NOT NULL,
  current_price   NUMERIC NOT NULL,
  currency        TEXT NOT NULL DEFAULT 'USD',
  in_stock        BOOLEAN NOT NULL,
  last_scraped_at TIMESTAMPTZ NOT NULL,
  created_at      TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS products;
