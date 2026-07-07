package googlechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iPFSoftwares/custos/internal/domain"
)

const requestTimeout = 10 * time.Second

// Notifier implements domain.Notifier by posting a card message to a Google Chat webhook.
type Notifier struct {
	webhookURL string
	client     *http.Client
}

// New creates a Google Chat Notifier using the given webhook URL.
func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: requestTimeout},
	}
}

// Google Chat card payload structures.
type cardPayload struct {
	Cards []card `json:"cards"`
}

type card struct {
	Header   cardHeader    `json:"header"`
	Sections []cardSection `json:"sections"`
}

type cardHeader struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type cardSection struct {
	Widgets []widget `json:"widgets"`
}

type widget struct {
	KeyValue *keyValue `json:"keyValue,omitempty"`
}

type keyValue struct {
	TopLabel string `json:"topLabel"`
	Content  string `json:"content"`
}

// Notify sends a formatted card alert to the configured Google Chat webhook.
func (n *Notifier) Notify(ctx context.Context, payload domain.AlertPayload) error {
	title := fmt.Sprintf("[%s] Error in %s", strings.ToUpper(payload.Analysis.Severity), payload.Issue.Service)

	checks := strings.Join(payload.Analysis.SuggestedChecks, "\n• ")
	if checks != "" {
		checks = "• " + checks
	}

	msg := cardPayload{
		Cards: []card{
			{
				Header: cardHeader{
					Title:    title,
					Subtitle: fmt.Sprintf("Environment: %s", payload.Issue.Environment),
				},
				Sections: []cardSection{
					{
						Widgets: []widget{
							{KeyValue: &keyValue{TopLabel: "Explanation", Content: payload.Analysis.Explanation}},
							{KeyValue: &keyValue{TopLabel: "Likely Cause", Content: payload.Analysis.LikelyCause}},
							{KeyValue: &keyValue{TopLabel: "Suggested Checks", Content: checks}},
							{KeyValue: &keyValue{TopLabel: "Occurrences", Content: fmt.Sprintf("%d", payload.Issue.OccurrenceCount)}},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("googlechat: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("googlechat: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("googlechat: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("googlechat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
