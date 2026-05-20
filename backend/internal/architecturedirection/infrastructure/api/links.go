package api

import (
	"errors"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
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

func (h *DirectionLinks) DirectionForActor(enterpriseCapabilityID string, status string, actor sharedctx.Actor) sharedAPI.Links {
	base := directionResourcePath(enterpriseCapabilityID)
	links := sharedAPI.Links{
		"self": h.Get(base),
		"up":   h.Get(enterpriseCapabilityResourcePath(enterpriseCapabilityID)),
	}
	h.addWriteAffordances(links, base, status, actor)
	return links
}

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
	return status != valueobjects.DirectionStatusRejected && status != valueobjects.DirectionStatusAgreed
}

func canReject(status string) bool {
	return status != valueobjects.DirectionStatusRejected
}

func nextAdvanceRel(status string) (rel, target string) {
	switch status {
	case valueobjects.DirectionStatusDraft:
		return "x-propose", "propose"
	case valueobjects.DirectionStatusProposed:
		return "x-agree", "agree"
	default:
		return "", ""
	}
}

func directionResourcePath(enterpriseCapabilityID string) string {
	return string(enterpriseCapabilitiesPath) + "/" + enterpriseCapabilityID + string(directionSubPath)
}

func enterpriseCapabilityResourcePath(enterpriseCapabilityID string) string {
	return string(enterpriseCapabilitiesPath) + "/" + enterpriseCapabilityID
}
