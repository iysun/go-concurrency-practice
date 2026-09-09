package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
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
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}

// Expire sets a key's TTL. Returns false if the key is missing or expired.
func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.data[key]
	if !ok {
		return false
	}
	if e.expired() {
		delete(s.data, key)
		return false
	}
	if ttl == 0 {
		e.expiresAt = time.Now().Add(-time.Nanosecond)
	} else {
		e.expiresAt = time.Now().Add(ttl)
	}
	s.data[key] = e
	return true
}

// Exists reports whether a key exists and has not expired.
func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.data[key]
	return ok && !e.expired()
}

// Keys returns all non-expired keys.
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
//	EXPIRE key seconds
//	EXISTS key
type Server struct {
	store *Store
	addr  string
}

func NewServer(addr string) *Server {
	return &Server{store: NewStore(), addr: addr}
}

// handleConn parses commands from a single client connection.
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
			if len(parts) != 3 && len(parts) != 4 {
				fmt.Fprintln(conn, "ERR usage: SET key value [ttl_seconds]")
				continue
			}
			var ttl time.Duration
			if len(parts) == 4 {
				var err error
				ttl, err = parseTTLSeconds(parts[3])
				if err != nil {
					fmt.Fprintln(conn, "ERR invalid ttl")
					continue
				}
			}
			srv.store.Set(parts[1], parts[2], ttl)
			fmt.Fprintln(conn, "OK")
		case "DEL":
			if len(parts) != 2 {
				fmt.Fprintln(conn, "ERR usage: DEL key")
				continue
			}
			srv.store.Delete(parts[1])
			fmt.Fprintln(conn, "OK")
		case "KEYS":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR usage: KEYS")
				continue
			}
			keys := srv.store.Keys()
			sort.Strings(keys)
			fmt.Fprintln(conn, strings.Join(keys, " "))
		case "EXPIRE":
			if len(parts) != 3 {
				fmt.Fprintln(conn, "ERR usage: EXPIRE key seconds")
				continue
			}
			seconds, err := parseTTLSeconds(parts[2])
			if err != nil {
				fmt.Fprintln(conn, "ERR invalid ttl")
				continue
			}
			if srv.store.Expire(parts[1], seconds) {
				fmt.Fprintln(conn, "OK")
			} else {
				fmt.Fprintln(conn, "NIL")
			}
		case "EXISTS":
			if len(parts) != 2 {
				fmt.Fprintln(conn, "ERR usage: EXISTS key")
				continue
			}
			if srv.store.Exists(parts[1]) {
				fmt.Fprintln(conn, "1")
			} else {
				fmt.Fprintln(conn, "0")
			}
		default:
			fmt.Fprintf(conn, "ERR unknown command %q\n", cmd)
		}
	}
}

func parseTTLSeconds(raw string) (time.Duration, error) {
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("invalid ttl")
	}
	if seconds > int64((1<<63-1)/int64(time.Second)) {
		return 0, fmt.Errorf("invalid ttl")
	}
	return time.Duration(seconds) * time.Second, nil
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
		defer ticker.Stop()
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
