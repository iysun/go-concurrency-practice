package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// entry wraps a stored value with an optional expiry.
type entry struct {
	value     string
	expiresAt time.Time // zero means no expiry
}

func (e entry) expired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

// Store is a thread-safe in-memory key-value store.
type Store struct {
	mu   sync.RWMutex
	data map[string]entry
}

func NewStore() *Store {
	return &Store{data: make(map[string]entry)}
}

// Get returns the value for key. Second return is false if missing or expired.
// TODO: implement
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.data[key]
	if !ok || e.expired() {
		return "", false
	}
	return e.value, true
}

// Set stores key=value with an optional TTL (0 = no expiry).
// TODO: implement
func (s *Store) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := entry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	s.data[key] = e
}

// Delete removes a key. Returns true if the key existed.
// TODO: implement
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}

// Keys returns all non-expired keys.
// TODO: implement
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var keys []string
	for k, e := range s.data {
		if !e.expired() {
			keys = append(keys, k)
		}
	}
	return keys
}

// evict removes expired entries. Called by the background janitor.
func (s *Store) evict() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	var removed int
	for k, e := range s.data {
		if e.expired() {
			delete(s.data, k)
			removed++
		}
	}
	return removed
}

// Server accepts TCP connections and speaks a simple line protocol:
//
//	SET key value [ttl_seconds]
//	GET key
//	DEL key
//	KEYS
type Server struct {
	store *Store
	addr  string
}

func NewServer(addr string) *Server {
	return &Server{store: NewStore(), addr: addr}
}

// handleConn parses commands from a single client connection.
// TODO: implement full command parsing
func (srv *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}
		cmd := strings.ToUpper(parts[0])
		switch cmd {
		case "GET":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR usage: GET key")
				continue
			}
			val, ok := srv.store.Get(parts[1])
			if !ok {
				fmt.Fprintln(conn, "NIL")
			} else {
				fmt.Fprintln(conn, val)
			}
		case "SET":
			// TODO: parse optional TTL argument
			if len(parts) < 3 {
				fmt.Fprintln(conn, "ERR usage: SET key value [ttl_seconds]")
				continue
			}
			srv.store.Set(parts[1], parts[2], 0)
			fmt.Fprintln(conn, "OK")
		case "DEL":
			// TODO: implement
			fmt.Fprintln(conn, "ERR not implemented")
		case "KEYS":
			// TODO: implement
			fmt.Fprintln(conn, "ERR not implemented")
		default:
			fmt.Fprintf(conn, "ERR unknown command %q\n", cmd)
		}
	}
}

func (srv *Server) Start() error {
	ln, err := net.Listen("tcp", srv.addr)
	if err != nil {
		return err
	}
	log.Printf("kv-store listening on %s", srv.addr)

	// Background janitor: evict expired keys every 5 seconds
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			n := srv.store.evict()
			if n > 0 {
				log.Printf("janitor: evicted %d expired keys", n)
			}
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go srv.handleConn(conn)
	}
}

func main() {
	srv := NewServer(":6380")
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
