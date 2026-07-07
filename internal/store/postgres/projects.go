package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/store"
	generated "github.com/iPFSoftwares/custos/internal/store/postgres/generated"
)

// Compile-time interface checks.
var (
	_ store.ProjectStore = (*ProjectStore)(nil)
	_ store.APIKeyStore  = (*APIKeyStore)(nil)
)

// ProjectStore is the PostgreSQL implementation of store.ProjectStore.
type ProjectStore struct {
	q *generated.Queries
}

// NewProjectStore creates a new ProjectStore backed by the given *sql.DB.
func NewProjectStore(db *sql.DB) *ProjectStore {
	return &ProjectStore{q: generated.New(db)}
}

func (s *ProjectStore) Create(ctx context.Context, project *domain.Project) error {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	return mapError(s.q.CreateProject(ctx, generated.CreateProjectParams{
		ID:        project.ID,
		Name:      project.Name,
		Slug:      project.Slug,
		OwnerID:   project.OwnerID,
		CreatedAt: project.CreatedAt,
	}))
}

func (s *ProjectStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	row, err := s.q.GetProjectByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	return rowToProject(row), nil
}

func (s *ProjectStore) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	row, err := s.q.GetProjectBySlug(ctx, slug)
	if err != nil {
		return nil, mapError(err)
	}
	return rowToProject(row), nil
}

func (s *ProjectStore) List(ctx context.Context) ([]*domain.Project, error) {
	rows, err := s.q.ListProjects(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	projects := make([]*domain.Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, rowToProject(row))
	}
	return projects, nil
}

func rowToProject(row generated.Project) *domain.Project {
	return &domain.Project{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		OwnerID:   row.OwnerID,
		CreatedAt: row.CreatedAt,
	}
}

// APIKeyStore is the PostgreSQL implementation of store.APIKeyStore.
type APIKeyStore struct {
	q *generated.Queries
}

// NewAPIKeyStore creates a new APIKeyStore backed by the given *sql.DB.
func NewAPIKeyStore(db *sql.DB) *APIKeyStore {
	return &APIKeyStore{q: generated.New(db)}
}

func (s *APIKeyStore) Create(ctx context.Context, key *domain.APIKey) error {
	if key.ID == uuid.Nil {
		key.ID = uuid.New()
	}
	var expiresAt sql.NullTime
	if key.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *key.ExpiresAt, Valid: true}
	}
	return mapError(s.q.CreateAPIKey(ctx, generated.CreateAPIKeyParams{
		ID:        key.ID,
		KeyHash:   key.KeyHash,
		ProjectID: key.ProjectID,
		Label:     key.Label,
		CreatedAt: key.CreatedAt,
		ExpiresAt: expiresAt,
	}))
}

func (s *APIKeyStore) GetByHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	row, err := s.q.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		return nil, mapError(err)
	}
	k := &domain.APIKey{
		ID:        row.ID,
		KeyHash:   row.KeyHash,
		ProjectID: row.ProjectID,
		Label:     row.Label,
		CreatedAt: row.CreatedAt,
	}
	if row.ExpiresAt.Valid {
		t := row.ExpiresAt.Time
		k.ExpiresAt = &t
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		k.RevokedAt = &t
	}
	return k, nil
}

func (s *APIKeyStore) Revoke(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.RevokeAPIKey(ctx, id)
	if err != nil {
		return mapError(err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
