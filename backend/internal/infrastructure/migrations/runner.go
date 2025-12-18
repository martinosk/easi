package migrations

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

var validFilenamePattern = regexp.MustCompile(`^[a-zA-Z0-9._]+\.sql$`)

func validateMigrationFilename(filename string) error {
	if !validFilenamePattern.MatchString(filename) {
		return errors.New("migration filename must end with .sql and contain only alphanumeric, dot, and underscore characters")
	}
	return nil
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

func getSortedSQLFiles(dirPath string) ([]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}
		sqlFiles = append(sqlFiles, file.Name())
	}

	sort.Strings(sqlFiles)
	return sqlFiles, nil
}

func (r *Runner) getMigrationFiles() ([]string, error) {
	files, err := getSortedSQLFiles(r.migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}
	return files, nil
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
	if err := validateMigrationFilename(filename); err != nil {
		return fmt.Errorf("invalid migration filename: %w", err)
	}

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
	defer func() { _ = tx.Rollback() }()

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

// RunAlwaysScripts executes all SQL scripts in the given directory on every run.
// Scripts are executed in sorted order. Environment variables in the format
// ${VAR_NAME} are substituted before execution.
func (r *Runner) RunAlwaysScripts(scriptsPath string) error {
	if _, err := os.Stat(scriptsPath); os.IsNotExist(err) {
		return nil
	}

	scriptFiles, err := getSortedSQLFiles(scriptsPath)
	if err != nil {
		return fmt.Errorf("failed to read scripts directory: %w", err)
	}

	for _, file := range scriptFiles {
		if err := r.executeAlwaysScript(scriptsPath, file); err != nil {
			return fmt.Errorf("failed to execute script %s: %w", file, err)
		}
		log.Printf("Successfully executed run-always script: %s", file)
	}

	return nil
}

func (r *Runner) executeAlwaysScript(scriptsPath, filename string) error {
	if err := validateMigrationFilename(filename); err != nil {
		return fmt.Errorf("invalid script filename: %w", err)
	}

	filePath := filepath.Join(scriptsPath, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read script file: %w", err)
	}

	sqlContent := substituteEnvVars(string(content))

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(sqlContent); err != nil {
		return fmt.Errorf("failed to execute script SQL: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func substituteEnvVars(content string) string {
	return envVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		varName := envVarPattern.FindStringSubmatch(match)[1]
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match
	})
}
