package server

import (
	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/Knightshrestha/Secret-Injector/server/server_sse"
	"github.com/gofiber/fiber/v2"
)

func RegisterApiRoutes(app *fiber.App, customDb database.CustomDB) {
	apiGroup := app.Group("/api")
	RegisterReadOnlyProjectRoute(apiGroup, customDb.ReadQueries)
	RegisterWriteProjectRoute(apiGroup, customDb.WriteDB, customDb.WriteQueries)

	RegisterReadOnlySecretRoute(apiGroup, customDb.ReadQueries)
	RegisterWriteSecretRoute(apiGroup, customDb.WriteQueries)

	sseGroup := app.Group("/events")
	server_sse.RegisterSSERoutes(sseGroup)
}
