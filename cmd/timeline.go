package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Show global activity timeline",
	Long: `Display the most recent commits across all indexed repositories,
sorted newest first.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := db.GetTimeline(50)
		if err != nil {
			return fmt.Errorf("get timeline: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
				"No activity yet. Run 'gitli scan' to index repositories.",
			))
			return nil
		}

		// Styles
		accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		repoColor := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
		hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
		msgStyle := lipgloss.NewStyle().Padding(0, 0, 0, 2)
		metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 0, 0, 2)
		sep := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("│")

		fmt.Println(accent.Render(fmt.Sprintf("Timeline — %d recent commits\n", len(entries))))

		for _, e := range entries {
			shortHash := e.Hash
			if len(shortHash) > 8 {
				shortHash = shortHash[:8]
			}

			line := e.Message
			if idx := strings.Index(line, "\n"); idx > 0 {
				line = line[:idx]
			}

			// Timestamp
			ago := formatTimeAgo(e.CommittedAt)

			// Build the entry
			fmt.Printf("  %s %s %s %s\n",
				subtle.Render(ago),
				sep,
				repoColor.Render(e.RepoName),
				hashStyle.Render(shortHash),
			)
			fmt.Println(msgStyle.Render(line))

			author := e.Author
			if e.Email != "" && !strings.Contains(e.Author, "@") {
				author = fmt.Sprintf("%s <%s>", e.Author, e.Email)
			}
			fmt.Println(metaStyle.Render(fmt.Sprintf("  by %s", author)))
			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(timelineCmd)
}