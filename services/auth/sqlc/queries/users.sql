-- name: CreateUser :one

INSERT INTO auth.users (
    id,
    email,
    password_hash
)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetUserByID :one

SELECT *
FROM auth.users
WHERE id = $1;

-- name: GetUserByEmail :one

SELECT *
FROM auth.users
WHERE email = $1;

-- name: VerifyEmail :exec

UPDATE auth.users
SET
    email_verified = TRUE,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdatePassword :exec

UPDATE auth.users
SET
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserStatus :exec

UPDATE auth.users
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec

DELETE
FROM auth.users
WHERE id = $1;