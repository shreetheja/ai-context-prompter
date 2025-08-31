package pgsqlvec

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shreetheja/ai-contextual-prompter/vector-db"
)

type Entity struct {
	db        *pgxpool.Pool
	table     string
	col       string
	idColname interface{}
}

func NewEntity(cfg vector.Config) (*Entity, error) {
	pool, err := Connect(cfg.User, cfg.Password, cfg.Host, cfg.Database)
	if err != nil {
		return nil, err
	}
	return &Entity{db: pool, table: cfg.Table, col: cfg.Col, idColname: cfg.IdColName}, nil
}

func (e *Entity) Type(ctx context.Context) string {
	return "pgsql"
}

func (e *Entity) Add(ctx context.Context, emb vector.Embedding) error {
	metaJson, _ := json.Marshal(emb.Meta)
	vecStr := floatSliceToPgvector(emb.Vec)
	_, err := e.db.Exec(ctx,
		fmt.Sprintf("INSERT INTO %s (%s, %s, meta) VALUES ($1, $2, $3) ON CONFLICT (%s) DO UPDATE SET %s = $2, meta = $3",
			e.table, e.idColname, e.col, e.idColname, e.col),
		emb.ID, vecStr, metaJson)
	return err
}

func (e *Entity) AddN(ctx context.Context, embs []vector.Embedding) error {
	for _, emb := range embs {
		if err := e.Add(ctx, emb); err != nil {
			return err
		}
	}
	return nil
}

func (e *Entity) Search(ctx context.Context, query []float64, topK int) ([]vector.Embedding, error) {
	vecStr := floatSliceToPgvector(query)
	q := fmt.Sprintf(`SELECT %s, %s, meta FROM %s ORDER BY (%s <#> $1::vector) ASC LIMIT $2`, e.idColname, e.col, e.table, e.col)
	rows, err := e.db.Query(ctx, q, vecStr, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []vector.Embedding
	for rows.Next() {
		var id string
		var vecStr string
		var metaJson []byte
		if err := rows.Scan(&id, &vecStr, &metaJson); err != nil {
			return nil, err
		}
		vec, err := parsePgvectorString(vecStr)
		if err != nil {
			return nil, err
		}
		var meta map[string]interface{}
		json.Unmarshal(metaJson, &meta)
		out = append(out, vector.Embedding{ID: id, Vec: vec, Meta: meta})
	}
	return out, nil
}

func (e *Entity) Count(ctx context.Context) (int, error) {
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s", e.table)
	var count int
	err := e.db.QueryRow(ctx, q).Scan(&count)
	return count, err
}

func (e *Entity) Delete(ctx context.Context, id string) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", e.table, e.idColname)
	_, err := e.db.Exec(ctx, q, id)
	return err
}

func (e *Entity) Clear(ctx context.Context) error {
	q := fmt.Sprintf("DELETE FROM %s", e.table)
	_, err := e.db.Exec(ctx, q)
	return err
}

// floatSliceToPgvector converts a []float64 to a pgvector string literal: [0.1, 0.2, 0.3]
func floatSliceToPgvector(vec []float64) string {
	s := make([]string, len(vec))
	for i, v := range vec {
		s[i] = fmt.Sprintf("%g", v)
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ", "))
}

// parsePgvectorString parses a pgvector string literal like "[0.1, 0.2, 0.3]" to []float64
func parsePgvectorString(s string) ([]float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]float64, len(parts))
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, err
		}
		out[i] = f
	}
	return out, nil
}

var _ vector.VectorDB = &Entity{}
