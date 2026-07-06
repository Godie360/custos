# AI Provider Integration

## Context

The AI Analysis Engine is provider-neutral. Custos ships multiple adapter implementations; users configure which provider to use and supply their own API key. The analysis worker only ever talks to the `AIAnalyzer` interface — it has no knowledge of which provider is behind it.

## Requirements

- No provider is baked in as a default. If `CUSTOS_AI_PROVIDER` is not set, analysis is disabled until configured.
- API key must never appear in logs or error messages.
- Analysis must complete within 30 seconds or the worker times out and marks the issue `analysis_failed`.
- Adding a new provider adapter must not require changes to the analysis worker or ingestion code.
- All shipped adapters are open source (MIT or Apache 2.0).

## Decisions

- **Pure adapter pattern** — `AIAnalyzer` interface is the only contract the worker uses.
- **Four shipped adapters**: Claude (Anthropic), OpenAI, Gemini (Google), Ollama (local/self-hosted).
- **User-configured**: provider and key are set via env vars; no default.
- **Provider-neutral prompt contract**: adapters are responsible for translating the shared `AnalysisEvent` into provider-specific API calls and normalizing responses back to `AnalysisResult`.

## Contracts

### Configuration Env Vars

| Env Var | Purpose |
|---|---|
| `CUSTOS_AI_PROVIDER` | Which adapter to load: `claude`, `openai`, `gemini`, `ollama` |
| `CUSTOS_AI_API_KEY` | API key for the configured provider |
| `CUSTOS_AI_MODEL` | Optional model override (e.g. `gpt-4o`, `claude-sonnet-4-6`) |
| `CUSTOS_AI_BASE_URL` | Optional base URL override (for Ollama or self-hosted endpoints) |

### `AIAnalyzer` Interface (Go)

```go
type AIAnalyzer interface {
    Analyze(ctx context.Context, event AnalysisEvent) (AnalysisResult, error)
}

type AnalysisEvent struct {
    ErrorType   string
    Message     string
    StackTrace  []string
    Service     string
    Environment string
}

type AnalysisResult struct {
    Explanation     string
    Severity        string   // "low" | "medium" | "high" | "critical"
    LikelyCause     string
    SuggestedChecks []string
}
```

### Adapter Registry

```go
// internal/analysis/provider/registry.go
func Load(cfg Config) (AIAnalyzer, error)
// Returns the adapter matching cfg.Provider, or error if provider is unknown or unconfigured.
```

### Shipped Adapters

| Provider | Adapter | License |
|---|---|---|
| Anthropic Claude | `provider/claude` | Calls Anthropic REST API |
| OpenAI | `provider/openai` | Calls OpenAI REST API |
| Google Gemini | `provider/gemini` | Calls Gemini REST API |
| Ollama (local) | `provider/ollama` | Calls local Ollama HTTP endpoint; no external key needed |

### Prompt Contract (shared across adapters)

Input to every adapter:
- Error type and message.
- Normalized stack trace (top 10 frames).
- Service name and environment.

Expected structured output (JSON, adapter parses per-provider response format):
- `explanation` — plain-language description of what happened.
- `likely_cause` — probable root cause.
- `severity` — `low | medium | high | critical`.
- `suggested_checks` — up to 3 actionable next steps.

## Acceptance Criteria

- [ ] `AIAnalyzer` interface defined; all four adapters implement it.
- [ ] Analysis worker loads adapter via registry from env config; no provider-specific code in worker.
- [ ] If `CUSTOS_AI_PROVIDER` is unset, worker logs a clear message and skips analysis — no panic.
- [ ] Each adapter handles provider API errors with 3-attempt exponential backoff.
- [ ] `CUSTOS_AI_API_KEY` never appears in application logs or error output.
- [ ] `analysis_failed` issues surfaced in dashboard and retriable manually.
- [ ] Ollama adapter works with no API key (local endpoint only).
