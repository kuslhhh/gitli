package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kush/gitli/internal/git"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo [name]",
	Short: "Show repository details",
	Long: `Display detailed information about a repository including its current branch,
latest commits, stash count, and dirty status.

Accepts a repository name (partial match, case-insensitive).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Find the repo in the database
		repo, err := db.GetRepoByName(name)
		if err != nil {
			return err
		}

		// Styles
		accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
		hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
		msgStyle := lipgloss.NewStyle().Padding(0, 0, 0, 2)
		metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 0, 0, 2)

		// Header
		fmt.Println()
		fmt.Println(accent.Render("  " + repo.Name))
		fmt.Println(subtle.Render(fmt.Sprintf("  %s\n", repo.Path)))

		// Get live data from git
		currentBranch, _ := git.GetDefaultBranch(repo.Path)
		isDirty, _ := git.IsDirty(repo.Path)
		stashCount, _ := git.StashCount(repo.Path)

		// Status summary line
		dirtyStatus := green.Render("clean")
		if isDirty {
			dirtyStatus = yellow.Render("dirty")
		}

		branchDisplay := label.Render(fmt.Sprintf("  %s", currentBranch))
		if isDirty {
			branchDisplay = label.Render(fmt.Sprintf("  %s", currentBranch)) + " " + dirtyStatus
		}

		fmt.Printf("  %s  %s  %s\n\n",
			label.Render("branch"),
			branchDisplay,
			subtle.Render(fmt.Sprintf("stashes: %d", stashCount)),
		)

		// Latest commits from database
		commits, err := db.GetLatestCommits(repo.ID, 10)
		if err != nil {
			return fmt.Errorf("get commits: %w", err)
		}

		if len(commits) == 0 {
			fmt.Println(subtle.Render("  No commits indexed yet. Run 'gitli scan' first."))
			return nil
		}

		fmt.Println(label.Render("  latest commits"))
		for _, c := range commits {
			shortHash := c.Hash
			if len(shortHash) > 8 {
				shortHash = shortHash[:8]
			}

			line := fmt.Sprintf("%s %s", hashStyle.Render(shortHash), c.Message)
			if idx := strings.Index(line, "\n"); idx > 0 {
				line = line[:idx]
			}

			fmt.Println(msgStyle.Render(line))
			fmt.Println(metaStyle.Render(fmt.Sprintf("  %s • %s", formatTimeAgo(c.CommittedAt), c.Author)))
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
}