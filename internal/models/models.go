package models

import "time"

type Repository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	DefaultBranch string    `json:"default_branch"`
	LastScannedAt time.Time `json:"last_scanned_at"`
}

type Commit struct {
	ID          int64     `json:"id"`
	RepoID      int64     `json:"repo_id"`
	Hash        string    `json:"hash"`
	Author      string    `json:"author"`
	Email       string    `json:"email"`
	Message     string    `json:"message"`
	CommittedAt time.Time `json:"committed_at"`
}

type Branch struct {
	ID        int64  `json:"id"`
	RepoID    int64  `json:"repo_id"`
	Name      string `json:"name"`
	IsCurrent bool   `json:"is_current"`
}

type Stash struct {
	ID        int64  `json:"id"`
	RepoID    int64  `json:"repo_id"`
	StashName string `json:"stash_name"`
}
