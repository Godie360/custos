package domain

import "context"

// Notifier is implemented by each notification channel adapter.
type Notifier interface {
	Notify(ctx context.Context, payload AlertPayload) error
}

// AlertPayload bundles an issue with its AI analysis for notifications.
type AlertPayload struct {
	Issue    Issue
	Analysis AnalysisResult
}
