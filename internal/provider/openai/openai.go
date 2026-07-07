package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/iPFSoftwares/custos/internal/config"
	"github.com/iPFSoftwares/custos/internal/domain"
)

const (
	defaultModel   = "gpt-4o"
	apiURL         = "https://api.openai.com/v1/chat/completions"
	maxRetries     = 3
	requestTimeout = 30 * time.Second
)

// Analyzer implements domain.AIAnalyzer using the OpenAI chat completions API.
type Analyzer struct {
	cfg    config.Config
	client *http.Client
}

// New creates an Analyzer for the OpenAI provider.
func New(cfg config.Config) *Analyzer {
	return &Analyzer{
		cfg:    cfg,
		client: &http.Client{Timeout: requestTimeout},
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Analyze sends an error event to OpenAI and returns a structured AnalysisResult.
func (a *Analyzer) Analyze(ctx context.Context, event domain.AnalysisEvent) (domain.AnalysisResult, error) {
	model := a.cfg.AI.Model
	if model == "" {
		model = defaultModel
	}

	prompt := buildPrompt(event)
	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are an expert software reliability engineer. Always respond with valid JSON only."},
			{Role: "user", Content: prompt},
		},
	}

	var result domain.AnalysisResult
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second //nolint:gosec // G115: attempt is bounded by maxRetries (<5)
			select {
			case <-ctx.Done():
				return result, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		payload, err := json.Marshal(reqBody)
		if err != nil {
			return result, fmt.Errorf("openai: marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
		if err != nil {
			return result, fmt.Errorf("openai: create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+a.cfg.AI.APIKey)

		resp, err := a.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("openai: http: %w", err)
			continue
		}
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("openai: read response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("openai: status %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return result, fmt.Errorf("openai: status %d: %s", resp.StatusCode, string(respBytes))
		}

		var cr chatResponse
		if err := json.Unmarshal(respBytes, &cr); err != nil {
			return result, fmt.Errorf("openai: unmarshal response: %w", err)
		}
		if len(cr.Choices) == 0 {
			return result, fmt.Errorf("openai: empty choices")
		}

		return parseResult(cr.Choices[0].Message.Content)
	}

	return result, fmt.Errorf("openai: max retries exceeded: %w", lastErr)
}

func buildPrompt(event domain.AnalysisEvent) string {
	var sb strings.Builder
	sb.WriteString("Analyze the following application error and respond with a JSON object:\n")
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
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var aj analysisJSON
	if err := json.Unmarshal([]byte(text), &aj); err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("openai: parse result JSON: %w", err)
	}
	return domain.AnalysisResult{
		Explanation:     aj.Explanation,
		Severity:        aj.Severity,
		LikelyCause:     aj.LikelyCause,
		SuggestedChecks: aj.SuggestedChecks,
	}, nil
}
