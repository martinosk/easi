package api

import (
	"time"

	adPL "easi/backend/internal/accessdelegation/publishedlanguage"
	directionPL "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/stewardship/application/home"
	homeRM "easi/backend/internal/stewardship/application/home/readmodels"
)

func setupHome(deps RoutesDeps) {
	caches := homeRM.NewCaches(deps.DB)
	subscribeMany(deps.EventBus, home.NewCapabilityCacheProjector(caches),
		capPL.CapabilityCreated, capPL.CapabilityUpdated, capPL.CapabilityMetadataUpdated,
		capPL.CapabilityParentChanged, capPL.CapabilityLevelChanged, capPL.CapabilityDeleted)
	subscribeMany(deps.EventBus, home.NewDomainAssignmentCacheProjector(caches),
		capPL.CapabilityAssignedToDomain, capPL.CapabilityUnassignedFromDomain, capPL.CapabilityDeleted, capPL.BusinessDomainDeleted)
	subscribeMany(deps.EventBus, home.NewApplicationCacheProjector(caches),
		amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated, amPL.ApplicationComponentDeleted,
		amPL.ApplicationOwnerNominated, amPL.ApplicationOwnershipConfirmed, amPL.ApplicationOwnerAssigned, amPL.ApplicationOwnershipCleared)
	subscribeMany(deps.EventBus, home.NewRealizationCacheProjector(caches),
		capPL.SystemLinkedToCapability, capPL.SystemRealizationDeleted, capPL.CapabilityDeleted, amPL.ApplicationComponentDeleted)
	subscribeMany(deps.EventBus, home.NewTimeAssessmentCacheProjector(caches),
		directionPL.TimeAssessmentRecorded, directionPL.TimeAssessmentRemoved)
	subscribeMany(deps.EventBus, home.NewEditGrantCacheProjector(caches),
		adPL.EditGrantActivated, adPL.EditGrantRevoked, adPL.EditGrantExpired)

	handlers := NewHomeHandlers(home.NewView(homeRM.NewQueries(deps.DB), time.Now), deps.HATEOAS)
	deps.Router.Get(string(homePath), handlers.GetHome)
}
