package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/easi/backend/internal/infrastructure/api"
	"github.com/easi/backend/internal/infrastructure/eventstore"
	_ "github.com/lib/pq"
)

// @title EASI Architecture Modeling API
// @version 1.0
// @description API for Enterprise Architecture System Integration - Component Modeling
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.easi.io/support
// @contact.email support@easi.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "easi")
	dbPassword := getEnv("DB_PASSWORD", "easi")
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

	log.Println("Connected to database successfully")

	// Initialize event store
	eventStore := eventstore.NewPostgresEventStore(db)

	// Initialize event store schema
	if err := eventStore.InitializeSchema(); err != nil {
		log.Fatalf("Failed to initialize event store schema: %v", err)
	}

	// Create HTTP server
	router := api.NewRouter(eventStore, db)

	port := getEnv("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
