package vector

import (
	"context"
)

// Embedding represents a vector and its associated metadata.
type Embedding struct {
	ID   string
	Vec  []float64
	Meta map[string]interface{}
}

// VectorDB defines the interface for a vector database.
type VectorDB interface {
	// Returns the type of vector DB ( inmem or online defined in directories)
	Type(ctx context.Context) string

	// Add stores an embedding in the database.
	Add(ctx context.Context, emb Embedding) error

	// Search returns the top K most similar embeddings to the query vector.
	Search(ctx context.Context, query []float64, topK int) ([]Embedding, error)

	// Count returns the number of embeddings stored.
	Count(ctx context.Context) (int, error)

	// Delete removes an embedding by ID.
	Delete(ctx context.Context, id string) error

	// Clear removes all embeddings.
	Clear(ctx context.Context) error
}

// Vector DB type
var (
	IN_MEMORY = "in_mem"
	PG_SQL    = "pg_sql"
)

// Config for selecting and configuring a vector DB
type Config struct {
	Type string
	// In-memory: no extra fields
	// PGSQL:
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Table    string
	Col      string
}

// Useful vector math helpers
func CosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float64) float64 {
	// Simple Newton's method for sqrt
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
