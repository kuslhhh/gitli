package models

import (
	"testing"
	"time"
)

func TestRepositoryDefaults(t *testing.T) {
	r := Repository{
		Name:          "test-repo",
		Path:          "/tmp/test-repo",
		DefaultBranch: "main",
	}
	if r.ID != 0 {
		t.Errorf("expected zero ID, got %d", r.ID)
	}
	if r.Name != "test-repo" {
		t.Errorf("expected test-repo, got %s", r.Name)
	}
	if r.DefaultBranch != "main" {
		t.Errorf("expected main, got %s", r.DefaultBranch)
	}
}

func TestRepositoryWithLastScanned(t *testing.T) {
	now := time.Now()
	r := Repository{
		Name:          "repo",
		Path:          "/tmp/repo",
		DefaultBranch: "master",
		LastScannedAt: now,
	}
	if r.LastScannedAt.IsZero() {
		t.Error("expected non-zero LastScannedAt")
	}
}

func TestCommitFields(t *testing.T) {
	now := time.Now()
	c := Commit{
		Hash:        "abc123",
		Author:      "Alice",
		Email:       "alice@example.com",
		Message:     "initial commit",
		CommittedAt: now,
	}
	if c.Hash != "abc123" {
		t.Errorf("expected abc123, got %s", c.Hash)
	}
	if c.Author != "Alice" {
		t.Errorf("expected Alice, got %s", c.Author)
	}
}

func TestBranchDefaults(t *testing.T) {
	b := Branch{Name: "feature-x", IsCurrent: true}
	if !b.IsCurrent {
		t.Error("expected IsCurrent to be true")
	}
	b2 := Branch{Name: "main", IsCurrent: false}
	if b2.IsCurrent {
		t.Error("expected IsCurrent to be false")
	}
}

func TestStashFields(t *testing.T) {
	s := Stash{StashName: "WIP on main: abc123 work in progress"}
	if s.StashName == "" {
		t.Error("expected non-empty StashName")
	}
}
