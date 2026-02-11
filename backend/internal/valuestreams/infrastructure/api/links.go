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
		"self":       h.Get(p),
		"collection": h.Get("/value-streams"),
	}
	if actor.CanWrite("valuestreams") {
		links["edit"] = h.Put(p)
		links["x-add-stage"] = h.Post(p + "/stages")
	}
	if actor.CanDelete("valuestreams") {
		links["delete"] = h.Del(p)
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
