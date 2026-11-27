package postgres

import (
	"context"
	"errors"
	"taskflow/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{pool: pool}
}

func (r *TaskRepo) Create(ctx context.Context, t domain.Task) (domain.Task, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (title, done, created_at) VALUES ($1, $2, $3) RETURNING id`,
		t.Title, t.Done, t.CreatedAt,
	).Scan(&t.ID)
	return t, err
}

func (r *TaskRepo) List(ctx context.Context) ([]domain.Task, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, title, done, created_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []domain.Task{}
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *TaskRepo) Get(ctx context.Context, id int64) (domain.Task, error) {
	var t domain.Task
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE id = $1`, id,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, domain.ErrNotFound
	}
	return t, err
}
