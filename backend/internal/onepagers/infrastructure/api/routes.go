package api

import (
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/onepagers/application/handlers"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/queries"
	"easi/backend/internal/onepagers/application/readmodels"
	opevents "easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type OnePagersRoutesDeps struct {
	Router          chi.Router
	CommandBus      *cqrs.InMemoryCommandBus
	EventStore      eventstore.EventStore
	EventBus        events.EventBus
	DB              *database.TenantAwareDB
	Hateoas         *sharedAPI.HATEOASLinks
	AuthMiddleware  AuthMiddleware
	SessionProvider authPL.SessionProvider
	Subjects        ports.SubjectExistenceChecker
	BuiltInFields   map[string]ports.BuiltInFieldSource
	MaturityScale   ports.MaturityScaleSource
}

func SetupOnePagersRoutes(deps OnePagersRoutesDeps) error {
	repo := repositories.NewOnePagerConfigurationRepository(deps.EventStore)
	readModel := readmodels.NewOnePagerConfigurationReadModel(deps.DB)

	projector := projectors.NewOnePagerConfigurationProjector(readModel)
	for _, eventType := range opevents.ConfigurationEventTypes() {
		deps.EventBus.Subscribe(eventType, projector)
	}

	registerCommands(deps.CommandBus, repo, readModel)

	factsRepo := repositories.NewOnePagerFactsRepository(deps.EventStore)
	factsReadModel := readmodels.NewOnePagerFactsReadModel(deps.DB)

	factsProjector := projectors.NewOnePagerFactsProjector(factsReadModel)
	for _, eventType := range opevents.FactsEventTypes() {
		deps.EventBus.Subscribe(eventType, factsProjector)
	}

	deletionReactor := projectors.NewSubjectDeletedReactor(factsReadModel, deps.CommandBus)
	for _, eventType := range projectors.SubjectDeletionEventTypes() {
		deps.EventBus.Subscribe(eventType, deletionReactor)
	}

	registerFactsCommands(deps, factsRepo, readModel, factsReadModel)

	links := NewOnePagerLinks(deps.Hateoas)
	configHandlers := NewOnePagerConfigurationHandlers(deps.CommandBus, readModel, links, deps.SessionProvider)
	impactPreviewQuery := queries.NewImpactPreviewQuery(queries.ImpactPreviewDeps{
		Configurations: readModel,
		Facts:          factsReadModel,
		Subjects:       deps.BuiltInFields,
	})
	impactPreviewHandlers := NewImpactPreviewHandlers(impactPreviewQuery, links)
	registerRoutes(deps.Router, configHandlers, impactPreviewHandlers, deps.AuthMiddleware)

	factsHandlers := NewOnePagerFactsHandlers(OnePagerFactsHandlersDeps{
		CommandBus:      deps.CommandBus,
		Facts:           factsReadModel,
		Configs:         readModel,
		Links:           links,
		SessionProvider: deps.SessionProvider,
	})
	onePagerQuery := queries.NewOnePagerQuery(queries.OnePagerQueryDeps{
		Configurations: readModel,
		Facts:          factsReadModel,
		Subjects:       deps.BuiltInFields,
		MaturityScale:  deps.MaturityScale,
	})
	viewHandlers := NewOnePagerViewHandlers(onePagerQuery, links)
	registerSubjectRoutes(deps.Router, viewHandlers, factsHandlers, deps.AuthMiddleware)

	return nil
}

func registerFactsCommands(
	deps OnePagersRoutesDeps,
	factsRepo *repositories.OnePagerFactsRepository,
	configReadModel *readmodels.OnePagerConfigurationReadModel,
	factsReadModel *readmodels.OnePagerFactsReadModel,
) {
	deps.CommandBus.Register("RecordFieldValue", handlers.NewRecordFieldValueHandler(factsRepo, configReadModel, factsReadModel, deps.Subjects))
	deps.CommandBus.Register("ClearFieldValue", handlers.NewClearFieldValueHandler(factsRepo, configReadModel, factsReadModel))
	deps.CommandBus.Register("ArchiveOnePagerFacts", handlers.NewArchiveOnePagerFactsHandler(factsRepo))
}

type subjectRoutePermissions struct {
	read  authPL.Permission
	write authPL.Permission
}

var subjectPermissionsByType = map[string]subjectRoutePermissions{
	"capability":            {read: authPL.PermCapabilitiesRead, write: authPL.PermCapabilitiesWrite},
	"enterprise-capability": {read: authPL.PermEnterpriseArchRead, write: authPL.PermEnterpriseArchWrite},
	"application":           {read: authPL.PermComponentsRead, write: authPL.PermComponentsWrite},
	"acquired-entity":       {read: authPL.PermComponentsRead, write: authPL.PermComponentsWrite},
	"vendor":                {read: authPL.PermComponentsRead, write: authPL.PermComponentsWrite},
	"internal-team":         {read: authPL.PermComponentsRead, write: authPL.PermComponentsWrite},
}

func registerSubjectRoutes(router chi.Router, viewHandlers *OnePagerViewHandlers, factsHandlers *OnePagerFactsHandlers, authMiddleware AuthMiddleware) {
	for _, subjectType := range valueobjects.AllSubjectTypes() {
		permissions := subjectPermissionsByType[subjectType.Value()]
		router.Route("/one-pagers/"+subjectType.Value()+"/{subjectID}", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequirePermission(permissions.read))
				r.Get("/", viewHandlers.GetOnePager(subjectType))
			})

			r.Route("/facts", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(authMiddleware.RequirePermission(permissions.read))
					r.Get("/", factsHandlers.GetFacts(subjectType))
				})

				r.Group(func(r chi.Router) {
					r.Use(authMiddleware.RequirePermission(permissions.write))
					r.Put("/{fieldID}", factsHandlers.RecordValue(subjectType))
					r.Delete("/{fieldID}", factsHandlers.ClearValue(subjectType))
				})
			})
		})
	}
}

func registerCommands(
	commandBus *cqrs.InMemoryCommandBus,
	repo *repositories.OnePagerConfigurationRepository,
	readModel *readmodels.OnePagerConfigurationReadModel,
) {
	commandBus.Register("CreateOnePagerConfiguration", handlers.NewCreateOnePagerConfigurationHandler(repo, readModel))
	commandBus.Register("DefineCustomField", handlers.NewDefineCustomFieldHandler(repo))
	commandBus.Register("RenameCustomField", handlers.NewRenameCustomFieldHandler(repo))
	commandBus.Register("ChangeCustomFieldRequirement", handlers.NewChangeCustomFieldRequirementHandler(repo))
	commandBus.Register("RetireCustomField", handlers.NewRetireCustomFieldHandler(repo))
	commandBus.Register("ReactivateCustomField", handlers.NewReactivateCustomFieldHandler(repo))
	commandBus.Register("IncludeBuiltInField", handlers.NewIncludeBuiltInFieldHandler(repo))
	commandBus.Register("ExcludeBuiltInField", handlers.NewExcludeBuiltInFieldHandler(repo))
	commandBus.Register("ReorderOnePagerFields", handlers.NewReorderOnePagerFieldsHandler(repo))
	commandBus.Register("AddSelectionOption", handlers.NewAddSelectionOptionHandler(repo))
	commandBus.Register("RetireSelectionOption", handlers.NewRetireSelectionOptionHandler(repo))
}

func registerRoutes(router chi.Router, h *OnePagerConfigurationHandlers, previewHandlers *ImpactPreviewHandlers, authMiddleware AuthMiddleware) {
	router.Route("/one-pagers/configurations/{subjectType}", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermMetaModelRead))
			r.Get("/", h.GetConfiguration)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermMetaModelWrite))
			r.Get("/impact-preview", previewHandlers.GetImpactPreview)
			r.Post("/custom-fields", h.DefineCustomField)
			r.Put("/custom-fields/{fieldID}", h.RenameCustomField)
			r.Put("/custom-fields/{fieldID}/requirement", h.ChangeCustomFieldRequirement)
			r.Post("/custom-fields/{fieldID}/retire", h.RetireCustomField)
			r.Post("/custom-fields/{fieldID}/reactivate", h.ReactivateCustomField)
			r.Post("/custom-fields/{fieldID}/options", h.AddSelectionOption)
			r.Post("/custom-fields/{fieldID}/options/{optionID}/retire", h.RetireSelectionOption)
			r.Post("/built-in-fields/{entryID}/include", h.IncludeBuiltInField)
			r.Post("/built-in-fields/{entryID}/exclude", h.ExcludeBuiltInField)
			r.Put("/display-order", h.ReorderFields)
		})
	})
}
