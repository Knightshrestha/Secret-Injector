package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/Knightshrestha/Secret-Injector/server"
	"github.com/Knightshrestha/Secret-Injector/server/server_sse"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
)

func StartServer(port int, logging bool) {
	log.Println("Starting Novel Server...")

	// Open DB
	mainDb := database.OpenDatabase()

	// Start SSE hub
	go server_sse.SSE_ProjectHub.Run()
	go server_sse.SSE_SecretHub.Run()

	log.Println("SSE Hub started")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
	})

	// Fiber Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", // Vite's default port
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Use(compress.New())
	app.Use(healthcheck.New(healthcheck.Config{
		LivenessEndpoint:  "/live",
		ReadinessEndpoint: "/ready",
	}))

	EmbedWebsite(app)

	// Routes
	server.RegisterApiRoutes(app, mainDb)

	// Setup graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(
			fmt.Sprintf(":%d", port),
		); err != nil {
			log.Fatal("Fiber server error:", err)
		}
	}()

	if !logging {
		log.SetOutput(io.Discard)
	}

	// Wait for shutdown signal
	<-shutdownChan
	if !logging {
		log.SetOutput(os.Stdout)
	}

	log.Println("\nReceived shutdown signal. Gracefully stopping...")

	// Perform graceful shutdown
	Shutdown(mainDb, app)
}
