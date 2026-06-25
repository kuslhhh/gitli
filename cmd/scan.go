package cmd

import (
	"fmt"

	"github.com/kush/gitli/internal/scanner"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan for Git repositories",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		fmt.Println("Scanning:", path)

		results, err := scanner.Scan(path)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		repos := scanner.ToModels(results, "main")
		for _, repo := range repos {
			id, err := db.UpsertRepository(&repo)
			if err != nil {
				return fmt.Errorf("store repository %s: %w", repo.Path, err)
			}
			fmt.Printf("  %s (id=%d)\n", repo.Path, id)
		}

		fmt.Printf("Found %d repositories\n", len(repos))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
