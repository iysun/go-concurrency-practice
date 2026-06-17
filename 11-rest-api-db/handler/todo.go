package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go-concurrency-practice/11-rest-api-db/model"
	"go-concurrency-practice/11-rest-api-db/store"
)

// TodoHandler wires the HTTP routes for the Todo resource.
type TodoHandler struct {
	store store.Store
}

func NewTodoHandler(s store.Store) *TodoHandler {
	return &TodoHandler{store: s}
}

// RegisterRoutes registers all /todos routes on mux.
// Routes:
//
//	POST   /todos        -> CreateTodo
//	GET    /todos        -> ListTodos
//	GET    /todos/{id}   -> GetTodo
//	PATCH  /todos/{id}   -> UpdateTodo
//	DELETE /todos/{id}   -> DeleteTodo
func (h *TodoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/todos", h.handleCollection)
	mux.HandleFunc("/todos/", h.handleItem)
}

func (h *TodoHandler) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTodo(w, r)
	case http.MethodGet:
		h.listTodos(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TodoHandler) handleItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getTodo(w, r, id)
	case http.MethodPatch:
		h.updateTodo(w, r, id)
	case http.MethodDelete:
		h.deleteTodo(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TodoHandler) createTodo(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		jsonError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	todo, err := h.store.CreateTodo(r.Context(), req.Title)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, todo, http.StatusCreated)
}

// listTodos returns all todos.
// TODO: implement — call h.store.ListTodos, render JSON
func (h *TodoHandler) listTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.ListTodos(r.Context())
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, todos, http.StatusOK)
}

// getTodo returns a single todo.
// TODO: implement — handle ErrNotFound -> 404
func (h *TodoHandler) getTodo(w http.ResponseWriter, r *http.Request, id int64) {
	todo, err := h.store.GetTodo(r.Context(), id)
	if err != nil {
		var nf *store.ErrNotFound
		if errors.As(err, &nf) {
			jsonError(w, "not found", http.StatusNotFound)
			return
		}
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, todo, http.StatusOK)
}

// updateTodo toggles the done field.
// TODO: implement
func (h *TodoHandler) updateTodo(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Done == nil {
		jsonError(w, "done field is required", http.StatusUnprocessableEntity)
		return
	}
	// TODO: call h.store.UpdateTodo and return updated todo
	jsonError(w, "not implemented", http.StatusNotImplemented)
}

// deleteTodo removes a todo.
// TODO: implement
func (h *TodoHandler) deleteTodo(w http.ResponseWriter, r *http.Request, id int64) {
	jsonError(w, "not implemented", http.StatusNotImplemented)
}

// parseID extracts the trailing ID from a path like "/todos/42"
func parseID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return strconv.ParseInt(parts[len(parts)-1], 10, 64)
}

func jsonOK(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
