-- name: CreateIssue :exec
INSERT INTO issues (
    id, fingerprint, project_id, service, environment,
    first_seen, last_seen, occurrence_count, status, severity,
    ai_explanation, ai_likely_cause, ai_suggested_checks
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13
);

-- name: GetIssueByID :one
SELECT id, fingerprint, project_id, service, environment,
    first_seen, last_seen, occurrence_count, status, severity,
    ai_explanation, ai_likely_cause, ai_suggested_checks
FROM issues
WHERE id = $1;

-- name: GetIssueByFingerprint :one
SELECT id, fingerprint, project_id, service, environment,
    first_seen, last_seen, occurrence_count, status, severity,
    ai_explanation, ai_likely_cause, ai_suggested_checks
FROM issues
WHERE project_id = $1 AND fingerprint = $2;

-- name: UpdateIssue :exec
UPDATE issues SET
    last_seen          = $1,
    occurrence_count   = $2,
    status             = $3,
    severity           = $4,
    ai_explanation     = $5,
    ai_likely_cause    = $6,
    ai_suggested_checks = $7
WHERE id = $8;

-- name: ListIssues :many
SELECT id, fingerprint, project_id, service, environment,
    first_seen, last_seen, occurrence_count, status, severity,
    ai_explanation, ai_likely_cause, ai_suggested_checks
FROM issues
WHERE
    (project_id  = sqlc.narg('project_id')  OR sqlc.narg('project_id')  IS NULL) AND
    (service     = sqlc.narg('service')      OR sqlc.narg('service')     IS NULL) AND
    (environment = sqlc.narg('environment')  OR sqlc.narg('environment') IS NULL) AND
    (severity    = sqlc.narg('severity')     OR sqlc.narg('severity')    IS NULL)
ORDER BY last_seen DESC
LIMIT  sqlc.narg('limit_val')
OFFSET sqlc.narg('offset_val');

-- name: CountIssuesBySeverity :many
SELECT severity, COUNT(*)::int AS count
FROM issues
GROUP BY severity;

-- name: TopServicesByOccurrences :many
SELECT service, SUM(occurrence_count)::int AS total
FROM issues
GROUP BY service
ORDER BY total DESC
LIMIT 5;

-- name: TotalIssues :one
SELECT COUNT(*)::int AS count FROM issues;
