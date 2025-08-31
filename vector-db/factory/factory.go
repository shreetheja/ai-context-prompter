package factory

import (
	"fmt"

	"github.com/shreetheja/ai-contextual-prompter/vector-db"
	"github.com/shreetheja/ai-contextual-prompter/vector-db/local"
	pgsqlvec "github.com/shreetheja/ai-contextual-prompter/vector-db/pgsql-vec"
)

// NewVectorDB returns a VectorDB implementation based on config.Type
func NewVectorDB(cfg vector.Config) (vector.VectorDB, error) {
	switch cfg.Type {
	case vector.IN_MEMORY:
		// import path: "github.com/shreetheja/ai-contextual-prompter/vector/local"
		return local.NewInMemoryVectorDB(), nil
	case vector.PG_SQL:
		// import path: "github.com/shreetheja/ai-contextual-prompter/vector/pgsql-vec"
		entity, err := pgsqlvec.NewEntity(cfg)
		if err != nil {
			return nil, err
		}
		return entity, nil
	default:
		return nil, fmt.Errorf("unknown vector db type: %s", cfg.Type)
	}
}
