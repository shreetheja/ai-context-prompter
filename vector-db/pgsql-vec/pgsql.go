package pgsqlvec

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shreetheja/ai-contextual-prompter/vector-db"
)

type Entity struct {
	db    *pgxpool.Pool
	table string
	col   string
}

func NewEntity(cfg vector.Config) (*Entity, error) {
	pool, err := Connect(cfg.User, cfg.Password, cfg.Host, cfg.Database)
	if err != nil {
		return nil, err
	}
	return &Entity{db: pool, table: cfg.Table, col: cfg.Col}, nil
}

func (e *Entity) Type(ctx context.Context) string {
	return "pgsql"
}

func (e *Entity) Add(ctx context.Context, emb vector.Embedding) error {
	metaJson, _ := json.Marshal(emb.Meta)
	_, err := e.db.Exec(ctx,
		fmt.Sprintf("INSERT INTO %s (id, %s, meta) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET %s = $2, meta = $3", e.table, e.col, e.col),
		emb.ID, emb.Vec, metaJson)
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
	q := fmt.Sprintf(`SELECT id, %s, meta FROM %s ORDER BY (%s <#> $1::vector) ASC LIMIT $2`, e.col, e.table, e.col)
	rows, err := e.db.Query(ctx, q, query, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []vector.Embedding
	for rows.Next() {
		var id string
		var vec []float64
		var metaJson []byte
		if err := rows.Scan(&id, &vec, &metaJson); err != nil {
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
	q := fmt.Sprintf("DELETE FROM %s WHERE id = $1", e.table)
	_, err := e.db.Exec(ctx, q, id)
	return err
}

func (e *Entity) Clear(ctx context.Context) error {
	q := fmt.Sprintf("DELETE FROM %s", e.table)
	_, err := e.db.Exec(ctx, q)
	return err
}
