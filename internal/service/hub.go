package service

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	UserID uint
	Send   chan []byte
	Done   chan struct{}
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[uint][]*Client
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uint][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			h.removeClient(client)
			close(client.Done)
			h.mu.Unlock()
		}
	}
}

func (h *Hub) SendToUser(userID uint, msg []byte) {
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()
	for _, client := range clients {
		select {
		case client.Send <- msg:
		default:
			log.Printf("WebSocket send buffer full for user %d, skipping", userID)
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	clients := h.clients[client.UserID]
	for i, c := range clients {
		if c == client {
			h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	if len(h.clients[client.UserID]) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Conn.Close()
		c.Hub.Unregister <- c
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.Done:
			return
		}
	}
}
