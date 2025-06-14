-- Users queries for admin and authentication operations

-- name: GetUserByID :one
SELECT id, username, email, is_staff, created_at, updated_at
FROM users 
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, is_staff, password_hash, created_at, updated_at
FROM users 
WHERE username = $1;

-- name: ListUsers :many
SELECT id, username, email, is_staff, created_at, updated_at
FROM users
ORDER BY id;

-- name: CreateUser :one
INSERT INTO users (username, email, is_staff, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, username, email, is_staff, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET username = $2,
    email = $3,
    is_staff = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING id, username, email, is_staff, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetCurrentUser :one
SELECT id, username, email
FROM users
WHERE id = $1;