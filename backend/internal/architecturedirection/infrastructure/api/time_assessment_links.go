package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const (
	capabilitiesPath      sharedAPI.ResourcePath = "/capabilities"
	timeAssessmentsPath   sharedAPI.ResourcePath = "/time-assessments"
	timeAssessmentSubPath sharedAPI.ResourcePath = "/time-assessment"
)

type TimeAssessmentLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewTimeAssessmentLinks(h *sharedAPI.HATEOASLinks) *TimeAssessmentLinks {
	return &TimeAssessmentLinks{HATEOASLinks: h}
}

func (h *TimeAssessmentLinks) ItemLinks(pair timeAssessmentPairID, actor sharedctx.Actor) sharedAPI.Links {
	return h.itemLinks(pair, actor, true)
}

func (h *TimeAssessmentLinks) UnassessedItemLinks(pair timeAssessmentPairID, actor sharedctx.Actor) sharedAPI.Links {
	return h.itemLinks(pair, actor, false)
}

func (h *TimeAssessmentLinks) itemLinks(pair timeAssessmentPairID, actor sharedctx.Actor, assessed bool) sharedAPI.Links {
	base := timeAssessmentItemResourcePath(pair)
	links := sharedAPI.Links{"self": h.Get(base)}
	if !actor.CanWrite(ArchitectureDirectionResource) {
		return links
	}
	links["edit"] = h.Put(base)
	if assessed {
		links["delete"] = h.Del(base)
	}
	return links
}

func (h *TimeAssessmentLinks) CollectionLinks(selfPath string, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get(selfPath)}
	if actor.CanWrite(ArchitectureDirectionResource) {
		links["x-assess"] = h.Put(templatedTimeAssessmentItemPath())
	}
	return links
}

func timeAssessmentItemResourcePath(pair timeAssessmentPairID) string {
	return string(capabilitiesPath) + "/" + pair.CapabilityID + "/components/" + pair.ComponentID + string(timeAssessmentSubPath)
}

func templatedTimeAssessmentItemPath() string {
	return string(capabilitiesPath) + "/{capabilityId}/components/{componentId}" + string(timeAssessmentSubPath)
}
