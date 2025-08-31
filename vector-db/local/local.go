package local

import (
	"context"
	"sort"

	"github.com/shreetheja/ai-contextual-prompter/vector-db"
)

// InMemoryVectorDB is an in-memory implementation of VectorDB.
type InMemoryVectorDB struct {
	store map[string]vector.Embedding
}

func NewInMemoryVectorDB() *InMemoryVectorDB {
	return &InMemoryVectorDB{store: make(map[string]vector.Embedding)}
}

// AddN adds multiple vector.embeddings to the database.
func (db *InMemoryVectorDB) AddN(ctx context.Context, embs []vector.Embedding) error {
	for _, emb := range embs {
		db.store[emb.ID] = emb
	}
	return nil
}

func (db *InMemoryVectorDB) Type(ctx context.Context) string {
	return vector.IN_MEMORY
}

func (db *InMemoryVectorDB) Add(ctx context.Context, emb vector.Embedding) error {
	db.store[emb.ID] = emb
	return nil
}

func (db *InMemoryVectorDB) Search(ctx context.Context, query []float64, topK int) ([]vector.Embedding, error) {
	type scored struct {
		emb   vector.Embedding
		score float64
	}
	var scoredList []scored
	for _, emb := range db.store {
		sim := vector.CosineSimilarity(query, emb.Vec)
		scoredList = append(scoredList, scored{emb: emb, score: sim})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].score > scoredList[j].score
	})
	var result []vector.Embedding
	for i := 0; i < topK && i < len(scoredList); i++ {
		result = append(result, scoredList[i].emb)
	}
	return result, nil
}

func (db *InMemoryVectorDB) Count(ctx context.Context) (int, error) {
	return len(db.store), nil
}

func (db *InMemoryVectorDB) Delete(ctx context.Context, id string) error {
	delete(db.store, id)
	return nil
}

func (db *InMemoryVectorDB) Clear(ctx context.Context) error {
	db.store = make(map[string]vector.Embedding)
	return nil
}
