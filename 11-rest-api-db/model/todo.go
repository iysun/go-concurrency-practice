package model

import "time"

// Todo is the domain model.
type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTodoRequest is the JSON body for POST /todos.
type CreateTodoRequest struct {
	Title string `json:"title"`
}

func (r *CreateTodoRequest) Validate() error {
	if r.Title == "" {
		return &ValidationError{Field: "title", Message: "is required"}
	}
	if len(r.Title) > 255 {
		return &ValidationError{Field: "title", Message: "must be at most 255 characters"}
	}
	return nil
}

// UpdateTodoRequest is the JSON body for PATCH /todos/:id.
type UpdateTodoRequest struct {
	Done *bool `json:"done"` // pointer so we can detect "not provided"
}

// ValidationError is returned when request input fails validation.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + " " + e.Message
}
