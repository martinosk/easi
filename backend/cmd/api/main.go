package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "easi/backend/docs"
	"easi/backend/internal/infrastructure/api"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	_ "easi/backend/internal/shared/api"
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
	// Database connection using app credentials
	connStr := getEnv("DB_CONN_STRING", "")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Wrap database connection with tenant-aware wrapper for RLS
	tenantDB := database.NewTenantAwareDB(db)

	// Initialize event store with tenant-aware DB
	eventStore := eventstore.NewPostgresEventStore(tenantDB)

	// Create HTTP server with tenant-aware DB
	router := api.NewRouter(eventStore, tenantDB)

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
