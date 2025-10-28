package cmd

import (
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize and select projects",
	Long:  `Fetch available projects and allow user to select one or multiple projects.`,
	Run: func(cmd *cobra.Command, args []string) {
	
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}