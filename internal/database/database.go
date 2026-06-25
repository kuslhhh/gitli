package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

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
