package core

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/Knightshrestha/Secret-Injector/server/server_sse"
	"github.com/gofiber/fiber/v2"
)

func Shutdown(db database.CustomDB, app *fiber.App) {
	log.Println("Starting graceful shutdown...")

	log.Println("Closing SSE hubs...")
	server_sse.SSE_ProjectHub.Close()
	server_sse.SSE_SecretHub.Close()
	log.Println("✓ SSE hubs closed")

	time.Sleep(500 * time.Millisecond)

	if app != nil {
		log.Println("Shutting down Fiber server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			if err == context.DeadlineExceeded {
				log.Println("⚠ Fiber shutdown timeout - forcing close")
			} else {
				log.Println("⚠ Fiber shutdown error:", err)
			}
		} else {
			log.Println("✓ Fiber server stopped gracefully")
		}
	}

	// Close database connections
	if db.WriteDB != nil || db.ReadDB != nil {
		log.Println("Closing database connections...")
		if err := database.CloseDatabase(db); err != nil {
			log.Println("Error closing database:", err)
		} else {
			log.Println("✓ Database connections closed successfully")
		}
	}

	log.Println("✓ Shutdown complete")
	os.Exit(0)
}