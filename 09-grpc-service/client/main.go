//go:build ignore
// +build ignore

// Remove build tags after generating the pb package (see server/main.go).

package main

import (
	"context"
	"log"
	"time"

	"go-concurrency-practice/09-grpc-service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Dial the server (insecure for local dev; use TLS in production)
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create
	resp, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{Title: "Learn gRPC"})
	if err != nil {
		log.Fatal("CreateTodo:", err)
	}
	log.Printf("created: %+v", resp.Todo)

	// Get
	// TODO: call GetTodo with resp.Todo.Id

	// List
	// TODO: call ListTodos and print all

	// Update
	// TODO: call UpdateTodo to mark the todo as done

	// Delete
	// TODO: call DeleteTodo
}
