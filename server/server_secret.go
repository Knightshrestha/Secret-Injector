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

func RegisterReadOnlySecretRoute(router fiber.Router, readOnlyDatabase *generated.Queries) {
	// Get all secrets
	router.Get("/secrets", func(c *fiber.Ctx) error {
		allSecrets, err := readOnlyDatabase.GetAllSecrets(c.Context())
		if err != nil {
			log.Printf("Error fetching secrets: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch secrets",
			})
		}
		return c.JSON(allSecrets)
	})

	// Get secrets by project ID
	router.Get("/projects/:projectId/secrets", func(c *fiber.Ctx) error {
		projectId := c.Params("projectId")

		if projectId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID cannot be empty",
			})
		}

		secrets, err := readOnlyDatabase.GetSecretsByProjectID(c.Context(), projectId)
		if err != nil {
			log.Printf("Error fetching secrets for project %s: %v", projectId, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch secrets",
			})
		}
		return c.JSON(secrets)
	})

	// Get secret by ID
	router.Get("/secrets/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if id == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret ID cannot be empty",
			})
		}

		secret, err := readOnlyDatabase.GetSecretByID(c.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Secret not found",
				})
			}
			log.Printf("Failed to fetch secret %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch secret",
			})
		}
		return c.JSON(secret)
	})
}

func RegisterWriteSecretRoute(router fiber.Router, readWriteDatabase *generated.Queries) {
	// Create secret
	router.Post("/secrets", func(c *fiber.Ctx) error {
		var body struct {
			ProjectID   string  `json:"project_id"`
			Key         string  `json:"key"`
			Value       string  `json:"value"`
			Description *string `json:"description"`
		}

		if err := c.BodyParser(&body); err != nil {
			log.Printf("Body parse error: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validation
		if strings.TrimSpace(body.ProjectID) == "" {
			log.Printf("Id: %s", body.ProjectID)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Project ID is required",
			})
		}

		if strings.TrimSpace(body.Key) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret key is required",
			})
		}

		if strings.TrimSpace(body.Value) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret value is required",
			})
		}

		// Verify project exists
		_, err := readWriteDatabase.GetProjectByID(c.Context(), body.ProjectID)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Project not found",
				})
			}
			log.Printf("Failed to fetch project %s: %v", body.ProjectID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify project",
			})
		}

		id := uuid.New().String()

		newSecret := generated.CreateSecretParams{
			ID:          id,
			ProjectID:   body.ProjectID,
			Key:         body.Key,
			Value:       body.Value,
			Description: body.Description,
		}

		secret, err := readWriteDatabase.CreateSecret(c.Context(), newSecret)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Secret with this name already exists in the project",
				})
			}
			if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Project not found",
				})
			}
			log.Printf("Failed to create secret: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create secret",
			})
		}

		server_sse.BroadcastSecretChange(server_sse.EventCreate, secret)

		return c.Status(fiber.StatusCreated).JSON(secret)
	})

	// Update secret
	router.Patch("/secrets/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if id == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret ID is required",
			})
		}

		var body struct {
			ProjectID   string  `json:"project_id"`
			Key         *string  `json:"key"`
			Value       *string  `json:"value"`
			Description *string `json:"description"`
		}

		if err := c.BodyParser(&body); err != nil {
			log.Printf("Body parse error: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validation - at least one field should be provided
		if body.Key == nil && body.Value == nil && body.Description == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "At least one field (key or value or description) must be provided",
			})
		}

		// Validate non-empty if provided
		if body.Key != nil && strings.TrimSpace(*body.Key) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret key cannot be empty",
			})
		}
		if body.Value != nil && strings.TrimSpace(*body.Value) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret value cannot be empty",
			})
		}
		// if body.Description != nil && strings.TrimSpace(*body.Description) == "" {
		// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 		"error": "Secret description cannot be empty",
		// 	})
		// }

		updatedSecret := generated.UpdateSecretParams{
			ID:     id,
			Key:   body.Key,
			Value: body.Value,
			Description: body.Description,
		}

		secret, err := readWriteDatabase.UpdateSecret(c.Context(), updatedSecret)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Secret not found",
				})
			}
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Secret with this name already exists in the project",
				})
			}
			log.Printf("Failed to update secret %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update secret",
			})
		}

		server_sse.BroadcastSecretChange(server_sse.EventUpdate, secret)

		return c.Status(fiber.StatusOK).JSON(secret)
	})

	// Delete secret
	router.Delete("/secrets/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		if id == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Secret ID cannot be empty",
			})
		}

		// Check if secret exists
		secret, err := readWriteDatabase.GetSecretByID(c.Context(), id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Secret not found",
				})
			}
			log.Printf("Failed to fetch secret %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch secret",
			})
		}

		err = readWriteDatabase.DeleteSecret(c.Context(), id)
		if err != nil {
			log.Printf("Failed to delete secret %s: %v", id, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete secret",
			})
		}

		server_sse.BroadcastSecretChange(server_sse.EventDelete, secret)

		return c.SendStatus(fiber.StatusNoContent)
	})
}
