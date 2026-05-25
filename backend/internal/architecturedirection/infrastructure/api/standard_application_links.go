package api

import (
	"errors"

	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const standardApplicationSubPath sharedAPI.ResourcePath = "/standard-application"
const standardApplicationHistorySubPath sharedAPI.ResourcePath = "/standard-application/history"

var ErrNoStandardApplication = errors.New("no standard application on this enterprise capability")

type StandardApplicationLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewStandardApplicationLinks(h *sharedAPI.HATEOASLinks) *StandardApplicationLinks {
	return &StandardApplicationLinks{HATEOASLinks: h}
}

func (h *StandardApplicationLinks) EnvelopeLinks(enterpriseCapabilityID string, standardExists bool, actor sharedctx.Actor) sharedAPI.Links {
	base := standardApplicationResourcePath(enterpriseCapabilityID)
	links := sharedAPI.Links{
		"self":      h.Get(base),
		"up":        h.Get(enterpriseCapabilityResourcePath(enterpriseCapabilityID)),
		"x-history": h.Get(standardApplicationHistoryResourcePath(enterpriseCapabilityID)),
	}
	if !actor.CanWrite(ArchitectureDirectionResource) {
		return links
	}
	if standardExists {
		links["edit"] = h.Put(base)
	} else {
		links["x-set-standard"] = h.Put(base)
	}
	return links
}

func (h *StandardApplicationLinks) StandardLinks(enterpriseCapabilityID string, actor sharedctx.Actor) sharedAPI.Links {
	base := standardApplicationResourcePath(enterpriseCapabilityID)
	links := sharedAPI.Links{
		"self":      h.Get(base),
		"up":        h.Get(enterpriseCapabilityResourcePath(enterpriseCapabilityID)),
		"x-history": h.Get(standardApplicationHistoryResourcePath(enterpriseCapabilityID)),
	}
	if actor.CanWrite(ArchitectureDirectionResource) {
		links["edit"] = h.Put(base)
	}
	return links
}

func (h *StandardApplicationLinks) HistoryLinks(enterpriseCapabilityID string) sharedAPI.Links {
	return sharedAPI.Links{
		"self": h.Get(standardApplicationHistoryResourcePath(enterpriseCapabilityID)),
		"up":   h.Get(standardApplicationResourcePath(enterpriseCapabilityID)),
	}
}

func standardApplicationResourcePath(enterpriseCapabilityID string) string {
	return string(enterpriseCapabilitiesPath) + "/" + enterpriseCapabilityID + string(standardApplicationSubPath)
}

func standardApplicationHistoryResourcePath(enterpriseCapabilityID string) string {
	return string(enterpriseCapabilitiesPath) + "/" + enterpriseCapabilityID + string(standardApplicationHistorySubPath)
}
