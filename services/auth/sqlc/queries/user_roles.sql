-- name: AssignRoleToUser :exec
INSERT INTO auth.user_roles (
    user_id,
    role_id,
    granted_by
)
VALUES (
    $1,
    $2,
    $3
)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE
FROM auth.user_roles
WHERE user_id = $1
AND role_id = $2;

-- name: GetUserRoles :many
SELECT r.*
FROM auth.roles r
JOIN auth.user_roles ur
ON r.id = ur.role_id
WHERE ur.user_id = $1;