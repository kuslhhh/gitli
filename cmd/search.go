package cmd

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search commit history",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		println("Searching for:", args[0])
		println("No results (not implemented yet)")
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}