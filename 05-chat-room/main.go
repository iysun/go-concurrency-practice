package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// Client represents a connected chat user.
type Client struct {
	conn     net.Conn
	nickname string
	send     chan string // outgoing messages for this client
}

func (c *Client) String() string {
	return c.nickname
}

// Room manages all connected clients and broadcasts messages.
type Room struct {
	clients   map[*Client]struct{}
	join      chan *Client
	leave     chan *Client
	message   chan string
	listQuery chan chan []string
}

func NewRoom() *Room {
	return &Room{
		clients:   make(map[*Client]struct{}),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		message:   make(chan string, 64),
		listQuery: make(chan chan []string),
	}
}

// Run is the Room's event loop — the single goroutine that mutates r.clients.
// All join/leave/broadcast actions pass through this loop to avoid data races.
func (r *Room) Run() {
	for {
		select {
		case c := <-r.join:
			r.clients[c] = struct{}{}
			r.broadcast(fmt.Sprintf("*** %s joined ***", c.nickname))
		case c := <-r.leave:
			delete(r.clients, c)
			close(c.send)
		case msg := <-r.message:
			r.broadcast(msg)
		case ch := <-r.listQuery:
			var names []string
			for c := range r.clients {
				names = append(names, c.nickname)
			}
			ch <- names
		}
	}
}

// broadcast sends msg to every connected client's send channel.
// Must only be called from within Run().
func (r *Room) broadcast(msg string) {
	for c := range r.clients {
		select {
		case c.send <- msg:
		default:
			// slow client — drop message
		}
	}
}

func writePump(c *Client) {
	for msg := range c.send {
		_, err := fmt.Fprintln(c.conn, msg)
		if err != nil {
			return
		}
	}
}

// readPump reads lines from the TCP connection and forwards them to the room.
// Signals departure on disconnect.
func readPump(c *Client, room *Room) {
	defer func() {
		room.leave <- c
		c.conn.Close()
	}()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		} else if line == "/quit" {
			return
		} else if line == "/list" {
			ch := make(chan []string)
			room.listQuery <- ch
			names := <-ch
			fmt.Fprintf(c.conn, "Online users (%d):\n", len(names))
			for _, name := range names {
				fmt.Fprintf(c.conn, "  - %s\n", name)
			}
			continue
		}
		room.message <- fmt.Sprintf("%s: %s", c.nickname, line)
	}
}

// handleConn negotiates nickname then starts read/write pumps.
func handleConn(conn net.Conn, room *Room) {
	fmt.Fprint(conn, "Enter nickname: ")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	nick := strings.TrimSpace(scanner.Text())
	if nick == "" {
		nick = conn.RemoteAddr().String()
	}

	c := &Client{
		conn:     conn,
		nickname: nick,
		send:     make(chan string, 32),
	}

	room.join <- c
	go writePump(c)
	readPump(c, room) // blocks until disconnect
}

func main() {
	room := NewRoom()
	go room.Run()

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Chat room listening on :9000")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConn(conn, room)
	}
}
