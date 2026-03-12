-- name: InsertWatchlistEntry :one
INSERT INTO watchlist (user_id, product_id, alert_sale, alert_restock)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetWatchlistByUser :many
SELECT
  w.id,
  w.user_id,
  w.product_id,
  w.alert_sale,
  w.alert_restock,
  w.created_at,
  p.name        AS product_name,
  p.slug        AS product_slug,
  p.image_url   AS product_image_url,
  p.current_price AS product_current_price,
  p.currency    AS product_currency,
  p.in_stock    AS product_in_stock
FROM watchlist w
INNER JOIN products p ON p.id = w.product_id
WHERE w.user_id = $1
ORDER BY w.created_at DESC;


-- name: GetWatchlistEntry :one
SELECT id, user_id, product_id, alert_sale, alert_restock, created_at
FROM watchlist
WHERE id = $1 AND user_id = $2;


-- name: UpdateWatchlistEntry :one
UPDATE watchlist
SET
  alert_sale    = $1,
  alert_restock = $2
WHERE id = $3 AND user_id = $4
RETURNING *;


-- name: DeleteWatchlistEntry :exec
DELETE FROM watchlist
WHERE id = $1 AND user_id = $2;


-- name: CountUserWatchlist :one
SELECT COUNT(*) FROM watchlist
WHERE user_id = $1;


-- name: GetWatchersByProduct :many
SELECT
  w.id          AS watchlist_id,
  w.user_id,
  w.alert_sale,
  w.alert_restock,
  u.email,
  u.name,
  u.notify_email
FROM watchlist w
INNER JOIN users u ON u.id = w.user_id
WHERE w.product_id = $1;
