-- Facilities queries for public and admin operations

-- name: ListFacilities :many
SELECT id, name, description, location, priority, is_active, created_at, updated_at
FROM facilities
WHERE is_active = true
ORDER BY priority ASC, name ASC;

-- name: ListAllFacilities :many
SELECT id, name, description, location, priority, is_active, created_at, updated_at
FROM facilities
ORDER BY priority ASC, name ASC;

-- name: GetFacilityByID :one
SELECT id, name, description, location, priority, is_active, created_at, updated_at
FROM facilities
WHERE id = $1;

-- name: CreateFacility :one
INSERT INTO facilities (name, description, location, priority, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, location, priority, is_active, created_at, updated_at;

-- name: UpdateFacility :one
UPDATE facilities
SET name = $2,
    description = $3,
    location = $4,
    priority = $5,
    is_active = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, description, location, priority, is_active, created_at, updated_at;

-- name: UpdateFacilityPartial :one
UPDATE facilities
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    location = COALESCE(sqlc.narg('location'), location),
    priority = COALESCE(sqlc.narg('priority'), priority),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, name, description, location, priority, is_active, created_at, updated_at;

-- name: DeleteFacility :exec
DELETE FROM facilities
WHERE id = $1;