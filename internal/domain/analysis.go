package domain

import "context"

// AIAnalyzer is implemented by each AI provider adapter.
type AIAnalyzer interface {
	Analyze(ctx context.Context, event AnalysisEvent) (AnalysisResult, error)
}

// AnalysisEvent carries the data sent to the AI provider.
type AnalysisEvent struct {
	ErrorType   string
	Message     string
	StackTrace  []string
	Service     string
	Environment string
}

// AnalysisResult holds the structured output from the AI provider.
type AnalysisResult struct {
	Explanation     string
	Severity        string
	LikelyCause     string
	SuggestedChecks []string
}
