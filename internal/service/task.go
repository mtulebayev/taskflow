package service

import (
	"context"
	"strings"
	"taskflow/internal/domain"
	"time"
)

type TaskRepository interface {
	Create(ctx context.Context, t domain.Task) (domain.Task, error)
	List(ctx context.Context) ([]domain.Task, error)
	Get(ctx context.Context, id int64) (domain.Task, error)
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, title string) (domain.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Task{}, domain.ErrEmptyTitle
	}

	return s.repo.Create(ctx, domain.Task{Title: title, CreatedAt: time.Now().UTC()})
}

func (s *TaskService) List(ctx context.Context) ([]domain.Task, error) {
	return s.repo.List(ctx)
}

func (s *TaskService) Get(ctx context.Context, id int64) (domain.Task, error) {
	return s.repo.Get(ctx, id)
}
