package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kush/gitli/internal/embed"
	"github.com/kush/gitli/internal/search"
	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask a natural language question about your code",
	Long: `Ask a natural language question and find semantically relevant commits.
Uses Ollama (nomic-embed-text) for embeddings. If Ollama is not available,
falls back to keyword search.

Examples:
  gitli ask "where did I implement jwt refresh?"
  gitli ask "when did I add redis caching?"
  gitli ask "which commit changed the authentication logic?"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := args[0]

		// Styles
		subtle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		repoCol := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
		hashSty := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))

		embedder := embed.New("", "")

		// Check if Ollama is available with a quick connection test
		if !embedder.IsAvailable() {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(
				"⚠ Ollama not available. Install it from https://ollama.com and run:",
			))
			fmt.Println(subtle.Render("  ollama pull nomic-embed-text"))
			fmt.Println(subtle.Render("  ollama serve"))
			fmt.Println()
			fmt.Println(subtle.Render("Falling back to keyword search..."))

			// Fall back to regular search
			searcher := search.New(db.Conn())
			results, err := searcher.Search(question)
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}

			if len(results) == 0 {
				fmt.Println(subtle.Render("No results found."))
				return nil
			}

			displayAskResults(question, results, subtle, accent, repoCol, hashSty)
			return nil
		}

		// Generate embedding for the question
		fmt.Println(accent.Render("  Generating embedding for your question..."))

		queryVec, err := embedder.Embed(question)
		if err != nil {
			return fmt.Errorf("generate embedding: %w", err)
		}

		fmt.Println(subtle.Render("  Searching semantically similar commits..."))

		matches, err := db.SearchByEmbedding(queryVec, 10)
		if err != nil {
			return fmt.Errorf("semantic search: %w", err)
		}

		if len(matches) == 0 {
			fmt.Println(subtle.Render("  No semantically similar commits found."))
			fmt.Println()
			fmt.Println(subtle.Render("  Try a keyword search instead: gitli search " + question))
			return nil
		}

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Padding(0, 1)
		countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		fmt.Println()
		fmt.Println(headerStyle.Render(fmt.Sprintf("  Semantic results for: %s", question)))
		fmt.Println(countStyle.Render(fmt.Sprintf("  Found %d semantically similar commit(s)\n", len(matches))))

		currentRepo := ""
		for _, m := range matches {
			if m.RepoName != currentRepo {
				currentRepo = m.RepoName
				fmt.Println(repoCol.Render("  " + currentRepo))
			}

			shortHash := m.Hash
			if len(shortHash) > 8 {
				shortHash = shortHash[:8]
			}

			line := m.Message
			if idx := strings.Index(line, "\n"); idx > 0 {
				line = line[:idx]
			}

			score := int(m.Score * 100)
			scoreStyle := lipgloss.NewStyle()
			switch {
			case score >= 50:
				scoreStyle = scoreStyle.Foreground(lipgloss.Color("42"))
			case score >= 30:
				scoreStyle = scoreStyle.Foreground(lipgloss.Color("220"))
			default:
				scoreStyle = scoreStyle.Foreground(lipgloss.Color("8"))
			}

			fmt.Printf("  %s %s %s\n",
				scoreStyle.Render(fmt.Sprintf("%3d%%", score)),
				hashSty.Render(shortHash),
				line,
			)

			by := m.Author
			if m.Email != "" && !strings.Contains(m.Author, "@") {
				by = fmt.Sprintf("%s <%s>", m.Author, m.Email)
			}
			ago := formatTimeAgo(m.CommittedAt)
			fmt.Printf("    %s • %s\n", ago, by)
		}

		return nil
	},
}

func displayAskResults(question string, results []search.Result, subtle, accent, repoCol, hashSty lipgloss.Style) {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Padding(0, 1)
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("  Keyword results for: %s", question)))
	fmt.Println(countStyle.Render(fmt.Sprintf("  Found %d result(s)\n", len(results))))

	currentRepo := ""
	for _, r := range results {
		if r.RepoName != currentRepo {
			currentRepo = r.RepoName
			fmt.Println(repoCol.Render("  " + currentRepo))
		}

		shortHash := r.CommitHash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}

		line := r.Message
		if idx := strings.Index(line, "\n"); idx > 0 {
			line = line[:idx]
		}

		fmt.Printf("  %s %s\n", hashSty.Render(shortHash), line)

		by := r.Author
		if r.Email != "" && !strings.Contains(r.Author, "@") {
			by = fmt.Sprintf("%s <%s>", r.Author, r.Email)
		}
		ago := formatTimeAgo(r.CommittedAt)
		fmt.Printf("    %s • %s\n", ago, by)
	}
}

func init() {
	rootCmd.AddCommand(askCmd)
}
