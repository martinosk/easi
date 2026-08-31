package api

import (
	"context"
	"log"
	"net/http"
	"os"

	"easi/backend/docs"
	accessdelegationAPI "easi/backend/internal/accessdelegation/infrastructure/api"
	archAssistantAPI "easi/backend/internal/archassistant/infrastructure/api"
	directionAPI "easi/backend/internal/architecturedirection/infrastructure/api"
	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	auditAPI "easi/backend/internal/audit/infrastructure/api"
	authAPI "easi/backend/internal/auth/infrastructure/api"
	capabilityAPI "easi/backend/internal/capabilitymapping/infrastructure/api"
	enterpriseArchAPI "easi/backend/internal/enterprisearchitecture/infrastructure/api"
	importingAPI "easi/backend/internal/importing/infrastructure/api"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	metamodelAPI "easi/backend/internal/metamodel/infrastructure/api"
	onepagersAPI "easi/backend/internal/onepagers/infrastructure/api"
	releasesAPI "easi/backend/internal/releases/infrastructure/api"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	valuestreamsAPI "easi/backend/internal/valuestreams/infrastructure/api"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

var Version = "0.7.0"

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
	appContext context.Context
}

// NewRouter creates and configures the HTTP router
func NewRouter(appContext context.Context, eventStore eventstore.EventStore, db *database.TenantAwareDB) http.Handler {
	r := chi.NewRouter()

	deps := initializeDependencies(appContext, eventStore, db)
	configureMiddleware(r, deps.authDeps)
	registerRootRoutes(r)
	registerAPIRoutes(r, deps)

	return r
}

func initializeDependencies(appContext context.Context, eventStore eventstore.EventStore, db *database.TenantAwareDB) routerDependencies {
	if appContext == nil {
		appContext = context.Background()
	}

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
		appContext: appContext,
	}
}

func compressionMiddleware() func(http.Handler) http.Handler {
	return chimiddleware.Compress(5)
}

func configureMiddleware(r chi.Router, authDeps *authAPI.AuthDependencies) {
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.ClientIP())
	r.Use(compressionMiddleware())
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Tenant-ID", "X-Platform-Admin-Key", "If-Match"},
		ExposedHeaders:   []string{"Link", "Location", "X-Request-Id", "ETag", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(authDeps.SCSManager.LoadAndSave)
}

func registerRootRoutes(r chi.Router) {
	r.Get("/health", healthHandler)
	r.Get("/swagger/*", swaggerHandlerWithDynamicBasePath())
}

func swaggerHandlerWithDynamicBasePath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		forwardedPrefix := r.Header.Get("X-Forwarded-Prefix")
		if forwardedPrefix != "" {
			docs.SwaggerInfo.BasePath = forwardedPrefix + "/api/v1"
		} else {
			docs.SwaggerInfo.BasePath = "/api/v1"
		}
		httpSwagger.Handler(httpSwagger.URL("doc.json"))(w, r)
	}
}

func registerAPIRoutes(r chi.Router, deps routerDependencies) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/version", versionHandler)
		registerPublicRoutes(r, deps)

		r.Group(func(r chi.Router) {
			r.Use(authAPI.SessionTenantMiddleware(deps.authDeps.SessionManager, deps.db))
			registerTenantRoutes(r, deps)
		})
	})
}

func registerPublicRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(authAPI.SetupAuthRoutes(authAPI.AuthRoutesDeps{
		Router:     r,
		RawDB:      deps.db.DB(),
		TenantDB:   deps.db,
		CommandBus: deps.commandBus,
		EventBus:   deps.eventBus,
		AuthDeps:   deps.authDeps,
	}), "auth routes")
}

func registerTenantRoutes(r chi.Router, deps routerDependencies) {
	adDeps, err := accessdelegationAPI.SetupAccessDelegationRoutes(accessdelegationAPI.AccessDelegationRoutesDeps{
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
	})
	mustSetup(err, "access delegation routes")
	r.Use(middleware.EditGrantEnrichment(adDeps.GrantResolver))
	adDeps.RegisterRoutes(r)
	setupModelingRoutes(r, deps)
	setupDomainRoutes(r, deps)
	setupValueStreamsRoutes(r, deps)
	setupSupportRoutes(r, deps)
	setupArchAssistantRoutes(r, deps)
	setupAuthRoutes(r, deps)
}

func setupModelingRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(architectureAPI.SetupArchitectureModelingRoutes(architectureAPI.RouteConfig{
		Router:         r,
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
	}), "architecture modeling routes")

	mustSetup(viewsAPI.SetupArchitectureViewsRoutes(viewsAPI.RouteConfig{
		Router:         r,
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
	}), "architecture views routes")

	mustSetup(capabilityAPI.SetupCapabilityMappingRoutes(&capabilityAPI.RouteConfig{
		Router:          r,
		CommandBus:      deps.commandBus,
		EventStore:      deps.eventStore,
		EventBus:        deps.eventBus,
		DB:              deps.db,
		HATEOAS:         deps.hateoas,
		SessionProvider: deps.authDeps.SessionManager,
		AuthMiddleware:  deps.authDeps.AuthMiddleware,
	}), "capability mapping routes")
}

func setupValueStreamsRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(valuestreamsAPI.SetupValueStreamsRoutes(&valuestreamsAPI.RouteConfig{
		Router:         r,
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
	}), "value streams routes")
}

func setupDomainRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(enterpriseArchAPI.SetupEnterpriseArchitectureRoutes(enterpriseArchAPI.EnterpriseArchRoutesDeps{
		Router:          r,
		CommandBus:      deps.commandBus,
		EventStore:      deps.eventStore,
		EventBus:        deps.eventBus,
		DB:              deps.db,
		AuthMiddleware:  deps.authDeps.AuthMiddleware,
		SessionProvider: deps.authDeps.SessionManager,
	}), "enterprise architecture routes")
	mustSetup(directionAPI.SetupRoutes(directionAPI.RoutesDeps{
		Router:         r,
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
	}), "architecture direction routes")

	mustSetup(metamodelAPI.SetupMetaModelRoutes(metamodelAPI.MetaModelRoutesDeps{
		Router:          r,
		CommandBus:      deps.commandBus,
		EventStore:      deps.eventStore,
		EventBus:        deps.eventBus,
		DB:              deps.db,
		Hateoas:         deps.hateoas,
		AuthMiddleware:  deps.authDeps.AuthMiddleware,
		SessionProvider: deps.authDeps.SessionManager,
	}), "metamodel routes")

	mustSetup(onepagersAPI.SetupOnePagersRoutes(onepagersAPI.OnePagersRoutesDeps{
		Router:          r,
		CommandBus:      deps.commandBus,
		EventStore:      deps.eventStore,
		EventBus:        deps.eventBus,
		DB:              deps.db,
		Hateoas:         deps.hateoas,
		AuthMiddleware:  deps.authDeps.AuthMiddleware,
		SessionProvider: deps.authDeps.SessionManager,
	}), "one-pagers routes")
}

func setupSupportRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(releasesAPI.SetupReleasesRoutes(r, deps.db.DB()), "releases routes")
	mustSetup(importingAPI.SetupImportingRoutes(r, importingAPI.ImportingRoutesDeps{
		CommandBus:       deps.commandBus,
		EventStore:       deps.eventStore,
		EventBus:         deps.eventBus,
		DB:               deps.db,
		ExecutionContext: deps.appContext,
	}), "importing routes")
	sharedAPI.SetupReferenceRoutes(r)
	mustSetup(auditAPI.SetupAuditRoutes(auditAPI.AuditRoutesDeps{
		Router:         r,
		DB:             deps.db,
		Hateoas:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
	}), "audit routes")
}

func setupAuthRoutes(r chi.Router, deps routerDependencies) {
	invDeps, err := authAPI.SetupInvitationRoutes(authAPI.InvitationRoutesDeps{
		Router:     r,
		CommandBus: deps.commandBus,
		EventStore: deps.eventStore,
		EventBus:   deps.eventBus,
		DB:         deps.db,
		AuthDeps:   deps.authDeps,
	})
	mustSetup(err, "invitation routes")
	authAPI.WireLoginService(deps.authDeps, invDeps)

	mustSetup(authAPI.SetupUserRoutes(authAPI.UserRoutesDeps{
		Router:     r,
		CommandBus: deps.commandBus,
		EventStore: deps.eventStore,
		EventBus:   deps.eventBus,
		DB:         deps.db,
		RawDB:      deps.db.DB(),
		AuthDeps:   deps.authDeps,
		InvDeps:    invDeps,
	}), "user routes")
}

func setupArchAssistantRoutes(r chi.Router, deps routerDependencies) {
	port := getEnv("PORT", "8080")
	mustSetup(archAssistantAPI.SetupArchAssistantRoutes(archAssistantAPI.ArchAssistantRoutesDeps{
		Router:          r,
		DB:              deps.db,
		EventBus:        deps.eventBus,
		HATEOAS:         deps.hateoas,
		AuthMiddleware:  deps.authDeps.AuthMiddleware,
		LoopbackBaseURL: "http://localhost:" + port + "/api/v1",
	}), "arch assistant routes")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	sharedAPI.RespondJSON(w, http.StatusOK, map[string]string{"version": appVersion})
}

func mustSetup(err error, name string) {
	if err != nil {
		log.Fatalf("Failed to setup %s: %v", name, err)
	}
}
