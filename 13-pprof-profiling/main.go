package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // side-effect import registers /debug/pprof routes
	"sync"
	"time"
)

// ---- Intentional performance problems for you to diagnose with pprof ----
//
// This server has THREE hidden issues. Find them using pprof:
//
//  1. CPU hot path   — one endpoint does excessive CPU work
//  2. Memory leak    — a goroutine accumulates data that is never released
//  3. Goroutine leak — a goroutine is started but never exits
//
// How to profile:
//
//	# Start this server
//	go run ./13-pprof-profiling/
//
//	# In another terminal — hit the endpoints a few times
//	curl http://localhost:6060/work
//	curl http://localhost:6060/alloc
//
//	# CPU profile (30-second sample)
//	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
//
//	# Heap profile
//	go tool pprof http://localhost:6060/debug/pprof/heap
//
//	# Goroutine profile — look for leaked goroutines
//	go tool pprof http://localhost:6060/debug/pprof/goroutine
//
//	# Flame graph (requires graphviz)
//	# Inside pprof shell: web

// ---- Problem 1: CPU hot path ----

// inefficientFib calculates Fibonacci recursively — O(2^n) instead of O(n).
// TODO: find this with CPU pprof, then fix with an iterative implementation.
func inefficientFib(n int) int {
	if n <= 1 {
		return n
	}
	return inefficientFib(n-1) + inefficientFib(n-2)
}

func workHandler(w http.ResponseWriter, r *http.Request) {
	result := inefficientFib(40) // intentionally expensive
	fmt.Fprintf(w, "fib(40) = %d\n", result)
}

// ---- Problem 2: Memory leak ----

var leakyCache = struct {
	mu   sync.Mutex
	data [][]byte
}{}

// allocHandler appends 1 MB to a global slice on every call and never trims it.
// TODO: find this with heap pprof, then add an eviction policy.
func allocHandler(w http.ResponseWriter, r *http.Request) {
	chunk := make([]byte, 1<<20) // 1 MB
	leakyCache.mu.Lock()
	leakyCache.data = append(leakyCache.data, chunk)
	size := len(leakyCache.data)
	leakyCache.mu.Unlock()
	fmt.Fprintf(w, "cache size: %d MB\n", size)
}

// ---- Problem 3: Goroutine leak ----

// startLeakyWorker launches a goroutine that blocks on an unbuffered channel
// that nobody ever closes or sends to. Every call adds one more stuck goroutine.
// TODO: find this with goroutine pprof, then fix by passing a context.
func startLeakyWorker() {
	ch := make(chan struct{}) // nobody ever sends here
	go func() {
		<-ch // blocks forever
		log.Println("worker done") // never reached
	}()
}

func leakHandler(w http.ResponseWriter, r *http.Request) {
	startLeakyWorker()
	fmt.Fprintln(w, "started a worker (check /debug/pprof/goroutine)")
}

// ---- Stats endpoint ----

func statsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: report runtime.NumGoroutine(), runtime.MemStats here
	fmt.Fprintln(w, "TODO: implement runtime stats")
}

func main() {
	http.HandleFunc("/work", workHandler)
	http.HandleFunc("/alloc", allocHandler)
	http.HandleFunc("/leak", leakHandler)
	http.HandleFunc("/stats", statsHandler)

	// /debug/pprof/* routes are registered automatically by the pprof import above
	log.Println("pprof server listening on :6060")
	log.Println("  /work  — CPU hot path")
	log.Println("  /alloc — memory leak")
	log.Println("  /leak  — goroutine leak")
	log.Println("  /debug/pprof/ — pprof dashboard")

	// Warm up: trigger the leak once at startup to make it visible immediately
	go func() {
		time.Sleep(time.Second)
		startLeakyWorker()
	}()

	log.Fatal(http.ListenAndServe(":6060", nil))
}
