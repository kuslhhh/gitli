package cmd

import (
	"fmt"

	"github.com/kush/gitli/internal/git"
	"github.com/kush/gitli/internal/scanner"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan and index Git repositories",
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

		for _, r := range results {
			fmt.Printf("\n  Repo: %s\n", r.Path)

			// Detect the actual default branch
			defBranch, err := git.GetDefaultBranch(r.Path)
			if err != nil {
				defBranch = "main"
			}

			repo := scanner.ToModels([]scanner.Result{r}, defBranch)[0]
			repoID, err := db.UpsertRepository(&repo)
			if err != nil {
				return fmt.Errorf("store repository %s: %w", r.Path, err)
			}
			fmt.Printf("    id: %d  branch: %s\n", repoID, defBranch)

			// Index branches
			branches, err := git.GetBranches(r.Path)
			if err != nil {
				return fmt.Errorf("get branches %s: %w", r.Path, err)
			}
			bCount, err := db.InsertBranches(repoID, branches)
			if err != nil {
				return fmt.Errorf("store branches %s: %w", r.Path, err)
			}
			fmt.Printf("    branches: %d\n", bCount)

			// Index commits
			commits, err := git.GetCommits(r.Path, 1000)
			if err != nil {
				return fmt.Errorf("get commits %s: %w", r.Path, err)
			}
			cCount, err := db.InsertCommits(repoID, commits)
			if err != nil {
				return fmt.Errorf("store commits %s: %w", r.Path, err)
			}
			fmt.Printf("    commits: %d new, %d total\n", cCount, len(commits))

			// Index stashes
			stashes, err := git.GetStashes(r.Path)
			if err != nil {
				return fmt.Errorf("get stashes %s: %w", r.Path, err)
			}
			sCount, err := db.InsertStashes(repoID, stashes)
			if err != nil {
				return fmt.Errorf("store stashes %s: %w", r.Path, err)
			}
			fmt.Printf("    stashes: %d\n", sCount)
		}

		fmt.Printf("\nFound and indexed %d repositories\n", len(results))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
