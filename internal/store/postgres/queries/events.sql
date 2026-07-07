-- name: CreateEvent :exec
INSERT INTO events (
    id, project_id, service, environment, error_type,
    raw_body, redacted_body, received_at, retention_delete_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9
);

-- name: GetEventByID :one
SELECT id, issue_id, project_id, service, environment, error_type,
    raw_body, redacted_body, received_at, retention_delete_at
FROM events
WHERE id = $1;
