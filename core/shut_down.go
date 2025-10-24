package core

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/gofiber/fiber/v2"
)

func Shutdown(db database.CustomDB, app *fiber.App) {
	log.Println("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Fiber
	if app != nil {
		log.Println("Shutting down Fiber server...")
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Println("Fiber shutdown error:", err)
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