//go:build ignore
// +build ignore

// Run `go generate ../` first to produce the pb package from todo.proto,
// then remove the build tags above to compile this file.
//
// Generate command (requires protoc + protoc-gen-go + protoc-gen-go-grpc):
//
//	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//	protoc --go_out=../pb --go-grpc_out=../pb --proto_path=../proto ../proto/todo.proto

package main

import (
	"context"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go-concurrency-practice/09-grpc-service/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// todoServer implements pb.TodoServiceServer.
// In-memory store; replace with a real database in production.
type todoServer struct {
	pb.UnimplementedTodoServiceServer
	mu      sync.RWMutex
	todos   map[int64]*pb.Todo
	counter int64
}

func newTodoServer() *todoServer {
	return &todoServer{todos: make(map[int64]*pb.Todo)}
}

func (s *todoServer) CreateTodo(_ context.Context, req *pb.CreateTodoRequest) (*pb.TodoResponse, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	id := atomic.AddInt64(&s.counter, 1)
	todo := &pb.Todo{
		Id:        id,
		Title:     req.Title,
		Done:      false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.mu.Lock()
	s.todos[id] = todo
	s.mu.Unlock()
	return &pb.TodoResponse{Todo: todo}, nil
}

// GetTodo returns a single todo by ID.
// TODO: implement — use s.mu.RLock, return codes.NotFound if missing
func (s *todoServer) GetTodo(_ context.Context, req *pb.GetTodoRequest) (*pb.TodoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ListTodos returns all todos.
// TODO: implement
func (s *todoServer) ListTodos(_ context.Context, _ *pb.ListTodosRequest) (*pb.ListTodosResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// UpdateTodo toggles the done field.
// TODO: implement
func (s *todoServer) UpdateTodo(_ context.Context, req *pb.UpdateTodoRequest) (*pb.TodoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// DeleteTodo removes a todo by ID.
// TODO: implement
func (s *todoServer) DeleteTodo(_ context.Context, req *pb.DeleteTodoRequest) (*pb.DeleteTodoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// loggingInterceptor is a unary server interceptor that logs each RPC call.
// TODO: implement — log method name, duration, and error if any
func loggingInterceptor(
	ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("rpc=%s duration=%v err=%v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func main() {
	ln, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	pb.RegisterTodoServiceServer(srv, newTodoServer())

	log.Println("gRPC server listening on :50051")
	if err := srv.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
