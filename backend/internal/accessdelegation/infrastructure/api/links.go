package api

import (
	"easi/backend/internal/accessdelegation/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type EditGrantLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewEditGrantLinks(h *sharedAPI.HATEOASLinks) *EditGrantLinks {
	return &EditGrantLinks{HATEOASLinks: h}
}

func (h *EditGrantLinks) EditGrantLinksForActor(grant readmodels.EditGrantDTO, actor sharedctx.Actor) sharedAPI.Links {
	p := "/edit-grants/" + grant.ID
	links := sharedAPI.Links{
		"self":       h.Get(p),
		"collection": h.Get("/edit-grants"),
	}
	if canRevokeEditGrant(grant, actor) {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *EditGrantLinks) AddArtifactLink(links sharedAPI.Links, grant readmodels.EditGrantDTO) {
	resourcePath := "/" + sharedctx.PluralResourceName(grant.ArtifactType) + "/" + grant.ArtifactID
	links["artifact"] = h.Get(resourcePath)
}

func canRevokeEditGrant(grant readmodels.EditGrantDTO, actor sharedctx.Actor) bool {
	return grant.Status == "active" && (grant.GrantorID == actor.ID || actor.HasPermission("edit-grants:manage"))
}

func (h *EditGrantLinks) EditGrantCollectionLinksForActor(actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get("/edit-grants")}
	if actor.HasPermission("edit-grants:manage") {
		links["create"] = h.Post("/edit-grants")
	}
	return links
}

func (h *EditGrantLinks) EditGrantArtifactCollectionLinks(artifactType, artifactID string) sharedAPI.Links {
	return sharedAPI.Links{
		"self":       h.Get("/edit-grants/artifact/" + artifactType + "/" + artifactID),
		"collection": h.Get("/edit-grants"),
		"x-artifact": h.Get("/" + sharedctx.PluralResourceName(artifactType) + "/" + artifactID),
	}
}
