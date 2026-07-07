package provider

import (
	"fmt"

	"github.com/Godie360/custos/internal/config"
	"github.com/Godie360/custos/internal/domain"
	"github.com/Godie360/custos/internal/provider/claude"
	"github.com/Godie360/custos/internal/provider/gemini"
	"github.com/Godie360/custos/internal/provider/ollama"
	"github.com/Godie360/custos/internal/provider/openai"
)

// Load returns the AI analyzer configured via cfg.AI.Provider.
// Returns an error if the provider is empty or unrecognised.
func Load(cfg config.Config) (domain.AIAnalyzer, error) {
	switch cfg.AI.Provider {
	case "claude":
		return claude.New(cfg), nil
	case "openai":
		return openai.New(cfg), nil
	case "gemini":
		return gemini.New(cfg), nil
	case "ollama":
		return ollama.New(cfg), nil
	case "":
		return nil, fmt.Errorf("provider: CUSTOS_AI_PROVIDER is not set")
	default:
		return nil, fmt.Errorf("provider: unknown provider %q", cfg.AI.Provider)
	}
}
