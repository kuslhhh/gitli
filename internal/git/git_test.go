package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func gitAdd(t *testing.T, dir string, files ...string) {
	t.Helper()
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
}

func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", msg)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@test.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

func gitCheckout(t *testing.T, dir, branch string) {
	t.Helper()
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout: %v\n%s", err, out)
	}
}

func gitStash(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "stash")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git stash: %v\n%s", err, out)
	}
}

func TestGetDefaultBranch(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	branch, err := GetDefaultBranch(dir)
	if err != nil {
		t.Fatalf("GetDefaultBranch: %v", err)
	}
	// git init defaults vary; just check it's non-empty
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
}

func TestGetBranches(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	gitCommit(t, dir, "initial")
	gitCheckout(t, dir, "feature-x")

	branches, err := GetBranches(dir)
	if err != nil {
		t.Fatalf("GetBranches: %v", err)
	}
	if len(branches) == 0 {
		t.Fatal("expected at least 1 branch")
	}

	found := false
	for _, b := range branches {
		if b.Name == "feature-x" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected feature-x branch among %v", branches)
	}
}

func TestGetCommits(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	gitCommit(t, dir, "first commit")
	gitCommit(t, dir, "second commit")
	gitCommit(t, dir, "third commit")

	commits, err := GetCommits(dir, 10)
	if err != nil {
		t.Fatalf("GetCommits: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(commits))
	}
	// Commits are returned newest first
	if commits[0].Message != "third commit" {
		t.Errorf("expected 'third commit', got '%s'", commits[0].Message)
	}
	if commits[0].Author != "Test" {
		t.Errorf("expected Test, got '%s'", commits[0].Author)
	}
	if commits[0].Email != "test@test.com" {
		t.Errorf("expected test@test.com, got '%s'", commits[0].Email)
	}
}

func TestGetCommitsMaxCount(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	for i := 0; i < 5; i++ {
		gitCommit(t, dir, "commit")
	}

	commits, err := GetCommits(dir, 2)
	if err != nil {
		t.Fatalf("GetCommits: %v", err)
	}
	if len(commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(commits))
	}
}

func TestGetStatusClean(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	gitCommit(t, dir, "initial")

	status, err := GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if len(status) != 0 {
		t.Errorf("expected clean status, got %d entries", len(status))
	}
}

func TestGetStatusDirty(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	gitCommit(t, dir, "initial")
	writeFile(t, dir, "untracked.txt", "hello")

	status, err := GetStatus(dir)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if len(status) == 0 {
		t.Fatal("expected dirty status")
	}
}

func TestIsDirty(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	gitCommit(t, dir, "initial")

	dirty, err := IsDirty(dir)
	if err != nil {
		t.Fatalf("IsDirty: %v", err)
	}
	if dirty {
		t.Error("expected clean repo")
	}

	writeFile(t, dir, "new.txt", "content")
	dirty, err = IsDirty(dir)
	if err != nil {
		t.Fatalf("IsDirty: %v", err)
	}
	if !dirty {
		t.Error("expected dirty repo")
	}
}

func TestGetStashes(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	gitCommit(t, dir, "initial")
	writeFile(t, dir, "work.txt", "wip")
	gitAdd(t, dir, "work.txt")

	stashes, err := GetStashes(dir)
	if err != nil {
		t.Fatalf("GetStashes: %v", err)
	}
	if len(stashes) != 0 {
		t.Errorf("expected 0 stashes initially, got %d", len(stashes))
	}

	gitStash(t, dir)

	stashes, err = GetStashes(dir)
	if err != nil {
		t.Fatalf("GetStashes: %v", err)
	}
	if len(stashes) != 1 {
		t.Errorf("expected 1 stash, got %d", len(stashes))
	}
}

func TestStashCount(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	gitCommit(t, dir, "initial")

	count, err := StashCount(dir)
	if err != nil {
		t.Fatalf("StashCount: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestGetCommitsEmptyRepo(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)

	commits, err := GetCommits(dir, 10)
	if err != nil {
		t.Fatalf("GetCommits: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("expected 0 commits, got %d", len(commits))
	}
}
