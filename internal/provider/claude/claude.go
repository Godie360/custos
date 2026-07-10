package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Godie360/custos/internal/config"
	"github.com/Godie360/custos/internal/domain"
)

const (
	defaultModel   = "claude-3-5-sonnet-20241022"
	apiURL         = "https://api.anthropic.com/v1/messages"
	maxRetries     = 3
	requestTimeout = 30 * time.Second
)

// Analyzer implements domain.AIAnalyzer using the Anthropic Claude API.
type Analyzer struct {
	cfg    config.Config
	client *http.Client
}

// New creates an Analyzer for the Claude provider.
func New(cfg config.Config) *Analyzer {
	return &Analyzer{
		cfg:    cfg,
		client: &http.Client{Timeout: requestTimeout},
	}
}

type requestBody struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseBody struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Analyze sends an error event to Claude and returns a structured AnalysisResult.
func (a *Analyzer) Analyze(ctx context.Context, event domain.AnalysisEvent) (domain.AnalysisResult, error) {
	model := a.cfg.AI.Model
	if model == "" {
		model = defaultModel
	}

	prompt := buildPrompt(event)
	body := requestBody{
		Model:     model,
		MaxTokens: 1024,
		Messages:  []message{{Role: "user", Content: prompt}},
	}

	var result domain.AnalysisResult
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return result, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		payload, err := json.Marshal(body)
		if err != nil {
			return result, fmt.Errorf("claude: marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
		if err != nil {
			return result, fmt.Errorf("claude: create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", a.cfg.AI.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		respBytes, status, err := a.doRequest(req)
		if err != nil {
			lastErr = fmt.Errorf("claude: http: %w", err)
			continue
		}

		if status == http.StatusTooManyRequests || status >= 500 {
			lastErr = fmt.Errorf("claude: status %d", status)
			continue
		}

		if status != http.StatusOK {
			return result, fmt.Errorf("claude: status %d: %s", status, string(respBytes))
		}

		var rb responseBody
		if err := json.Unmarshal(respBytes, &rb); err != nil {
			return result, fmt.Errorf("claude: unmarshal response: %w", err)
		}

		if len(rb.Content) == 0 {
			return result, fmt.Errorf("claude: empty response content")
		}

		return parseResult(rb.Content[0].Text)
	}

	return result, fmt.Errorf("claude: max retries exceeded: %w", lastErr)
}

func (a *Analyzer) doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // response body drain; error not actionable
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read body: %w", err)
	}
	return b, resp.StatusCode, nil
}

func buildPrompt(event domain.AnalysisEvent) string {
	var sb strings.Builder
	sb.WriteString("Analyze the following application error and respond with a JSON object containing:\n")
	sb.WriteString(`{"explanation":"<plain English explanation>","severity":"<critical|high|medium|low>","likely_cause":"<root cause>","suggested_checks":["<check1>","<check2>"]}\n\n`)
	fmt.Fprintf(&sb, "Service: %s\nEnvironment: %s\nError Type: %s\nMessage: %s\n",
		event.Service, event.Environment, event.ErrorType, event.Message)
	if len(event.StackTrace) > 0 {
		sb.WriteString("Stack Trace:\n")
		for _, frame := range event.StackTrace {
			fmt.Fprintf(&sb, "  %s\n", frame)
		}
	}
	sb.WriteString("\nRespond ONLY with the JSON object, no other text.")
	return sb.String()
}

type analysisJSON struct {
	Explanation     string   `json:"explanation"`
	Severity        string   `json:"severity"`
	LikelyCause     string   `json:"likely_cause"`
	SuggestedChecks []string `json:"suggested_checks"`
}

func parseResult(text string) (domain.AnalysisResult, error) {
	// Strip markdown code fences if present.
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var aj analysisJSON
	if err := json.Unmarshal([]byte(text), &aj); err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("claude: parse result JSON: %w", err)
	}
	return domain.AnalysisResult{
		Explanation:     aj.Explanation,
		Severity:        aj.Severity,
		LikelyCause:     aj.LikelyCause,
		SuggestedChecks: aj.SuggestedChecks,
	}, nil
}
