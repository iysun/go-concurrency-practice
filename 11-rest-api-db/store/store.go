package store

import (
	"context"
	"go-concurrency-practice/11-rest-api-db/model"
)

// Store is the data access interface. Swap implementations (SQLite, Postgres, in-memory)
// without changing handler code — the handlers only depend on this interface.
type Store interface {
	CreateTodo(ctx context.Context, title string) (*model.Todo, error)
	GetTodo(ctx context.Context, id int64) (*model.Todo, error)
	ListTodos(ctx context.Context) ([]*model.Todo, error)
	UpdateTodo(ctx context.Context, id int64, done bool) (*model.Todo, error)
	DeleteTodo(ctx context.Context, id int64) error
}

// ErrNotFound is returned when a requested record does not exist.
type ErrNotFound struct {
	ID int64
}

func (e *ErrNotFound) Error() string {
	return "todo not found"
}
