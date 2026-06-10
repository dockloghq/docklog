package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return err
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT,
		is_admin BOOLEAN,
		password_changed BOOLEAN DEFAULT 0,
		can_start BOOLEAN DEFAULT 0,
		can_stop BOOLEAN DEFAULT 0,
		can_restart BOOLEAN DEFAULT 0,
		can_delete BOOLEAN DEFAULT 0,
		can_shell BOOLEAN DEFAULT 0,
		is_restricted_access BOOLEAN DEFAULT 1,
		allowed_containers TEXT DEFAULT '.*',
		is_active BOOLEAN DEFAULT 1,
		password_version INTEGER DEFAULT 1
	);
	CREATE TABLE IF NOT EXISTS stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		container_id TEXT,
		cpu REAL,
		memory INTEGER,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_stats_container_time ON stats(container_id, timestamp);
	CREATE TABLE IF NOT EXISTS system_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cpu REAL,
		memory INTEGER,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_system_stats_time ON system_stats(timestamp);
	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		username TEXT,
		action TEXT,
		resource TEXT,
		status TEXT,
		message TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_time ON audit_logs(timestamp);
	`
	_, err = DB.Exec(schema)
	if err != nil {
		return err
	}

	// Auto-migration: add password_version column for existing databases
	_, migErr := DB.Exec("ALTER TABLE users ADD COLUMN password_version INTEGER DEFAULT 1")
	if migErr != nil {
		// Column likely already exists — safe to ignore
		log.Printf("Migration note (safe to ignore): %v", migErr)
	}

	_, migErr = DB.Exec("ALTER TABLE users ADD COLUMN can_shell BOOLEAN DEFAULT 0")
	if migErr != nil {
		log.Printf("Migration note (safe to ignore): %v", migErr)
	}

	return nil
}

