# Storage Strategy

## Context

Two distinct access patterns: relational config/account data with strong integrity requirements, and high-ingestion-rate append-only log event data needing fast aggregate queries.

## Requirements

- Events table must support filtering by project, service, environment, severity, and time range.
- Issue fingerprint must be stable across re-deployments of the same service.
- Raw event bodies must be independently deletable without cascading to issues.
- PostgreSQL alone must be sufficient for small-scale office deployment.
- Migration to ClickHouse must not require SDK or API contract changes.

## Decisions

| Store | Purpose | License | Why |
|---|---|---|---|
| PostgreSQL | Users, projects, API keys, alert rules, issue metadata | PostgreSQL License (OSS) | Strong relational integrity; simple to self-host. |
| ClickHouse | Raw log/event storage and time-series analytics | Apache 2.0 (OSS) | Columnar storage for high-ingestion-rate append-only data and fast aggregate queries. |

For an internal office deployment with modest log volume, PostgreSQL alone (with a well-indexed events table) is sufficient to start. ClickHouse is introduced later without changing SDKs or API contracts.

Migrations tool: **`golang-migrate/migrate`** (MIT) — CLI and Go library for versioned schema migrations.

## PostgreSQL Schema (initial)

Core tables:

- `projects` — id, name, slug, owner_id, created_at
- `api_keys` — id, key_hash, project_id, label, created_at, expires_at, revoked_at
- `users` — id, email, role, created_at
- `issues` — id, fingerprint, project_id, service, environment, first_seen, last_seen, occurrence_count, status, severity, ai_explanation, ai_likely_cause, ai_suggested_checks
- `events` — id, issue_id, project_id, service, environment, error_type, raw_body, redacted_body, received_at, retention_delete_at

## Contracts

- `issues.fingerprint` is immutable after creation.
- `events.raw_body` is the only field targeted by the retention deletion job.
- Adding ClickHouse does not change the `events` write path — the ingestion API writes to both stores (or only ClickHouse for raw events at that point) via a storage interface abstraction.

## Acceptance Criteria

- [ ] PostgreSQL schema handles all MVP queries without ClickHouse.
- [ ] Migration files are versioned and runnable via `golang-migrate`.
- [ ] Adding ClickHouse requires no SDK or API contract changes.
- [ ] Retention deletion job removes `raw_body` without touching `issues` or `ai_explanation`.
