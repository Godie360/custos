package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	custosmetrics "github.com/iPFSoftwares/custos/internal/api/metrics"
	"github.com/iPFSoftwares/custos/internal/config"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/queue"
	"github.com/iPFSoftwares/custos/internal/store"
)

// IngestionService handles incoming raw events: persisting, deduplicating,
// and publishing analysis work items.
type IngestionService struct {
	events   store.EventStore
	issues   store.IssueStore
	producer queue.Producer
	cfg      config.Config
}

// NewIngestionService creates a new IngestionService.
func NewIngestionService(
	events store.EventStore,
	issues store.IssueStore,
	producer queue.Producer,
	cfg config.Config,
) *IngestionService {
	return &IngestionService{
		events:   events,
		issues:   issues,
		producer: producer,
		cfg:      cfg,
	}
}

// Ingest processes a single raw event: persists it, deduplicates against
// existing issues, and publishes an analysis event for new issues.
func (s *IngestionService) Ingest(ctx context.Context, event *domain.RawEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = time.Now().UTC()
	}

	if err := s.events.Create(ctx, event); err != nil {
		return fmt.Errorf("ingestion: save event: %w", err)
	}
	custosmetrics.EventsIngestedTotal.Inc()

	fingerprint := computeFingerprint(event.ErrorType, event.StackTrace)

	existing, err := s.issues.GetByFingerprint(ctx, event.ProjectID, fingerprint)
	if err == nil {
		// Issue exists — increment occurrence count and update last seen.
		existing.OccurrenceCount++
		existing.LastSeen = event.ReceivedAt
		if updateErr := s.issues.Update(ctx, existing); updateErr != nil {
			return fmt.Errorf("ingestion: update issue: %w", updateErr)
		}
		return nil
	}

	if err != domain.ErrNotFound {
		return fmt.Errorf("ingestion: lookup issue: %w", err)
	}

	// New issue.
	issue := &domain.Issue{
		ID:              uuid.New(),
		Fingerprint:     fingerprint,
		ProjectID:       event.ProjectID,
		Service:         event.Service,
		Environment:     event.Environment,
		FirstSeen:       event.ReceivedAt,
		LastSeen:        event.ReceivedAt,
		OccurrenceCount: 1,
		Status:          domain.IssueStatusOpen,
		Severity:        "error",
	}

	if err := s.issues.Create(ctx, issue); err != nil {
		return fmt.Errorf("ingestion: create issue: %w", err)
	}
	custosmetrics.IssuesCreatedTotal.Inc()

	analysisEvent := domain.AnalysisEvent{
		ErrorType:   event.ErrorType,
		Message:     event.Message,
		StackTrace:  event.StackTrace,
		Service:     event.Service,
		Environment: event.Environment,
	}

	payload, err := json.Marshal(analysisEvent)
	if err != nil {
		return fmt.Errorf("ingestion: marshal analysis event: %w", err)
	}

	if err := s.producer.Publish(ctx, s.cfg.Kafka.AnalysisTopic, []byte(issue.ID.String()), payload); err != nil {
		// Non-fatal: log in the caller; analysis can catch up from the DB.
		return fmt.Errorf("ingestion: publish analysis event: %w", err)
	}

	return nil
}

// normaliseFrame strips memory addresses and line numbers to produce a stable
// canonical form of a stack frame for fingerprinting.
var addrPattern = regexp.MustCompile(`0x[0-9a-fA-F]+`)
var linePattern = regexp.MustCompile(`:\d+$`)

func normaliseFrame(frame string) string {
	frame = addrPattern.ReplaceAllString(frame, "0x?")
	frame = linePattern.ReplaceAllString(frame, "")
	return strings.TrimSpace(frame)
}

// computeFingerprint returns a stable SHA256 hex digest for the error.
func computeFingerprint(errorType string, stackTrace []string) string {
	var sb strings.Builder
	sb.WriteString(errorType)
	sb.WriteByte('\n')
	limit := 5
	if len(stackTrace) < limit {
		limit = len(stackTrace)
	}
	for _, frame := range stackTrace[:limit] {
		sb.WriteString(normaliseFrame(frame))
		sb.WriteByte('\n')
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}
