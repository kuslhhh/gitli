package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanEmptyDir(t *testing.T) {
	dir := t.TempDir()
	results, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan empty dir: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestScanStandardRepo(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, "repo1", ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}

	results, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "repo1" {
		t.Errorf("expected repo1, got %s", results[0].Name)
	}
}

func TestScanWorktree(t *testing.T) {
	dir := t.TempDir()
	gitFile := filepath.Join(dir, "worktree-repo", ".git")
	if err := os.MkdirAll(filepath.Dir(gitFile), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(gitFile, []byte("gitdir: /some/where/.git/worktrees/foo"), 0644); err != nil {
		t.Fatalf("write .git file: %v", err)
	}

	results, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "worktree-repo" {
		t.Errorf("expected worktree-repo, got %s", results[0].Name)
	}
}

func TestScanMultipleRepos(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"repo-a", "repo-b", "repo-c"} {
		gitDir := filepath.Join(dir, name, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", gitDir, err)
		}
	}

	results, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestScanSkipsNestedGit(t *testing.T) {
	dir := t.TempDir()
	// Create a repo with a nested .git inside submodules
	outer := filepath.Join(dir, "outer", ".git")
	inner := filepath.Join(dir, "outer", "submodule", ".git")
	if err := os.MkdirAll(outer, 0755); err != nil {
		t.Fatalf("mkdir outer: %v", err)
	}
	if err := os.MkdirAll(inner, 0755); err != nil {
		t.Fatalf("mkdir inner: %v", err)
	}

	results, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	// Should find both since outer/.git is skipped but inner is already inside outer
	// Actually, WalkDir would skip the outer/.git directory, but inner/.git is inside
	// outer/ (not inside .git/). Both should be found.
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestToModels(t *testing.T) {
	results := []Result{
		{Path: "/tmp/a", Name: "a"},
		{Path: "/tmp/b", Name: "b"},
	}
	repos := ToModels(results, "main")
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].DefaultBranch != "main" {
		t.Errorf("expected main, got %s", repos[0].DefaultBranch)
	}
	if repos[0].Path != "/tmp/a" {
		t.Errorf("expected /tmp/a, got %s", repos[0].Path)
	}
}

func TestToModelsDedup(t *testing.T) {
	results := []Result{
		{Path: "/tmp/a", Name: "a"},
		{Path: "/tmp/a/", Name: "a"},
	}
	repos := ToModels(results, "main")
	if len(repos) != 1 {
		t.Errorf("expected 1 repo after dedup, got %d", len(repos))
	}
}

func TestScanNonexistentPath(t *testing.T) {
	results, err := Scan("/nonexistent/path/that/does/not/exist")
	if err != nil {
		// err is acceptable — some platforms report it, some don't
		return
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent path, got %d", len(results))
	}
}
