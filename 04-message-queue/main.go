package main

import (
	"fmt"
	"sync"
	"time"
)

// Message is the unit passed between publisher and subscribers.
type Message struct {
	Topic   string
	Payload any
}

// Subscriber wraps a channel that receives messages for its subscribed topics.
type Subscriber struct {
	id     int
	ch     chan Message
	topics map[string]struct{}
}

func (s *Subscriber) Receive() <-chan Message {
	return s.ch
}

// Broker routes messages from publishers to matching subscribers (fan-out).
type Broker struct {
	mu          sync.RWMutex
	subscribers map[int]*Subscriber
	nextID      int
}

func NewBroker() *Broker {
	return &Broker{subscribers: make(map[int]*Subscriber)}
}

// Subscribe registers a new subscriber for the given topics.
// Returns the Subscriber so the caller can read from it.
// TODO: implement
func (b *Broker) Subscribe(bufSize int, topics ...string) *Subscriber {
	b.mu.Lock()
	defer b.mu.Unlock()
	// hint: create Subscriber, add to b.subscribers map
	_ = bufSize
	return nil
}

// Unsubscribe removes a subscriber and closes its channel.
// TODO: implement
func (b *Broker) Unsubscribe(sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// hint: delete from map, close sub.ch
}

// Publish sends msg to every subscriber whose topics include msg.Topic.
// Non-blocking: if a subscriber's channel is full, the message is dropped for that subscriber.
// TODO: implement
func (b *Broker) Publish(msg Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sub := range b.subscribers {
		if _, ok := sub.topics[msg.Topic]; ok {
			select {
			case sub.ch <- msg:
			default:
				// drop — subscriber too slow
			}
		}
	}
}

// Close shuts down the broker and all subscriber channels.
// TODO: implement
func (b *Broker) Close() {}

func main() {
	broker := NewBroker()

	// Two subscribers on overlapping topics
	sub1 := broker.Subscribe(16, "news", "sports")
	sub2 := broker.Subscribe(16, "sports")

	if sub1 == nil || sub2 == nil {
		fmt.Println("Subscribe not yet implemented")
		return
	}

	var wg sync.WaitGroup

	consume := func(name string, sub *Subscriber) {
		defer wg.Done()
		for msg := range sub.Receive() {
			fmt.Printf("[%s] topic=%s payload=%v\n", name, msg.Topic, msg.Payload)
		}
	}

	wg.Add(2)
	go consume("sub1", sub1)
	go consume("sub2", sub2)

	broker.Publish(Message{Topic: "news", Payload: "Go 1.24 released"})
	broker.Publish(Message{Topic: "sports", Payload: "World Cup 2026"})
	broker.Publish(Message{Topic: "news", Payload: "New concurrency features"})

	time.Sleep(100 * time.Millisecond)
	broker.Close()
	wg.Wait()
}
