package framework

import (
	"net/http"
)

// Engine is the top-level HTTP handler. It holds global middleware and the router.
//
// Usage:
//
//	e := framework.New()
//	e.Use(framework.Logger(), framework.Recovery())
//	e.GET("/ping", func(c *framework.Context) { c.String(200, "pong") })
//	e.POST("/user/:id", handler)
//	http.ListenAndServe(":8080", e)
type Engine struct {
	router     *router
	middleware []HandlerFunc // global middleware applied to every request
}

func New() *Engine {
	return &Engine{router: newRouter()}
}

// Use registers global middleware. Middleware runs in registration order
// before the route handler.
func (e *Engine) Use(middleware ...HandlerFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// GET registers a handler for GET requests matching pattern.
func (e *Engine) GET(pattern string, handlers ...HandlerFunc) {
	e.addRoute(http.MethodGet, pattern, handlers)
}

// POST registers a handler for POST requests matching pattern.
func (e *Engine) POST(pattern string, handlers ...HandlerFunc) {
	e.addRoute(http.MethodPost, pattern, handlers)
}

// PUT registers a handler for PUT requests.
func (e *Engine) PUT(pattern string, handlers ...HandlerFunc) {
	e.addRoute(http.MethodPut, pattern, handlers)
}

// DELETE registers a handler for DELETE requests.
func (e *Engine) DELETE(pattern string, handlers ...HandlerFunc) {
	e.addRoute(http.MethodDelete, pattern, handlers)
}

func (e *Engine) addRoute(method, pattern string, handlers []HandlerFunc) {
	// Combine global middleware with route-specific handlers
	all := make([]HandlerFunc, 0, len(e.middleware)+len(handlers))
	all = append(all, e.middleware...)
	all = append(all, handlers...)
	e.router.addRoute(method, pattern, all)
}

// ServeHTTP makes Engine implement http.Handler.
// It looks up the route, builds a Context, and runs the handler chain.
// TODO: return 404 when no route matches, 405 when method not allowed
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handlers, params := e.router.match(r.Method, r.URL.Path)
	if handlers == nil {
		http.NotFound(w, r)
		return
	}
	c := newContext(w, r)
	c.Params = params
	c.handlers = handlers
	c.Next()
}
