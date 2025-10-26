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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
