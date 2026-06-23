package cmd

import (
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo [name]",
	Short: "Show repository details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		println("Repository:", args[0])
		println("Not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
}