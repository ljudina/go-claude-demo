-- name: CreateUser :one
INSERT INTO users (email, name, created_at, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, email, name, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, name, created_at, updated_at
FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT id, email, name, created_at, updated_at
FROM users
WHERE email = ?;

-- name: ListUsers :many
SELECT id, email, name, created_at, updated_at
FROM users
ORDER BY id;

-- name: UpdateUser :one
UPDATE users
SET email = ?, name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, email, name, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;
