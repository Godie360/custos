package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Godie360/custos/internal/domain"
)

// Compile-time interface check.
var _ domain.Notifier = (*NotificationService)(nil)

// NotificationService fans out alert payloads to all configured notifiers.
type NotificationService struct {
	notifiers []domain.Notifier
}

// NewNotificationService creates a NotificationService with the given notifiers.
func NewNotificationService(notifiers ...domain.Notifier) *NotificationService {
	return &NotificationService{notifiers: notifiers}
}

// Notify calls every registered notifier. All notifiers are called even if one
// fails; errors are joined so the caller sees all failures at once.
func (s *NotificationService) Notify(ctx context.Context, payload domain.AlertPayload) error {
	errs := make([]error, 0, len(s.notifiers))
	for _, n := range s.notifiers {
		if err := n.Notify(ctx, payload); err != nil {
			errs = append(errs, fmt.Errorf("%T: %w", n, err))
		}
	}
	return errors.Join(errs...)
}
