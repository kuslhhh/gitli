package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/kush/gitli/internal/embed"
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

		subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		fmt.Println("Scanning:", path)

		results, err := scanner.Scan(path)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		// Check Ollama availability once for embedding generation
		embedder := embed.New("", "")
		ollamaAvailable := embedder.IsAvailable()
		if ollamaAvailable {
			fmt.Println(subtle.Render("  Ollama detected — will generate semantic embeddings"))
		} else {
			fmt.Println(subtle.Render("  Ollama not available — skipping semantic embeddings"))
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

			// Generate semantic embeddings for new commits
			if cCount > 0 && ollamaAvailable {
				fmt.Printf("    embeddings: generating...\n")
				eCount := 0
				for _, c := range commits {
					var commitID int64
					err := db.Conn().QueryRow(
						"SELECT id FROM commits WHERE repo_id = ? AND hash = ?",
						repoID, c.Hash,
					).Scan(&commitID)
					if err != nil {
						continue
					}

					text := c.Message
					vec, err := embedder.Embed(text)
					if err != nil {
						continue
					}

					if err := db.StoreEmbedding(commitID, vec, embed.DefaultModel); err != nil {
						continue
					}
					eCount++
				}
				if eCount > 0 {
					fmt.Printf("    embeddings: %d generated\n", eCount)
				}
			}

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
