package server_sse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Knightshrestha/Secret-Injector/database/generated"
	"github.com/gofiber/fiber/v2"
)

// SecretChange represents a secret change event with full data
type SecretChange struct {
	Type      EventType             `json:"type"` // "create", "update", "delete", "ping"
	Timestamp time.Time          `json:"timestamp"`
	Data      generated.SecretList `json:"data"` // Full secret object
}

// SecretClient represents an SSE client for secrets
type SecretClient struct {
	ID       string
	Chan     chan SecretChange
	Context  context.Context
	Cancel   context.CancelFunc
	LastPing time.Time
	mu       sync.Mutex // Protect LastPing
}

// SecretHub manages all secret SSE clients
type SecretHub struct {
	clients    map[string]*SecretClient
	register   chan *SecretClient
	unregister chan *SecretClient
	broadcast  chan SecretChange
	mu         sync.RWMutex // Protect clients map
}

var SSE_SecretHub = &SecretHub{
	clients:    make(map[string]*SecretClient),
	register:   make(chan *SecretClient),
	unregister: make(chan *SecretClient),
	broadcast:  make(chan SecretChange, 100),
}


// BroadcastSecretChange sends a secret change with full data to all clients
func BroadcastSecretChange(changeType EventType, secretData generated.SecretList) {
	SSE_SecretHub.broadcast <- SecretChange{
		Type:      changeType,
		Timestamp: time.Now(),
		Data:      secretData,
	}
}

// Run starts the secret hub with cleanup
func (h *SecretHub) Run() {
	ticker := time.NewTicker(SSEPingInterval)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			client.mu.Lock()
			client.LastPing = time.Now()
			client.mu.Unlock()
			log.Printf("Secret client registered: %s (Total: %d)", client.ID, len(h.clients))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Chan)
				client.Cancel()
				log.Printf("Secret client unregistered: %s (Total: %d)", client.ID, len(h.clients))
			}
			h.mu.Unlock()

		case change := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Chan <- change:
				default:
					log.Printf("Secret client %s is slow, skipping message", client.ID)
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.mu.Lock()
			now := time.Now()
			for id, client := range h.clients {
				client.mu.Lock()
				lastPing := client.LastPing
				client.mu.Unlock()

				if now.Sub(lastPing) > SSEClientTimeout {
					log.Printf("Secret client %s timed out, removing", id)
					delete(h.clients, id)
					close(client.Chan)
					client.Cancel()
					continue
				}

				// Send ping
				select {
				case client.Chan <- SecretChange{
					Type:      EventPing,
					Timestamp: now,
				}:
					// Ping sent successfully
				default:
					// Client channel full, probably dead
					log.Printf("Secret client %s channel full, removing", id)
					delete(h.clients, id)
					close(client.Chan)
					client.Cancel()
				}
			}
			h.mu.Unlock()
		}
	}
}

// handleSecretSSE handles SSE connections for secrets
func handleSecretSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ctx, cancel := context.WithCancel(c.UserContext())

	client := &SecretClient{
		ID:       fmt.Sprintf("secret-%d", time.Now().UnixNano()),
		Chan:     make(chan SecretChange, 10),
		Context:  ctx,
		Cancel:   cancel,
		LastPing: time.Now(),
	}

	SSE_SecretHub.register <- client

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			SSE_SecretHub.unregister <- client
			defer cancel()
		}()

		// Send initial connection message with named event
		initMsg := map[string]string{"status": "connected", "channel": "secret_list"}
		data, _ := json.Marshal(initMsg)
		fmt.Fprintf(w, "event: connected\n")
		fmt.Fprintf(w, "data: %s\n\n", data)
		if err := w.Flush(); err != nil {
			return
		}

		for {
			select {
			case <-ctx.Done():
				return

			case change, ok := <-client.Chan:
				if !ok {
					return
				}

				// Use named events based on change type
				if change.Type == EventPing {
					// Send ping as named event with minimal data
					fmt.Fprintf(w, "event: ping\n")
					fmt.Fprintf(w, "data: {}\n\n")
				} else {
					// Send actual changes with event type
					data, err := json.Marshal(change)
					if err != nil {
						continue
					}
					fmt.Fprintf(w, "event: %s\n", change.Type)
					fmt.Fprintf(w, "data: %s\n\n", data)
				}

				if err := w.Flush(); err != nil {
					return
				}

				// Update LastPing after successful write
				client.mu.Lock()
				client.LastPing = time.Now()
				client.mu.Unlock()
			}
		}
	})

	return nil
}
