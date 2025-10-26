package db_ro

import (
	"context"

	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/Knightshrestha/Secret-Injector/database/generated"
)

func FetchProjects() ([]generated.ProjectList, error) {
	mainDb, err := database.OpenReadDatabase()
	if err != nil {
		return nil, err
	}

	allProjects, err := mainDb.Queries.GetAllProjects(context.Background())
	if err != nil {
		return nil, err
	}

	if err := database.CloseReadDatabase(mainDb.DB); err != nil {
		return nil, err
	}

	return allProjects, nil
}
