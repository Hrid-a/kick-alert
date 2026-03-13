-- name: InsertUser :one
INSERT INTO users (email, name, password_hash) 
        VALUES ($1, $2, $3)
RETURNING *;


-- name: GetUserByEmail :one
SELECT id, created_at, name, email, password_hash, updated_at, activated, tier
        FROM users
        WHERE email = $1;


-- name: GetUserById :one
SELECT id, created_at, name, email, updated_at, activated, tier
        FROM users
        WHERE id = $1;


-- name: UpdateUser :exec
UPDATE users
        SET name = $1, email = $2, password_hash = $3, activated = $4
        WHERE id = $5;

-- name: UpdateNotifyEmail :exec
UPDATE users SET notify_email = $1 WHERE id = $2;