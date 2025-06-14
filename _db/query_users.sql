-- Users queries for Phase 1 token-based authentication

-- name: GetUserByToken :one
SELECT u.id, u.username, u.is_staff
FROM users u
JOIN user_tokens t ON u.id = t.user_id
WHERE t.token = $1 
  AND (t.expires_at IS NULL OR t.expires_at > NOW());

-- name: GetUserByID :one
SELECT id, username, is_staff, created_at
FROM users 
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, is_staff, created_at
FROM users 
WHERE username = $1;

-- name: ListUsers :many
SELECT id, username, is_staff, created_at
FROM users
ORDER BY created_at;

-- name: CreateUser :one
INSERT INTO users (username, is_staff)
VALUES ($1, $2)
RETURNING id, username, is_staff, created_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CreateToken :one
INSERT INTO user_tokens (user_id, token, name, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token, name, expires_at, created_at;

-- name: ListUserTokens :many
SELECT id, user_id, token, name, expires_at, created_at
FROM user_tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteToken :exec
DELETE FROM user_tokens
WHERE id = $1;