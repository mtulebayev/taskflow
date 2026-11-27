package domain

import (
	"errors"
	"time"
)

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	ErrNotFound   = errors.New("task not found")
	ErrEmptyTitle = errors.New("title must not be empty")
)
