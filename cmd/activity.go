package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var activityCmd = &cobra.Command{
	Use:   "activity",
	Short: "Show developer activity analytics",
	Long: `Display productivity insights including commit counts over time,
most active repositories, and branch activity.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		stats, err := db.GetActivityStats()
		if err != nil {
			return fmt.Errorf("get activity stats: %w", err)
		}

		if stats.TotalRepos == 0 {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
				"No data yet. Run 'gitli scan' to index repositories.",
			))
			return nil
		}

		// Styles
		accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		label := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
		subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		numStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
		barColor := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		repoColor := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)

		fmt.Println()
		fmt.Println(accent.Render("  Activity Overview"))
		fmt.Println()

		// Summary cards
		fmt.Printf("  %s  %s\n",
			label.Render("repositories"),
			numStyle.Render(fmt.Sprintf("%d", stats.TotalRepos)),
		)
		fmt.Printf("  %s  %s\n",
			label.Render("total commits"),
			numStyle.Render(fmt.Sprintf("%d", stats.TotalCommits)),
		)
		fmt.Printf("  %s  %s  %s\n",
			label.Render("commits (7d)"),
			formatCommitCount(stats.Commits7d, numStyle, subtle, green, yellow),
			subtle.Render(fmt.Sprintf("(%d in 30d)", stats.Commits30d)),
		)
		fmt.Println()

		// Top repos
		if len(stats.TopRepos) > 0 {
			fmt.Println(label.Render("  most active repositories"))
			maxCount := stats.TopRepos[0].Count
			if maxCount == 0 {
				maxCount = 1
			}
			for _, r := range stats.TopRepos {
				if r.Count == 0 {
					continue
				}
				barLen := r.Count * 20 / maxCount
				bar := ""
				for i := 0; i < barLen; i++ {
					bar += "█"
				}
				if bar == "" {
					bar = "▏"
				}

				fmt.Printf("  %s %s %s\n",
					repoColor.Render(fmt.Sprintf("%-20s", r.Name)),
					barColor.Render(bar),
					subtle.Render(fmt.Sprintf(" %d", r.Count)),
				)
			}
			fmt.Println()
		}

		// Branch activity — just list the branches (commit counts per branch
		// aren't stored in the data model, so we show branch names per repo)
		if len(stats.BranchCounts) > 0 {
			fmt.Println(label.Render("  branch activity"))
			for _, b := range stats.BranchCounts {
				if b.Branch == "" {
					continue
				}
				fmt.Printf("  %s/%s  %s\n",
					repoColor.Render(b.RepoName),
					lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(b.Branch),
					subtle.Render(fmt.Sprintf("%s commits in repo", fmt.Sprint(b.Commits))),
				)
			}
			fmt.Println()
		}

		return nil
	},
}

func formatCommitCount(n int, numStyle, subtle, green, yellow lipgloss.Style) string {
	color := subtle
	if n > 0 {
		color = green
	}
	if n >= 10 {
		color = yellow
	}
	return color.Render(fmt.Sprintf("%d", n))
}

func init() {
	rootCmd.AddCommand(activityCmd)
}
