package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kush/gitli/internal/database"
	"github.com/kush/gitli/internal/search"
)

// ─── Styles ────────────────────────────────────────────────────────────────

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	tabStyle = lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(lipgloss.Color("8"))

	tabActiveStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Underline(true)

	accent   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	subtle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	label    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	hashSty  = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	repoCol  = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	green    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	yellow   = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	numStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// ─── Tab constants ─────────────────────────────────────────────────────────

const (
	tabTimeline = iota
	tabSearch
	tabRepos
	tabActivity
)

var tabNames = []string{"1 Timeline", "2 Search", "3 Repos", "4 Activity"}

// ─── Model ─────────────────────────────────────────────────────────────────

type model struct {
	db       *database.DB
	searcher *search.Searcher
	width    int
	height   int
	tabIndex int
	ready    bool
	loading  bool
	err      error

	timeline []database.TimelineEntry
	offset   int

	// Search
	searchInput  textinput.Model
	searchActive bool
	searchQuery  string
	searchRes    []search.Result
	searchCur    int

	// Repos
	repos        []database.RepoCount
	repoCur      int
	repoDetail   string
	repoCommits  []database.TimelineEntry

	// Activity
	activity *database.ActivityStats
}

func New(db *database.DB) tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Search commits..."
	ti.CharLimit = 100
	ti.Width = 40

	return &model{
		db:          db,
		searcher:    search.New(db.Conn()),
		tabIndex:    tabTimeline,
		searchInput: ti,
	}
}

// ─── Init ──────────────────────────────────────────────────────────────────

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		loadTimeline(m.db),
		loadRepos(m.db),
		loadActivity(m.db),
	)
}

func loadTimeline(db *database.DB) tea.Cmd {
	return func() tea.Msg {
		entries, err := db.GetTimeline(100)
		if err != nil {
			return errMsg{err}
		}
		return timelineLoadedMsg{entries}
	}
}

func loadRepos(db *database.DB) tea.Cmd {
	return func() tea.Msg {
		stats, err := db.GetActivityStats()
		if err != nil {
			return errMsg{err}
		}
		return reposLoadedMsg{stats.TopRepos}
	}
}

func loadActivity(db *database.DB) tea.Cmd {
	return func() tea.Msg {
		stats, err := db.GetActivityStats()
		if err != nil {
			return errMsg{err}
		}
		return activityLoadedMsg{stats}
	}
}

func loadRepoDetail(db *database.DB, name string) tea.Cmd {
	return func() tea.Msg {
		repo, err := db.GetRepoByName(name)
		if err != nil {
			return errMsg{err}
		}
		commits, err := db.GetLatestCommits(repo.ID, 10)
		if err != nil {
			return errMsg{err}
		}
		entries := make([]database.TimelineEntry, len(commits))
		for i, c := range commits {
			entries[i] = database.TimelineEntry{
				Commit:   c,
				RepoName: repo.Name,
				RepoPath: repo.Path,
			}
		}
		return repoDetailLoadedMsg{repo.Name, entries}
	}
}

func searchCommits(s *search.Searcher, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := s.Search(query)
		if err != nil {
			return errMsg{err}
		}
		return searchResultsMsg{results}
	}
}

// ─── Messages ──────────────────────────────────────────────────────────────

type errMsg struct{ err error }
type timelineLoadedMsg struct{ entries []database.TimelineEntry }
type reposLoadedMsg struct{ repos []database.RepoCount }
type activityLoadedMsg struct{ stats *database.ActivityStats }
type searchResultsMsg struct{ results []search.Result }
type repoDetailLoadedMsg struct {
	name    string
	commits []database.TimelineEntry
}

// ─── Update ────────────────────────────────────────────────────────────────

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case timelineLoadedMsg:
		m.timeline = msg.entries
		m.loading = false
		m.clampOffset()

	case reposLoadedMsg:
		m.repos = msg.repos
		m.loading = false
		if m.repoCur >= len(m.repos) && len(m.repos) > 0 {
			m.repoCur = len(m.repos) - 1
		}

	case activityLoadedMsg:
		m.activity = msg.stats
		m.loading = false

	case searchResultsMsg:
		m.searchRes = msg.results
		m.loading = false
		if m.searchCur >= len(m.searchRes) {
			m.searchCur = max(0, len(m.searchRes)-1)
		}

	case repoDetailLoadedMsg:
		m.repoDetail = msg.name
		m.repoCommits = msg.commits
		m.loading = false

	case errMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, nil
}

func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "1":
		m.tabIndex = tabTimeline
		m.offset = 0
		return m, nil
	case "2":
		m.tabIndex = tabSearch
		m.searchCur = 0
		m.offset = 0
		return m, nil
	case "3":
		m.tabIndex = tabRepos
		m.repoCur = 0
		m.repoDetail = ""
		m.offset = 0
		return m, nil
	case "4":
		m.tabIndex = tabActivity
		m.offset = 0
		return m, loadActivity(m.db)
	}

	switch m.tabIndex {
	case tabTimeline:
		return m.handleTimelineKey(msg)
	case tabSearch:
		return m.handleSearchKey(msg)
	case tabRepos:
		return m.handleReposKey(msg)
	case tabActivity:
		return m.handleActivityKey(msg)
	}

	return m, nil
}

func (m *model) clampOffset() {
	maxOffset := max(0, len(m.timeline)-1)
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

// ─── Timeline tab ──────────────────────────────────────────────────────────

func (m *model) handleTimelineKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.offset > 0 {
			m.offset--
		}
	case "down", "j":
		if m.offset < len(m.timeline)-1 {
			m.offset++
		}
	}
	return m, nil
}

func (m *model) viewTimeline() string {
	if m.timeline == nil {
		return subtle.Render("Loading timeline...")
	}
	if len(m.timeline) == 0 {
		return subtle.Render("No commits yet. Run 'gitli scan' to index repositories.")
	}

	contentHeight := m.height - 6
	if contentHeight < 1 {
		contentHeight = 10
	}

	visible := m.timeline
	end := m.offset + contentHeight
	if end > len(visible) {
		end = len(visible)
	}
	visible = visible[m.offset:end]

	var b strings.Builder
	for _, e := range visible {
		shortHash := e.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		line := e.Message
		if idx := strings.Index(line, "\n"); idx > 0 {
			line = line[:idx]
		}

		ago := formatTimeAgo(e.CommittedAt)

		fmt.Fprintf(&b, "  %s %s %s %s\n",
			subtle.Render(ago),
			subtle.Render("│"),
			repoCol.Render(e.RepoName),
			hashSty.Render(shortHash),
		)
		fmt.Fprintf(&b, "    %s\n", line)
		fmt.Fprintf(&b, "    %s\n\n", subtle.Render("by "+e.Author))
	}

	return b.String()
}

// ─── Search tab ────────────────────────────────────────────────────────────

func (m *model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchActive = !m.searchActive
		if !m.searchActive {
			m.searchInput.Blur()
		} else {
			m.searchInput.Focus()
		}
		return m, nil

	case "up", "k":
		if !m.searchActive && m.searchCur > 0 {
			m.searchCur--
		}
	case "down", "j":
		if !m.searchActive && m.searchCur < len(m.searchRes)-1 {
			m.searchCur++
		}
	case "esc":
		if m.searchActive {
			m.searchActive = false
			m.searchInput.Blur()
			return m, nil
		}
	}

	if m.searchActive {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		if query := strings.TrimSpace(m.searchInput.Value()); query != "" && query != m.searchQuery {
			m.searchQuery = query
			m.loading = true
			return m, tea.Batch(cmd, searchCommits(m.searcher, query))
		}
		return m, cmd
	}

	return m, nil
}

func (m *model) viewSearch() string {
	var b strings.Builder

	if m.searchActive {
		fmt.Fprintf(&b, "  %s %s\n\n", label.Render("search:"), m.searchInput.View())
	} else {
		query := m.searchInput.Value()
		if query == "" {
			fmt.Fprintf(&b, "  %s\n\n", subtle.Render("Press Enter to start typing, then type to search"))
		} else {
			fmt.Fprintf(&b, "  %s %s\n\n", label.Render("search:"), subtle.Render(query))
		}
	}

	if m.searchRes == nil {
		if m.loading {
			b.WriteString("  " + subtle.Render("Searching...") + "\n")
		}
		return b.String()
	}

	if len(m.searchRes) == 0 {
		b.WriteString("  " + subtle.Render("No results found.") + "\n")
		return b.String()
	}

	contentHeight := m.height - 8
	if contentHeight > len(m.searchRes) {
		contentHeight = len(m.searchRes)
	}

	start := m.searchCur - contentHeight/2
	if start < 0 {
		start = 0
	}
	end := start + contentHeight
	if end > len(m.searchRes) {
		end = len(m.searchRes)
	}

	for i := start; i < end; i++ {
		r := m.searchRes[i]
		shortHash := r.CommitHash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		line := r.Message
		if idx := strings.Index(line, "\n"); idx > 0 {
			line = line[:idx]
		}

		cursor := " "
		if i == m.searchCur {
			cursor = "▸"
		}

		ago := formatTimeAgo(r.CommittedAt)

		fmt.Fprintf(&b, "  %s %s %s %s\n",
			subtle.Render(cursor),
			repoCol.Render(r.RepoName),
			hashSty.Render(shortHash),
			subtle.Render(ago),
		)
		fmt.Fprintf(&b, "    %s\n", line)
	}

	return b.String()
}

// ─── Repos tab ─────────────────────────────────────────────────────────────

func (m *model) handleReposKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.repoDetail != "" {
		switch msg.String() {
		case "esc", "backspace":
			m.repoDetail = ""
			m.offset = 0
			return m, nil
		case "up", "k":
			if m.offset > 0 {
				m.offset--
			}
		case "down", "j":
			maxOffset := max(0, len(m.repoCommits)-1)
			if m.offset < maxOffset {
				m.offset++
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.repoCur > 0 {
			m.repoCur--
		}
	case "down", "j":
		if m.repoCur < len(m.repos)-1 {
			m.repoCur++
		}
	case "enter":
		if len(m.repos) > 0 && m.repoCur < len(m.repos) && m.repos[m.repoCur].Count > 0 {
			m.loading = true
			return m, loadRepoDetail(m.db, m.repos[m.repoCur].Name)
		}
	}

	return m, nil
}

func (m *model) viewRepos() string {
	if m.repoDetail != "" {
		return m.viewRepoDetail()
	}

	if m.repos == nil {
		return subtle.Render("Loading repositories...")
	}
	if len(m.repos) == 0 {
		return subtle.Render("No repositories indexed. Run 'gitli scan' first.")
	}

	var b strings.Builder
	maxCount := m.repos[0].Count
	if maxCount == 0 {
		maxCount = 1
	}

	contentHeight := m.height - 6
	if contentHeight > len(m.repos) {
		contentHeight = len(m.repos)
	}

	start := m.repoCur - contentHeight/2
	if start < 0 {
		start = 0
	}
	end := start + contentHeight
	if end > len(m.repos) {
		end = len(m.repos)
	}

	for i := start; i < end; i++ {
		r := m.repos[i]
		cursor := " "
		if i == m.repoCur {
			cursor = "▸"
		}

		barLen := r.Count * 20 / maxCount
		bar := strings.Repeat("█", barLen)
		if bar == "" && r.Count > 0 {
			bar = "▏"
		}

		fmt.Fprintf(&b, "  %s %-25s %s %s\n",
			subtle.Render(cursor),
			repoCol.Render(r.Name),
			accent.Render(bar),
			numStyle.Render(fmt.Sprintf("%d", r.Count)),
		)
	}

	if len(m.repos) > 0 {
		fmt.Fprintf(&b, "\n  %s\n", subtle.Render("Select a repo with ↑↓, press Enter to view details"))
	}

	return b.String()
}

func (m *model) viewRepoDetail() string {
	var b strings.Builder
	fmt.Fprintf(&b, "  %s %s\n\n", accent.Render("Repository:"), repoCol.Render(m.repoDetail))
	fmt.Fprintf(&b, "  %s\n\n", subtle.Render("Press Esc to go back"))

	if len(m.repoCommits) == 0 {
		b.WriteString("  " + subtle.Render("No commits.") + "\n")
		return b.String()
	}

	contentHeight := m.height - 8
	off := m.offset
	if off < 0 {
		off = 0
	}
	if off >= len(m.repoCommits) {
		off = max(0, len(m.repoCommits)-1)
	}

	end := off + contentHeight
	if end > len(m.repoCommits) {
		end = len(m.repoCommits)
	}
	visible := m.repoCommits[off:end]

	for _, c := range visible {
		shortHash := c.Hash
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		line := c.Message
		if idx := strings.Index(line, "\n"); idx > 0 {
			line = line[:idx]
		}

		ago := formatTimeAgo(c.CommittedAt)

		fmt.Fprintf(&b, "  %s %s\n", hashSty.Render(shortHash), line)
		fmt.Fprintf(&b, "    %s • %s\n\n", ago, c.Author)
	}

	return b.String()
}

// ─── Activity tab ──────────────────────────────────────────────────────────

func (m *model) handleActivityKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.offset > 0 {
			m.offset--
		}
	}
	return m, nil
}

func (m *model) viewActivity() string {
	if m.activity == nil {
		return subtle.Render("Loading activity...")
	}
	if m.activity.TotalRepos == 0 {
		return subtle.Render("No data yet. Run 'gitli scan' to index repositories.")
	}

	var b strings.Builder

	fmt.Fprintf(&b, "  %s  %s\n", label.Render("repositories"), numStyle.Render(fmt.Sprintf("%d", m.activity.TotalRepos)))
	fmt.Fprintf(&b, "  %s  %s\n", label.Render("total commits"), numStyle.Render(fmt.Sprintf("%d", m.activity.TotalCommits)))
	fmt.Fprintf(&b, "  %s  %s  %s\n\n", label.Render("commits (7d)"),
		activityColor(m.activity.Commits7d).Render(fmt.Sprintf("%d", m.activity.Commits7d)),
		subtle.Render(fmt.Sprintf("(%d in 30d)", m.activity.Commits30d)),
	)

	if len(m.activity.TopRepos) > 0 {
		fmt.Fprintf(&b, "  %s\n", label.Render("most active"))
		maxCount := m.activity.TopRepos[0].Count
		if maxCount == 0 {
			maxCount = 1
		}
		for _, r := range m.activity.TopRepos {
			if r.Count == 0 {
				continue
			}
			barLen := r.Count * 20 / maxCount
			bar := strings.Repeat("█", barLen)
			if bar == "" {
				bar = "▏"
			}
			fmt.Fprintf(&b, "  %-20s %s %s\n", repoCol.Render(r.Name), accent.Render(bar), numStyle.Render(fmt.Sprintf("%d", r.Count)))
		}
		fmt.Fprintf(&b, "\n")
	}

	if len(m.activity.BranchCounts) > 0 {
		fmt.Fprintf(&b, "  %s\n", label.Render("branches"))
		for _, br := range m.activity.BranchCounts {
			if br.Branch == "" {
				continue
			}
			fmt.Fprintf(&b, "  %s/%s  %s\n",
				repoCol.Render(br.RepoName),
				yellow.Render(br.Branch),
				subtle.Render(fmt.Sprintf("%d commits in repo", br.Commits)),
			)
		}
	}

	return b.String()
}

func activityColor(n int) lipgloss.Style {
	if n == 0 {
		return subtle
	}
	if n < 10 {
		return green
	}
	return yellow
}

// ─── View ──────────────────────────────────────────────────────────────────

func (m *model) View() string {
	if !m.ready {
		return "\n  Loading gitli..."
	}

	var b strings.Builder

	var tabRow []string
	for i, name := range tabNames {
		if i == m.tabIndex {
			tabRow = append(tabRow, tabActiveStyle.Render(name))
		} else {
			tabRow = append(tabRow, tabStyle.Render(name))
		}
	}
	fmt.Fprintf(&b, "  %s\n", lipgloss.JoinHorizontal(lipgloss.Top, tabRow...))
	b.WriteString(strings.Repeat("─", m.width-4) + "\n")

	var content string
	switch m.tabIndex {
	case tabTimeline:
		content = m.viewTimeline()
	case tabSearch:
		content = m.viewSearch()
	case tabRepos:
		content = m.viewRepos()
	case tabActivity:
		content = m.viewActivity()
	}

	if m.err != nil {
		content += "\n  " + errStyle.Render("Error: "+m.err.Error())
	}

	b.WriteString(content)

	help := m.footerHelp()
	b.WriteString("\n" + helpStyle.Render(help))

	return appStyle.Render(b.String())
}

func (m *model) footerHelp() string {
	switch m.tabIndex {
	case tabTimeline:
		return "↑↓ scroll • 1-4 switch tab • q quit"
	case tabSearch:
		if m.searchActive {
			return "type to search • Enter stop • Esc cancel • ↑↓ results"
		}
		return "Enter start typing • ↑↓ navigate results • 1-4 switch tab • q quit"
	case tabRepos:
		if m.repoDetail != "" {
			return "↑↓ scroll • Esc back • q quit"
		}
		return "↑↓ select • Enter view details • 1-4 switch tab • q quit"
	case tabActivity:
		return "↑↓ scroll • 1-4 switch tab • q quit"
	}
	return "1-4 switch tab • q quit"
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	case d < 30*24*time.Hour:
		day := int(d.Hours() / 24)
		if day == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", day)
	default:
		return t.Format("Jan 2")
	}
}
