package factory

import (
	"fmt"

	llmproviders "github.com/shreetheja/ai-contextual-prompter/llm-providers"
	"github.com/shreetheja/ai-contextual-prompter/llm-providers/openai"
)

// OpenAIConfig holds config for OpenAI client
type OpenAIConfig struct {
	SecKey string
	OrgId  string
	AsstId *string // pointer: nil for classic, value for assistant
}

// NewLLM returns an LLM implementation based on provider type
func NewLLM(provider string, cfg ...llmproviders.PromptOption) (llmproviders.LLM, error) {
	switch provider {
	case "openai":
		return openai.New(cfg...)
	// Add more providers here (e.g., "claude", "gemini")
	default:
		return nil, fmt.Errorf("unknown llm provider: %s", provider)
	}
}
