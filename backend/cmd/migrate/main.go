package main

import (
	"database/sql"
	"log"
	"os"

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

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Run migrations
	migrationsPath := getEnv("MIGRATIONS_PATH", "./migrations")
	migrationRunner := migrations.NewRunner(db, migrationsPath)

	if err := migrationRunner.Run(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("All migrations completed successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
