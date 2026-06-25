package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kush/gitli/internal/models"
)

// Result holds a discovered repository.
type Result struct {
	Path string
	Name string
}

// Scan walks rootPath and discovers Git repositories.
// It detects both .git/ directories (standard repos) and .git files (worktrees).
func Scan(rootPath string) ([]Result, error) {
	rootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	var results []Result

	err = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't read
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Name() != ".git" {
			return nil
		}

		// The parent directory is the repo root
		repoPath := filepath.Dir(path)
		repoName := filepath.Base(repoPath)

		// Check if it's a file (worktree) or directory (standard repo)
		if d.IsDir() {
			results = append(results, Result{
				Path: repoPath,
				Name: repoName,
			})
		} else {
			// It's a file — likely a git worktree pointer
			// Still treat the parent directory as the repo root
			results = append(results, Result{
				Path: repoPath,
				Name: repoName,
			})
		}

		// Don't walk into .git directories
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})

	return results, err
}

// ToModels converts scan results to database models with the given default branch.
func ToModels(results []Result, defaultBranch string) []models.Repository {
	repos := make([]models.Repository, 0, len(results))
	seen := make(map[string]bool)

	for _, r := range results {
		normalized := strings.TrimSuffix(r.Path, "/")
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		repos = append(repos, models.Repository{
			Name:          r.Name,
			Path:          normalized,
			DefaultBranch: defaultBranch,
		})
	}

	return repos
}
