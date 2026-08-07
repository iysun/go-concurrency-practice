package main

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestStoreSetGetDeleteKeysAndTTL(t *testing.T) {
	store := NewStore()

	store.Set("name", "alice", 0)
	got, ok := store.Get("name")
	if !ok {
		t.Fatal("Get(name) returned ok=false")
	}
	if got != "alice" {
		t.Fatalf("Get(name) = %q, want %q", got, "alice")
	}

	store.Set("temp", "value", time.Nanosecond)
	time.Sleep(time.Millisecond)
	if got, ok := store.Get("temp"); ok {
		t.Fatalf("Get(temp) = %q, want missing after TTL expiry", got)
	}

	keys := sortedKeys(store.Keys())
	if want := []string{"name"}; !reflect.DeepEqual(keys, want) {
		t.Fatalf("Keys() = %v, want %v", keys, want)
	}

	if !store.Delete("name") {
		t.Fatal("Delete(name) returned false, want true")
	}
	if store.Delete("name") {
		t.Fatal("Delete(name) returned true for a missing key")
	}
}

func TestStoreEvictRemovesExpiredEntries(t *testing.T) {
	store := NewStore()
	store.Set("keep", "value", 0)
	store.Set("expire", "value", time.Nanosecond)

	time.Sleep(time.Millisecond)
	removed := store.evict()
	if removed != 1 {
		t.Fatalf("evict() removed %d entries, want 1", removed)
	}

	keys := sortedKeys(store.Keys())
	if want := []string{"keep"}; !reflect.DeepEqual(keys, want) {
		t.Fatalf("Keys() after evict = %v, want %v", keys, want)
	}
}

func TestServerHandleConnSetGetAndErrors(t *testing.T) {
	client, reader, cleanup := startTestConn(t, NewServer(":0"))
	defer cleanup()

	assertExchange(t, client, reader, "SET name alice", "OK")
	assertExchange(t, client, reader, "GET name", "alice")
	assertExchange(t, client, reader, "GET missing", "NIL")
	assertExchange(t, client, reader, "GET", "ERR usage: GET key")
	assertExchange(t, client, reader, "SET onlykey", "ERR usage: SET key value [ttl_seconds]")
	assertExchange(t, client, reader, "NOPE", `ERR unknown command "NOPE"`)
}

func TestServerHandleConnUnimplementedCommands(t *testing.T) {
	client, reader, cleanup := startTestConn(t, NewServer(":0"))
	defer cleanup()

	assertExchange(t, client, reader, "DEL name", "ERR not implemented")
	assertExchange(t, client, reader, "KEYS", "ERR not implemented")
}

func startTestConn(t *testing.T, srv *Server) (net.Conn, *bufio.Reader, func()) {
	t.Helper()

	serverConn, clientConn := net.Pipe()
	done := make(chan struct{})
	go func() {
		srv.handleConn(serverConn)
		close(done)
	}()

	cleanup := func() {
		if err := clientConn.Close(); err != nil {
			t.Errorf("closing client connection: %v", err)
		}
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("server handler did not exit")
		}
	}

	return clientConn, bufio.NewReader(clientConn), cleanup
}

func assertExchange(t *testing.T, conn net.Conn, reader *bufio.Reader, command, want string) {
	t.Helper()

	if _, err := fmt.Fprintln(conn, command); err != nil {
		t.Fatalf("writing command %q: %v", command, err)
	}

	got, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading response for %q: %v", command, err)
	}
	got = strings.TrimRight(got, "\r\n")
	if got != want {
		t.Fatalf("response for %q = %q, want %q", command, got, want)
	}
}

func sortedKeys(keys []string) []string {
	sort.Strings(keys)
	return keys
}
