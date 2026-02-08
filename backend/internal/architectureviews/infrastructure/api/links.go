package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type ViewLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewViewLinks(h *sharedAPI.HATEOASLinks) *ViewLinks {
	return &ViewLinks{HATEOASLinks: h}
}

type ViewInfo struct {
	ID          string
	IsPrivate   bool
	IsDefault   bool
	OwnerUserID *string
}

func (v ViewInfo) isOwnedBy(actorID string) bool {
	return v.OwnerUserID != nil && *v.OwnerUserID == actorID
}

func (v ViewInfo) canBeEditedBy(actor sharedctx.Actor) bool {
	return (!v.IsPrivate || v.isOwnedBy(actor.ID)) && actor.CanWrite("views")
}

func (v ViewInfo) canBeDeletedBy(actor sharedctx.Actor) bool {
	return (!v.IsPrivate || v.isOwnedBy(actor.ID)) && actor.CanDelete("views") && !v.IsDefault
}

func (h *ViewLinks) ViewLinksForActor(v ViewInfo, actor sharedctx.Actor) sharedAPI.Links {
	p := "/views/" + v.ID
	links := sharedAPI.Links{
		"self":         h.Get(p),
		"x-components": h.Get(p + "/components"),
		"collection":   h.Get("/views"),
	}
	if v.canBeEditedBy(actor) {
		links["edit"] = h.Patch(p + "/name")
		links["x-change-visibility"] = h.Patch(p + "/visibility")
	} else if actor.HasEditGrant("views", v.ID) {
		links["edit"] = h.Patch(p + "/name")
		links["x-change-visibility"] = h.Patch(p + "/visibility")
	}
	if v.canBeDeletedBy(actor) {
		links["delete"] = h.Del(p)
	}
	h.AddEditGrantsLink(links, actor, "views")
	return links
}
