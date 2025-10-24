package core

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/Knightshrestha/Secret-Injector/frontend"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

func EmbedWebsite(app *fiber.App) {
	distFS, err := fs.Sub(frontend.Website, "build")
	if err != nil {
		log.Fatal(err)
	}

	// Add cache headers middleware
	app.Use("/ui", func(c *fiber.Ctx) error {
		path := c.Path()

		// Aggressive caching for SvelteKit's immutable assets
		if strings.Contains(path, "/_app/immutable/") {
			c.Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasPrefix(path, "/ui/_app/") {
			c.Set("Cache-Control", "public, max-age=86400")
		} else if strings.HasSuffix(path, ".js") ||
			strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".png") ||
			strings.HasSuffix(path, ".jpg") ||
			strings.HasSuffix(path, ".jpeg") ||
			strings.HasSuffix(path, ".svg") ||
			strings.HasSuffix(path, ".woff") ||
			strings.HasSuffix(path, ".woff2") ||
			strings.HasSuffix(path, ".ico") {
			// Static assets outside _app - cache for 1 hour
			c.Set("Cache-Control", "public, max-age=3600")
		} else {
			// Don't cache HTML files (index.html, fallback)
			c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
		}
		return c.Next()
	})

	app.Use("/ui", filesystem.New(filesystem.Config{
		Root:         http.FS(distFS),
		Browse:       false,
		Index:        "index.html",
		NotFoundFile: "index.html",
	}))
}
