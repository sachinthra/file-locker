package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq" // PostgreSQL driver
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations executes all pending database migrations
// dbURL format: postgres://username:password@host:port/database?sslmode=disable
func RunMigrations(dbURL string, logger *slog.Logger) error {
	logger.Info("Starting database migrations")

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

	// Create migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create source from embedded filesystem
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		logger.Warn("Database is in dirty state, attempting to force version",
			slog.Uint64("version", uint64(version)))
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("failed to force migration version: %w", err)
		}
	}

	logger.Info("Current migration version",
		slog.Uint64("version", uint64(version)),
		slog.Bool("dirty", dirty))

	// Run all up migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("No new migrations to apply")
	} else {
		newVersion, _, _ := m.Version()
		logger.Info("Migrations applied successfully",
			slog.Uint64("new_version", uint64(newVersion)))
	}

	return nil
}
