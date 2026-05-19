package api

import (
	"errors"

	"easi/backend/internal/architecturedirection/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const ArchitectureDirectionResource sharedctx.ResourceName = "architecture-direction"

const (
	enterpriseCapabilitiesPath sharedAPI.ResourcePath = "/enterprise-capabilities"
	directionSubPath           sharedAPI.ResourcePath = "/direction"
)

var ErrNoActiveDirection = errors.New("no active direction on this enterprise capability")

type DirectionLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewDirectionLinks(h *sharedAPI.HATEOASLinks) *DirectionLinks {
	return &DirectionLinks{HATEOASLinks: h}
}

// DirectionForActor returns _links for a Direction DTO. The direction is always
// addressed as a singleton sub-resource of its parent EC, so the URLs are
// derived from the EC ID, not the direction ID.
func (h *DirectionLinks) DirectionForActor(enterpriseCapabilityID string, status string, actor sharedctx.Actor) sharedAPI.Links {
	base := directionResourcePath(enterpriseCapabilityID)
	links := sharedAPI.Links{
		"self": h.Get(base),
		"up":   h.Get(enterpriseCapabilityResourcePath(enterpriseCapabilityID)),
	}
	h.addWriteAffordances(links, base, status, actor)
	return links
}

// EnterpriseCapabilityDirectionLinks builds the _links for the EC-direction
// envelope: the singleton sub-resource itself plus a capture affordance when
// no direction exists and the actor may write.
func (h *DirectionLinks) EnterpriseCapabilityDirectionLinks(enterpriseCapabilityID string, direction *readmodels.DirectionDTO, actor sharedctx.Actor) sharedAPI.Links {
	base := directionResourcePath(enterpriseCapabilityID)
	links := sharedAPI.Links{
		"self": h.Get(base),
		"up":   h.Get(enterpriseCapabilityResourcePath(enterpriseCapabilityID)),
	}
	if direction == nil && actor.CanWrite(ArchitectureDirectionResource) {
		links["x-capture-direction"] = h.Post(base)
	}
	return links
}

func (h *DirectionLinks) addWriteAffordances(links sharedAPI.Links, base, status string, actor sharedctx.Actor) {
	if !actor.CanWrite(ArchitectureDirectionResource) {
		return
	}
	if canEdit(status) {
		links["edit"] = h.Put(base)
	}
	if rel, target := nextAdvanceRel(status); rel != "" {
		links[rel] = h.Post(base + "/" + target)
	}
	if canReject(status) {
		links["x-reject"] = h.Post(base + "/reject")
	}
}

func canEdit(status string) bool {
	return status != "rejected" && status != "agreed"
}

func canReject(status string) bool {
	return status != "rejected"
}

// nextAdvanceRel returns the link relation and URL segment for the next legal
// advance transition. Returns ("", "") if no further advance is possible.
func nextAdvanceRel(status string) (rel, target string) {
	switch status {
	case "draft":
		return "x-propose", "propose"
	case "proposed":
		return "x-agree", "agree"
	default:
		return "", ""
	}
}

// directionResourcePath returns the relative URL of the direction singleton
// sub-resource (no /api/v1 prefix; the HATEOASLinks helper adds that).
func directionResourcePath(enterpriseCapabilityID string) string {
	return string(enterpriseCapabilitiesPath) + "/" + enterpriseCapabilityID + string(directionSubPath)
}

func enterpriseCapabilityResourcePath(enterpriseCapabilityID string) string {
	return string(enterpriseCapabilitiesPath) + "/" + enterpriseCapabilityID
}
