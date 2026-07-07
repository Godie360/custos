package service

import (
	"context"
	"fmt"

	"github.com/iPFSoftwares/custos/internal/domain"
)

// NotificationService fans out alert payloads to all configured notifiers.
type NotificationService struct {
	notifiers []domain.Notifier
}

// NewNotificationService creates a NotificationService with the given notifiers.
func NewNotificationService(notifiers ...domain.Notifier) *NotificationService {
	return &NotificationService{notifiers: notifiers}
}

// Notify calls every registered notifier in sequence. Errors are accumulated
// and returned together so that a single failure does not block others.
func (s *NotificationService) Notify(ctx context.Context, payload domain.AlertPayload) error {
	var errs []string
	for _, n := range s.notifiers {
		if err := n.Notify(ctx, payload); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notification: %d notifier(s) failed: %v", len(errs), errs)
	}
	return nil
}
