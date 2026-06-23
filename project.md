# gitli- Agent Driven Development Plan

## Project Overview

gitli is a local-first developer memory system.

It continuously indexes Git repositories on a machine and provides fast search, timeline, repository activity, branch history, stash history, and developer insights through a CLI interface.

The goal is to answer questions such as:

* What was I working on last week?
* Which repository contains authentication changes?
* When did I implement Redis caching?
* What projects have been most active recently?

---

# Technical Stack

## Language

* Go 1.24+

## Database

* SQLite

## CLI

* Cobra

## Output Styling

* Lipgloss

## Configuration

* Viper

## Future

* BubbleTea TUI
* SQLite FTS5
* Embedding Search
* Background Indexer

---

# Development Rules

## Rule 1

Agent must only implement the current phase.

Do not work on future phases.

## Rule 2

Before changing code:

* Read existing files
* Explain intended changes
* Generate implementation plan

## Rule 3

After implementation:

* Run tests
* Verify build
* Summarize changes

## Rule 4

Never generate placeholder code.

All code must compile.

## Rule 5

Prefer small commits.

One feature per commit.

---

# Architecture

```text
User Command
      ↓
CLI Layer
      ↓
Service Layer
      ↓
Git Adapter
      ↓
SQLite Storage
      ↓
Search Layer
```

---

# Folder Structure

```text
gitli/

├── cmd/
│   ├── root.go
│   ├── scan.go
│   ├── search.go
│   ├── repo.go
│   └── timeline.go
│
├── internal/
│   ├── database/
│   ├── scanner/
│   ├── git/
│   ├── search/
│   ├── models/
│   └── services/
│
├── migrations/
├── db/
├── scripts/
├── Makefile
├── main.go
└── README.md
```

---

# Phase 1 - Project Bootstrap

## Goal

Create production-ready project foundation.

## Tasks

### Initialize Project

```bash
go mod init github.com/yourusername/gitli
```

### Install Dependencies

* cobra
* sqlite
* lipgloss
* viper

### Create CLI Root Command

Commands:

* gitli
* gitli version
* gitli help

### Create Configuration Loader

Support:

```yaml
database:
  path: ~/.gitli/gitli.db
```

### Deliverables

* Build succeeds
* CLI runs
* Config loads

### Acceptance Criteria

```bash
gitli version
```

works successfully.

---

# Phase 2 - Database Layer

## Goal

Create SQLite storage system.

## Tasks

Create migrations.

### repositories

Fields:

* id
* name
* path
* default_branch
* last_scanned_at

### commits

Fields:

* id
* repo_id
* hash
* author
* email
* message
* committed_at

### branches

Fields:

* id
* repo_id
* name
* is_current

### stashes

Fields:

* id
* repo_id
* stash_name

### Deliverables

* Auto migration
* DB initialization

### Acceptance Criteria

Database created automatically.

---

# Phase 3 - Repository Discovery

## Goal

Find git repositories.

## Tasks

Implement filesystem scanner.

Detect:

```text
.git/
```

and

```text
.git
```

(worktree support)

Input:

```bash
gitli scan ~/projects
```

Output:

```text
Found 14 repositories
```

### Acceptance Criteria

All repositories detected.

---

# Phase 4 - Git Adapter

## Goal

Read repository metadata.

## Tasks

Implement:

### GetBranches()

Uses:

```bash
git branch
```

### GetCommits()

Uses:

```bash
git log
```

### GetStashes()

Uses:

```bash
git stash list
```

### GetStatus()

Uses:

```bash
git status --porcelain
```

### Acceptance Criteria

Can retrieve repository information.

---

# Phase 5 - Indexing Engine

## Goal

Store git information.

## Tasks

For every repository:

* insert repository
* insert commits
* insert branches
* insert stashes

Avoid duplicate commits.

### Acceptance Criteria

Repeated scans do not duplicate data.

---

# Phase 6 - Search Command

## Goal

Search commit history.

## Command

```bash
gitli search redis
```

## Tasks

Search:

* commit messages
* repository names

### Acceptance Criteria

Results returned in less than 1 second for normal datasets.

---

# Phase 7 - Repository View

## Goal

Show repository details.

## Command

```bash
gitli repo authify
```

Display:

* current branch
* latest commits
* stash count
* dirty status

### Acceptance Criteria

Information displayed cleanly.

---

# Phase 8 - Timeline

## Goal

Global activity feed.

## Command

```bash
gitli timeline
```

Display:

* timestamp
* repository
* commit message

Sorted newest first.

---

# Phase 9 - Full Text Search

## Goal

Replace LIKE queries.

## Tasks

Implement:

```sql
FTS5
```

### Acceptance Criteria

Search remains fast with 100k+ commits.

---

# Phase 10 - Developer Analytics

## Goal

Generate productivity insights.

## Command

```bash
gitli activity
```

Metrics:

* commits last 7 days
* commits last 30 days
* active repositories
* top repository
* branch activity

---

# Phase 11 - BubbleTea TUI

## Goal

Interactive interface.

## Command

```bash
gitli ui
```

Features:

* search
* timeline
* repository browser
* activity dashboard

---

# Phase 12 - Semantic Search

## Goal

Natural language repository memory.

Example:

```bash
gitli ask "where did I implement jwt refresh?"
```

Pipeline:

Question
→ Embedding
→ Vector Search
→ Matching Commits

---

# Definition Of Done

Project is complete when:

* Repository discovery works
* Git indexing works
* Search works
* Timeline works
* Activity analytics works
* Tests pass
* CI passes
* Binary builds on Linux, macOS, Windows

Target Release:

v1.0.0
