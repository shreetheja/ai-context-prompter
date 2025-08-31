package context_prompter

import (
	"context"
	"errors"
	"fmt"

	llmproviders "github.com/shreetheja/ai-contextual-prompter/llm-providers"
	"github.com/shreetheja/ai-contextual-prompter/vector-db"
)

// Prompter manages context and LLM for contextual prompting.
type Prompter struct {
	VectorDB   vector.VectorDB
	LLM        llmproviders.LLM
	MaxContext int // max context items/tokens to use in prompt
}

// NewPrompter returns an empty Prompter with MaxContext set.
func NewPrompter(maxContext int) *Prompter {
	return &Prompter{MaxContext: maxContext}
}

// NewPrompterWithLLM returns a Prompter with LLM and MaxContext set.
func NewPrompterWithLLM(llm llmproviders.LLM, maxContext int) *Prompter {
	return &Prompter{LLM: llm, MaxContext: maxContext}
}

// NewPrompterWithVector returns a Prompter with VectorDB and MaxContext set.
func NewPrompterWithVector(vdb vector.VectorDB, maxContext int) *Prompter {
	return &Prompter{VectorDB: vdb, MaxContext: maxContext}
}

// SetLLM sets the LLM provider for the Prompter.
func (p *Prompter) SetLLM(llm llmproviders.LLM) {
	p.LLM = llm
}

// SetVector sets the VectorDB for the Prompter.
func (p *Prompter) SetVector(vdb vector.VectorDB) {
	p.VectorDB = vdb
}

// AddContext adds a new context item (text + metadata) and stores its embedding.
func (p *Prompter) AddContext(ctx context.Context, text string, meta map[string]interface{}) error {
	if p.LLM == nil || p.VectorDB == nil {
		return errors.New("LLM and VectorDB must be set")
	}
	embedding, err := p.LLM.Embed(ctx, text)
	if err != nil {
		return err
	}
	emb := vector.Embedding{
		ID:   text, // You may want to use a hash or UUID here
		Vec:  embedding,
		Meta: meta,
	}
	return p.VectorDB.Add(ctx, emb)
}

// SimilarContext returns the top K most relevant context items for a query.
func (p *Prompter) SimilarContext(ctx context.Context, query string, topK int) ([]vector.Embedding, error) {
	if p.LLM == nil || p.VectorDB == nil {
		return nil, errors.New("LLM and VectorDB must be set")
	}
	queryVec, err := p.LLM.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	return p.VectorDB.Search(ctx, queryVec, topK)
}

// Query builds a prompt using the most relevant context and queries the LLM.
func (p *Prompter) Query(ctx context.Context, prompt string, topK int, opts ...llmproviders.PromptOption) (string, error) {
	contexts, err := p.SimilarContext(ctx, prompt, topK)
	if err != nil {
		return "", err
	}
	fmt.Printf("len(contexts): %v\n", len(contexts))
	var contextItems []string
	for _, emb := range contexts {
		if metaText, ok := emb.Meta["text"].(string); ok {
			contextItems = append(contextItems, metaText)
		} else {
			contextItems = append(contextItems, emb.ID)
		}
	}
	return p.LLM.PromptWithContext(ctx, prompt, contextItems, opts...)
}

// ClearContext removes all stored context.
func (p *Prompter) ClearContext(ctx context.Context) error {
	if p.VectorDB == nil {
		return errors.New("VectorDB must be set")
	}
	return p.VectorDB.Clear(ctx)
}
