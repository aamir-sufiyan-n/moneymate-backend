-- name: CreateEmailVerification :one
INSERT INTO auth.email_verifications (
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

-- name: GetEmailVerification :one
SELECT *
FROM auth.email_verifications
WHERE token_hash = $1;

-- name: MarkEmailVerificationUsed :exec
UPDATE auth.email_verifications
SET used_at = NOW()
WHERE token_hash = $1;

-- name: DeleteExpiredEmailVerifications :exec
DELETE
FROM auth.email_verifications
WHERE expires_at < NOW();