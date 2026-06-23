package cmd

import (
	"github.com/spf13/cobra"
)

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Show global activity timeline",
	Run: func(cmd *cobra.Command, args []string) {
		println("Timeline (not implemented yet)")
	},
}

func init() {
	rootCmd.AddCommand(timelineCmd)
}