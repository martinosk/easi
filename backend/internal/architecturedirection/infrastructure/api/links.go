package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const ArchitectureDirectionResource sharedctx.ResourceName = "architecture-direction"

type DirectionLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewDirectionLinks(h *sharedAPI.HATEOASLinks) *DirectionLinks {
	return &DirectionLinks{HATEOASLinks: h}
}

func (h *DirectionLinks) DirectionForActor(directionID, enterpriseCapabilityID string, status string, actor sharedctx.Actor) sharedAPI.Links {
	base := "/directions/" + directionID
	links := sharedAPI.Links{
		"self":                    h.Get(base),
		"x-enterprise-capability": h.Get("/enterprise-capabilities/" + enterpriseCapabilityID),
	}
	h.addWriteAffordances(links, base, status, actor)
	return links
}

func (h *DirectionLinks) addWriteAffordances(links sharedAPI.Links, base, status string, actor sharedctx.Actor) {
	if !actor.CanWrite(ArchitectureDirectionResource) {
		return
	}
	if canEdit(status) {
		links["edit"] = h.Put(base)
	}
	if next := nextAdvanceRel(status); next != "" {
		links[next] = h.Post(base + "/advance/" + advanceTarget(status))
	}
	if canReject(status) {
		links["x-reject"] = h.Post(base + "/reject")
	}
}

func canEdit(status string) bool {
	return status != "rejected" && status != "agreed"
}

func canReject(status string) bool {
	return status == "draft" || status == "proposed"
}

func nextAdvanceRel(status string) string {
	switch status {
	case "draft":
		return "x-advance-proposed"
	case "proposed":
		return "x-advance-agreed"
	default:
		return ""
	}
}

func advanceTarget(status string) string {
	if status == "draft" {
		return "proposed"
	}
	return "agreed"
}

func (h *DirectionLinks) EnterpriseCapabilityDirectionLinks(enterpriseCapabilityID string, hasActive bool, actor sharedctx.Actor) sharedAPI.Links {
	base := "/enterprise-capabilities/" + enterpriseCapabilityID + "/direction"
	links := sharedAPI.Links{
		"self": h.Get(base),
	}
	if !hasActive && actor.CanWrite(ArchitectureDirectionResource) {
		links["x-capture-direction"] = h.Post(base)
	}
	return links
}
