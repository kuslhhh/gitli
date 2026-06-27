package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kush/gitli/internal/models"
)

func runGit(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\nstderr: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// GetBranches returns all branches for a repository.
// The current branch is marked by a leading '*' in `git branch` output.
func GetBranches(repoPath string) ([]models.Branch, error) {
	out, err := runGit(repoPath, "branch")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	branches := make([]models.Branch, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isCurrent := strings.HasPrefix(line, "* ")
		name := strings.TrimPrefix(line, "* ")
		name = strings.TrimPrefix(name, "  ")

		if name == "" {
			continue
		}

		branches = append(branches, models.Branch{
			Name:      name,
			IsCurrent: isCurrent,
		})
	}

	return branches, nil
}

// GetCommits returns recent commits from all branches.
// Uses a pipe-delimited format for reliable parsing.
func GetCommits(repoPath string, maxCount int) ([]models.Commit, error) {
	if maxCount <= 0 {
		maxCount = 1000
	}

	out, err := runGit(repoPath, "log", "--all", fmt.Sprintf("--max-count=%d", maxCount),
		"--format=%H|%an|%ae|%s|%aI")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	commits := make([]models.Commit, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		committedAt, err := time.Parse(time.RFC3339, parts[4])
		if err != nil {
			continue
		}

		commits = append(commits, models.Commit{
			Hash:        parts[0],
			Author:      parts[1],
			Email:       parts[2],
			Message:     parts[3],
			CommittedAt: committedAt,
		})
	}

	return commits, nil
}

// GetStashes returns all stashes for a repository.
func GetStashes(repoPath string) ([]models.Stash, error) {
	out, err := runGit(repoPath, "stash", "list", "--format=%gD|%gs")
	if err != nil {
		// `git stash list` may return non-zero on some git versions when empty.
		// Treat any error as "no stashes" rather than failing entirely.
		return nil, nil
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}

	lines := strings.Split(out, "\n")
	stashes := make([]models.Stash, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		name := parts[0]
		if len(parts) > 1 {
			name = parts[1] // Use the description as stash name
		}

		stashes = append(stashes, models.Stash{
			StashName: name,
		})
	}

	return stashes, nil
}

// StatusEntry represents a single file status from `git status --porcelain`.
type StatusEntry struct {
	X    string // index status
	Y    string // worktree status
	Path string // file path
}

// GetStatus returns the working tree status.
func GetStatus(repoPath string) ([]StatusEntry, error) {
	out, err := runGit(repoPath, "status", "--porcelain")
	if err != nil {
		return nil, err
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return nil, nil
	}

	lines := strings.Split(out, "\n")
	entries := make([]StatusEntry, 0, len(lines))

	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		entries = append(entries, StatusEntry{
			X:    string(line[0]),
			Y:    string(line[1]),
			Path: strings.TrimSpace(line[2:]),
		})
	}

	return entries, nil
}

// GetDefaultBranch returns the name of the currently checked-out branch (HEAD).
func GetDefaultBranch(repoPath string) (string, error) {
	out, err := runGit(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "main", nil // fallback
	}

	name := strings.TrimSpace(out)
	if name == "" || name == "HEAD" {
		return "main", nil // detached HEAD
	}
	return name, nil
}

// IsDirty returns true if the repository has uncommitted changes.
func IsDirty(repoPath string) (bool, error) {
	entries, err := GetStatus(repoPath)
	if err != nil {
		return false, err
	}
	return len(entries) > 0, nil
}

// StashCount returns the number of stashes.
func StashCount(repoPath string) (int, error) {
	stashes, err := GetStashes(repoPath)
	if err != nil {
		return 0, err
	}
	return len(stashes), nil
}
