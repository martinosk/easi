package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	migrationsTable = "schema_migrations"
)

// Runner handles database migrations
type Runner struct {
	db             *sql.DB
	migrationsPath string
}

// NewRunner creates a new migration runner
func NewRunner(db *sql.DB, migrationsPath string) *Runner {
	return &Runner{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// Run executes all pending migrations
func (r *Runner) Run() error {
	// Create migrations tracking table if it doesn't exist
	if err := r.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := r.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get already executed migrations
	executedMigrations, err := r.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Execute pending migrations
	for _, file := range migrationFiles {
		if executedMigrations[file] {
			log.Printf("Skipping already executed migration: %s", file)
			continue
		}

		if err := r.executeMigration(file); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		log.Printf("Successfully executed migration: %s", file)
	}

	log.Printf("All migrations completed successfully")
	return nil
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func (r *Runner) createMigrationsTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			migration_name VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`, migrationsTable)

	_, err := r.db.Exec(query)
	return err
}

// getMigrationFiles returns a sorted list of migration files
func (r *Runner) getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir(r.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only process .sql files
		if strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Sort files to ensure they run in order
	sort.Strings(migrationFiles)

	return migrationFiles, nil
}

// getExecutedMigrations returns a map of already executed migration names
func (r *Runner) getExecutedMigrations() (map[string]bool, error) {
	query := fmt.Sprintf("SELECT migration_name FROM %s", migrationsTable)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		executed[name] = true
	}

	return executed, rows.Err()
}

// executeMigration executes a single migration file
func (r *Runner) executeMigration(filename string) error {
	// Read migration file
	filePath := filepath.Join(r.migrationsPath, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration in tracking table
	query := fmt.Sprintf("INSERT INTO %s (migration_name) VALUES ($1)", migrationsTable)
	if _, err := tx.Exec(query, filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
