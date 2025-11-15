package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"easi/backend/internal/infrastructure/migrations"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting database migration...")

	// Database connection using admin credentials
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_ADMIN_USER", getEnv("DB_USER", "easi"))
	dbPassword := getEnv("DB_ADMIN_PASSWORD", getEnv("DB_PASSWORD", "easi"))
	dbName := getEnv("DB_NAME", "easi")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Connected to database at %s:%s as user %s", dbHost, dbPort, dbUser)

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
