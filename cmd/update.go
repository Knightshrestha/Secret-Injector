/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Knightshrestha/Secret-Injector/config"
	"github.com/Knightshrestha/Secret-Injector/updater"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Fetch update from github",
	Long: `This command will fetch the latest release from github, check if new version is available and if so, download and update the app version. The old file is appended with ".old" so that we can go back in case something has gone wrong.`,
	Run: func(cmd *cobra.Command, args []string) {
		updateStruct := &updater.Updater{
			Owner:      config.Owner, // Replace with your GitHub username
			Repo:       config.Repo,  // Replace with your repo name
			CurrentVer: config.AppVersion,
			ExeName:    "secret_injector", // Your exe name without .exe
		}

		if err := updateStruct.Update(); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
