package api

import (
	"context"
	"log"
	"net/http"
	"os"

	"easi/backend/docs"
	accessdelegationAPI "easi/backend/internal/accessdelegation/infrastructure/api"
	adServices "easi/backend/internal/accessdelegation/infrastructure/services"
	adPL "easi/backend/internal/accessdelegation/publishedlanguage"
	archAssistantHandlers "easi/backend/internal/archassistant/application/handlers"
	archAssistantAdapters "easi/backend/internal/archassistant/infrastructure/adapters"
	archAssistantAPI "easi/backend/internal/archassistant/infrastructure/api"
	archAssistantRepos "easi/backend/internal/archassistant/infrastructure/repositories"
	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	archAdapters "easi/backend/internal/architecturemodeling/infrastructure/adapters"
	architectureAPI "easi/backend/internal/architecturemodeling/infrastructure/api"
	viewReadModels "easi/backend/internal/architectureviews/application/readmodels"
	viewAdapters "easi/backend/internal/architectureviews/infrastructure/adapters"
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	authProjectors "easi/backend/internal/auth/application/projectors"
	authReadModels "easi/backend/internal/auth/application/readmodels"
	authAdapters "easi/backend/internal/auth/infrastructure/adapters"
	authAPI "easi/backend/internal/auth/infrastructure/api"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	capAdapters "easi/backend/internal/capabilitymapping/infrastructure/adapters"
	capabilityAPI "easi/backend/internal/capabilitymapping/infrastructure/api"
	enterpriseArchAPI "easi/backend/internal/enterprisearchitecture/infrastructure/api"
	importingAPI "easi/backend/internal/importing/infrastructure/api"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	metamodelAPI "easi/backend/internal/metamodel/infrastructure/api"
	platformAPI "easi/backend/internal/platform/infrastructure/api"
	platformPL "easi/backend/internal/platform/publishedlanguage"
	releasesAPI "easi/backend/internal/releases/infrastructure/api"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/audit"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	vsAdapters "easi/backend/internal/valuestreams/infrastructure/adapters"
	valuestreamsAPI "easi/backend/internal/valuestreams/infrastructure/api"
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
	eventStore            eventstore.EventStore
	db                    *database.TenantAwareDB
	authDeps              *authAPI.AuthDependencies
	commandBus            *cqrs.InMemoryCommandBus
	eventBus              *events.InMemoryEventBus
	hateoas               *sharedAPI.HATEOASLinks
	userReadModel         *authReadModels.UserReadModel
	aiConfigStatusChecker *archAssistantAdapters.AIConfigStatusAdapter
	appContext            context.Context
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
	userReadModel := authReadModels.NewUserReadModel(db)

	if pgStore, ok := eventStore.(*eventstore.PostgresEventStore); ok {
		pgStore.SetEventBus(eventBus)
	}

	aiConfigStatusChecker := archAssistantAdapters.NewAIConfigStatusAdapter(db)

	return routerDependencies{
		eventStore:            eventStore,
		db:                    db,
		authDeps:              authDeps,
		commandBus:            commandBus,
		eventBus:              eventBus,
		hateoas:               sharedAPI.NewHATEOASLinks("/api/v1"),
		userReadModel:         userReadModel,
		aiConfigStatusChecker: aiConfigStatusChecker,
		appContext:            appContext,
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
	}), "platform routes")
	mustSetup(authAPI.SetupAuthRoutes(r, deps.db.DB(), deps.authDeps, deps.aiConfigStatusChecker), "auth routes")
}

func registerTenantRoutes(r chi.Router, deps routerDependencies) {
	adDeps := setupAccessDelegation(deps)
	r.Use(middleware.EditGrantEnrichment(adDeps.GrantResolver))
	adDeps.RegisterRoutes(r)
	setupModelingRoutes(r, deps)
	setupDomainRoutes(r, deps)
	setupValueStreamsRoutes(r, deps)
	setupSupportRoutes(r, deps)
	setupArchAssistantRoutes(r, deps)
	setupAuthRoutes(r, deps)
	wireAutoInvitationProjector(deps)
}

func setupAccessDelegation(deps routerDependencies) *accessdelegationAPI.AccessDelegationDependencies {
	adDeps, adErr := accessdelegationAPI.SetupAccessDelegationRoutes(accessdelegationAPI.AccessDelegationRoutesDeps{
		CommandBus:     deps.commandBus,
		EventStore:     deps.eventStore,
		EventBus:       deps.eventBus,
		DB:             deps.db,
		HATEOAS:        deps.hateoas,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
		NameLookups: adServices.ArtifactNameResolverDeps{
			Capabilities:     capAdapters.NewCapabilityNameAdapter(capReadModels.NewCapabilityReadModel(deps.db)),
			Components:       archAdapters.NewComponentNameAdapter(archReadModels.NewApplicationComponentReadModel(deps.db)),
			Views:            viewAdapters.NewViewNameAdapter(viewReadModels.NewArchitectureViewReadModel(deps.db)),
			Domains:          capAdapters.NewDomainNameAdapter(capReadModels.NewBusinessDomainReadModel(deps.db)),
			Vendors:          archAdapters.NewVendorNameAdapter(archReadModels.NewVendorReadModel(deps.db)),
			AcquiredEntities: archAdapters.NewAcquiredEntityNameAdapter(archReadModels.NewAcquiredEntityReadModel(deps.db)),
			InternalTeams:    archAdapters.NewInternalTeamNameAdapter(archReadModels.NewInternalTeamReadModel(deps.db)),
		},
		UserLookup:    authAdapters.NewUserEmailLookupAdapter(deps.userReadModel),
		InvChecker:    authAdapters.NewInvitationCheckerAdapter(authReadModels.NewInvitationReadModel(deps.db)),
		DomainChecker: authAdapters.NewDomainAllowlistCheckerAdapter(authReadModels.NewTenantDomainChecker(deps.db)),
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
	userRoleChecker := authAdapters.NewUserRoleCheckerAdapter(deps.userReadModel)
	viewsAPI.RegisterCommands(deps.commandBus, deps.eventStore, deps.db, userRoleChecker)
	viewHandlers := viewsAPI.NewHTTPHandlers(deps.commandBus, deps.db, deps.hateoas)
	viewsAPI.RegisterRoutes(r, viewHandlers, deps.authDeps.AuthMiddleware)

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

	viewlayoutsAPI.SubscribeEvents(deps.eventBus, deps.db)
	viewlayoutsAPI.RegisterRoutes(r, deps.db, deps.hateoas, deps.authDeps.AuthMiddleware)

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
}

func setupSupportRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(releasesAPI.SetupReleasesRoutes(r, deps.db.DB()), "releases routes")
	mustSetup(importingAPI.SetupImportingRoutes(r, importingAPI.ImportingRoutesDeps{
		CommandBus:         deps.commandBus,
		EventStore:         deps.eventStore,
		EventBus:           deps.eventBus,
		DB:                 deps.db,
		ComponentGateway:   archAdapters.NewImportComponentGateway(deps.commandBus),
		CapabilityGateway:  capAdapters.NewImportCapabilityGateway(deps.commandBus),
		ValueStreamGateway: vsAdapters.NewImportValueStreamGateway(deps.commandBus),
		ExecutionContext:   deps.appContext,
	}), "importing routes")
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

func setupArchAssistantRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(archAssistantAPI.SetupArchAssistantRoutes(archAssistantAPI.ArchAssistantRoutesDeps{
		Router:         r,
		DB:             deps.db,
		AuthMiddleware: deps.authDeps.AuthMiddleware,
	}), "arch assistant routes")

	aiConfigRepo := archAssistantRepos.NewAIConfigurationRepository(deps.db)
	tenantCreatedHandler := archAssistantHandlers.NewTenantCreatedHandler(aiConfigRepo)
	deps.eventBus.Subscribe(platformPL.TenantCreated, tenantCreatedHandler)
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
