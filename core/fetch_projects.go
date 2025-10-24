package core

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/Knightshrestha/Secret-Injector/database/generated"
)

func FetchProjects() []generated.ProjectList {
	log.SetOutput(io.Discard)
	mainDb := database.OpenDatabase()
	log.SetOutput(os.Stdout)

	allProjects, err := mainDb.ReadQueries.GetAllProjects(context.Background())
	if err != nil {
		log.Fatalf("Could not fetch project list: %s", err)

	}

	if mainDb.WriteDB != nil || mainDb.ReadDB != nil {
		// log.Println("Closing database connections...")
		if err := database.CloseDatabase(mainDb); err != nil {
			log.Println("Error closing database:", err)
		} else {
			// log.Println("✓ Database connections closed successfully")
		}
	}

	return allProjects

}
