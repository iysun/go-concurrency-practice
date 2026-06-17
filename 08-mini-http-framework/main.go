package main

import (
	"log"
	"net/http"

	"go-concurrency-practice/08-mini-http-framework/framework"
)

func main() {
	e := framework.New()

	// Global middleware: every request goes through Logger then Recovery
	e.Use(framework.Logger(), framework.Recovery())

	// Static route
	e.GET("/ping", func(c *framework.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Path parameter: /user/:id
	e.GET("/user/:id", func(c *framework.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"id": c.Param("id"),
		})
	})

	// POST with JSON body binding
	e.POST("/user", func(c *framework.Context) {
		var body struct {
			Name string `json:"name"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, map[string]string{"name": body.Name})
	})

	// Demonstrate Recovery: this handler panics but the server keeps running
	e.GET("/panic", func(c *framework.Context) {
		panic("intentional panic — Recovery middleware should catch this")
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", e))
}
