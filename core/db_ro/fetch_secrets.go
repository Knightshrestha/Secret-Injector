package db_ro

import (
	"context"
	"fmt"

	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/Knightshrestha/Secret-Injector/database/generated"
)

func FetchSecrets(projectIds []string) ([]generated.SecretList, error) {
	mainDb, err := database.OpenReadDatabase()
	if err != nil {
		return nil, err
	}
	defer database.CloseReadDatabase(mainDb.DB)

	var allSecrets []generated.SecretList

	for _, projectId := range projectIds {
		secrets, err := mainDb.Queries.GetSecretsByProjectID(context.Background(), projectId)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch secrets for project %s: %w", projectId, err)
		}

		allSecrets = append(allSecrets, secrets...)
	}

	return allSecrets, nil
}

