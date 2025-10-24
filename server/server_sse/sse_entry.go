package server_sse

import (
	"time"
	"github.com/gofiber/fiber/v2"
)

type EventType string

const (
	EventCreate EventType = "create"
	EventUpdate EventType = "update"
	EventDelete EventType = "delete"
	EventPing   EventType = "ping"
)

// Constant Time
const (
	SSEPingInterval  = 10 * time.Second // how often to send pings
	SSEClientTimeout = 20 * time.Second // client inactivity timeout
)

// RegisterSSERoutes adds SSE endpoints to your app
func RegisterSSERoutes(router fiber.Router) {
	router.Get("/projects", handleProjectSSE)
	router.Get("/secrets", handleSecretSSE)
}
