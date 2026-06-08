package main

import (
	"os"
	"path/filepath"
	"testing"

	"docklog/db"

	"golang.org/x/crypto/bcrypt"
)

func TestResetPasswordCLI(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	os.Setenv("DB_PATH", dbPath)
	defer os.Unsetenv("DB_PATH")

	if err := db.InitDB(dbPath); err != nil {
		t.Fatalf("init db: %v", err)
	}
	_, err := db.DB.Exec(
		`INSERT INTO users (username, password, is_admin, password_changed, password_version) VALUES (?, ?, 1, 1, 1)`,
		"admin", "old",
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	if err := runResetPasswordCLI([]string{"admin", "newpassword1"}); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	var hash string
	var changed bool
	var version int
	err = db.DB.QueryRow(
		`SELECT password, password_changed, password_version FROM users WHERE username = 'admin'`,
	).Scan(&hash, &changed, &version)
	if err != nil {
		t.Fatalf("query user: %v", err)
	}
	if changed {
		t.Fatal("expected password_changed = 0 after CLI reset")
	}
	if version != 2 {
		t.Fatalf("expected password_version 2, got %d", version)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte("newpassword1")) != nil {
		t.Fatal("password hash does not match new password")
	}
}
