# gitli

**Local-first developer memory system.** Index your Git repositories and search, explore, and analyze your development activity — all from the command line.

```bash
gitli scan ~/projects          # Index all repos
gitli search "redis cache"     # Search commit messages
gitli timeline                 # Global activity feed
gitli repo my-project          # View repo details
gitli activity                 # Developer analytics
gitli ask "jwt implementation" # Semantic search (requires Ollama)
gitli ui                       # Interactive TUI
```

---

## Quick Start

### 1. Build

```bash
git clone git@github.com:kuslhhh/gitli.git
cd gitli
make build
```

Or without `make`:

```bash
go build -o gitli .
```

### 2. Index a repository

```bash
./gitli scan ~/projects/my-project
```

Scans a directory for Git repositories (`.git/` directories and worktrees), reads branches, commits, and stashes, and stores everything in a local SQLite database.

### 3. Search

```bash
./gitli search "implement bootstrap"
```

### 4. Explore

```bash
./gitli timeline        # Recent commits across all repos
./gitli repo my-project # Details for a specific repo
./gitli activity        # Commit counts, top repos, branch activity
```

### 5. Interactive TUI

```bash
./gitli ui
```

A terminal UI with four tabs:
| Tab | Keys | Description |
|-----|------|-------------|
| Timeline | `↑↓` scroll | Recent commits across all repos |
| Search | `Enter` to type, `↑↓` results | Interactive keyword search |
| Repos | `↑↓` select, `Enter` detail | Repo list with activity bars |
| Activity | `↑↓` scroll | Analytics dashboard |

Press `1`-`4` to switch tabs, `q` or `Ctrl+C` to quit.

### 6. Semantic Search (optional)

Requires [Ollama](https://ollama.com):

```bash
ollama pull nomic-embed-text
ollama serve
./gitli scan ~/projects       # Generates embeddings during scan
./gitli ask "where did I implement jwt refresh?"
./gitli ask "when did I add redis caching?"
```

If Ollama is not available, `gitli ask` falls back to keyword search.

---

## Commands

### `gitli scan [path]`

Scan a directory for Git repositories and index them.

```bash
gitli scan                    # Scan current directory
gitli scan ~/projects         # Scan a specific directory
gitli scan ~/projects ~/work  # Scan multiple directories
```

Indexes: repositories, branches, commits (up to 1000 per repo), stashes.  
Safe to run repeatedly — duplicate commits are automatically skipped.

### `gitli search <query>`

Search commit messages and repository names using SQLite FTS5 full-text search.

```bash
gitli search "redis"          # Find commits mentioning "redis"
gitli search "implement jwt"  # Multi-word search
```

### `gitli repo <name>`

Display details for a specific repository (partial name match).

```bash
gitli repo gitli              # Shows branch, dirty status, stash count, recent commits
```

### `gitli timeline`

Display the most recent commits across all indexed repositories.

```bash
gitli timeline                # Shows last 50 commits, newest first
```

### `gitli activity`

Developer analytics dashboard.

```bash
gitli activity                # Shows commit counts, top repos, branch activity
```

### `gitli ask <question>`

Natural language semantic search (requires Ollama).

```bash
gitli ask "what changed in the auth system?"
gitli ask "where did I implement rate limiting?"
```

### `gitli ui`

Launch the interactive terminal UI.

```bash
gitli ui
```

### `gitli version`

Print version information.

```bash
gitli version
```

---

## Configuration

Configuration is loaded from `~/.gitli.yaml`:

```yaml
database:
  path: ~/.gitli/gitli.db  # Default: ~/.gitli/gitli.db
```

You can specify a custom config with `--config`:

```bash
gitli --config /path/to/config.yaml scan
```

---

## Data Storage

All data is stored in a local SQLite database at `~/.gitli/gitli.db`.

### Database Tables

| Table | Contents |
|-------|----------|
| `repositories` | Name, path, default branch, last scanned at |
| `commits` | Hash, author, email, message, committed_at |
| `branches` | Name, is_current (per repo) |
| `stashes` | Stash name (per repo) |
| `commits_fts` | FTS5 full-text index on commit messages |
| `commit_embeddings` | Vector embeddings for semantic search |

---

## Dependencies

| Package | Purpose |
|---------|---------|
| [Cobra](https://github.com/spf13/cobra) | CLI framework |
| [Viper](https://github.com/spf13/viper) | Configuration |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework |
| [Bubbles](https://github.com/charmbracelet/bubbles) | TUI components (text input) |
| [SQLite (modernc)](https://modernc.org/sqlite) | Embedded database (no CGo) |

---

## Development

### Prerequisites

- Go 1.24+
- Git

### Build

```bash
make build     # Build binary
make test      # Run tests
make clean     # Clean build artifacts
make install   # Build and install to /usr/local/bin
make dev ARGS="scan ."  # Run with arguments
```

### Project Structure

```
gitli/
├── cmd/              # CLI commands
│   ├── root.go       # Root command, config, DB init
│   ├── scan.go       # Repo scanning & indexing
│   ├── search.go     # Keyword search
│   ├── repo.go       # Repo details
│   ├── timeline.go   # Global activity feed
│   ├── activity.go   # Developer analytics
│   ├── ask.go        # Semantic search
│   └── ui.go         # Interactive TUI
├── internal/
│   ├── config/       # Configuration loader
│   ├── database/     # SQLite storage & queries
│   ├── scanner/      # Filesystem repo discovery
│   ├── git/          # Git command adapter
│   ├── search/       # FTS5 & LIKE search
│   ├── embed/        # Ollama embedding client
│   ├── models/       # Data structs
│   └── tui/          # BubbleTea TUI
├── main.go
└── Makefile
```

### Testing the App

```bash
# 1. Start by scanning the project itself
go run . scan .

# 2. Search
go run . search "implement"

# 3. View repo details
go run . repo gitli

# 4. Check timeline
go run . timeline

# 5. View activity dashboard
go run . activity

# 6. Launch the TUI
go run . ui

# 7. Run the built binary on a larger directory
go build -o gitli .
./gitli scan ~/projects
./gitli search "redis"
./gitli activity
```

---

## License

MIT
