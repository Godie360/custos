package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/store"
	generated "github.com/iPFSoftwares/custos/internal/store/postgres/generated"
)

// Compile-time interface check.
var _ store.EventStore = (*EventStore)(nil)

// EventStore is the PostgreSQL implementation of store.EventStore.
type EventStore struct {
	q *generated.Queries
}

// NewEventStore creates a new EventStore backed by the given *sql.DB.
func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{q: generated.New(db)}
}

func (s *EventStore) Create(ctx context.Context, event *domain.RawEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	var retentionDeleteAt sql.NullTime
	if event.RetentionDeleteAt != nil {
		retentionDeleteAt = sql.NullTime{Time: *event.RetentionDeleteAt, Valid: true}
	}
	return mapError(s.q.CreateEvent(ctx, generated.CreateEventParams{
		ID:                event.ID,
		ProjectID:         event.ProjectID,
		Service:           event.Service,
		Environment:       event.Environment,
		ErrorType:         event.ErrorType,
		RawBody:           event.RawBody,
		RedactedBody:      event.RedactedBody,
		ReceivedAt:        event.ReceivedAt,
		RetentionDeleteAt: retentionDeleteAt,
	}))
}

func (s *EventStore) GetByID(ctx context.Context, id uuid.UUID) (*domain.RawEvent, error) {
	row, err := s.q.GetEventByID(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}

	e := &domain.RawEvent{
		ID:           row.ID,
		ProjectID:    row.ProjectID,
		Service:      row.Service,
		Environment:  row.Environment,
		ErrorType:    row.ErrorType,
		RawBody:      row.RawBody,
		RedactedBody: row.RedactedBody,
		ReceivedAt:   row.ReceivedAt,
	}
	if row.RetentionDeleteAt.Valid {
		t := row.RetentionDeleteAt.Time
		e.RetentionDeleteAt = &t
	}
	if row.IssueID.Valid {
		e.IssueID = &row.IssueID.UUID
	}
	return e, nil
}

// unused — satisfies future needs; time.Now kept for reference
var _ = time.Now
