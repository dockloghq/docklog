package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"docklog/db"

	"golang.org/x/crypto/bcrypt"
)

func runResetPasswordCLI(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: docklog reset-password <username> <new-password>")
	}

	username := strings.TrimSpace(args[0])
	password := args[1]
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if !isPasswordStrongEnough(password) {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "docklog.db"
	}
	if dbPath == ":memory:" {
		return fmt.Errorf("cannot reset password on an in-memory database")
	}

	if err := db.InitDB(dbPath); err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	res, err := db.DB.Exec(
		`UPDATE users SET password = ?, password_changed = 0, password_version = COALESCE(password_version, 1) + 1 WHERE username = ?`,
		string(h), username,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user %q not found", username)
	}

	fmt.Printf("Password reset for %q. Existing sessions are invalidated.\n", username)
	return nil
}

func maybeRunCLI() bool {
	if len(os.Args) < 2 {
		return false
	}

	switch os.Args[1] {
	case "reset-password":
		if err := runResetPasswordCLI(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return true
	default:
		return false
	}
}
