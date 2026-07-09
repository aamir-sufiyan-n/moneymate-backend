-- name: CreateRole :one
INSERT INTO auth.roles (
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

-- name: GetRoleByID :one
SELECT *
FROM auth.roles
WHERE id = $1;

-- name: GetRoleByName :one
SELECT *
FROM auth.roles
WHERE name = $1;

-- name: ListRoles :many
SELECT *
FROM auth.roles
ORDER BY name;

-- name: UpdateRole :exec
UPDATE auth.roles
SET
    name = $2,
    description = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteRole :exec
DELETE FROM auth.roles
WHERE id = $1;