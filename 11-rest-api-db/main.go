package main

import (
	"log"
	"net/http"

	"go-concurrency-practice/11-rest-api-db/handler"
	"go-concurrency-practice/11-rest-api-db/store"
)

func main() {
	// Open SQLite database (file-based; use ":memory:" for tests)
	s, err := store.NewSQLiteStore("todos.db")
	if err != nil {
		log.Fatal("open db:", err)
	}

	mux := http.NewServeMux()
	handler.NewTodoHandler(s).RegisterRoutes(mux)

	// TODO: wrap mux with logging and recovery middleware
	// (reuse the patterns from 08-mini-http-framework)

	log.Println("REST API listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
