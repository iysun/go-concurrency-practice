package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// 客户端负责维护链接, send 接收来自 room 的广播
// 然后将接收到的message通过 Fprintf 往 conn 中写
type Client struct {
	nickname string
	send     chan string
	conn     net.Conn
}

func (c *Client) String() string {
	return c.nickname
}

// room 维护所有的 client 集合
// 监听 leave join message 事件
// listQuery 用于解决 clients 读写冲突问题
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

// room 核心 run 函数
// 通过 leave join message 三个 channal 监听所有客户端状态和信息
// listQuery 分支通过将遍历读取 clients 逻辑放在修改 clients 逻辑的同一个 select 中解决读写冲突问题
func (r *Room) Run() {
	for {
		select {
		case c := <-r.join:
			r.clients[c] = struct{}{}
			r.broadcast(fmt.Sprintf("*** %s joined ***", c.nickname))
		case c := <-r.leave:
			delete(r.clients, c)
			close(c.send)
			r.broadcast(fmt.Sprintf("*** %s leaved ***", c.nickname))
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

// 遍历 clients 向所有 client 的 send 中发送消息
func (r *Room) broadcast(msg string) {
	for c := range r.clients {
		select {
		case c.send <- msg:
		default:
		}

	}
}

// client 接收到消息后想入 conn
func writePump(c *Client) {
	for msg := range c.send {
		_, err := fmt.Fprintln(c.conn, msg)
		if err != nil {
			return
		}
	}
}

// 接收 client conn 的输入，基于输入进行不同的处理
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
		} else {
			room.message <- fmt.Sprintf("%s: %s", c.nickname, line)
		}
	}
}

// 首次链接 注册 client
func handleConn(conn net.Conn, room *Room) {
	fmt.Fprint(conn, "Enter nickname: ")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	nick := strings.TrimSpace(scanner.Text())
	if nick == "" {
		nick = conn.RemoteAddr().String()
	}

	c := &Client{
		nickname: nick,
		conn:     conn,
		send:     make(chan string, 32),
	}

	room.join <- c
	go writePump(c)
	readPump(c, room)
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
			log.Println("err: ", err)
			continue
		}
		go handleConn(conn, room)
	}

}
