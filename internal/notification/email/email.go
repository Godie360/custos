package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/Godie360/custos/internal/config"
	"github.com/Godie360/custos/internal/domain"
)

// Notifier implements domain.Notifier by sending email via SMTP.
type Notifier struct {
	cfg config.NotificationConfig
	to  []string
}

// New creates an email Notifier using config SMTP settings.
func New(cfg config.NotificationConfig, recipients []string) *Notifier {
	return &Notifier{cfg: cfg, to: recipients}
}

// Notify sends an alert email to all configured recipients.
func (n *Notifier) Notify(_ context.Context, payload domain.AlertPayload) error {
	if n.cfg.SMTPHost == "" {
		return fmt.Errorf("email: SMTP host not configured")
	}

	subject := fmt.Sprintf("[Custos] %s severity issue in %s",
		payload.Analysis.Severity, payload.Issue.Service)

	var body strings.Builder
	fmt.Fprintf(&body, "Issue ID: %s\n", payload.Issue.ID)
	fmt.Fprintf(&body, "Service: %s\n", payload.Issue.Service)
	fmt.Fprintf(&body, "Environment: %s\n", payload.Issue.Environment)
	fmt.Fprintf(&body, "Severity: %s\n", payload.Analysis.Severity)
	fmt.Fprintf(&body, "Explanation: %s\n", payload.Analysis.Explanation)
	fmt.Fprintf(&body, "Likely Cause: %s\n", payload.Analysis.LikelyCause)
	if len(payload.Analysis.SuggestedChecks) > 0 {
		fmt.Fprintf(&body, "Suggested Checks:\n")
		for _, c := range payload.Analysis.SuggestedChecks {
			fmt.Fprintf(&body, "  - %s\n", c)
		}
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		n.cfg.SMTPFrom,
		strings.Join(n.to, ", "),
		subject,
		body.String(),
	)

	addr := fmt.Sprintf("%s:%s", n.cfg.SMTPHost, n.cfg.SMTPPort)
	if err := smtp.SendMail(addr, nil, n.cfg.SMTPFrom, n.to, []byte(msg)); err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	return nil
}
