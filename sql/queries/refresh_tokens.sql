-- name: CreateToken :one
INSERT INTO refresh_tokens(id, expires_at, user_id)
VALUES($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET
    id = EXCLUDED.id,
    expires_at = EXCLUDED.expires_at,
    user_id = EXCLUDED.user_id
RETURNING *;

-- name: GetTokenByID :one
SELECT *
FROM refresh_tokens
WHERE id = $1;

-- name: GetTokenByUserID :one
SELECT *
FROM refresh_tokens
WHERE user_id = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET expires_at = $1, updated_at = $2
WHERE id = $3;
