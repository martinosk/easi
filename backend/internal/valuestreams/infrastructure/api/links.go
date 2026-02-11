package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type ValueStreamsLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewValueStreamsLinks(h *sharedAPI.HATEOASLinks) *ValueStreamsLinks {
	return &ValueStreamsLinks{HATEOASLinks: h}
}

func (h *ValueStreamsLinks) ValueStreamLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/value-streams/" + id
	links := sharedAPI.Links{
		"self":           h.Get(p),
		"collection":     h.Get("/value-streams"),
		"x-capabilities": h.Get(p + "/capabilities"),
	}
	if actor.CanWrite("valuestreams") {
		links["edit"] = h.Put(p)
		links["x-add-stage"] = h.Post(p + "/stages")
		links["x-reorder-stages"] = h.Put(p + "/stages/positions")
	}
	if actor.CanDelete("valuestreams") {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *ValueStreamsLinks) StageLinksForActor(vsID, stageID string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/value-streams/" + vsID + "/stages/" + stageID
	links := sharedAPI.Links{
		"self": h.Get(p),
	}
	if actor.CanWrite("valuestreams") {
		links["edit"] = h.Put(p)
		links["x-add-capability"] = h.Post(p + "/capabilities")
	}
	if actor.CanDelete("valuestreams") {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *ValueStreamsLinks) StageCapabilityLinksForActor(vsID, stageID, capID string, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{}
	if actor.CanWrite("valuestreams") {
		links["delete"] = h.Del("/value-streams/" + vsID + "/stages/" + stageID + "/capabilities/" + capID)
	}
	return links
}

func (h *ValueStreamsLinks) ValueStreamCollectionLinksForActor(actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get("/value-streams")}
	if actor.CanWrite("valuestreams") {
		links["x-create"] = h.Post("/value-streams")
	}
	return links
}
