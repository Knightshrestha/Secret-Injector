package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// injectCmd represents the inject command
var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Inject secrets and run commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("inject called")
	},
}

func init() {
	rootCmd.AddCommand(injectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// injectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// injectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
