-- name: CreateChirp :one
INSERT INTO chirps(body, user_id)
VALUES($1, $2)
RETURNING *;

-- name: GetChirpById :one
SELECT * FROM chirps Where id = $1;

-- name: GetAllChirps :many
SELECT * FROM chirps;

-- name: DeleteChirpByUserID :exec
DELETE FROM chirps WHERE id = $1 AND user_id = $2;

