package cmd

import (
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan for Git repositories",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		println("Scanning:", path)
		println("Found 0 repositories (not implemented yet)")
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}