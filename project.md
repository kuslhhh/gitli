# gitli

**Local-first developer memory system** — all 12 phases complete ✅

gitli continuously indexes Git repositories and provides fast search, timeline, repository activity, branch history, stash history, developer insights, and semantic search through a CLI interface and interactive TUI.

It answers questions like:
- What was I working on last week?
- Which repository contains authentication changes?
- When did I implement Redis caching?
- What projects have been most active recently?

---

## Technical Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25+ |
| Database | SQLite (modernc.org/sqlite, no CGo) |
| CLI | Cobra |
| Styling | Lipgloss |
| Config | Viper |
| TUI | BubbleTea + Bubbles |
| Full-Text Search | SQLite FTS5 |
| Semantic Search | Ollama (nomic-embed-text) |

---

## Architecture

```
User Command
      ↓
CLI Layer (cmd/)
      ↓
Internal Packages (internal/)
      ↓
Git Adapter (internal/git/) → SQLite Storage (internal/database/)
                                    ↓
                            Search Layer (internal/search/ + FTS5)
                                    ↓
                            Embedding Search (internal/embed/)
```

## Current Folder Structure

```
gitli/
├── cmd/
│   ├── root.go        # Root command, config, DB init
│   ├── scan.go        # Repository scanning & indexing
│   ├── search.go      # Keyword search (FTS5)
│   ├── repo.go        # Repository detail view
│   ├── timeline.go    # Global activity feed
│   ├── activity.go    # Developer analytics
│   ├── ask.go         # Semantic search (Ollama)
│   └── ui.go          # Interactive TUI launcher
├── internal/
│   ├── config/        # Viper configuration loader
│   ├── database/      # SQLite storage, migrations, queries
│   ├── scanner/       # Filesystem repository discovery
│   ├── git/           # Git command adapter (branches, commits, stashes, status)
│   ├── search/        # FTS5 + LIKE search with fallback
│   ├── embed/         # Ollama embedding client & cosine similarity
│   ├── models/        # Data structs (Repository, Commit, Branch, Stash)
│   └── tui/           # BubbleTea terminal UI (4 tabs)
├── main.go
├── Makefile
├── README.md
└── go.mod / go.sum
```

---

## Implemented Phases

All 12 phases have been completed:

| Phase | Feature | Status |
|-------|---------|--------|
| 1 | Project Bootstrap (Cobra, Viper, CLI skeleton) | ✅ |
| 2 | Database Layer (SQLite, auto-migration) | ✅ |
| 3 | Repository Discovery (filesystem scanner) | ✅ |
| 4 | Git Adapter (branches, commits, stashes, status) | ✅ |
| 5 | Indexing Engine (dedup, bulk inserts) | ✅ |
| 6 | Search Command (keyword search) | ✅ |
| 7 | Repository View (repo details) | ✅ |
| 8 | Timeline (global activity feed) | ✅ |
| 9 | Full-Text Search (FTS5) | ✅ |
| 10 | Developer Analytics (activity dashboard) | ✅ |
| 11 | BubbleTea TUI (interactive interface) | ✅ |
| 12 | Semantic Search (Ollama embeddings) | ✅ |

---

## Commands

| Command | Description |
|---------|-------------|
| `gitli scan [path]` | Scan and index Git repositories |
| `gitli search <query>` | Search commit messages (FTS5) |
| `gitli repo <name>` | Show repository details |
| `gitli timeline` | Global activity feed |
| `gitli activity` | Developer analytics dashboard |
| `gitli ask <question>` | Semantic search (requires Ollama) |
| `gitli ui` | Interactive terminal UI |
| `gitli version` | Print version |
| `gitli --help` | Show help |

---

## Definition Of Done

- ✅ Repository discovery works
- ✅ Git indexing works
- ✅ Search works (keyword + FTS5 + semantic)
- ✅ Timeline works
- ✅ Activity analytics works
- ✅ Interactive TUI works
- ✅ Binary builds
- ✅ Documentation (README)

Target Release: **v1.0.0**
