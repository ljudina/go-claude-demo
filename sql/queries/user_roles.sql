-- name: AssignRoleToUser :exec
INSERT OR IGNORE INTO user_roles (user_id, role_id)
VALUES (?, ?);

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = ? AND role_id = ?;

-- name: GetUserRoles :many
SELECT r.id, r.name, r.description
FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = ?
ORDER BY r.name;

-- name: GetUsersWithRole :many
SELECT u.id, u.email, u.name, u.created_at, u.updated_at
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role_id = ?
ORDER BY u.id;

-- name: ClearUserRoles :exec
DELETE FROM user_roles
WHERE user_id = ?;
