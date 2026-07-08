-- name: CreateOAuthAccount :one
INSERT INTO auth.oauth_accounts (
    id,
    user_id,
    provider,
    provider_user_id
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetOAuthAccount :one
SELECT *
FROM auth.oauth_accounts
WHERE provider = $1
AND provider_user_id = $2;

-- name: GetOAuthAccountsByUser :many
SELECT *
FROM auth.oauth_accounts
WHERE user_id = $1;