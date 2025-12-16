-- name: CreateUser :one
INSERT INTO users(id, email, password)
VALUES($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserByEmailAndPassword :exec
UPDATE users
SET email = $1, password = $2
WHERE id = $3;

-- name: DeleteAllUsers :exec
DELETE FROM users;
