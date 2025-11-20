package api

import (
	"log"
	"net/http"

	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	capabilityAPI "easi/backend/internal/capabilitymapping/infrastructure/api"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter creates and configures the HTTP router
func NewRouter(eventStore eventstore.EventStore, db *database.TenantAwareDB) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link", "Location"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	// Tenant context middleware - injects tenant from header (dev) or OAuth (prod)
	r.Use(middleware.TenantMiddleware())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	// Initialize CQRS buses and event bus
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Wire event store to event bus
	if pgStore, ok := eventStore.(*eventstore.PostgresEventStore); ok {
		pgStore.SetEventBus(eventBus)
	}

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Architecture Modeling Context
		if err := architectureAPI.SetupArchitectureModelingRoutes(r, commandBus, eventStore, eventBus, db, hateoas); err != nil {
			log.Fatalf("Failed to setup architecture modeling routes: %v", err)
		}

		// Architecture Views Context
		if err := viewsAPI.SetupArchitectureViewsRoutes(r, commandBus, eventStore, eventBus, db, hateoas); err != nil {
			log.Fatalf("Failed to setup architecture views routes: %v", err)
		}

		// Capability Mapping Context
		if err := capabilityAPI.SetupCapabilityMappingRoutes(r, commandBus, eventStore, eventBus, db, hateoas); err != nil {
			log.Fatalf("Failed to setup capability mapping routes: %v", err)
		}
	})

	return r
}
