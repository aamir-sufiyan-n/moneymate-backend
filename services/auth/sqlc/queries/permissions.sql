-- name: CreatePermission :one
INSERT INTO auth.permissions (
    id,
    name,
    description
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetPermissionByID :one
SELECT *
FROM auth.permissions
WHERE id = $1;

-- name: GetPermissionByName :one
SELECT *
FROM auth.permissions
WHERE name = $1;

-- name: ListPermissions :many
SELECT *
FROM auth.permissions
ORDER BY name;

-- name: DeletePermission :exec
DELETE FROM auth.permissions
WHERE id = $1;