package api

import (
	"log"
	"net/http"
	"os"

	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	capabilityAPI "easi/backend/internal/capabilitymapping/infrastructure/api"
	importingAPI "easi/backend/internal/importing/infrastructure/api"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	platformAPI "easi/backend/internal/platform/infrastructure/api"
	releasesAPI "easi/backend/internal/releases/infrastructure/api"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	viewlayoutsAPI "easi/backend/internal/viewlayouts/infrastructure/api"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

var Version = "0.7.0" // Set via ldflags at build time: -ldflags "-X 'easi/backend/internal/infrastructure/api.Version=x.y.z'"

var appVersion = getEnv("APP_VERSION", Version)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-ID", "X-Platform-Admin-Key"},
		ExposedHeaders:   []string{"Link", "Location"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Version endpoint (outside /api/v1 to avoid tenant middleware)
	r.Get("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		sharedAPI.RespondJSON(w, http.StatusOK, map[string]string{
			"version": appVersion,
		})
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	// Platform API routes (no tenant context - uses API key authentication)
	if err := platformAPI.SetupPlatformRoutes(r, db.DB()); err != nil {
		log.Fatalf("Failed to setup platform routes: %v", err)
	}

	// Initialize CQRS buses and event bus
	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	hateoas := sharedAPI.NewHATEOASLinks("/api/v1")

	// Wire event store to event bus
	if pgStore, ok := eventStore.(*eventstore.PostgresEventStore); ok {
		pgStore.SetEventBus(eventBus)
	}

	// Tenant-scoped API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Tenant context middleware - injects tenant from header (dev) or OAuth (prod)
		r.Use(middleware.TenantMiddleware())
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

		// Releases Context (system-wide, no tenancy)
		if err := releasesAPI.SetupReleasesRoutes(r, db.DB()); err != nil {
			log.Fatalf("Failed to setup releases routes: %v", err)
		}

		// ViewLayouts Context
		if err := viewlayoutsAPI.SetupViewLayoutsRoutes(r, eventBus, db, hateoas); err != nil {
			log.Fatalf("Failed to setup view layouts routes: %v", err)
		}

		// Importing Context
		if err := importingAPI.SetupImportingRoutes(r, commandBus, eventStore, eventBus, db); err != nil {
			log.Fatalf("Failed to setup importing routes: %v", err)
		}

		// Reference Documentation (static)
		sharedAPI.SetupReferenceRoutes(r)
	})

	return r
}
