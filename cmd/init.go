package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	Long:  `Create the database file and setup tables.`,
	Run: func(cmd *cobra.Command, args []string) {
		
		fmt.Println("Init method called")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
