-- name: InsertProduct :one
INSERT INTO products (
    slug,
    name,
    sku,
    external_id,
    category,
    url,
    image_url,
    current_price,
    currency,
    in_stock,
    last_scraped_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;


-- name: GetProductById :one
SELECT id, name, slug, sku, last_scraped_at, external_id, category, url, image_url, currency, current_price, in_stock FROM products 
WHERE id=$1;


-- name: SearchProducts :many
SELECT * FROM products
WHERE
  ($1::text = '' OR to_tsvector('simple', name) @@ plainto_tsquery('simple', $1::text))
  AND ($2::text = '' OR category = $2::text)
  AND ($3::text = '' OR in_stock = ($3::text)::bool)
  AND ($4::text = '' OR current_price >= ($4::text)::numeric)
  AND ($5::text = '' OR current_price <= ($5::text)::numeric)
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;


-- name: UpdateProduct :exec
UPDATE products
SET current_price = $2, in_stock = $3, last_scraped_at = now()
WHERE id = $1;


-- name: GetProductByExternalID :one
SELECT * FROM products WHERE external_id = $1;


-- name: GetStaleProducts :many
SELECT DISTINCT p.id, p.slug, p.name, p.sku, p.external_id, p.category, p.url, p.image_url, p.current_price, p.currency, p.in_stock, p.last_scraped_at, p.created_at
FROM products p
INNER JOIN watchlist w ON w.product_id = p.id
WHERE p.last_scraped_at < now() - interval '5 minutes';