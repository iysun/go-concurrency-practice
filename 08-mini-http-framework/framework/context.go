package framework

import (
	"encoding/json"
	"net/http"
)

// HandlerFunc is the function signature for route handlers and middleware.
type HandlerFunc func(*Context)

// Context wraps the standard http.ResponseWriter and *http.Request,
// providing helper methods and carrying per-request state (params, middleware chain).
type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string // URL path parameters, e.g. /user/:id -> {"id": "42"}

	// middleware chain state
	handlers []HandlerFunc
	index    int
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
		index:   -1,
	}
}

// Next executes the next handler in the middleware chain.
// Call this inside a middleware to pass control forward.
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort stops the middleware chain. Handlers after the current one will not run.
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

// Status writes the HTTP status code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// JSON serializes v as JSON and writes it with Content-Type: application/json.
// TODO: implement
func (c *Context) JSON(code int, v any) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	if err := json.NewEncoder(c.Writer).Encode(v); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// String writes a plain-text response.
func (c *Context) String(code int, format string, values ...any) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(code)
	// TODO: use fmt.Fprintf(c.Writer, format, values...)
}

// BindJSON decodes the request body into v.
// TODO: implement
func (c *Context) BindJSON(v any) error {
	return json.NewDecoder(c.Request.Body).Decode(v)
}

// Param returns a URL path parameter by name.
func (c *Context) Param(name string) string {
	return c.Params[name]
}

// Query returns a URL query parameter by name.
func (c *Context) Query(name string) string {
	return c.Request.URL.Query().Get(name)
}
