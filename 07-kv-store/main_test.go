package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
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

func TestStoreExpireAndExists(t *testing.T) {
	store := NewStore()
	store.Set("keep", "value", 0)
	store.Set("expire", "value", 0)

	if !store.Exists("keep") {
		t.Fatal("Exists(keep) = false, want true")
	}
	if !store.Expire("expire", time.Nanosecond) {
		t.Fatal("Expire(expire) = false, want true")
	}
	time.Sleep(time.Millisecond)
	if store.Exists("expire") {
		t.Fatal("Exists(expire) = true after expiry, want false")
	}
	if store.Expire("missing", time.Second) {
		t.Fatal("Expire(missing) = true, want false")
	}
	if !store.Expire("keep", 0) {
		t.Fatal("Expire(keep, 0) = false, want true")
	}
	if store.Exists("keep") {
		t.Fatal("Exists(keep) = true after zero TTL, want false")
	}
}

func TestServerHandleConnSetGetAndErrors(t *testing.T) {
	client, reader, cleanup := startTestConn(t, NewServer(":0"))
	defer cleanup()

	assertExchange(t, client, reader, "SET name alice", "OK")
	assertExchange(t, client, reader, "SET temporary value 1", "OK")
	assertExchange(t, client, reader, "GET name", "alice")
	assertExchange(t, client, reader, "GET missing", "NIL")
	assertExchange(t, client, reader, "GET temporary", "value")
	assertExchange(t, client, reader, "GET", "ERR usage: GET key")
	assertExchange(t, client, reader, "SET onlykey", "ERR usage: SET key value [ttl_seconds]")
	assertExchange(t, client, reader, "SET name alice nope", "ERR invalid ttl")
	assertExchange(t, client, reader, "SET name alice -1", "ERR invalid ttl")
	assertExchange(t, client, reader, "SET name alice 0 extra", "ERR usage: SET key value [ttl_seconds]")
	assertExchange(t, client, reader, "EXISTS name", "1")
	assertExchange(t, client, reader, "EXPIRE name 1", "OK")
	assertExchange(t, client, reader, "EXISTS name", "1")
	assertExchange(t, client, reader, "EXPIRE missing 1", "NIL")
	assertExchange(t, client, reader, "EXPIRE name nope", "ERR invalid ttl")
	assertExchange(t, client, reader, "EXPIRE name -1", "ERR invalid ttl")
	assertExchange(t, client, reader, "EXPIRE name", "ERR usage: EXPIRE key seconds")
	assertExchange(t, client, reader, "EXISTS", "ERR usage: EXISTS key")
	assertExchange(t, client, reader, "NOPE", `ERR unknown command "NOPE"`)
}

func TestServerHandleConnDeleteAndKeys(t *testing.T) {
	client, reader, cleanup := startTestConn(t, NewServer(":0"))
	defer cleanup()

	assertExchange(t, client, reader, "SET z last", "OK")
	assertExchange(t, client, reader, "SET a first", "OK")
	assertExchange(t, client, reader, "KEYS", "a z")
	assertExchange(t, client, reader, "DEL a", "OK")
	assertExchange(t, client, reader, "KEYS", "z")
	assertExchange(t, client, reader, "DEL", "ERR usage: DEL key")
	assertExchange(t, client, reader, "KEYS extra", "ERR usage: KEYS")
}

func TestServerAOFAppendAndReplay(t *testing.T) {
	aofPath := filepath.Join(t.TempDir(), "aof.log")
	srv := NewServerWithAOF(":0", aofPath)
	client, reader, cleanup := startTestConn(t, srv)
	defer cleanup()

	assertExchange(t, client, reader, "SET name alice", "OK")
	assertExchange(t, client, reader, "SET temporary value 60", "OK")
	assertExchange(t, client, reader, "EXPIRE name 60", "OK")
	assertExchange(t, client, reader, "DEL temporary", "OK")
	assertExchange(t, client, reader, "DEL missing", "OK")

	logData, err := os.ReadFile(aofPath)
	if err != nil {
		t.Fatalf("reading AOF: %v", err)
	}
	wantLog := "SET name alice\nSET temporary value 60\nEXPIRE name 60\nDEL temporary\n"
	if string(logData) != wantLog {
		t.Fatalf("AOF = %q, want %q", string(logData), wantLog)
	}

	restored := NewServerWithAOF(":0", aofPath)
	if err := restored.replayAOF(); err != nil {
		t.Fatalf("replaying AOF: %v", err)
	}
	if got, ok := restored.store.Get("name"); !ok || got != "alice" {
		t.Fatalf("restored name = %q, %v; want %q, true", got, ok, "alice")
	}
	if restored.store.Exists("temporary") {
		t.Fatal("restored temporary exists after DEL")
	}
}

func TestServerReplayMissingAOF(t *testing.T) {
	srv := NewServerWithAOF(":0", filepath.Join(t.TempDir(), "missing.log"))
	if err := srv.replayAOF(); err != nil {
		t.Fatalf("replaying missing AOF: %v", err)
	}
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
