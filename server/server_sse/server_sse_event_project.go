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
	mu       sync.Mutex
}

// ProjectHub manages all project SSE clients
type ProjectHub struct {
	clients    map[string]*ProjectClient
	register   chan *ProjectClient
	unregister chan *ProjectClient
	broadcast  chan ProjectChange
	shutdown   chan struct{}
	mu         sync.RWMutex
	running    bool
}

func NewProjectHub() *ProjectHub {
	return &ProjectHub{
		clients:    make(map[string]*ProjectClient),
		register:   make(chan *ProjectClient),
		unregister: make(chan *ProjectClient),
		broadcast:  make(chan ProjectChange, 100),
		shutdown:   make(chan struct{}),
		running:    false,
	}
}

var SSE_ProjectHub = NewProjectHub()

// Run starts the project hub with cleanup
func (h *ProjectHub) Run() {
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()

	ticker := time.NewTicker(SSEPingInterval)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if !h.running {
				close(client.Chan)
				client.Cancel()
				h.mu.Unlock()
				continue
			}
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
					log.Printf("Project client %s is slow, skipping message", client.ID)
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.mu.Lock()
			if !h.running {
				h.mu.Unlock()
				continue
			}
			now := time.Now()
			for id, client := range h.clients {
				client.mu.Lock()
				lastPing := client.LastPing
				client.mu.Unlock()

				if now.Sub(lastPing) > SSEClientTimeout {
					log.Printf("Project client %s timed out, removing", id)
					delete(h.clients, id)
					close(client.Chan)
					client.Cancel()
					continue
				}

				select {
				case client.Chan <- ProjectChange{
					Type:      EventPing,
					Timestamp: now,
				}:
				default:
					log.Printf("Project client %s channel full, removing", id)
					delete(h.clients, id)
					close(client.Chan)
					client.Cancel()
				}
			}
			h.mu.Unlock()

		case <-h.shutdown:
			h.mu.Lock()
			log.Println("Project hub shutting down...")
			for id, client := range h.clients {
				close(client.Chan)
				client.Cancel()
				delete(h.clients, id)
			}
			h.running = false
			h.mu.Unlock()
			log.Println("Project hub stopped")
			return
		}
	}
}

// Close gracefully shuts down the hub
func (h *ProjectHub) Close() {
	h.mu.RLock()
	isRunning := h.running
	h.mu.RUnlock()

	if isRunning {
		close(h.shutdown)
		time.Sleep(100 * time.Millisecond) // Give it time to process
	}
}

// BroadcastProjectChange sends a project change to all clients
func BroadcastProjectChange(changeType EventType, projectData generated.ProjectList) {
	SSE_ProjectHub.mu.RLock()
	isRunning := SSE_ProjectHub.running
	SSE_ProjectHub.mu.RUnlock()

	if !isRunning {
		return
	}

	select {
	case SSE_ProjectHub.broadcast <- ProjectChange{
		Type:      changeType,
		Timestamp: time.Now(),
		Data:      projectData,
	}:
	default:
		log.Println("Project broadcast channel full, skipping")
	}
}

// handleProjectSSE handles SSE connections for projects
func handleProjectSSE(c *fiber.Ctx) error {
	// Check if hub is running
	SSE_ProjectHub.mu.RLock()
	running := SSE_ProjectHub.running
	SSE_ProjectHub.mu.RUnlock()

	if !running {
		return c.Status(503).SendString("SSE service shutting down")
	}

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
			cancel()
		}()

		// Send initial connection message
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

				if change.Type == EventPing {
					fmt.Fprintf(w, "event: ping\n")
					fmt.Fprintf(w, "data: {}\n\n")
				} else {
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

				client.mu.Lock()
				client.LastPing = time.Now()
				client.mu.Unlock()
			}
		}
	})

	return nil
}