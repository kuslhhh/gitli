package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kush/gitli/internal/search"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search commit history across all repositories",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		searcher := search.New(db.Conn())
		results, err := searcher.Search(query)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) == 0 {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("No results found for: " + query))
			return nil
		}

		// Styles
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(0, 1)

		countStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

		repoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

		hashStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))

		msgStyle := lipgloss.NewStyle().
			Padding(0, 0, 0, 2)

		metaStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(0, 0, 0, 2)

		fmt.Println(headerStyle.Render(fmt.Sprintf("Search results for: %s", query)))
		fmt.Println(countStyle.Render(fmt.Sprintf("Found %d result(s)\n", len(results))))

		// Group results by repo
		currentRepo := ""
		for _, r := range results {
			if r.RepoName != currentRepo {
				currentRepo = r.RepoName
				fmt.Println(repoStyle.Render("  " + currentRepo))
			}

			shortHash := r.CommitHash
			if len(shortHash) > 8 {
				shortHash = shortHash[:8]
			}

			// Format the commit line
			line := fmt.Sprintf("%s %s", hashStyle.Render(shortHash), r.Message)

			// Clean up multi-line messages — show only the first line
			if idx := strings.Index(line, "\n"); idx > 0 {
				line = line[:idx]
			}

			fmt.Println(msgStyle.Render(line))

			by := r.Author
			if r.Email != "" && !strings.Contains(r.Author, "@") {
				by = fmt.Sprintf("%s <%s>", r.Author, r.Email)
			}
			ago := formatTimeAgo(r.CommittedAt)
			fmt.Println(metaStyle.Render(fmt.Sprintf("%s • %s", ago, by)))
		}

		return nil
	},
}

func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 30*24*time.Hour:
		day := int(d.Hours() / 24)
		if day == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", day)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func init() {
	rootCmd.AddCommand(searchCmd)
}