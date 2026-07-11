package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Godie360/custos/internal/domain"
	"github.com/Godie360/custos/internal/store"
)

// Compile-time interface check.
var _ store.FilterStore = (*FilterStore)(nil)

// FilterStore is the PostgreSQL implementation of store.FilterStore.
type FilterStore struct {
	db *sql.DB
}

// NewFilterStore creates a new FilterStore backed by the given *sql.DB.
func NewFilterStore(db *sql.DB) *FilterStore {
	return &FilterStore{db: db}
}

// Create inserts a new filter rule and populates its ID and CreatedAt.
func (s *FilterStore) Create(ctx context.Context, rule *domain.FilterRule) error {
	const q = `
		INSERT INTO filter_rules (project_id, field, operator, value)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	row := s.db.QueryRowContext(ctx, q,
		rule.ProjectID,
		string(rule.Field),
		string(rule.Operator),
		rule.Value,
	)
	if err := row.Scan(&rule.ID, &rule.CreatedAt); err != nil {
		return fmt.Errorf("filter store: create: %w", err)
	}
	return nil
}

// ListByProject returns all filter rules for a project, newest first.
func (s *FilterStore) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*domain.FilterRule, error) {
	const q = `
		SELECT id, project_id, field, operator, value, created_at
		FROM filter_rules
		WHERE project_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("filter store: list: %w", err)
	}
	defer rows.Close() //nolint:errcheck // read-only query; close error not actionable

	var rules []*domain.FilterRule
	for rows.Next() {
		r := &domain.FilterRule{}
		var field, operator string
		if err := rows.Scan(&r.ID, &r.ProjectID, &field, &operator, &r.Value, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("filter store: scan: %w", err)
		}
		r.Field = domain.FilterField(field)
		r.Operator = domain.FilterOperator(operator)
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("filter store: rows: %w", err)
	}
	return rules, nil
}

// Delete removes a filter rule by id, scoped to the project to prevent cross-project deletion.
func (s *FilterStore) Delete(ctx context.Context, id uuid.UUID, projectID uuid.UUID) error {
	const q = `DELETE FROM filter_rules WHERE id = $1 AND project_id = $2`

	res, err := s.db.ExecContext(ctx, q, id, projectID)
	if err != nil {
		return fmt.Errorf("filter store: delete: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("filter store: rows affected: %w", err)
	}
	if n == 0 {
		return errors.Join(domain.ErrNotFound, fmt.Errorf("filter rule %s not found", id))
	}
	return nil
}
