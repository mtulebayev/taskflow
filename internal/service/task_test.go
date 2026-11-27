package service_test

import (
	"context"
	"errors"
	"testing"

	"taskflow/internal/domain"
	"taskflow/internal/repository/memory"
	"taskflow/internal/service"
)

func TestCreate_EmptyTitle(t *testing.T) {
	svc := service.NewTaskService(memory.NewTaskRepo())

	_, err := svc.Create(context.Background(), "   ")

	if !errors.Is(err, domain.ErrEmptyTitle) {
		t.Fatalf("want ErrEmptyTitle, got %v", err)
	}
}

func TestCreate_TrimsTitleAndAssignsID(t *testing.T) {
	svc := service.NewTaskService(memory.NewTaskRepo())

	task, err := svc.Create(context.Background(), "  buy milk  ")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != 1 || task.Title != "buy milk" {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := service.NewTaskService(memory.NewTaskRepo())

	_, err := svc.Get(context.Background(), 42)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
