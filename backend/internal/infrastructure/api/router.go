package api

import (
	"database/sql"
	"log"
	"net/http"

	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter creates and configures the HTTP router
func NewRouter(eventStore eventstore.EventStore, db *sql.DB) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "Location"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/swagger.json"),
	))

	// Initialize CQRS buses
	commandBus := cqrs.NewInMemoryCommandBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Architecture Modeling Context
		if err := architectureAPI.SetupArchitectureModelingRoutes(r, commandBus, eventStore, db, hateoas); err != nil {
			log.Fatalf("Failed to setup architecture modeling routes: %v", err)
		}

		// Architecture Views Context
		if err := viewsAPI.SetupArchitectureViewsRoutes(r, commandBus, eventStore, db, hateoas); err != nil {
			log.Fatalf("Failed to setup architecture views routes: %v", err)
		}
	})

	return r
}
