package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateDefaultAdmin creates the default admin user if it doesn't exist
func CreateDefaultAdmin(dbURL string, username, email, password string, logger *slog.Logger) error {
	logger.Info("Checking default admin user")

	// Open database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Check if admin already exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check admin existence: %w", err)
	}

	if exists {
		logger.Info("Default admin user already exists", slog.String("username", username))
		return nil
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	adminID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role) 
		VALUES ($1, $2, $3, $4, 'admin')
	`, adminID, username, email, string(hashedPassword))

	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	logger.Info("✅ Default admin user created successfully",
		slog.String("username", username),
		slog.String("email", email))

	return nil
}
