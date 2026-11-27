package memory

import (
	"context"
	"sort"
	"sync"
	"taskflow/internal/domain"
)

type TaskRepo struct {
	mu     sync.Mutex
	nextID int64
	tasks  map[int64]domain.Task
}

func NewTaskRepo() *TaskRepo {
	return &TaskRepo{nextID: 1, tasks: make(map[int64]domain.Task)}
}

func (r *TaskRepo) Create(_ context.Context, t domain.Task) (domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t.ID = r.nextID
	r.nextID++
	r.tasks[t.ID] = t
	return t, nil
}

func (r *TaskRepo) List(_ context.Context) ([]domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *TaskRepo) Get(_ context.Context, id int64) (domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return domain.Task{}, domain.ErrNotFound
	}

	return t, nil
}
