package api

import (
	"net/http"

	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/projectors"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type RoutesDeps struct {
	Router             chi.Router
	CommandBus         *cqrs.InMemoryCommandBus
	EventStore         eventstore.EventStore
	EventBus           events.EventBus
	DB                 *database.TenantAwareDB
	HATEOAS            *sharedAPI.HATEOASLinks
	AuthMiddleware     AuthMiddleware
	ReferenceChecker   *services.ReferenceChecker
	SourceEligibility  services.SourceEligibility
	CompositionPreview CompositionPreviewProvider
	DirectRealization  services.DirectRealizationLookup

	CapabilityExists              services.CapabilityExists
	ComponentExists               services.ComponentExists
	DomainExists                  services.DomainExists
	CapabilityEffectivelyInDomain services.CapabilityEffectivelyInDomain
}

func SetupRoutes(deps RoutesDeps) error {
	readModel := readmodels.NewDirectionReadModel(deps.DB)
	repo := repositories.NewDirectionRepository(deps.EventStore)

	subscribeEvents(deps.EventBus, readModel)
	deps.EventBus.Subscribe(eaPL.EnterpriseCapabilityDeleted,
		projectors.NewEnterpriseCapabilityDeletedReactor(readModel, deps.CommandBus))
	registerCommandHandlers(commandHandlerDeps{
		commandBus:  deps.CommandBus,
		repo:        repo,
		readModel:   readModel,
		refs:        deps.ReferenceChecker,
		eligibility: deps.SourceEligibility,
	})

	links := NewDirectionLinks(deps.HATEOAS)
	httpHandlers := NewDirectionHandlers(deps.CommandBus, readModel, links)
	previewHandlers := NewCompositionPreviewHandlers(deps.CompositionPreview, deps.HATEOAS)

	registerRoutes(deps.Router, httpHandlers, previewHandlers, deps.AuthMiddleware)

	setupStandardApplicationRoutes(deps)
	setupTimeAssessmentRoutes(deps)
	setupRealizationRoleRoutes(deps)
	setupCapabilityJourneyRoutes(deps)
	return nil
}

func setupStandardApplicationRoutes(deps RoutesDeps) {
	readModel := readmodels.NewStandardApplicationReadModel(deps.DB)
	repo := repositories.NewStandardApplicationRepository(deps.EventStore)

	subscribeStandardApplicationEvents(deps.EventBus, readModel)
	deps.CommandBus.Register("SetStandardApplication", handlers.NewSetStandardApplicationHandler(repo, readModel, deps.ReferenceChecker))

	links := NewStandardApplicationLinks(deps.HATEOAS)
	httpHandlers := NewStandardApplicationHandlers(deps.CommandBus, readModel, links)

	registerStandardApplicationRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func subscribeStandardApplicationEvents(eventBus events.EventBus, rm *readmodels.StandardApplicationReadModel) {
	eventBus.Subscribe(pl.StandardApplicationSet, projectors.NewStandardApplicationProjector(rm))
	subscribeMany(eventBus, projectors.NewStaleApplicationProjector(rm),
		amPL.ApplicationComponentDeleted, amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated)
}

func subscribeMany(eventBus events.EventBus, handler events.EventHandler, eventTypes ...string) {
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, handler)
	}
}

func registerStandardApplicationRoutes(r chi.Router, h *StandardApplicationHandlers, authMiddleware AuthMiddleware) {
	r.Route("/enterprise-capabilities/{id}/standard-application", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionRead))
			r.Get("/", h.GetStandardApplicationForEnterpriseCapability)
			r.Get("/history", h.GetStandardApplicationHistory)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Put("/", h.SetStandardApplication)
		})
	})
}

func setupTimeAssessmentRoutes(deps RoutesDeps) {
	readModel := readmodels.NewTimeAssessmentReadModel(deps.DB)
	repo := repositories.NewTimeAssessmentRepository(deps.EventStore)

	subscribeTimeAssessmentEvents(deps.EventBus, readModel)
	deps.CommandBus.Register("AssessRealization", handlers.NewAssessRealizationHandler(repo, readModel, deps.DirectRealization))
	deps.CommandBus.Register("RemoveTimeAssessment", handlers.NewRemoveTimeAssessmentHandler(repo, readModel))
	deps.EventBus.Subscribe(cmPL.SystemRealizationDeleted, projectors.NewTimeAssessmentDeletionReactor(readModel, deps.CommandBus))

	links := NewTimeAssessmentLinks(deps.HATEOAS)
	httpHandlers := NewTimeAssessmentHandlers(deps.CommandBus, readModel, links)

	registerTimeAssessmentRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func subscribeTimeAssessmentEvents(eventBus events.EventBus, rm *readmodels.TimeAssessmentReadModel) {
	subscribeMany(eventBus, projectors.NewTimeAssessmentProjector(rm),
		pl.TimeAssessmentRecorded, pl.TimeAssessmentRemoved)
	subscribeMany(eventBus, projectors.NewTimeAssessmentReferenceProjector(rm),
		cmPL.CapabilityDeleted, cmPL.CapabilityCreated, cmPL.CapabilityUpdated,
		amPL.ApplicationComponentDeleted, amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated,
		authPL.UserCreated)
}

func registerTimeAssessmentRoutes(r chi.Router, h *TimeAssessmentHandlers, authMiddleware AuthMiddleware) {
	registerDomainReadCollection(r, "/time-assessments", authMiddleware, func(r chi.Router) {
		r.Get("/", h.GetTimeAssessments)
		r.Get("/rollups", h.GetTimeAssessmentRollups)
	})
	registerPairResourceRoutes(r, "/capabilities/{id}/components/{componentId}/time-assessment", authMiddleware,
		pairResourceHandlers{get: h.GetTimeAssessment, put: h.PutTimeAssessment, delete: h.DeleteTimeAssessment})
}

func registerDomainReadCollection(r chi.Router, pattern string, authMiddleware AuthMiddleware, register func(chi.Router)) {
	r.Route(pattern, func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			register(r)
		})
	})
}

type pairResourceHandlers struct {
	get    http.HandlerFunc
	put    http.HandlerFunc
	delete http.HandlerFunc
}

func registerPairResourceRoutes(r chi.Router, pattern string, authMiddleware AuthMiddleware, h pairResourceHandlers) {
	r.Route(pattern, func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			r.Get("/", h.get)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Put("/", h.put)
			r.Delete("/", h.delete)
		})
	})
}

func setupRealizationRoleRoutes(deps RoutesDeps) {
	readModel := readmodels.NewRealizationRoleReadModel(deps.DB)
	repo := repositories.NewRealizationRolesRepository(deps.EventStore)

	subscribeRealizationRoleEvents(deps.EventBus, readModel)
	deps.CommandBus.Register("AssignRealizationRole", handlers.NewAssignRealizationRoleHandler(repo, readModel, deps.DirectRealization))
	deps.CommandBus.Register("ClearRealizationRole", handlers.NewClearRealizationRoleHandler(repo, readModel))
	deps.EventBus.Subscribe(cmPL.SystemRealizationDeleted, projectors.NewRealizationRoleDeletionReactor(readModel, deps.CommandBus))

	links := NewRealizationRoleLinks(deps.HATEOAS)
	httpHandlers := NewRealizationRoleHandlers(deps.CommandBus, readModel, links)

	registerRealizationRoleRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func subscribeRealizationRoleEvents(eventBus events.EventBus, rm *readmodels.RealizationRoleReadModel) {
	subscribeMany(eventBus, projectors.NewRealizationRoleProjector(rm),
		pl.RealizationRoleAssigned, pl.RealizationRoleCleared)
	subscribeMany(eventBus, projectors.NewRealizationRoleReferenceProjector(rm),
		cmPL.CapabilityDeleted, cmPL.CapabilityCreated, cmPL.CapabilityUpdated,
		amPL.ApplicationComponentDeleted, amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated)
}

func setupCapabilityJourneyRoutes(deps RoutesDeps) {
	readModel := readmodels.NewCapabilityJourneyReadModel(deps.DB)
	repo := repositories.NewCapabilityJourneyRepository(deps.EventStore)

	subscribeCapabilityJourneyEvents(deps.EventBus, readModel)

	refs := handlers.JourneyReferenceChecks{
		CapabilityExists:              deps.CapabilityExists,
		ComponentExists:               deps.ComponentExists,
		DomainExists:                  deps.DomainExists,
		CapabilityEffectivelyInDomain: deps.CapabilityEffectivelyInDomain,
	}
	deps.CommandBus.Register("PlanJourney", handlers.NewPlanJourneyHandler(repo, readModel, refs))
	deps.CommandBus.Register("StartJourney", handlers.NewStartJourneyHandler(repo))
	deps.CommandBus.Register("CompleteJourney", handlers.NewCompleteJourneyHandler(repo))
	deps.CommandBus.Register("AbandonJourney", handlers.NewAbandonJourneyHandler(repo))
	deps.CommandBus.Register("UpdateJourneyProgress", handlers.NewUpdateJourneyProgressHandler(repo))
	deps.CommandBus.Register("UpdateJourneyDetails", handlers.NewUpdateJourneyDetailsHandler(repo))
	deps.CommandBus.Register("ChangeJourneySourceApplications", handlers.NewChangeJourneySourceApplicationsHandler(repo, deps.ComponentExists))
	deps.CommandBus.Register("AddJourneyMilestone", handlers.NewAddJourneyMilestoneHandler(repo))
	deps.CommandBus.Register("UpdateJourneyMilestone", handlers.NewUpdateJourneyMilestoneHandler(repo))
	deps.CommandBus.Register("RemoveJourneyMilestone", handlers.NewRemoveJourneyMilestoneHandler(repo))
	deps.CommandBus.Register("ReorderJourneyMilestones", handlers.NewReorderJourneyMilestonesHandler(repo))

	links := NewCapabilityJourneyLinks(deps.HATEOAS)
	httpHandlers := NewCapabilityJourneyHandlers(deps.CommandBus, readModel, links)

	registerCapabilityJourneyRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func subscribeCapabilityJourneyEvents(eventBus events.EventBus, rm *readmodels.CapabilityJourneyReadModel) {
	subscribeMany(eventBus, projectors.NewCapabilityJourneyProjector(rm),
		pl.JourneyPlanned, pl.JourneyStarted, pl.JourneyCompleted, pl.JourneyAbandoned,
		pl.JourneyProgressUpdated, pl.JourneyDetailsUpdated, pl.JourneySourceApplicationsChanged,
		pl.JourneyMilestoneAdded, pl.JourneyMilestoneUpdated, pl.JourneyMilestoneRemoved, pl.JourneyMilestonesReordered)
	subscribeMany(eventBus, projectors.NewCapabilityJourneyReferenceProjector(rm),
		cmPL.CapabilityCreated, cmPL.CapabilityUpdated, cmPL.CapabilityDeleted,
		cmPL.BusinessDomainCreated, cmPL.BusinessDomainUpdated, cmPL.BusinessDomainDeleted,
		amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated, amPL.ApplicationComponentDeleted,
		authPL.UserCreated)
}

func registerCapabilityJourneyRoutes(r chi.Router, h *CapabilityJourneyHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capabilities/{id}/journey", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			r.Get("/", h.GetJourneyForCapability)
			r.Get("/history", h.GetJourneyHistory)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Post("/", h.CaptureJourney)
		})
	})

	registerDomainReadCollection(r, "/capability-journeys", authMiddleware, func(r chi.Router) {
		r.Get("/", h.GetCapabilityJourneys)
	})

	r.Route("/capability-journeys/{journeyId}", func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
		r.Post("/start", h.StartJourney)
		r.Post("/complete", h.CompleteJourney)
		r.Post("/abandon", h.AbandonJourney)
		r.Put("/details", h.PutJourneyDetails)
		r.Put("/progress", h.PutJourneyProgress)
		r.Put("/source-applications", h.PutJourneySourceApplications)
		r.Post("/milestones", h.PostJourneyMilestone)
		r.Put("/milestones/{milestoneId}", h.PutJourneyMilestone)
		r.Delete("/milestones/{milestoneId}", h.DeleteJourneyMilestone)
		r.Put("/milestone-order", h.PutJourneyMilestoneOrder)
	})
}

func registerRealizationRoleRoutes(r chi.Router, h *RealizationRoleHandlers, authMiddleware AuthMiddleware) {
	registerDomainReadCollection(r, "/realization-roles", authMiddleware, func(r chi.Router) {
		r.Get("/", h.GetRealizationRoles)
	})
	registerPairResourceRoutes(r, "/capabilities/{id}/components/{componentId}/realization-role", authMiddleware,
		pairResourceHandlers{get: h.GetRealizationRole, put: h.PutRealizationRole, delete: h.DeleteRealizationRole})
}

func subscribeEvents(eventBus events.EventBus, rm *readmodels.DirectionReadModel) {
	subscribeMany(eventBus, projectors.NewDirectionProjector(rm),
		pl.DirectionDrafted, pl.DirectionProposed, pl.DirectionAgreed, pl.DirectionRejected,
		pl.DirectionNarrativeUpdated, pl.DirectionHorizonChanged, pl.DirectionPlacementsChanged,
		pl.DirectionSourceCapabilitiesChanged)
	subscribeMany(eventBus, projectors.NewStaleReferenceProjector(rm),
		cmPL.CapabilityDeleted, cmPL.CapabilityCreated, cmPL.CapabilityUpdated,
		cmPL.BusinessDomainCreated, cmPL.BusinessDomainUpdated,
		cmPL.CapabilityAssignedToDomain, cmPL.CapabilityUnassignedFromDomain)
}

type commandHandlerDeps struct {
	commandBus  *cqrs.InMemoryCommandBus
	repo        *repositories.DirectionRepository
	readModel   *readmodels.DirectionReadModel
	refs        *services.ReferenceChecker
	eligibility services.SourceEligibility
}

func registerCommandHandlers(deps commandHandlerDeps) {
	policy := services.NewDirectionReferenceService(deps.refs, deps.readModel, deps.eligibility)
	deps.commandBus.Register("CaptureDirection", handlers.NewCaptureDirectionHandler(deps.repo, policy))
	deps.commandBus.Register("AdvanceDirection", handlers.NewAdvanceDirectionHandler(deps.repo))
	deps.commandBus.Register("RejectDirection", handlers.NewRejectDirectionHandler(deps.repo))
	deps.commandBus.Register("UpdateDirection", handlers.NewUpdateDirectionHandler(deps.repo))
	deps.commandBus.Register("AddDirectionSource", handlers.NewAddDirectionSourceHandler(deps.repo, policy))
	deps.commandBus.Register("RemoveDirectionSource", handlers.NewRemoveDirectionSourceHandler(deps.repo))
}

func registerRoutes(r chi.Router, h *DirectionHandlers, preview *CompositionPreviewHandlers, authMiddleware AuthMiddleware) {
	r.Route("/enterprise-capabilities/{id}/direction", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionRead))
			r.Get("/", h.GetDirectionForEnterpriseCapability)
			r.Post("/composition-preview", preview.PreviewComposition)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Post("/", h.CaptureDirection)
			r.Put("/", h.UpdateDirection)
			r.Post("/propose", h.ProposeDirection)
			r.Post("/agree", h.AgreeDirection)
			r.Post("/reject", h.RejectDirection)
			r.Post("/sources", h.AddDirectionSource)
			r.Delete("/sources/{capabilityId}", h.RemoveDirectionSource)
		})
	})
}
