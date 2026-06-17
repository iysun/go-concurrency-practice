package framework

import (
	"log"
	"net/http"
	"time"
)

// Logger logs method, path, status code, and elapsed time for every request.
// TODO: implement — use c.Next() to measure time around downstream handlers
func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s %v", c.Request.Method, c.Request.URL.Path, time.Since(start))
	}
}

// Recovery catches panics in downstream handlers and returns 500.
// Without this, a panic would crash the whole server.
// TODO: implement — use defer + recover(), call c.Abort() before returning 500
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(c.Writer, "Internal Server Error", http.StatusInternalServerError)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS adds Access-Control-Allow-* headers for cross-origin requests.
// TODO: implement — handle OPTIONS preflight and set appropriate headers
func CORS(allowOrigin string) HandlerFunc {
	return func(c *Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}
