package api

import (
	"slices"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const (
	journeySubPath            sharedAPI.ResourcePath = "/journey"
	journeyHistorySubPath     sharedAPI.ResourcePath = "/journey/history"
	capabilityJourneysPath    sharedAPI.ResourcePath = "/capability-journeys"
	journeyMilestonesPath     sharedAPI.ResourcePath = "/milestones"
	journeyMilestoneOrderPath sharedAPI.ResourcePath = "/milestone-order"
)

type CapabilityJourneyLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewCapabilityJourneyLinks(h *sharedAPI.HATEOASLinks) *CapabilityJourneyLinks {
	return &CapabilityJourneyLinks{HATEOASLinks: h}
}

func (h *CapabilityJourneyLinks) ForCapability(capabilityID string, activeJourneys []readmodels.CapabilityJourneyDTO, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{
		"self":      h.Get(journeyResourcePath(capabilityID)),
		"x-history": h.Get(journeyHistoryResourcePath(capabilityID)),
	}
	if !actor.CanWrite(ArchitectureDirectionResource) {
		return links
	}
	if !hasActiveMaturityJourney(activeJourneys) {
		links["x-capture-maturity"] = h.Post(journeyResourcePath(capabilityID))
	}
	if !hasActiveApplicationJourney(activeJourneys) {
		links["x-capture"] = h.Post(journeyResourcePath(capabilityID))
	}
	return links
}

func hasActiveMaturityJourney(journeys []readmodels.CapabilityJourneyDTO) bool {
	return slices.ContainsFunc(journeys, isMaturityJourney)
}

func hasActiveApplicationJourney(journeys []readmodels.CapabilityJourneyDTO) bool {
	return slices.ContainsFunc(journeys, func(journey readmodels.CapabilityJourneyDTO) bool {
		return !isMaturityJourney(journey)
	})
}

func isMaturityJourney(journey readmodels.CapabilityJourneyDTO) bool {
	return journey.Kind == valueobjects.JourneyKindMaturity
}

func (h *CapabilityJourneyLinks) ItemLinks(journey *readmodels.CapabilityJourneyDTO, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{
		"self":      h.Get(journeyResourcePath(journey.CapabilityID)),
		"x-history": h.Get(journeyHistoryResourcePath(journey.CapabilityID)),
	}
	if !actor.CanWrite(ArchitectureDirectionResource) {
		return links
	}
	h.addTransitionLinks(links, journey)
	if !isActiveJourneyStatus(journey.Status) {
		return links
	}
	itemBase := journeyItemResourcePath(journey.ID)
	links["edit"] = h.Put(itemBase + "/details")
	links["x-progress"] = h.Put(itemBase + "/progress")
	if !isMaturityJourney(*journey) {
		links["x-change-sources"] = h.Put(itemBase + "/source-applications")
	}
	links["x-add-milestone"] = h.Post(itemBase + string(journeyMilestonesPath))
	if len(journey.Milestones) > 1 {
		links["x-reorder-milestones"] = h.Put(itemBase + string(journeyMilestoneOrderPath))
	}
	return links
}

func (h *CapabilityJourneyLinks) addTransitionLinks(links sharedAPI.Links, journey *readmodels.CapabilityJourneyDTO) {
	itemBase := journeyItemResourcePath(journey.ID)
	switch journey.Status {
	case valueobjects.JourneyStatusPlanned:
		links["x-start"] = h.Post(itemBase + "/start")
		links["x-abandon"] = h.Post(itemBase + "/abandon")
	case valueobjects.JourneyStatusInFlight:
		links["x-complete"] = h.Post(itemBase + "/complete")
		links["x-abandon"] = h.Post(itemBase + "/abandon")
	}
}

func (h *CapabilityJourneyLinks) MilestoneLinks(journey *readmodels.CapabilityJourneyDTO, milestoneID string, actor sharedctx.Actor) sharedAPI.Links {
	if !actor.CanWrite(ArchitectureDirectionResource) || !isActiveJourneyStatus(journey.Status) {
		return sharedAPI.Links{}
	}
	base := journeyItemResourcePath(journey.ID) + string(journeyMilestonesPath) + "/" + milestoneID
	return sharedAPI.Links{
		"edit":   h.Put(base),
		"delete": h.Del(base),
	}
}

func (h *CapabilityJourneyLinks) BulkLinks(selfPath string, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get(selfPath)}
	if actor.CanWrite(ArchitectureDirectionResource) {
		links["x-capture"] = h.Post(string(capabilitiesPath) + "/{capabilityId}" + string(journeySubPath))
	}
	return links
}

func isActiveJourneyStatus(status string) bool {
	return status == valueobjects.JourneyStatusPlanned || status == valueobjects.JourneyStatusInFlight
}

func journeyResourcePath(capabilityID string) string {
	return string(capabilitiesPath) + "/" + capabilityID + string(journeySubPath)
}

func journeyHistoryResourcePath(capabilityID string) string {
	return string(capabilitiesPath) + "/" + capabilityID + string(journeyHistorySubPath)
}

func journeyItemResourcePath(journeyID string) string {
	return string(capabilityJourneysPath) + "/" + journeyID
}
