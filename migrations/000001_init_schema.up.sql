CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    email      TEXT        NOT NULL UNIQUE,
    role       TEXT        NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE projects (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       TEXT        NOT NULL,
    slug       TEXT        NOT NULL UNIQUE,
    owner_id   UUID        NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE api_keys (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash   TEXT        NOT NULL UNIQUE,
    project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    label      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE TABLE issues (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    fingerprint         TEXT        NOT NULL,
    project_id          UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    service             TEXT        NOT NULL,
    environment         TEXT        NOT NULL DEFAULT 'production',
    first_seen          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    occurrence_count    INT         NOT NULL DEFAULT 1,
    status              TEXT        NOT NULL DEFAULT 'open',
    severity            TEXT        NOT NULL DEFAULT 'error',
    ai_explanation      TEXT        NOT NULL DEFAULT '',
    ai_likely_cause     TEXT        NOT NULL DEFAULT '',
    ai_suggested_checks JSONB       NOT NULL DEFAULT '[]',
    UNIQUE (fingerprint, project_id)
);

CREATE TABLE events (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id            UUID        REFERENCES issues(id) ON DELETE SET NULL,
    project_id          UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    service             TEXT        NOT NULL,
    environment         TEXT        NOT NULL,
    error_type          TEXT        NOT NULL,
    raw_body            TEXT        NOT NULL DEFAULT '',
    redacted_body       TEXT        NOT NULL DEFAULT '',
    received_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retention_delete_at TIMESTAMPTZ
);

CREATE INDEX idx_issues_project_id  ON issues(project_id);
CREATE INDEX idx_issues_fingerprint ON issues(fingerprint);
CREATE INDEX idx_events_project_id  ON events(project_id);
CREATE INDEX idx_events_received_at ON events(received_at);
CREATE INDEX idx_events_retention   ON events(retention_delete_at) WHERE retention_delete_at IS NOT NULL;
