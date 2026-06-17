package store

import (
	"context"
	"database/sql"
	"time"

	"go-concurrency-practice/11-rest-api-db/model"
	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

// SQLiteStore implements Store using SQLite via database/sql.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) the SQLite file at dsn and runs migrations.
func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// migrate creates tables if they don't exist.
// For production use golang-migrate or goose for versioned migrations.
func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS todos (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			title      TEXT    NOT NULL,
			done       BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL
		)
	`)
	return err
}

func (s *SQLiteStore) CreateTodo(ctx context.Context, title string) (*model.Todo, error) {
	now := time.Now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO todos (title, done, created_at) VALUES (?, ?, ?)`,
		title, false, now,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &model.Todo{ID: id, Title: title, Done: false, CreatedAt: now}, nil
}

// GetTodo fetches a single todo by primary key.
// TODO: implement — use QueryRowContext, scan columns, return ErrNotFound if sql.ErrNoRows
func (s *SQLiteStore) GetTodo(ctx context.Context, id int64) (*model.Todo, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, title, done, created_at FROM todos WHERE id = ?`, id)
	var t model.Todo
	if err := row.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, &ErrNotFound{ID: id}
		}
		return nil, err
	}
	return &t, nil
}

// ListTodos returns all todos ordered by creation time descending.
// TODO: implement — use QueryContext and iterate rows
func (s *SQLiteStore) ListTodos(ctx context.Context) ([]*model.Todo, error) {
	return nil, nil
}

// UpdateTodo sets the done field for a given ID and returns the updated record.
// TODO: implement — use ExecContext then call GetTodo to return fresh data
func (s *SQLiteStore) UpdateTodo(ctx context.Context, id int64, done bool) (*model.Todo, error) {
	return nil, nil
}

// DeleteTodo removes a todo. Returns ErrNotFound if the ID doesn't exist.
// TODO: implement — check RowsAffected to detect not-found
func (s *SQLiteStore) DeleteTodo(ctx context.Context, id int64) error {
	return nil
}
