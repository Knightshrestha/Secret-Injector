package cmd

import (
	"fmt"
	"os"

	"github.com/Knightshrestha/Secret-Injector/core"
	"github.com/spf13/cobra"
)

var port int
var logging bool

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Secret Injector UI Server",
	Long:  `This starts the main server`,
	Run: func(cmd *cobra.Command, args []string) {
		if port < 1024 || port > 65535 {
			fmt.Fprintf(os.Stderr, "Error: port must be between 1024 and 65535\n")
			os.Exit(1)
		}
		core.StartServer(port, logging)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVarP(&port, "port", "p", 5544, "Port to run the server on")
	serveCmd.Flags().BoolVarP(&logging, "debug", "d", false, "Enable Logging")
}
