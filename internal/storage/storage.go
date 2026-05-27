package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Graph struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      string    `json:"data"`
}

type Store interface {
	SaveGraph(ctx context.Context, name, data string) (*Graph, error)
	UpdateGraph(ctx context.Context, id int64, name, data string) (*Graph, error)
	RenameGraph(ctx context.Context, id int64, name string) (*Graph, error)
	ListGraphs(ctx context.Context, limit, offset int) ([]Graph, int, error)
	GetGraph(ctx context.Context, id int64) (*Graph, error)
	DeleteGraph(ctx context.Context, id int64) error
	Close()
}

type pgStore struct {
	pool *pgxpool.Pool
}

const schema = `
CREATE TABLE IF NOT EXISTS graphs (
	id         BIGSERIAL PRIMARY KEY,
	name       TEXT        NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	data       TEXT        NOT NULL
);`

func New(dsn string) (Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &pgStore{pool: pool}, nil
}

func (s *pgStore) SaveGraph(ctx context.Context, name, data string) (*Graph, error) {
	var g Graph
	err := s.pool.QueryRow(ctx,
		`INSERT INTO graphs (name, data) VALUES ($1, $2)
		 RETURNING id, name, created_at, updated_at, data`,
		name, data,
	).Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt, &g.Data)
	if err != nil {
		return nil, fmt.Errorf("insert graph: %w", err)
	}
	return &g, nil
}

func (s *pgStore) UpdateGraph(ctx context.Context, id int64, name, data string) (*Graph, error) {
	var g Graph
	err := s.pool.QueryRow(ctx,
		`UPDATE graphs SET name=$1, data=$2, updated_at=NOW()
		 WHERE id=$3
		 RETURNING id, name, created_at, updated_at, data`,
		name, data, id,
	).Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt, &g.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update graph: %w", err)
	}
	return &g, nil
}

func (s *pgStore) RenameGraph(ctx context.Context, id int64, name string) (*Graph, error) {
	var g Graph
	err := s.pool.QueryRow(ctx,
		`UPDATE graphs SET name=$1, updated_at=NOW()
		 WHERE id=$2
		 RETURNING id, name, created_at, updated_at, data`,
		name, id,
	).Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt, &g.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("rename graph: %w", err)
	}
	return &g, nil
}

func (s *pgStore) ListGraphs(ctx context.Context, limit, offset int) ([]Graph, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM graphs`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count graphs: %w", err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, name, created_at, updated_at, data
		 FROM graphs ORDER BY updated_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list graphs: %w", err)
	}
	defer rows.Close()

	var graphs []Graph
	for rows.Next() {
		var g Graph
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt, &g.Data); err != nil {
			return nil, 0, fmt.Errorf("scan graph: %w", err)
		}
		graphs = append(graphs, g)
	}
	return graphs, total, rows.Err()
}

func (s *pgStore) GetGraph(ctx context.Context, id int64) (*Graph, error) {
	var g Graph
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, created_at, updated_at, data FROM graphs WHERE id=$1`, id,
	).Scan(&g.ID, &g.Name, &g.CreatedAt, &g.UpdatedAt, &g.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get graph: %w", err)
	}
	return &g, nil
}

func (s *pgStore) DeleteGraph(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM graphs WHERE id=$1`, id)
	return err
}

func (s *pgStore) Close() {
	s.pool.Close()
}
