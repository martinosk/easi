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

func (h *TimeAssessmentLinks) ItemLinks(capabilityID, componentID string, actor sharedctx.Actor) sharedAPI.Links {
	base := timeAssessmentItemResourcePath(capabilityID, componentID)
	links := sharedAPI.Links{"self": h.Get(base)}
	if actor.CanWrite(ArchitectureDirectionResource) {
		links["edit"] = h.Put(base)
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

func timeAssessmentItemResourcePath(capabilityID, componentID string) string {
	return string(capabilitiesPath) + "/" + capabilityID + "/components/" + componentID + string(timeAssessmentSubPath)
}

func templatedTimeAssessmentItemPath() string {
	return string(capabilitiesPath) + "/{capabilityId}/components/{componentId}" + string(timeAssessmentSubPath)
}
