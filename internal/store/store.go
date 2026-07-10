package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/Godie360/custos/internal/domain"
)

// ListIssuesFilter specifies optional filters for listing issues.
type ListIssuesFilter struct {
	ProjectID   uuid.UUID
	Service     string
	Environment string
	Severity    string
	Limit       int
	Offset      int
}

// SeverityCount holds an aggregated count per severity level.
type SeverityCount struct {
	Severity string
	Count    int
}

// ServiceCount holds an aggregated occurrence count per service.
type ServiceCount struct {
	Service string
	Total   int
}

// IssueStore defines persistence operations for issues.
type IssueStore interface {
	Create(ctx context.Context, issue *domain.Issue) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error)
	GetByFingerprint(ctx context.Context, projectID uuid.UUID, fingerprint string) (*domain.Issue, error)
	Update(ctx context.Context, issue *domain.Issue) error
	List(ctx context.Context, filter ListIssuesFilter) ([]*domain.Issue, error)
	CountBySeverity(ctx context.Context) ([]SeverityCount, error)
	TopServices(ctx context.Context) ([]ServiceCount, error)
	TotalCount(ctx context.Context) (int, error)
}

// EventStore defines persistence operations for raw events.
type EventStore interface {
	Create(ctx context.Context, event *domain.RawEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RawEvent, error)
}

// ProjectStore defines persistence operations for projects.
type ProjectStore interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Project, error)
	List(ctx context.Context) ([]*domain.Project, error)
}

// APIKeyStore defines persistence operations for API keys.
type APIKeyStore interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByHash(ctx context.Context, hash string) (*domain.APIKey, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}
