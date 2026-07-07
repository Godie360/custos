package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	custosmetrics "github.com/iPFSoftwares/custos/internal/api/metrics"
	"github.com/iPFSoftwares/custos/internal/domain"
	"github.com/iPFSoftwares/custos/internal/queue"
	"github.com/iPFSoftwares/custos/internal/store"
)

// AnalysisService consumes analysis work items from Kafka, calls the AI
// provider, persists the results, and triggers notifications for severe issues.
type AnalysisService struct {
	issues   store.IssueStore
	analyzer domain.AIAnalyzer
	notifier domain.Notifier
	consumer queue.Consumer
}

// NewAnalysisService creates an AnalysisService.
func NewAnalysisService(
	issues store.IssueStore,
	analyzer domain.AIAnalyzer,
	notifier domain.Notifier,
	consumer queue.Consumer,
) *AnalysisService {
	return &AnalysisService{
		issues:   issues,
		analyzer: analyzer,
		notifier: notifier,
		consumer: consumer,
	}
}

// Run subscribes to the analysis topic and processes messages until ctx is cancelled.
func (s *AnalysisService) Run(ctx context.Context, topic string) error {
	if err := s.consumer.Subscribe(ctx, topic, s.handleMessage); err != nil {
		return fmt.Errorf("analysis service: subscribe: %w", err)
	}
	return nil
}

func (s *AnalysisService) handleMessage(ctx context.Context, msg queue.Message) error {
	var event domain.AnalysisEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("analysis: decode message: %w", err)
	}

	issueID, err := uuid.Parse(string(msg.Key))
	if err != nil {
		return fmt.Errorf("analysis: parse issue id %q: %w", string(msg.Key), err)
	}

	issue, err := s.issues.GetByID(ctx, issueID)
	if err != nil {
		return fmt.Errorf("analysis: get issue %s: %w", issueID, err)
	}

	result, err := s.analyzer.Analyze(ctx, event)
	if err != nil {
		// Persist failure status (best-effort) then propagate the error upward.
		// The consumer will NOT commit the offset, so the message is retried.
		issue.Status = domain.IssueStatusAnalysisFailed
		if updateErr := s.issues.Update(ctx, issue); updateErr != nil {
			// Log the secondary failure; primary error returned below.
			slog.WarnContext(ctx, "analysis: could not persist failed status",
				slog.String("issue_id", issueID.String()),
				slog.String("error", updateErr.Error()),
			)
		}
		custosmetrics.AnalysisTotal.WithLabelValues("failed").Inc()
		return fmt.Errorf("analysis: AI provider: %w", err)
	}

	custosmetrics.AnalysisTotal.WithLabelValues("success").Inc()

	issue.AIExplanation = result.Explanation
	issue.AILikelyCause = result.LikelyCause
	issue.AISuggestedChecks = result.SuggestedChecks
	issue.Severity = result.Severity

	if err := s.issues.Update(ctx, issue); err != nil {
		return fmt.Errorf("analysis: update issue: %w", err)
	}

	if result.Severity == "high" || result.Severity == "critical" {
		payload := domain.AlertPayload{Issue: *issue, Analysis: result}
		if err := s.notifier.Notify(ctx, payload); err != nil {
			// Notification failure is non-fatal — log and continue.
			slog.WarnContext(ctx, "analysis: notification failed",
				slog.String("issue_id", issueID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}
