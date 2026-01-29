-- name: CreateNavItem :one
INSERT INTO nav_items (name, url, icon, parent_id, sort_order, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, name, url, icon, parent_id, sort_order, created_at, updated_at;

-- name: GetNavItemByID :one
SELECT id, name, url, icon, parent_id, sort_order, created_at, updated_at
FROM nav_items
WHERE id = ?;

-- name: ListNavItems :many
SELECT id, name, url, icon, parent_id, sort_order, created_at, updated_at
FROM nav_items
ORDER BY sort_order, id;

-- name: ListTopLevelNavItems :many
SELECT id, name, url, icon, parent_id, sort_order, created_at, updated_at
FROM nav_items
WHERE parent_id IS NULL
ORDER BY sort_order, id;

-- name: ListNavItemChildren :many
SELECT id, name, url, icon, parent_id, sort_order, created_at, updated_at
FROM nav_items
WHERE parent_id = ?
ORDER BY sort_order, id;

-- name: UpdateNavItem :one
UPDATE nav_items
SET name = ?, url = ?, icon = ?, parent_id = ?, sort_order = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, name, url, icon, parent_id, sort_order, created_at, updated_at;

-- name: DeleteNavItem :exec
DELETE FROM nav_items
WHERE id = ?;

-- name: AssignRoleToNavItem :exec
INSERT OR IGNORE INTO nav_item_roles (nav_item_id, role_id)
VALUES (?, ?);

-- name: RemoveRoleFromNavItem :exec
DELETE FROM nav_item_roles
WHERE nav_item_id = ? AND role_id = ?;

-- name: ClearNavItemRoles :exec
DELETE FROM nav_item_roles
WHERE nav_item_id = ?;

-- name: GetNavItemRoles :many
SELECT r.id, r.name, r.description
FROM roles r
JOIN nav_item_roles nir ON r.id = nir.role_id
WHERE nir.nav_item_id = ?
ORDER BY r.name;

-- name: GetNavItemsForRoles :many
SELECT DISTINCT ni.id, ni.name, ni.url, ni.icon, ni.parent_id, ni.sort_order, ni.created_at, ni.updated_at
FROM nav_items ni
JOIN nav_item_roles nir ON ni.id = nir.nav_item_id
WHERE nir.role_id IN (sqlc.slice('role_ids'))
ORDER BY ni.sort_order, ni.id;

-- name: GetNavItemByName :one
SELECT id, name, url, icon, parent_id, sort_order, created_at, updated_at
FROM nav_items
WHERE name = ?;
