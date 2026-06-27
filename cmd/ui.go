package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/kush/gitli/internal/tui"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch interactive TUI",
	Long: `Launch the interactive terminal user interface with tabs for timeline,
search, repository browser, and activity analytics.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(
			tui.New(db),
			tea.WithAltScreen(),
		)

		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
