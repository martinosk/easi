package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"easi/backend/internal/infrastructure/migrations"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting database migration...")

	// Database connection using admin credentials
	// Use DB_ADMIN_CONN_STRING for admin connection, fallback to DB_CONN_STRING
	connStr := getEnv("DB_ADMIN_CONN_STRING", "")
	if connStr == "" {
		connStr = getEnv("DB_CONN_STRING", "")
	}

	// Extract target database name from connection string
	targetDB, err := extractDatabaseName(connStr)
	if err != nil {
		log.Fatalf("Failed to extract database name: %v", err)
	}

	// Ensure the target database exists
	if err := ensureDatabaseExists(connStr, targetDB); err != nil {
		log.Fatalf("Failed to ensure database exists: %v", err)
	}

	// Connect to the target database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	migrationRunner := migrations.NewRunner(db, "./deploy-scripts/migrations")

	if err := migrationRunner.RunAlwaysScripts("./deploy-scripts/pre"); err != nil {
		log.Fatalf("Pre-migration scripts failed: %v", err)
	}

	if err := migrationRunner.Run(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	if err := migrationRunner.RunAlwaysScripts("./deploy-scripts/post"); err != nil {
		log.Fatalf("Post-migration scripts failed: %v", err)
	}

	log.Println("All migrations completed successfully")
}

// extractDatabaseName extracts the database name from a PostgreSQL connection string
func extractDatabaseName(connStr string) (string, error) {
	// Handle both URL format and key=value format
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		u, err := url.Parse(connStr)
		if err != nil {
			return "", err
		}
		dbName := strings.TrimPrefix(u.Path, "/")
		if dbName == "" {
			return "postgres", nil
		}
		return dbName, nil
	}

	// Parse key=value format
	parts := strings.Fields(connStr)
	for _, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname="), nil
		}
	}

	return "postgres", nil
}

// ensureDatabaseExists connects to the postgres database and creates the target database if needed
func ensureDatabaseExists(connStr, targetDB string) error {
	if targetDB == "postgres" {
		// Already connecting to postgres, no need to create
		return nil
	}

	// Modify connection string to connect to postgres database
	postgresConnStr, err := replaceDatabaseName(connStr, "postgres")
	if err != nil {
		return fmt.Errorf("failed to create postgres connection string: %w", err)
	}

	log.Printf("Connecting to postgres database to check if '%s' exists...", targetDB)

	// Connect to postgres database
	db, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping postgres database: %w", err)
	}

	// Check if database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	if err := db.QueryRow(query, targetDB).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		log.Printf("Database '%s' already exists", targetDB)
		return nil
	}

	// Create database
	log.Printf("Creating database '%s'...", targetDB)
	createQuery := fmt.Sprintf("CREATE DATABASE %s", targetDB)
	if _, err := db.Exec(createQuery); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database '%s' created successfully", targetDB)
	return nil
}

// replaceDatabaseName replaces the database name in a connection string
func replaceDatabaseName(connStr, newDB string) (string, error) {
	// Handle URL format
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		u, err := url.Parse(connStr)
		if err != nil {
			return "", err
		}
		u.Path = "/" + newDB
		return u.String(), nil
	}

	// Handle key=value format
	parts := strings.Fields(connStr)
	var newParts []string
	foundDB := false

	for _, part := range parts {
		if strings.HasPrefix(part, "dbname=") {
			newParts = append(newParts, "dbname="+newDB)
			foundDB = true
		} else {
			newParts = append(newParts, part)
		}
	}

	if !foundDB {
		newParts = append(newParts, "dbname="+newDB)
	}

	return strings.Join(newParts, " "), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
