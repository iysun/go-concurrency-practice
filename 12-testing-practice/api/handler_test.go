package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-concurrency-practice/12-testing-practice/api"
)

// newTestServer creates a real HTTP server backed by MemStore for testing.
// httptest.NewServer starts a local listener; use its URL in HTTP client calls.
func newTestServer(t *testing.T) (*httptest.Server, *api.MemStore) {
	t.Helper()
	store := api.NewMemStore()
	mux := http.NewServeMux()
	api.NewHandler(store).RegisterRoutes(mux)
	return httptest.NewServer(mux), store
}

// TestGetItem_NotFound demonstrates testing a 404 response.
func TestGetItem_NotFound(t *testing.T) {
	// httptest.NewRecorder lets you test a handler directly without a network call.
	store := api.NewMemStore()
	mux := http.NewServeMux()
	api.NewHandler(store).RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/items/99", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestPutAndGetItem exercises the full PUT -> GET round trip using a real server.
func TestPutAndGetItem(t *testing.T) {
	srv, _ := newTestServer(t)
	defer srv.Close()

	// PUT /items/1
	body := bytes.NewBufferString(`{"value":"hello"}`)
	resp, err := http.DefaultClient.Do(mustRequest(t,
		http.MethodPut, srv.URL+"/items/1", body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("PUT got %d, want 201", resp.StatusCode)
	}

	// GET /items/1 — should now return the item we just created
	resp, err = http.Get(srv.URL + "/items/1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET got %d, want 200", resp.StatusCode)
	}

	var item api.Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		t.Fatal("decode:", err)
	}
	if item.Value != "hello" {
		t.Errorf("got value %q, want %q", item.Value, "hello")
	}
}

// TestList_Empty verifies the list endpoint returns an empty JSON array, not null.
// This is a common real-world bug: json.Encode(nil slice) = "null", not "[]".
func TestList_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()

	store := api.NewMemStore()
	mux := http.NewServeMux()
	api.NewHandler(store).RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rec.Code)
	}

	// TODO: verify that the response body is "[]" and not "null"
	// hint: use bytes.TrimSpace(rec.Body.Bytes())
}

// TestMethodNotAllowed checks that unsupported methods get a 405.
// TODO: implement this test
func TestMethodNotAllowed(t *testing.T) {
	t.Skip("TODO: implement")
}

func mustRequest(t *testing.T, method, url string, body *bytes.Buffer) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}
