-- name: CreateRefreshToken :one
INSERT INTO auth.refresh_tokens (
    id,
    user_id,
    token_hash,
    expires_at
)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT *
FROM auth.refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE auth.refresh_tokens
SET revoked_at = NOW()
WHERE token_hash = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE
FROM auth.refresh_tokens
WHERE expires_at < NOW();

-- name: RevokeAllRefreshTokensForUser :exec
UPDATE auth.refresh_tokens
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;