package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kush/gitli/internal/models"
	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode and foreign keys
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			return nil, fmt.Errorf("set pragma: %w", err)
		}
	}

	db := &DB{conn: conn}
	if err := db.autoMigrate(); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}

// UpsertRepository inserts a repository or updates it if it already exists (matched by path).
// Returns the repository ID using RETURNING for reliability.
func (db *DB) UpsertRepository(repo *models.Repository) (int64, error) {
	query := `
		INSERT INTO repositories (name, path, default_branch, last_scanned_at)
		VALUES (?, ?, ?, datetime('now'))
		ON CONFLICT(path) DO UPDATE SET
			name = excluded.name,
			default_branch = excluded.default_branch,
			last_scanned_at = datetime('now')
		RETURNING id
	`
	var id int64
	err := db.conn.QueryRow(query, repo.Name, repo.Path, repo.DefaultBranch).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert repository: %w", err)
	}
	return id, nil
}

// InsertCommits bulk-inserts commits for a repository, skipping duplicates.
func (db *DB) InsertCommits(repoID int64, commits []models.Commit) (int, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO commits (repo_id, hash, author, email, message, committed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for _, c := range commits {
		result, err := stmt.Exec(repoID, c.Hash, c.Author, c.Email, c.Message, c.CommittedAt)
		if err != nil {
			return 0, fmt.Errorf("insert commit %s: %w", c.Hash, err)
		}
		n, _ := result.RowsAffected()
		inserted += int(n)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}
	return inserted, nil
}

// InsertBranches bulk-inserts branches for a repository, replacing existing ones.
func (db *DB) InsertBranches(repoID int64, branches []models.Branch) (int, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear old branches for this repo first, then re-insert
	if _, err := tx.Exec("DELETE FROM branches WHERE repo_id = ?", repoID); err != nil {
		return 0, fmt.Errorf("delete old branches: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO branches (repo_id, name, is_current)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for _, b := range branches {
		isCurrent := 0
		if b.IsCurrent {
			isCurrent = 1
		}
		_, err := stmt.Exec(repoID, b.Name, isCurrent)
		if err != nil {
			return 0, fmt.Errorf("insert branch %s: %w", b.Name, err)
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}
	return inserted, nil
}

// InsertStashes bulk-inserts stashes for a repository, skipping duplicates.
func (db *DB) InsertStashes(repoID int64, stashes []models.Stash) (int, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO stashes (repo_id, stash_name)
		VALUES (?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for _, s := range stashes {
		result, err := stmt.Exec(repoID, s.StashName)
		if err != nil {
			return 0, fmt.Errorf("insert stash %s: %w", s.StashName, err)
		}
		n, _ := result.RowsAffected()
		inserted += int(n)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}
	return inserted, nil
}

// GetRepoByName finds a repository by name (case-insensitive, partial match).
func (db *DB) GetRepoByName(name string) (*models.Repository, error) {
	query := `
		SELECT id, name, path, default_branch, last_scanned_at
		FROM repositories
		WHERE name LIKE '%' || ? || '%'
		ORDER BY
			CASE
				WHEN LOWER(name) = LOWER(?) THEN 0
				WHEN LOWER(name) LIKE LOWER(?) || '%' THEN 1
				ELSE 2
			END
		LIMIT 1
	`
	var repo models.Repository
	var lastScanned string
	err := db.conn.QueryRow(query, name, name, name).Scan(
		&repo.ID, &repo.Name, &repo.Path, &repo.DefaultBranch, &lastScanned,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no repository found matching '%s'", name)
	}
	if err != nil {
		return nil, fmt.Errorf("query repository: %w", err)
	}
	repo.LastScannedAt, _ = time.Parse("2006-01-02 15:04:05", lastScanned)
	return &repo, nil
}

// GetLatestCommits returns the most recent commits for a repository from the database.
func (db *DB) GetLatestCommits(repoID int64, limit int) ([]models.Commit, error) {
	query := `
		SELECT id, repo_id, hash, author, email, message, committed_at
		FROM commits
		WHERE repo_id = ?
		ORDER BY committed_at DESC
		LIMIT ?
	`
	rows, err := db.conn.Query(query, repoID, limit)
	if err != nil {
		return nil, fmt.Errorf("query commits: %w", err)
	}
	defer rows.Close()

	var commits []models.Commit
	for rows.Next() {
		var c models.Commit
		var committedAt string
		if err := rows.Scan(&c.ID, &c.RepoID, &c.Hash, &c.Author, &c.Email, &c.Message, &committedAt); err != nil {
			return nil, fmt.Errorf("scan commit: %w", err)
		}
		c.CommittedAt, _ = time.Parse("2006-01-02 15:04:05", committedAt)
		commits = append(commits, c)
	}
	return commits, rows.Err()
}

func (db *DB) autoMigrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			default_branch TEXT NOT NULL DEFAULT 'main',
			last_scanned_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS commits (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo_id INTEGER NOT NULL,
			hash TEXT NOT NULL,
			author TEXT NOT NULL,
			email TEXT NOT NULL,
			message TEXT NOT NULL,
			committed_at DATETIME NOT NULL,
			FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE,
			UNIQUE(repo_id, hash)
		)`,
		`CREATE TABLE IF NOT EXISTS branches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			is_current INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE,
			UNIQUE(repo_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS stashes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo_id INTEGER NOT NULL,
			stash_name TEXT NOT NULL,
			FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE,
			UNIQUE(repo_id, stash_name)
		)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}

	return nil
}
