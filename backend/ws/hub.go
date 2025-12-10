package ws

import (
    "sync"
)

type Hub struct {
    clients map[*Client]bool
    broadcast chan []byte
    register chan *Client
    unregister chan *Client
    mu sync.Mutex
}

func NewHub() *Hub {
    return &Hub{
        clients: make(map[*Client]bool),
        broadcast: make(chan []byte),
        register: make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case c := <-h.register:
            h.mu.Lock()
            h.clients[c] = true
            h.mu.Unlock()
        case c := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[c]; ok {
                delete(h.clients, c)
                close(c.send)
            }
            h.mu.Unlock()
        case message := <-h.broadcast:
            h.mu.Lock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.Unlock()
        }
    }
}

func (h *Hub) Broadcast(message []byte) {
    h.broadcast <- message
}
