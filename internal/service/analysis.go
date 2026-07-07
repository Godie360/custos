package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
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
	return s.consumer.Subscribe(ctx, topic, s.handleMessage)
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
		slog.ErrorContext(ctx, "analysis: AI provider error",
			slog.String("issue_id", issueID.String()),
			slog.String("error", err.Error()),
		)
		issue.Status = domain.IssueStatusAnalysisFailed
		if updateErr := s.issues.Update(ctx, issue); updateErr != nil {
			slog.ErrorContext(ctx, "analysis: update issue status",
				slog.String("issue_id", issueID.String()),
				slog.String("error", updateErr.Error()),
			)
		}
		return fmt.Errorf("analysis: analyze: %w", err)
	}

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
			slog.WarnContext(ctx, "analysis: notify failed",
				slog.String("issue_id", issueID.String()),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}
