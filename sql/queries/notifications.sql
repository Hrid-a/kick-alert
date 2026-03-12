-- name: InsertNotification :one
INSERT INTO notifications (user_id, product_id, watchlist_id, type, old_price, new_price)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;


-- name: GetNotificationsByUser :many
SELECT
  n.id,
  n.user_id,
  n.product_id,
  n.watchlist_id,
  n.type,
  n.old_price,
  n.new_price,
  n.read,
  n.created_at,
  p.name       AS product_name,
  p.image_url  AS product_image_url,
  p.slug       AS product_slug
FROM notifications n
INNER JOIN products p ON p.id = n.product_id
WHERE n.user_id = $1
  AND (sqlc.narg('read_filter')::boolean IS NULL OR n.read = sqlc.narg('read_filter'))
ORDER BY n.created_at DESC
LIMIT $2 OFFSET $3;


-- name: CountNotificationsByUser :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1
  AND (sqlc.narg('read_filter')::boolean IS NULL OR read = sqlc.narg('read_filter'));


-- name: MarkNotificationRead :one
UPDATE notifications
SET read = true
WHERE id = $1 AND user_id = $2
RETURNING *;


-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read = true
WHERE user_id = $1 AND read = false;
