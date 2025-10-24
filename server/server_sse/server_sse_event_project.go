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

// ProjectChange represents a project change event with full data
type ProjectChange struct {
	Type      EventType             `json:"type"` // "create", "update", "delete", "ping"
	Timestamp time.Time             `json:"timestamp"`
	Data      generated.ProjectList `json:"data"` // Full project object
}

// ProjectClient represents an SSE client for projects
type ProjectClient struct {
	ID       string
	Chan     chan ProjectChange
	Context  context.Context
	Cancel   context.CancelFunc
	LastPing time.Time
	mu       sync.Mutex // Protect LastPing
}

// ProjectHub manages all project SSE clients
type ProjectHub struct {
	clients    map[string]*ProjectClient
	register   chan *ProjectClient
	unregister chan *ProjectClient
	broadcast  chan ProjectChange
	mu         sync.RWMutex // Protect clients map
}

var SSE_ProjectHub = &ProjectHub{
	clients:    make(map[string]*ProjectClient),
	register:   make(chan *ProjectClient),
	unregister: make(chan *ProjectClient),
	broadcast:  make(chan ProjectChange, 100),
}

// Run starts the project hub with cleanup
func (h *ProjectHub) Run() {
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
			log.Printf("Project client registered: %s (Total: %d)", client.ID, len(h.clients))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Chan)
				client.Cancel()
				log.Printf("Project client unregistered: %s (Total: %d)", client.ID, len(h.clients))
			}
			h.mu.Unlock()

		case change := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Chan <- change:
					// Successfully sent
				default:
					// Client too slow, skip this message
					log.Printf("Project client %s is slow, skipping message", client.ID)
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

				// Check if client hasn't responded in SSEClientTimeout
				if now.Sub(lastPing) > SSEClientTimeout {
					log.Printf("Project client %s timed out, removing", id)
					delete(h.clients, id)
					close(client.Chan)
					client.Cancel()
					continue
				}

				// Send ping
				select {
				case client.Chan <- ProjectChange{
					Type:      EventPing,
					Timestamp: now,
				}:
					// Ping sent successfully
				default:
					// Client channel full, probably dead
					log.Printf("Project client %s channel full, removing", id)
					delete(h.clients, id)
					close(client.Chan)
					client.Cancel()
				}
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastProjectChange sends a project change with full data to all clients
func BroadcastProjectChange(changeType EventType, projectData generated.ProjectList) {
	SSE_ProjectHub.broadcast <- ProjectChange{
		Type:      changeType,
		Timestamp: time.Now(),
		Data:      projectData,
	}
}

// handleProjectSSE handles SSE connections for projects
func handleProjectSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ctx, cancel := context.WithCancel(c.UserContext())

	client := &ProjectClient{
		ID:       fmt.Sprintf("project-%d", time.Now().UnixNano()),
		Chan:     make(chan ProjectChange, 10),
		Context:  ctx,
		Cancel:   cancel,
		LastPing: time.Now(),
	}

	SSE_ProjectHub.register <- client

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			SSE_ProjectHub.unregister <- client
			defer cancel()
		}()

		// Send initial connection message with named event
		initMsg := map[string]string{"status": "connected", "channel": "project_list"}
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
