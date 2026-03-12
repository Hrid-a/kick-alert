-- name: InsertPriceHistory :one
INSERT INTO price_history (product_id, price, in_stock)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetPriceHistoryByProduct :many
SELECT id, product_id, price, in_stock, scraped_at
FROM price_history
WHERE product_id = $1
ORDER BY scraped_at DESC
LIMIT $2;
