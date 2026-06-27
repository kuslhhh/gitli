package search

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open mem db: %v", err)
	}

	schema := []string{
		`CREATE TABLE repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			default_branch TEXT NOT NULL DEFAULT 'main',
			last_scanned_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE commits (
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
		`CREATE VIRTUAL TABLE commits_fts USING fts5(
			message, author, email,
			content='commits',
			content_rowid='id'
		)`,
		`CREATE TRIGGER commits_fts_ai AFTER INSERT ON commits BEGIN
			INSERT INTO commits_fts(rowid, message, author, email)
			VALUES (new.id, new.message, new.author, new.email);
		END`,
	}
	for _, s := range schema {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("exec schema: %v\nSQL: %s", err, s)
		}
	}

	return db
}

func insertRepo(t *testing.T, db *sql.DB, name, path string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRow(
		"INSERT INTO repositories (name, path) VALUES (?, ?) RETURNING id",
		name, path,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert repo: %v", err)
	}
	return id
}

func insertCommit(t *testing.T, db *sql.DB, repoID int64, hash, author, email, msg, committedAt string) {
	t.Helper()
	_, err := db.Exec(
		"INSERT INTO commits (repo_id, hash, author, email, message, committed_at) VALUES (?, ?, ?, ?, ?, ?)",
		repoID, hash, author, email, msg, committedAt,
	)
	if err != nil {
		t.Fatalf("insert commit: %v", err)
	}
}

func TestToFTS5Query(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"redis", "redis*"},
		{"implement jwt", "implement* AND jwt*"},
		{"", ""},
	}
	for _, tc := range tests {
		got := toFTS5Query(tc.input)
		if got != tc.want {
			t.Errorf("toFTS5Query(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSearchFTS(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repoID := insertRepo(t, db, "my-project", "/tmp/my-project")
	insertCommit(t, db, repoID, "a1", "Alice", "alice@c.com", "implemented redis cache", "2024-01-15T10:00:00Z")
	insertCommit(t, db, repoID, "a2", "Bob", "bob@c.com", "fixed jwt auth", "2024-01-14T10:00:00Z")

	s := New(db)
	results, err := s.Search("redis")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Message != "implemented redis cache" {
		t.Errorf("expected 'implemented redis cache', got '%s'", results[0].Message)
	}
}

func TestSearchMultiWord(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repoID := insertRepo(t, db, "proj", "/tmp/proj")
	insertCommit(t, db, repoID, "a1", "A", "a@c.com", "implemented jwt authentication", "2024-01-15T10:00:00Z")
	insertCommit(t, db, repoID, "a2", "B", "b@c.com", "fixed css styling", "2024-01-14T10:00:00Z")

	s := New(db)
	results, err := s.Search("jwt authentication")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSearchNoResults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repoID := insertRepo(t, db, "proj", "/tmp/proj")
	insertCommit(t, db, repoID, "a1", "A", "a@c.com", "some work", "2024-01-15T10:00:00Z")

	s := New(db)
	results, err := s.Search("nonexistent")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchRepoNameMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repoID := insertRepo(t, db, "redis-cache", "/tmp/redis-cache")
	insertCommit(t, db, repoID, "a1", "A", "a@c.com", "initial commit", "2024-01-15T10:00:00Z")

	s := New(db)
	results, err := s.Search("redis-cache")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for repo name match")
	}
}

func TestSearchResultsOrdered(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repoID := insertRepo(t, db, "proj", "/tmp/proj")
	insertCommit(t, db, repoID, "a1", "A", "a@c.com", "redis: old", "2024-01-10T10:00:00Z")
	insertCommit(t, db, repoID, "a2", "B", "b@c.com", "redis: new", "2024-02-15T10:00:00Z")

	s := New(db)
	results, err := s.Search("redis")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Newest first
	if results[0].Message != "redis: new" {
		t.Errorf("expected newest first, got '%s'", results[0].Message)
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2024-01-15T10:30:00Z", "2024-01-15 10:30:00"},
		{"2024-01-15 10:30:00", "2024-01-15 10:30:00"},
	}
	for _, tc := range tests {
		got := parseTime(tc.input)
		if got.IsZero() {
			t.Errorf("parseTime(%q) returned zero time", tc.input)
		}
	}
}

func TestSearchEmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	s := New(db)
	results, err := s.Search("anything")
	if err != nil {
		t.Fatalf("Search on empty DB: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results on empty DB, got %d", len(results))
	}
}
