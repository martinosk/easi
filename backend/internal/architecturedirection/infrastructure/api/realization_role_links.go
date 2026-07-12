package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

const (
	realizationRolesPath   sharedAPI.ResourcePath = "/realization-roles"
	realizationRoleSubPath sharedAPI.ResourcePath = "/realization-role"
)

type RealizationRoleLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewRealizationRoleLinks(h *sharedAPI.HATEOASLinks) *RealizationRoleLinks {
	return &RealizationRoleLinks{HATEOASLinks: h}
}

func (h *RealizationRoleLinks) ItemLinks(capabilityID, componentID string, actor sharedctx.Actor) sharedAPI.Links {
	base := realizationRoleItemResourcePath(capabilityID, componentID)
	links := sharedAPI.Links{"self": h.Get(base)}
	if actor.CanWrite(ArchitectureDirectionResource) {
		links["edit"] = h.Put(base)
		links["delete"] = h.Del(base)
	}
	return links
}

func (h *RealizationRoleLinks) CollectionLinks(selfPath string, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get(selfPath)}
	if actor.CanWrite(ArchitectureDirectionResource) {
		links["x-assign"] = h.Put(templatedRealizationRoleItemPath())
	}
	return links
}

func realizationRoleItemResourcePath(capabilityID, componentID string) string {
	return string(capabilitiesPath) + "/" + capabilityID + "/components/" + componentID + string(realizationRoleSubPath)
}

func templatedRealizationRoleItemPath() string {
	return string(capabilitiesPath) + "/{capabilityId}/components/{componentId}" + string(realizationRoleSubPath)
}
