package search

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Result struct {
	RepoName    string    `json:"repo_name"`
	RepoPath    string    `json:"repo_path"`
	CommitHash  string    `json:"commit_hash"`
	Author      string    `json:"author"`
	Email       string    `json:"email"`
	Message     string    `json:"message"`
	CommittedAt time.Time `json:"committed_at"`
}

type Searcher struct {
	db *sql.DB
}

func New(db *sql.DB) *Searcher {
	return &Searcher{db: db}
}

// toFTS5Query converts a user query string to FTS5 MATCH syntax.
// Each word becomes a prefix match (word*) joined with AND.
func toFTS5Query(query string) string {
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}
	for i, w := range words {
		words[i] = w + "*"
	}
	return strings.Join(words, " AND ")
}

// parseTime tries multiple SQLite datetime formats.
func parseTime(s string) time.Time {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// Search searches commit messages, authors, and emails for the given query.
// Uses SQLite FTS5 for fast full-text search with LIKE as fallback.
// LIKE fallback also handles repository name matching, which FTS5 doesn't cover.
func (s *Searcher) Search(query string) ([]Result, error) {
	// Try FTS5 first for commit content search
	results, err := s.searchFTS(query)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// Fallback to LIKE-based search (handles repo name matches, partial substrings, etc.)
	return s.searchLIKE(query)
}

func (s *Searcher) searchFTS(query string) ([]Result, error) {
	ftsQuery := toFTS5Query(query)
	if ftsQuery == "" {
		return nil, fmt.Errorf("empty query")
	}

	sqlQuery := `
		SELECT c.hash, c.author, c.email, c.message, c.committed_at, r.name, r.path
		FROM commits_fts
		JOIN commits c ON commits_fts.rowid = c.id
		JOIN repositories r ON r.id = c.repo_id
		WHERE commits_fts MATCH ?
		ORDER BY c.committed_at DESC
		LIMIT 50
	`

	rows, err := s.db.Query(sqlQuery, ftsQuery)
	if err != nil {
		return nil, fmt.Errorf("FTS search: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		var committedAt string
		if err := rows.Scan(&r.CommitHash, &r.Author, &r.Email, &r.Message, &committedAt, &r.RepoName, &r.RepoPath); err != nil {
			return nil, fmt.Errorf("scan FTS result: %w", err)
		}
		r.CommittedAt = parseTime(committedAt)
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Searcher) searchLIKE(query string) ([]Result, error) {
	sqlQuery := `
		SELECT c.hash, c.author, c.email, c.message, c.committed_at, r.name, r.path
		FROM commits c
		JOIN repositories r ON r.id = c.repo_id
		WHERE c.message LIKE '%' || ? || '%'
		   OR r.name LIKE '%' || ? || '%'
		ORDER BY c.committed_at DESC
		LIMIT 50
	`

	rows, err := s.db.Query(sqlQuery, query, query)
	if err != nil {
		return nil, fmt.Errorf("LIKE search: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		var committedAt string
		if err := rows.Scan(&r.CommitHash, &r.Author, &r.Email, &r.Message, &committedAt, &r.RepoName, &r.RepoPath); err != nil {
			return nil, fmt.Errorf("scan LIKE result: %w", err)
		}
		r.CommittedAt = parseTime(committedAt)
		results = append(results, r)
	}

	return results, rows.Err()
}
