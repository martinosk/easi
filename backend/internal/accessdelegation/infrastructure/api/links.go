package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type EditGrantLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewEditGrantLinks(h *sharedAPI.HATEOASLinks) *EditGrantLinks {
	return &EditGrantLinks{HATEOASLinks: h}
}

func (h *EditGrantLinks) EditGrantLinksForActor(id, status, grantorID string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/edit-grants/" + id
	links := sharedAPI.Links{
		"self":       h.Get(p),
		"collection": h.Get("/edit-grants"),
	}
	if canRevokeEditGrant(status, grantorID, actor) {
		links["delete"] = h.Del(p)
	}
	return links
}

func canRevokeEditGrant(status, grantorID string, actor sharedctx.Actor) bool {
	return status == "active" && (grantorID == actor.ID || actor.Role == "admin")
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
