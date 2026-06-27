package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kush/gitli/internal/models"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("New DB: %v", err)
	}
	return db
}

func TestNewAndClose(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestUpsertRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{
		Name:          "test-repo",
		Path:          "/tmp/test-repo",
		DefaultBranch: "main",
	}

	id, err := db.UpsertRepository(&repo)
	if err != nil {
		t.Fatalf("UpsertRepository: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	// Upsert again should return same id
	id2, err := db.UpsertRepository(&repo)
	if err != nil {
		t.Fatalf("UpsertRepository (2nd): %v", err)
	}
	if id2 != id {
		t.Errorf("expected same id %d, got %d", id, id2)
	}
}

func TestInsertCommits(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{Name: "r", Path: "/tmp/r", DefaultBranch: "main"}
	repoID, err := db.UpsertRepository(&repo)
	if err != nil {
		t.Fatalf("UpsertRepository: %v", err)
	}

	now := time.Now().UTC()
	commits := []models.Commit{
		{Hash: "aaa", Author: "A", Email: "a@a.com", Message: "first", CommittedAt: now},
		{Hash: "bbb", Author: "B", Email: "b@b.com", Message: "second", CommittedAt: now},
	}

	n, err := db.InsertCommits(repoID, commits)
	if err != nil {
		t.Fatalf("InsertCommits: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 inserted, got %d", n)
	}

	// Insert again should be skipped (duplicate hashes)
	n, err = db.InsertCommits(repoID, commits)
	if err != nil {
		t.Fatalf("InsertCommits (dedup): %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 new inserts, got %d", n)
	}
}

func TestInsertBranches(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{Name: "r", Path: "/tmp/r", DefaultBranch: "main"}
	repoID, err := db.UpsertRepository(&repo)
	if err != nil {
		t.Fatalf("UpsertRepository: %v", err)
	}

	branches := []models.Branch{
		{Name: "main", IsCurrent: true},
		{Name: "dev", IsCurrent: false},
	}

	n, err := db.InsertBranches(repoID, branches)
	if err != nil {
		t.Fatalf("InsertBranches: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 branches, got %d", n)
	}
}

func TestInsertStashes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{Name: "r", Path: "/tmp/r", DefaultBranch: "main"}
	repoID, err := db.UpsertRepository(&repo)
	if err != nil {
		t.Fatalf("UpsertRepository: %v", err)
	}

	stashes := []models.Stash{
		{StashName: "WIP on main: abc123"},
		{StashName: "WIP on dev: def456"},
	}

	n, err := db.InsertStashes(repoID, stashes)
	if err != nil {
		t.Fatalf("InsertStashes: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 stashes, got %d", n)
	}
}

func TestGetRepoByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{Name: "my-project", Path: "/tmp/my-project", DefaultBranch: "main"}
	_, err := db.UpsertRepository(&repo)
	if err != nil {
		t.Fatalf("UpsertRepository: %v", err)
	}

	found, err := db.GetRepoByName("my-project")
	if err != nil {
		t.Fatalf("GetRepoByName: %v", err)
	}
	if found.Name != "my-project" {
		t.Errorf("expected my-project, got %s", found.Name)
	}

	// Partial match
	found, err = db.GetRepoByName("my-proj")
	if err != nil {
		t.Fatalf("GetRepoByName partial: %v", err)
	}
	if found == nil {
		t.Fatal("expected repo for partial match")
	}
}

func TestGetRepoByNameNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetRepoByName("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent repo")
	}
}

func TestGetLatestCommits(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{Name: "r", Path: "/tmp/r", DefaultBranch: "main"}
	repoID, err := db.UpsertRepository(&repo)
	if err != nil {
		t.Fatalf("UpsertRepository: %v", err)
	}

	now := time.Now().UTC()
	commits := []models.Commit{
		{Hash: "c1", Author: "A", Email: "a@a.com", Message: "first", CommittedAt: now.Add(-2 * time.Hour)},
		{Hash: "c2", Author: "B", Email: "b@b.com", Message: "second", CommittedAt: now.Add(-1 * time.Hour)},
		{Hash: "c3", Author: "C", Email: "c@c.com", Message: "third", CommittedAt: now},
	}

	_, err = db.InsertCommits(repoID, commits)
	if err != nil {
		t.Fatalf("InsertCommits: %v", err)
	}

	// Should return newest first
	latest, err := db.GetLatestCommits(repoID, 2)
	if err != nil {
		t.Fatalf("GetLatestCommits: %v", err)
	}
	if len(latest) != 2 {
		t.Fatalf("expected 2, got %d", len(latest))
	}
	if latest[0].Message != "third" {
		t.Errorf("expected third first, got %s", latest[0].Message)
	}
	if latest[1].Message != "second" {
		t.Errorf("expected second second, got %s", latest[1].Message)
	}
}

func TestGetTimeline(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo1 := models.Repository{Name: "repo-a", Path: "/tmp/a", DefaultBranch: "main"}
	repoID1, _ := db.UpsertRepository(&repo1)
	repo2 := models.Repository{Name: "repo-b", Path: "/tmp/b", DefaultBranch: "main"}
	repoID2, _ := db.UpsertRepository(&repo2)

	now := time.Now().UTC()
	db.InsertCommits(repoID1, []models.Commit{
		{Hash: "a1", Author: "A", Email: "a@a.com", Message: "repo a commit", CommittedAt: now.Add(-1 * time.Hour)},
	})
	db.InsertCommits(repoID2, []models.Commit{
		{Hash: "b1", Author: "B", Email: "b@b.com", Message: "repo b commit", CommittedAt: now},
	})

	timeline, err := db.GetTimeline(10)
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	if len(timeline) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(timeline))
	}
	// Newest first
	if timeline[0].RepoName != "repo-b" {
		t.Errorf("expected repo-b first, got %s", timeline[0].RepoName)
	}
	if timeline[1].RepoName != "repo-a" {
		t.Errorf("expected repo-a second, got %s", timeline[1].RepoName)
	}
}

func TestGetActivityStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{Name: "r", Path: "/tmp/r", DefaultBranch: "main"}
	repoID, _ := db.UpsertRepository(&repo)

	now := time.Now().UTC()
	db.InsertCommits(repoID, []models.Commit{
		{Hash: "c1", Author: "A", Email: "a@a.com", Message: "m1", CommittedAt: now.Add(-2 * 24 * time.Hour)},
		{Hash: "c2", Author: "B", Email: "b@b.com", Message: "m2", CommittedAt: now.Add(-15 * 24 * time.Hour)},
	})

	stats, err := db.GetActivityStats()
	if err != nil {
		t.Fatalf("GetActivityStats: %v", err)
	}
	if stats.TotalRepos != 1 {
		t.Errorf("expected 1 repo, got %d", stats.TotalRepos)
	}
	if stats.TotalCommits != 2 {
		t.Errorf("expected 2 commits, got %d", stats.TotalCommits)
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"2024-01-15T10:30:00Z", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		{"2024-01-15 10:30:00", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		{"2024-01-15T10:30:00", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
	}
	for _, tc := range tests {
		got, err := parseTime(tc.input)
		if err != nil {
			t.Errorf("parseTime(%q): %v", tc.input, err)
			continue
		}
		if !got.Equal(tc.want) {
			t.Errorf("parseTime(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseTimeInvalid(t *testing.T) {
	_, err := parseTime("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestStoreAndSearchEmbedding(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := models.Repository{Name: "r", Path: "/tmp/r", DefaultBranch: "main"}
	repoID, _ := db.UpsertRepository(&repo)

	now := time.Now().UTC()
	_, err := db.InsertCommits(repoID, []models.Commit{
		{Hash: "c1", Author: "A", Email: "a@a.com", Message: "implemented jwt auth", CommittedAt: now},
	})
	if err != nil {
		t.Fatalf("InsertCommits: %v", err)
	}

	var commitID int64
	db.conn.QueryRow("SELECT id FROM commits WHERE hash = 'c1'").Scan(&commitID)

	vec := []float64{0.1, 0.2, 0.3}
	if err := db.StoreEmbedding(commitID, vec, "test-model"); err != nil {
		t.Fatalf("StoreEmbedding: %v", err)
	}

	// Search with similar vector
	matches, err := db.SearchByEmbedding([]float64{0.1, 0.2, 0.3}, 10)
	if err != nil {
		t.Fatalf("SearchByEmbedding: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least 1 match")
	}
	if matches[0].Message != "implemented jwt auth" {
		t.Errorf("expected 'implemented jwt auth', got '%s'", matches[0].Message)
	}
	if matches[0].Score <= 0 {
		t.Errorf("expected positive score, got %f", matches[0].Score)
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		a, b []float64
		want float64
	}{
		{[]float64{1, 0}, []float64{1, 0}, 1.0},
		{[]float64{1, 0}, []float64{0, 1}, 0.0},
		{[]float64{1, 2, 3}, []float64{2, 4, 6}, 1.0},
		{[]float64{1, 0, 0}, []float64{0, 1, 0}, 0.0},
	}
	for _, tc := range tests {
		got := cosineSimilarity(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("cosineSimilarity(%v, %v) = %f, want %f", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCosineSimilarityMismatchedLength(t *testing.T) {
	got := cosineSimilarity([]float64{1, 2}, []float64{1, 2, 3})
	if got != 0 {
		t.Errorf("expected 0 for mismatched lengths, got %f", got)
	}
}

func TestCosineSimilarityZeroVector(t *testing.T) {
	got := cosineSimilarity([]float64{0, 0}, []float64{1, 1})
	if got != 0 {
		t.Errorf("expected 0 for zero vector, got %f", got)
	}
}

