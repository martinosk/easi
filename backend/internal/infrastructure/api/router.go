package api

import (
	"log"
	"net/http"
	"os"

	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	authAPI "easi/backend/internal/auth/infrastructure/api"
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

type routerDependencies struct {
	eventStore eventstore.EventStore
	db         *database.TenantAwareDB
	authDeps   *authAPI.AuthDependencies
	commandBus *cqrs.InMemoryCommandBus
	eventBus   *events.InMemoryEventBus
	hateoas    *sharedAPI.HATEOASLinks
}

// NewRouter creates and configures the HTTP router
func NewRouter(eventStore eventstore.EventStore, db *database.TenantAwareDB) http.Handler {
	r := chi.NewRouter()

	deps := initializeDependencies(eventStore, db)
	configureMiddleware(r, deps.authDeps)
	registerPublicRoutes(r, db, deps.authDeps)
	registerTenantRoutes(r, deps)

	return r
}

func initializeDependencies(eventStore eventstore.EventStore, db *database.TenantAwareDB) routerDependencies {
	authDeps, err := authAPI.SetupAuthDependencies(db.DB())
	if err != nil {
		log.Fatalf("Failed to setup auth dependencies: %v", err)
	}

	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()

	if pgStore, ok := eventStore.(*eventstore.PostgresEventStore); ok {
		pgStore.SetEventBus(eventBus)
	}

	return routerDependencies{
		eventStore: eventStore,
		db:         db,
		authDeps:   authDeps,
		commandBus: commandBus,
		eventBus:   eventBus,
		hateoas:    sharedAPI.NewHATEOASLinks("/api/v1"),
	}
}

func configureMiddleware(r chi.Router, authDeps *authAPI.AuthDependencies) {
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
	r.Use(authDeps.SCSManager.LoadAndSave)
}

func registerPublicRoutes(r chi.Router, db *database.TenantAwareDB, authDeps *authAPI.AuthDependencies) {
	r.Get("/health", healthHandler)
	r.Get("/api/v1/version", versionHandler)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("doc.json")))

	mustSetup(platformAPI.SetupPlatformRoutes(r, db.DB()), "platform routes")
	mustSetup(authAPI.SetupAuthRoutes(r, db.DB(), authDeps), "auth routes")
}

func registerTenantRoutes(r chi.Router, deps routerDependencies) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.TenantMiddlewareWithSession(deps.authDeps.SessionManager))

		mustSetup(architectureAPI.SetupArchitectureModelingRoutes(r, deps.commandBus, deps.eventStore, deps.eventBus, deps.db, deps.hateoas), "architecture modeling routes")
		mustSetup(viewsAPI.SetupArchitectureViewsRoutes(r, deps.commandBus, deps.eventStore, deps.eventBus, deps.db, deps.hateoas), "architecture views routes")
		mustSetup(capabilityAPI.SetupCapabilityMappingRoutes(r, deps.commandBus, deps.eventStore, deps.eventBus, deps.db, deps.hateoas), "capability mapping routes")
		mustSetup(releasesAPI.SetupReleasesRoutes(r, deps.db.DB()), "releases routes")
		mustSetup(viewlayoutsAPI.SetupViewLayoutsRoutes(r, deps.eventBus, deps.db, deps.hateoas), "view layouts routes")
		mustSetup(importingAPI.SetupImportingRoutes(r, deps.commandBus, deps.eventStore, deps.eventBus, deps.db), "importing routes")

		sharedAPI.SetupReferenceRoutes(r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	sharedAPI.RespondJSON(w, http.StatusOK, map[string]string{"version": appVersion})
}

func mustSetup(err error, name string) {
	if err != nil {
		log.Fatalf("Failed to setup %s: %v", name, err)
	}
}
