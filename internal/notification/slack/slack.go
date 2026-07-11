package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Godie360/custos/internal/domain"
)

const (
	requestTimeout   = 10 * time.Second
	blockTypeHeader  = "header"
	blockTypeSection = "section"
	textTypeMrkdwn   = "mrkdwn"
	textTypePlain    = "plain_text"
)

// Notifier implements domain.Notifier by posting a Block Kit message to a Slack webhook.
type Notifier struct {
	webhookURL string
	client     *http.Client
}

// New creates a Slack Notifier using the given incoming webhook URL.
func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: requestTimeout},
	}
}

// Slack Block Kit payload structures.
type slackPayload struct {
	Blocks []block `json:"blocks"`
}

type block struct {
	Type   string    `json:"type"`
	Text   *textObj  `json:"text,omitempty"`
	Fields []textObj `json:"fields,omitempty"`
}

type textObj struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Notify sends a Block Kit alert to the configured Slack webhook.
func (n *Notifier) Notify(ctx context.Context, payload domain.AlertPayload) error {
	title := fmt.Sprintf("[%s] Error in %s", strings.ToUpper(payload.Analysis.Severity), payload.Issue.Service)

	checks := ""
	if len(payload.Analysis.SuggestedChecks) > 0 {
		checks = "• " + strings.Join(payload.Analysis.SuggestedChecks, "\n• ")
	}

	msg := slackPayload{
		Blocks: []block{
			{
				Type: blockTypeHeader,
				Text: &textObj{Type: textTypePlain, Text: title},
			},
			{
				Type: blockTypeSection,
				Fields: []textObj{
					{Type: textTypeMrkdwn, Text: fmt.Sprintf("*Environment:*\n%s", payload.Issue.Environment)},
					{Type: textTypeMrkdwn, Text: fmt.Sprintf("*Occurrences:*\n%d", payload.Issue.OccurrenceCount)},
				},
			},
			{
				Type: blockTypeSection,
				Text: &textObj{Type: textTypeMrkdwn, Text: fmt.Sprintf("*Explanation:*\n%s", payload.Analysis.Explanation)},
			},
			{
				Type: blockTypeSection,
				Text: &textObj{Type: textTypeMrkdwn, Text: fmt.Sprintf("*Likely Cause:*\n%s", payload.Analysis.LikelyCause)},
			},
			{
				Type: blockTypeSection,
				Text: &textObj{Type: textTypeMrkdwn, Text: fmt.Sprintf("*Suggested Checks:*\n%s", checks)},
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack: post: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response body drain; error not actionable

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
