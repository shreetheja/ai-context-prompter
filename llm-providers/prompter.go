package llmproviders

import (
	"context"
)

// LLM is an interface for large language models that support context-aware prompting.
type LLM interface {
	// Name returns the name of the model/provider (e.g., "openai", "claude", "gemini").
	Name() string

	// Embed returns the embedding vector for a given text.
	Embed(ctx context.Context, text string) ([]float64, error)

	// PromptWithContext sends a prompt and context to the model and returns the response.
	PromptWithContext(ctx context.Context, prompt string, contextItems []string, opts ...PromptOption) (string, error)

	// MaxContext returns the maximum number of context tokens supported by the model.
	MaxContext() int
}

// PromptOption is a marker interface for provider-specific options.
type PromptOption interface{}

// List LLM's Supported
const OPEN_AI = "openai"
