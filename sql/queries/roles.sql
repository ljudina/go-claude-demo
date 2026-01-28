-- name: CreateRole :one
INSERT INTO roles (name, description)
VALUES (?, ?)
RETURNING id, name, description;

-- name: GetRoleByID :one
SELECT id, name, description
FROM roles
WHERE id = ?;

-- name: GetRoleByName :one
SELECT id, name, description
FROM roles
WHERE name = ?;

-- name: ListRoles :many
SELECT id, name, description
FROM roles
ORDER BY name;

-- name: UpdateRole :one
UPDATE roles
SET name = ?, description = ?
WHERE id = ?
RETURNING id, name, description;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = ?;
