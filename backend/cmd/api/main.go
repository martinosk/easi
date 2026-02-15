package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// @schemes http https
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Session-based authentication using httpOnly cookies. Obtain session via /auth/sessions endpoint.
func main() {
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run() error {
	// Database connection using app credentials
	connStr := getEnv("DB_CONN_STRING", "")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database successfully")

	// Wrap database connection with tenant-aware wrapper for RLS
	tenantDB := database.NewTenantAwareDB(db)

	// Initialize event store with tenant-aware DB
	eventStore := eventstore.NewPostgresEventStore(tenantDB)
	appContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create HTTP server with tenant-aware DB
	router := api.NewRouter(appContext, eventStore, tenantDB)

	port := getEnv("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		<-appContext.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown failed: %v", err)
		}
	}()

	log.Printf("Server starting on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
