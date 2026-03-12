-- name: InsertToken :exec
INSERT INTO tokens (hash, user_id, expiry, scope) 
        VALUES ($1, $2, $3, $4);



-- name: DeleteAllForUser :exec
DELETE FROM tokens 
        WHERE scope = $1 AND user_id = $2;


-- name: GetUserToken :one
SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated
FROM users
INNER JOIN tokens
ON users.id = tokens.user_id
WHERE tokens.hash = $1
AND tokens.scope = $2 
AND tokens.expiry > $3;