package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iPFSoftwares/custos/internal/domain"
)

const requestTimeout = 10 * time.Second

// Notifier implements domain.Notifier by POSTing an AlertPayload to a webhook URL.
type Notifier struct {
	webhookURL string
	client     *http.Client
}

// New creates a webhook Notifier targeting the given URL.
func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: requestTimeout},
	}
}

// Notify serialises the AlertPayload to JSON and POSTs it to the webhook URL.
func (n *Notifier) Notify(ctx context.Context, payload domain.AlertPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: post: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response body drain; error not actionable

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
