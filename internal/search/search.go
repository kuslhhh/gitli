package search

import (
	"database/sql"
	"fmt"
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

// Search searches commit messages and repository names for the given query.
// Returns results sorted by most recent commit first, limited to 50.
func (s *Searcher) Search(query string) ([]Result, error) {
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
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		var committedAt string
		if err := rows.Scan(&r.CommitHash, &r.Author, &r.Email, &r.Message, &committedAt, &r.RepoName, &r.RepoPath); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		r.CommittedAt, _ = time.Parse("2006-01-02 15:04:05", committedAt)
		results = append(results, r)
	}

	return results, rows.Err()
}
