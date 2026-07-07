package postgres

import (
	"context"
	"database/sql"
	"encoding/json"


	"github.com/google/uuid"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/store"
	generated "github.com/iPFSoftwares/custos/internal/store/postgres/generated"

)

// Compile-time interface check.
var _ store.IssueStore = (*IssueStore)(nil)

// IssueStore is the PostgreSQL implementation of store.IssueStore.
type IssueStore struct {
	q *generated.Queries
}

// NewIssueStore creates a new IssueStore backed by the given *sql.DB.
func NewIssueStore(db *sql.DB) *IssueStore {
	return &IssueStore{q: generated.New(db)}
}

func (s *IssueStore) Create(ctx context.Context, issue *domain.Issue) error {
	if issue.ID == uuid.Nil {
		issue.ID = uuid.New()
	}
	checks, err := json.Marshal(issue.AISuggestedChecks)
	if err != nil {
		return err
	}
	return mapError(s.q.CreateIssue(ctx, generated.CreateIssueParams{
		ID:                issue.ID,
		Fingerprint:       issue.Fingerprint,
		ProjectID:         issue.ProjectID,
		Service:           issue.Service,
		Environment:       issue.Environment,
		FirstSeen:         issue.FirstSeen,
		LastSeen:          issue.LastSeen,
		OccurrenceCount:   int32(issue.OccurrenceCount),
		Status:            string(issue.Status),
		Severity:          issue.Severity,
		AiExplanation:     issue.AIExplanation,
		AiLikelyCause:     issue.AILikelyCause,
		AiSuggestedChecks: checks,
	}))
}

func (s *IssueStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.Issue, error) {
	row, err := s.q.GetIssueByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	return rowToIssue(row)
}

func (s *IssueStore) GetByFingerprint(ctx context.Context, projectID uuid.UUID, fingerprint string) (*domain.Issue, error) {
	row, err := s.q.GetIssueByFingerprint(ctx, generated.GetIssueByFingerprintParams{
		ProjectID:   projectID,
		Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return rowToIssue(row)
}

func (s *IssueStore) Update(ctx context.Context, issue *domain.Issue) error {
	checks, err := json.Marshal(issue.AISuggestedChecks)
	if err != nil {
		return err
	}
	return mapError(s.q.UpdateIssue(ctx, generated.UpdateIssueParams{
		LastSeen:          issue.LastSeen,
		OccurrenceCount:   int32(issue.OccurrenceCount),
		Status:            string(issue.Status),
		Severity:          issue.Severity,
		AiExplanation:     issue.AIExplanation,
		AiLikelyCause:     issue.AILikelyCause,
		AiSuggestedChecks: checks,
		ID:                issue.ID,
	}))
}

func (s *IssueStore) List(ctx context.Context, filter store.ListIssuesFilter) ([]*domain.Issue, error) {
	limit := int32(filter.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int32(filter.Offset)

	var projectID uuid.NullUUID
	if filter.ProjectID != uuid.Nil {
		projectID = uuid.NullUUID{UUID: filter.ProjectID, Valid: true}
	}
	var service, environment, severity sql.NullString
	if filter.Service != "" {
		service = sql.NullString{String: filter.Service, Valid: true}
	}
	if filter.Environment != "" {
		environment = sql.NullString{String: filter.Environment, Valid: true}
	}
	if filter.Severity != "" {
		severity = sql.NullString{String: filter.Severity, Valid: true}
	}

	rows, err := s.q.ListIssues(ctx, generated.ListIssuesParams{
		ProjectID:   projectID,
		Service:     service,
		Environment: environment,
		Severity:    severity,
		LimitVal:    sql.NullInt32{Int32: limit, Valid: true},
		OffsetVal:   sql.NullInt32{Int32: offset, Valid: true},
	})
	if err != nil {
		return nil, mapError(err)
	}

	issues := make([]*domain.Issue, 0, len(rows))
	for _, row := range rows {
		issue, err := rowToIssue(row)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}
	return issues, nil
}

func (s *IssueStore) CountBySeverity(ctx context.Context) ([]store.SeverityCount, error) {
	rows, err := s.q.CountIssuesBySeverity(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	result := make([]store.SeverityCount, 0, len(rows))
	for _, r := range rows {
		result = append(result, store.SeverityCount{Severity: r.Severity, Count: int(r.Count)})
	}
	return result, nil
}

func (s *IssueStore) TopServices(ctx context.Context) ([]store.ServiceCount, error) {
	rows, err := s.q.TopServicesByOccurrences(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	result := make([]store.ServiceCount, 0, len(rows))
	for _, r := range rows {
		result = append(result, store.ServiceCount{Service: r.Service, Total: int(r.Total)})
	}
	return result, nil
}

func (s *IssueStore) TotalCount(ctx context.Context) (int, error) {
	n, err := s.q.TotalIssues(ctx)
	if err != nil {
		return 0, mapError(err)
	}
	return int(n), nil
}

func rowToIssue(row generated.Issue) (*domain.Issue, error) {
	var checks []string
	if err := json.Unmarshal(row.AiSuggestedChecks, &checks); err != nil {
		checks = []string{}
	}
	return &domain.Issue{
		ID:                row.ID,
		Fingerprint:       row.Fingerprint,
		ProjectID:         row.ProjectID,
		Service:           row.Service,
		Environment:       row.Environment,
		FirstSeen:         row.FirstSeen,
		LastSeen:          row.LastSeen,
		OccurrenceCount:   int(row.OccurrenceCount),
		Status:            domain.IssueStatus(row.Status),
		Severity:          row.Severity,
		AIExplanation:     row.AiExplanation,
		AILikelyCause:     row.AiLikelyCause,
		AISuggestedChecks: checks,
	}, nil
}
