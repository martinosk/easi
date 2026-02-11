package api

import (
	"log"
	"net/http"
	"os"

	"easi/backend/docs"
	adPL "easi/backend/internal/accessdelegation/publishedlanguage"
	accessdelegationAPI "easi/backend/internal/accessdelegation/infrastructure/api"
	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	authProjectors "easi/backend/internal/auth/application/projectors"
	authAPI "easi/backend/internal/auth/infrastructure/api"
	authReadModels "easi/backend/internal/auth/application/readmodels"
	capabilityAPI "easi/backend/internal/capabilitymapping/infrastructure/api"
	valuestreamsAPI "easi/backend/internal/valuestreams/infrastructure/api"
	enterpriseArchAPI "easi/backend/internal/enterprisearchitecture/infrastructure/api"
	importingAPI "easi/backend/internal/importing/infrastructure/api"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	metamodelAPI "easi/backend/internal/metamodel/infrastructure/api"
	platformAPI "easi/backend/internal/platform/infrastructure/api"
	releasesAPI "easi/backend/internal/releases/infrastructure/api"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/audit"
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
	eventStore    eventstore.EventStore
	db            *database.TenantAwareDB
	authDeps      *authAPI.AuthDependencies
	commandBus    *cqrs.InMemoryCommandBus
	eventBus      *events.InMemoryEventBus
	hateoas       *sharedAPI.HATEOASLinks
	userReadModel *authReadModels.UserReadModel
}

// NewRouter creates and configures the HTTP router
func NewRouter(eventStore eventstore.EventStore, db *database.TenantAwareDB) http.Handler {
	r := chi.NewRouter()

	deps := initializeDependencies(eventStore, db)
	configureMiddleware(r, deps.authDeps)
	registerRootRoutes(r)
	registerAPIRoutes(r, deps)

	return r
}

func initializeDependencies(eventStore eventstore.EventStore, db *database.TenantAwareDB) routerDependencies {
	authDeps, err := authAPI.SetupAuthDependencies(db.DB())
	if err != nil {
		log.Fatalf("Failed to setup auth dependencies: %v", err)
	}

	commandBus := cqrs.NewInMemoryCommandBus()
	eventBus := events.NewInMemoryEventBus()
	userReadModel := authReadModels.NewUserReadModel(db)

	if pgStore, ok := eventStore.(*eventstore.PostgresEventStore); ok {
		pgStore.SetEventBus(eventBus)
	}

	return routerDependencies{
		eventStore:    eventStore,
		db:            db,
		authDeps:      authDeps,
		commandBus:    commandBus,
		eventBus:      eventBus,
		hateoas:       sharedAPI.NewHATEOASLinks("/api/v1"),
		userReadModel: userReadModel,
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
			r.Use(middleware.TenantMiddlewareWithSession(deps.authDeps.SessionManager, deps.userReadModel))
			registerTenantRoutes(r, deps)
		})
	})
}

func registerPublicRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(platformAPI.SetupPlatformRoutes(platformAPI.PlatformRoutesDeps{
		Router:     r,
		RawDB:      deps.db.DB(),
		TenantDB:   deps.db,
		CommandBus: deps.commandBus,
		EventStore: deps.eventStore,
	}), "platform routes")
	mustSetup(authAPI.SetupAuthRoutes(r, deps.db.DB(), deps.authDeps), "auth routes")
}

func registerTenantRoutes(r chi.Router, deps routerDependencies) {
	adDeps := setupAccessDelegation(deps)
	r.Use(middleware.EditGrantEnrichment(adDeps.GrantResolver))
	adDeps.RegisterRoutes(r)
	setupModelingRoutes(r, deps)
	setupDomainRoutes(r, deps)
	setupValueStreamsRoutes(r, deps)
	setupSupportRoutes(r, deps)
	setupAuthRoutes(r, deps)
	wireAutoInvitationProjector(deps)
}

func setupAccessDelegation(deps routerDependencies) *accessdelegationAPI.AccessDelegationDependencies {
	invReadModel := authReadModels.NewInvitationReadModel(deps.db)
	domainChecker := authReadModels.NewTenantDomainChecker(deps.db)

	adDeps, adErr := accessdelegationAPI.SetupAccessDelegationRoutes(accessdelegationAPI.AccessDelegationRoutesDeps{
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
		UserReadModel:  deps.userReadModel,
		InvReadModel:   invReadModel,
		DomainChecker:  domainChecker,
	})
	mustSetup(adErr, "access delegation routes")
	return adDeps
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

	viewsAPI.SubscribeEvents(deps.eventBus, deps.commandBus, deps.db)
	viewsAPI.RegisterCommands(deps.commandBus, deps.eventStore, deps.db)
	viewHandlers := viewsAPI.NewHTTPHandlers(deps.commandBus, deps.db, deps.hateoas)
	viewsAPI.RegisterRoutes(r, viewHandlers, deps.authDeps.AuthMiddleware)

	mustSetup(capabilityAPI.SetupCapabilityMappingRoutes(&capabilityAPI.RouteConfig{
		Router:         r,
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		SessionManager: deps.authDeps.SessionManager,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
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
		Router:         r,
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
		SessionManager: deps.authDeps.SessionManager,
	}), "enterprise architecture routes")

	viewlayoutsAPI.SubscribeEvents(deps.eventBus, deps.db)
	viewlayoutsAPI.RegisterRoutes(r, deps.db, deps.hateoas, deps.authDeps.AuthMiddleware)

	mustSetup(metamodelAPI.SetupMetaModelRoutes(metamodelAPI.MetaModelRoutesDeps{
		Router:         r,
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		Hateoas:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
		SessionManager: deps.authDeps.SessionManager,
	}), "metamodel routes")
}

func setupSupportRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(releasesAPI.SetupReleasesRoutes(r, deps.db.DB()), "releases routes")
	mustSetup(importingAPI.SetupImportingRoutes(r, deps.commandBus, deps.eventStore, deps.eventBus, deps.db), "importing routes")
	sharedAPI.SetupReferenceRoutes(r)
	mustSetup(audit.SetupAuditRoutes(audit.AuditRoutesDeps{
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

func wireAutoInvitationProjector(deps routerDependencies) {
	projector := authProjectors.NewInvitationAutoCreateProjector(deps.commandBus)
	deps.eventBus.Subscribe(adPL.EditGrantForNonUserCreated, projector)
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
