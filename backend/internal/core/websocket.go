package core

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for the prototype
		return true
	},
}

type Hub struct {
	mu          sync.Mutex
	connections map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket connection: %v\n", err)
		return
	}

	h.mu.Lock()
	h.connections[conn] = true
	h.mu.Unlock()

	log.Printf("New WebSocket client connected. Total clients: %d\n", len(h.connections))

	// Keep connection open and read messages (discard them as we only broadcast downstream)
	go func() {
		defer func() {
			conn.Close()
			h.mu.Lock()
			delete(h.connections, conn)
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d\n", len(h.connections))
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

func (h *Hub) Broadcast(msg interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Printf("Broadcasting WebSocket message to %d clients\n", len(h.connections))
	for conn := range h.connections {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Failed to write websocket message: %v\n", err)
			conn.Close()
			delete(h.connections, conn)
		}
	}
}
