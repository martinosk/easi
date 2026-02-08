package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type ArchitectureModelingLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewArchitectureModelingLinks(h *sharedAPI.HATEOASLinks) *ArchitectureModelingLinks {
	return &ArchitectureModelingLinks{HATEOASLinks: h}
}

var (
	acquiredEntityConfig = sharedAPI.ResourceConfig{Path: "/acquired-entities", Collection: "/acquired-entities", Permission: "components"}
	vendorConfig         = sharedAPI.ResourceConfig{Path: "/vendors", Collection: "/vendors", Permission: "components"}
	internalTeamConfig   = sharedAPI.ResourceConfig{Path: "/internal-teams", Collection: "/internal-teams", Permission: "components"}
)

func (h *ArchitectureModelingLinks) ComponentLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/components/" + id
	links := sharedAPI.Links{
		"self":           h.Get(p),
		"describedby":    h.Get("/reference/components"),
		"collection":     h.Get("/components"),
		"x-expert-roles": h.Get("/components/expert-roles"),
	}
	h.AddEditOrGrantLink(links, actor, "components", "components", id, h.Put(p), map[string]types.Link{
		"x-add-expert": h.Post(p + "/experts"),
	})
	if actor.CanDelete("components") {
		links["delete"] = h.Del(p)
	}
	h.AddEditGrantsLink(links, actor, "components")
	return links
}

func (h *ArchitectureModelingLinks) ComponentExpertLinksForActor(p sharedAPI.ExpertParams, actor sharedctx.Actor) sharedAPI.Links {
	return h.ExpertRemoveLink(p, actor, "components")
}

func (h *ArchitectureModelingLinks) RelationLinks(id string) sharedAPI.Links {
	links := h.Crud("/relations/" + id)
	links["describedby"] = h.Get("/reference/relations/generic")
	links["collection"] = h.Get("/relations")
	return links
}

func (h *ArchitectureModelingLinks) RelationTypeLinks(relationType string) sharedAPI.Links {
	doc := "relations/generic"
	if relationType == "Triggers" {
		doc = "relations/triggering"
	} else if relationType == "Serves" {
		doc = "relations/serving"
	}
	return sharedAPI.Links{"describedby": h.Get("/reference/" + doc)}
}

func (h *ArchitectureModelingLinks) AcquiredEntityLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	return h.SimpleResourceLinks(acquiredEntityConfig, id, actor)
}

func (h *ArchitectureModelingLinks) VendorLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	return h.SimpleResourceLinks(vendorConfig, id, actor)
}

func (h *ArchitectureModelingLinks) InternalTeamLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	return h.SimpleResourceLinks(internalTeamConfig, id, actor)
}

func (h *ArchitectureModelingLinks) OriginRelationshipLinksForActor(basePath, id, componentID string, extraLinks map[string]types.Link, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{
		"self":      h.Get(basePath + "/" + id),
		"component": h.Get("/components/" + componentID),
	}
	for k, v := range extraLinks {
		links[k] = v
	}
	if actor.CanDelete("components") {
		links["delete"] = h.Del(basePath + "/" + id)
	}
	return links
}
