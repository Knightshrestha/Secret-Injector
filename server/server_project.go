package server

import (
	"database/sql"
	"log"
	"strings"

	"github.com/Knightshrestha/Secret-Injector/database/generated"
	"github.com/Knightshrestha/Secret-Injector/server/server_sse"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterReadOnlyProjectRoute(router fiber.Router, readOnlyDatabase *generated.Queries) {
	router.Get("/projects", func(c *fiber.Ctx) error {
		allProjects, err := readOnlyDatabase.GetAllProjects(c.Context())
		if err != nil {
			log.Printf("Error fetching projects: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch projects",
			})
		}
		return c.JSON(allProjects)
	})

	router.Get("/projects/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if id == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project name cannot be empty",
			})
		}

		project, err := readOnlyDatabase.GetProjectByID(c.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Project not found",
				})
			}
			log.Printf("Failed to fetch project %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch project",
			})
		}
		return c.JSON(project)
	})
}

func RegisterWriteProjectRoute(
	router fiber.Router,
	readWriteDatabase *sql.DB,
	readWriteQueries *generated.Queries,
) {
	router.Post("/projects", func(c *fiber.Ctx) error {
		var body struct {
			Name        string  `json:"name"`
			Description *string `json:"description"`
		}
		if err := c.BodyParser(&body); err != nil {
			log.Printf("Body parse error: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validation
		if strings.TrimSpace(body.Name) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project name is required",
			})
		}

		id := uuid.New().String()

		newProject := generated.CreateProjectParams{
			ID:          id,
			Name:        body.Name,
			Description: body.Description,
		}

		project, err := readWriteQueries.CreateProject(c.Context(), newProject)
		if err != nil {

			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Project with this name already exists",
				})
			}
			log.Printf("Failed to create project %s: %v", id, err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create project",
			})
		}

		server_sse.BroadcastProjectChange(server_sse.EventCreate, project)

		return c.Status(fiber.StatusCreated).JSON(project)

	})

	router.Patch("/projects/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if id == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID is required",
			})
		}

		var body struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := c.BodyParser(&body); err != nil {
			log.Printf("Body parse error: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validation - at least one field should be provided
		if body.Name == nil && body.Description == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "At least one field (name or description) must be provided",
			})
		}

		// Validate non-empty if provided
		if body.Name != nil && strings.TrimSpace(*body.Name) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret name cannot be empty",
			})
		}
		if body.Description != nil && strings.TrimSpace(*body.Description) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret value cannot be empty",
			})
		}

		updatedProject := generated.UpdateProjectParams{
			ID:          id,
			Name:        body.Name,
			Description: body.Description,
		}

		project, err := readWriteQueries.UpdateProject(c.Context(), updatedProject)
		if err != nil {
			log.Printf("Failed to update project %s: %v", id, err)

			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Project with this name already exists",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update project",
			})
		}

		server_sse.BroadcastProjectChange(server_sse.EventUpdate, project)

		return c.Status(fiber.StatusOK).JSON(project)
	})

	router.Delete("/projects/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID cannot be empty",
			})
		}

		// Fetch project to return in SSE
		project, err := readWriteQueries.GetProjectByID(c.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Project not found",
				})
			}
			log.Printf("Failed to fetch project %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch project",
			})
		}

		// Begin transaction
		txn, err := readWriteDatabase.BeginTx(c.Context(), nil)
		if err != nil {
			log.Printf("Failed to begin transaction: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to begin transaction",
			})
		}
		defer txn.Rollback()
		queriesTx := readWriteQueries.WithTx(txn)

		// Delete all secrets in project
		if err := queriesTx.DeleteAllSecretsInProjects(c.Context(), id); err != nil {
			log.Printf("Failed to delete secrets for project %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete secrets",
			})
		}

		// Delete the project
		if err := queriesTx.DeleteProject(c.Context(), id); err != nil {
			log.Printf("Failed to delete project %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete project",
			})
		}

		// Commit transaction
		if err := txn.Commit(); err != nil {
			log.Printf("Failed to commit transaction for project %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to commit transaction",
			})
		}

		// Broadcast SSE event
		server_sse.BroadcastProjectChange(server_sse.EventDelete, project)

		return c.SendStatus(fiber.StatusNoContent)
	})

}
