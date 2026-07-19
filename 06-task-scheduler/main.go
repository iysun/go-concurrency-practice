package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// TaskFunc is the function signature for a scheduled task.
type TaskFunc func(ctx context.Context)

// Task holds the configuration for a scheduled job.
type Task struct {
	Name     string
	Interval time.Duration
	Fn       TaskFunc
	Timeout  time.Duration

	// 统计字段 —— 会被调度器更新，用 atomic 更安全
	LastRun    atomic.Value
	RunCount   atomic.Int64
	ErrorCount atomic.Int64
}

// 运行实例
type taskRunner struct {
	task   *Task
	cancel context.CancelFunc
	done   chan struct{} // runTask 退出时 close
}

// Scheduler runs registered tasks on their configured intervals.
// Each task fires in its own goroutine so a slow task doesn't block others.
type Scheduler struct {
	mu      sync.Mutex
	runners map[string]*taskRunner
	wg      sync.WaitGroup
	started bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		runners: make(map[string]*taskRunner),
	}
}

// Register adds a task to the scheduler.
// Returns an error if a task with the same name is already registered.
func (s *Scheduler) Register(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.runners[t.Name]; exists {
		return fmt.Errorf("task %q already registered", t.Name)
	}
	ctx, cancel := context.WithCancel(context.Background())
	runner := &taskRunner{
		task:   t,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	s.runners[t.Name] = runner
	if s.started {
		s.wg.Add(1)
		go s.runTask(runner, ctx)
	}
	return nil
}

// Unregister removes a task by name.
func (s *Scheduler) Unregister(name string) error {
	s.mu.Lock()
	runner, exists := s.runners[name]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("task %q not registered", name)
	}
	delete(s.runners, name)
	s.mu.Unlock()

	runner.cancel()
	<-runner.done
	return nil
}

const defaultTaskTimeout = 10 * time.Second

func (s *Scheduler) safeRun(t *Task, parent context.Context) {
	timeout := t.Timeout
	if timeout <= 0 {
		timeout = defaultTaskTimeout
	}
	runCtx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[scheduler] task=%s panic: %v", t.Name, r)
			t.ErrorCount.Add(1)
		}
	}()

	t.Fn(runCtx)

	t.LastRun.Store(time.Now())
	t.RunCount.Add(1)
}

// runTask manages the ticker loop for a single task.
// Fires t.Fn in a new goroutine each tick so the scheduler loop isn't blocked.
func (s *Scheduler) runTask(runner *taskRunner, ctx context.Context) {
	defer s.wg.Done()
	defer close(runner.done)

	t := runner.task

	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("[scheduler] task=%s tick", t.Name)
			go s.safeRun(t, ctx)
		}
	}
}

// Start launches goroutines for all registered tasks.
// Calling Start twice is a no-op for already-running tasks.
func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return
	}
	s.started = true

	for _, runner := range s.runners {
		ctx, cancel := context.WithCancel(context.Background())
		runner.cancel = cancel
		s.wg.Add(1)
		go s.runTask(runner, ctx)
	}
}

// Stop signals all task goroutines to exit and waits for them to finish.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	runners := s.runners
	s.started = false
	s.mu.Unlock()

	for _, runner := range runners {
		runner.cancel()
	}
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
		Timeout:  8 * time.Second, //
		Fn: func(ctx context.Context) {
			fmt.Println("cleanup start:", time.Now().Format(time.TimeOnly))
			// 模拟慢清理：实际需要 3 秒
			select {
			case <-time.After(3 * time.Second):
				fmt.Println("cleanup done")
			case <-ctx.Done():
				// Stop / Unregister / 单次超时 都会走到这里
				fmt.Println("cleanup cancelled:", ctx.Err())
			}
		},
	})

	sched.Start()

	// Run for 15 seconds then shut down gracefully
	time.Sleep(15 * time.Second)
	sched.Stop()
}
