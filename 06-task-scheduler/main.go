package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// TaskFunc is the function signature for a scheduled task.
type TaskFunc func(ctx context.Context)

// Task holds the configuration for a scheduled job.
type Task struct {
	Name     string
	Interval time.Duration
	Fn       TaskFunc
	// TODO: add fields for LastRun, RunCount, MaxRetries
}

// Scheduler runs registered tasks on their configured intervals.
// Each task fires in its own goroutine so a slow task doesn't block others.
type Scheduler struct {
	mu      sync.RWMutex
	tasks   map[string]*Task
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewScheduler() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		tasks:  make(map[string]*Task),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Register adds a task to the scheduler.
// Returns an error if a task with the same name is already registered.
// TODO: implement
func (s *Scheduler) Register(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[t.Name]; exists {
		return fmt.Errorf("task %q already registered", t.Name)
	}
	s.tasks[t.Name] = t
	return nil
}

// Unregister removes a task by name.
// TODO: implement (hint: stopping a running ticker requires tracking it)
func (s *Scheduler) Unregister(name string) {}

// runTask manages the ticker loop for a single task.
// Fires t.Fn in a new goroutine each tick so the scheduler loop isn't blocked.
// TODO: implement — respect s.ctx cancellation
func (s *Scheduler) runTask(t *Task) {
	defer s.wg.Done()
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case tick := <-ticker.C:
			log.Printf("[scheduler] task=%s tick=%s", t.Name, tick.Format(time.TimeOnly))
			// TODO: run t.Fn in its own goroutine with a per-run context + timeout
			// TODO: track RunCount and handle panics/retries
			go t.Fn(s.ctx)
		}
	}
}

// Start launches goroutines for all registered tasks.
// Calling Start twice is a no-op for already-running tasks.
func (s *Scheduler) Start() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		s.wg.Add(1)
		go s.runTask(t)
	}
}

// Stop signals all task goroutines to exit and waits for them to finish.
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	log.Println("[scheduler] stopped")
}

func main() {
	sched := NewScheduler()

	sched.Register(&Task{
		Name:     "heartbeat",
		Interval: 2 * time.Second,
		Fn: func(ctx context.Context) {
			fmt.Println("heartbeat:", time.Now().Format(time.TimeOnly))
		},
	})

	sched.Register(&Task{
		Name:     "cleanup",
		Interval: 5 * time.Second,
		Fn: func(ctx context.Context) {
			fmt.Println("running cleanup...")
			// TODO: simulate slow cleanup and handle ctx cancellation
		},
	})

	sched.Start()

	// Run for 15 seconds then shut down gracefully
	time.Sleep(15 * time.Second)
	sched.Stop()
}
