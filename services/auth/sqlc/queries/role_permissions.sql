-- name: AssignPermissionToRole :exec
INSERT INTO auth.role_permissions (
    role_id,
    permission_id
)
VALUES (
    $1,
    $2
);

-- name: RemovePermissionFromRole :exec
DELETE
FROM auth.role_permissions
WHERE role_id = $1
AND permission_id = $2;

-- name: GetRolePermissions :many
SELECT p.*
FROM auth.permissions p
JOIN auth.role_permissions rp
ON p.id = rp.permission_id
WHERE rp.role_id = $1;