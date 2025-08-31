package context_prompter

import (
	"context"

	llmproviders "github.com/shreetheja/ai-contextual-prompter/llm-providers"
	"github.com/shreetheja/ai-contextual-prompter/vector-db"
)

// ContextPrompter defines the interface for managing context and prompting LLMs.
type ContextPrompter interface {
	// AddContext adds a new context item (text + metadata) and stores its embedding.
	AddContext(ctx context.Context, text string, meta map[string]interface{}) error

	// Query builds a prompt using the most relevant context and queries the LLM.
	Query(ctx context.Context, prompt string, topK int, opts ...llmproviders.PromptOption) (string, error)

	// SimilarContext returns the top K most relevant context items for a query.
	SimilarContext(ctx context.Context, query string, topK int) ([]vector.Embedding, error)

	// ClearContext removes all stored context.
	ClearContext(ctx context.Context) error
}

// Optionally, you can define a struct for configuration.
type Config struct {
	VectorDB   vector.VectorDB
	LLM        llmproviders.LLM
	MaxContext int // max context items/tokens to use in prompt
}
