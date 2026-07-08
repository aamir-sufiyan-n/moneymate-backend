-- name: CreatePasswordResetToken :one
INSERT INTO auth.password_reset_tokens (
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

-- name: GetPasswordResetToken :one
SELECT *
FROM auth.password_reset_tokens
WHERE token_hash = $1;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE auth.password_reset_tokens
SET used_at = NOW()
WHERE token_hash = $1;

-- name: DeleteExpiredPasswordResetTokens :exec
DELETE
FROM auth.password_reset_tokens
WHERE expires_at < NOW();