-- name: CreateChirp :one
INSERT INTO chirps(body, user_id)
VALUES($1, $2)
RETURNING *;

-- name: DeleteUserChirp :exec
DELETE FROM chirps WHERE id = $1;
