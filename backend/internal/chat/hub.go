// Package chat — WebSocket-чат (Global / Alliance / PM).
//
// Hub — in-memory broadcast-сервер. Каждый подключённый клиент
// регистрирует себя на один channel (строку). Hub рассылает
// входящие сообщения всем клиентам этого channel.
package chat

import (
	"context"
	"sync"
)

// Message — сообщение в чате (отдано клиенту как JSON).
type Message struct {
	ID        string `json:"id"`
	Channel   string `json:"channel"`
	AuthorID  string `json:"author_id"`
	AuthorName string `json:"author_name"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type client struct {
	channel string
	send    chan Message
}

// Hub управляет подключёнными клиентами и рассылкой сообщений.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*client]struct{})}
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	close(c.send)
}

// Broadcast отправляет сообщение всем клиентам канала.
func (h *Hub) Broadcast(ctx context.Context, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.channel != msg.Channel {
			continue
		}
		select {
		case c.send <- msg:
		default:
			// медленный клиент — пропускаем, не блокируем broadcast
		}
	}
}
