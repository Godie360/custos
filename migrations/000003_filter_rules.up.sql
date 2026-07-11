CREATE TABLE filter_rules (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    field      TEXT        NOT NULL CHECK (field IN ('error_type', 'message', 'service', 'environment')),
    operator   TEXT        NOT NULL CHECK (operator IN ('equals', 'contains', 'starts_with')),
    value      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_filter_rules_project_id ON filter_rules(project_id);
