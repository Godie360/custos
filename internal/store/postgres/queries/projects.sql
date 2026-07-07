-- name: CreateProject :exec
INSERT INTO projects (id, name, slug, owner_id, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetProjectByID :one
SELECT id, name, slug, owner_id, created_at
FROM projects WHERE id = $1;

-- name: GetProjectBySlug :one
SELECT id, name, slug, owner_id, created_at
FROM projects WHERE slug = $1;

-- name: ListProjects :many
SELECT id, name, slug, owner_id, created_at
FROM projects
ORDER BY created_at DESC;

-- name: CreateAPIKey :exec
INSERT INTO api_keys (id, key_hash, project_id, label, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAPIKeyByHash :one
SELECT id, key_hash, project_id, label, created_at, expires_at, revoked_at
FROM api_keys WHERE key_hash = $1;

-- name: RevokeAPIKey :execrows
UPDATE api_keys SET revoked_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;
