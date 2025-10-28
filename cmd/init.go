package cmd

import (
	"fmt"
	"log"
	
	"github.com/Knightshrestha/Secret-Injector/database"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  `Create the database file and setup tables.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.SetupDatabase(); err != nil {
			log.Fatal("Failed to setup database:", err)
		}
		
		fmt.Println("✓ Database schema initialized")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
