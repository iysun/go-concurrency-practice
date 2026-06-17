package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// Item is a simple resource used to demonstrate handler testing.
type Item struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

// Store is an in-memory repository of Items, injectable for testing.
type Store interface {
	Get(id int) (*Item, bool)
	List() []*Item
	Set(id int, value string) *Item
}

// MemStore is the default in-memory Store implementation.
type MemStore struct {
	items map[int]*Item
}

func NewMemStore() *MemStore {
	return &MemStore{items: make(map[int]*Item)}
}

func (m *MemStore) Get(id int) (*Item, bool) {
	item, ok := m.items[id]
	return item, ok
}

func (m *MemStore) List() []*Item {
	result := make([]*Item, 0, len(m.items))
	for _, v := range m.items {
		result = append(result, v)
	}
	return result
}

func (m *MemStore) Set(id int, value string) *Item {
	item := &Item{ID: id, Value: value}
	m.items[id] = item
	return item
}

// Handler is an HTTP handler that serves Items.
type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/items", h.list)
	mux.HandleFunc("/items/", h.item)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	items := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) item(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/items/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, ok := h.store.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)

	case http.MethodPut:
		var body struct {
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		item := h.store.Set(id, body.Value)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
