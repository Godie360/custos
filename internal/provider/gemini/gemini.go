package gemini

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
	defaultModel   = "gemini-1.5-flash"
	baseURL        = "https://generativelanguage.googleapis.com/v1beta/models"
	maxRetries     = 3
	requestTimeout = 30 * time.Second
)

// Analyzer implements domain.AIAnalyzer using the Google Gemini API.
type Analyzer struct {
	cfg    config.Config
	client *http.Client
}

// New creates an Analyzer for the Gemini provider.
func New(cfg config.Config) *Analyzer {
	return &Analyzer{
		cfg:    cfg,
		client: &http.Client{Timeout: requestTimeout},
	}
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Analyze sends an error event to Gemini and returns a structured AnalysisResult.
func (a *Analyzer) Analyze(ctx context.Context, event domain.AnalysisEvent) (domain.AnalysisResult, error) {
	model := a.cfg.AI.Model
	if model == "" {
		model = defaultModel
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", baseURL, model, a.cfg.AI.APIKey)
	prompt := buildPrompt(event)
	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
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

		payload, err := json.Marshal(reqBody)
		if err != nil {
			return result, fmt.Errorf("gemini: marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return result, fmt.Errorf("gemini: create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		respBytes, status, err := a.doRequest(req)
		if err != nil {
			lastErr = fmt.Errorf("gemini: http: %w", err)
			continue
		}

		if status == http.StatusTooManyRequests || status >= 500 {
			lastErr = fmt.Errorf("gemini: status %d", status)
			continue
		}

		if status != http.StatusOK {
			return result, fmt.Errorf("gemini: status %d: %s", status, string(respBytes))
		}

		var gr geminiResponse
		if err := json.Unmarshal(respBytes, &gr); err != nil {
			return result, fmt.Errorf("gemini: unmarshal response: %w", err)
		}
		if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
			return result, fmt.Errorf("gemini: empty candidates")
		}

		return parseResult(gr.Candidates[0].Content.Parts[0].Text)
	}

	return result, fmt.Errorf("gemini: max retries exceeded: %w", lastErr)
}

func (a *Analyzer) doRequest(req *http.Request) ([]byte, int, error) {
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close() //nolint:errcheck // response body drain; error not actionable
	b, err := io.ReadAll(resp.Body)
	return b, resp.StatusCode, err
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
		return domain.AnalysisResult{}, fmt.Errorf("gemini: parse result JSON: %w", err)
	}
	return domain.AnalysisResult{
		Explanation:     aj.Explanation,
		Severity:        aj.Severity,
		LikelyCause:     aj.LikelyCause,
		SuggestedChecks: aj.SuggestedChecks,
	}, nil
}
