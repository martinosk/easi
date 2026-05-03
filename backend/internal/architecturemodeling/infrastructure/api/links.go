package api

import (
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
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

type originResourceConfig struct {
	sharedAPI.ResourceConfig
	ArtifactType sharedctx.ResourceName
}

var (
	acquiredEntityConfig = originResourceConfig{sharedAPI.ResourceConfig{Path: "/acquired-entities", Collection: "/acquired-entities", Permission: "components"}, "acquired_entities"}
	vendorConfig         = originResourceConfig{sharedAPI.ResourceConfig{Path: "/vendors", Collection: "/vendors", Permission: "components"}, "vendors"}
	internalTeamConfig   = originResourceConfig{sharedAPI.ResourceConfig{Path: "/internal-teams", Collection: "/internal-teams", Permission: "components"}, "internal_teams"}
)

func (h *ArchitectureModelingLinks) ComponentLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/components/" + id
	links := sharedAPI.Links{
		"self":           h.Get(p),
		"describedby":    h.Get("/reference/components"),
		"collection":     h.Get("/components"),
		"x-expert-roles": h.Get("/components/expert-roles"),
	}
	h.AddEditOrGrantLink(links, actor, sharedAPI.EditGrantParams{
		Permission:   "components",
		ArtifactType: "components",
		ArtifactID:   id,
		EditLink:     h.Put(p),
		ExtraWrite: map[string]types.Link{
			"x-add-expert": h.Post(p + "/experts"),
		},
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

type relatedLinkSpec struct {
	HrefSuffix   string
	Title        string
	TargetType   string
	RelationType string
}

var (
	componentTriggersSpec   = relatedLinkSpec{HrefSuffix: "/components", Title: "Component (triggers)", TargetType: "component", RelationType: "component-triggers"}
	componentServesSpec     = relatedLinkSpec{HrefSuffix: "/components", Title: "Component (serves)", TargetType: "component", RelationType: "component-serves"}
	componentAcquiredVia    = relatedLinkSpec{HrefSuffix: "/acquired-entities", Title: "Acquired Entity (acquired-via)", TargetType: "acquiredEntity", RelationType: "origin-acquired-via"}
	componentPurchasedFrom  = relatedLinkSpec{HrefSuffix: "/vendors", Title: "Vendor (purchased-from)", TargetType: "vendor", RelationType: "origin-purchased-from"}
	componentBuiltBy        = relatedLinkSpec{HrefSuffix: "/internal-teams", Title: "Internal Team (built-by)", TargetType: "internalTeam", RelationType: "origin-built-by"}
	acquiredEntitySpec      = relatedLinkSpec{HrefSuffix: "/components", Title: "Component (acquired-via)", TargetType: "component", RelationType: "origin-acquired-via"}
	vendorSpec              = relatedLinkSpec{HrefSuffix: "/components", Title: "Component (purchased-from)", TargetType: "component", RelationType: "origin-purchased-from"}
	internalTeamSpec        = relatedLinkSpec{HrefSuffix: "/components", Title: "Component (built-by)", TargetType: "component", RelationType: "origin-built-by"}
)

func (h *ArchitectureModelingLinks) buildRelated(spec relatedLinkSpec) types.RelatedLink {
	return types.RelatedLink{
		Href:         h.Base() + spec.HrefSuffix,
		Methods:      []string{"POST"},
		Title:        spec.Title,
		TargetType:   spec.TargetType,
		RelationType: spec.RelationType,
	}
}

func (h *ArchitectureModelingLinks) gatedRelated(specs []relatedLinkSpec, actor sharedctx.Actor) []types.RelatedLink {
	if !actor.CanWrite("components") {
		return []types.RelatedLink{}
	}
	related := make([]types.RelatedLink, 0, len(specs))
	for _, spec := range specs {
		related = append(related, h.buildRelated(spec))
	}
	return related
}

var componentXRelatedSpecs = []relatedLinkSpec{
	componentTriggersSpec,
	componentServesSpec,
	componentAcquiredVia,
	componentPurchasedFrom,
	componentBuiltBy,
}

func (h *ArchitectureModelingLinks) ComponentXRelatedForActor(actor sharedctx.Actor) []types.RelatedLink {
	return h.gatedRelated(componentXRelatedSpecs, actor)
}

func (h *ArchitectureModelingLinks) AcquiredEntityXRelatedForActor(actor sharedctx.Actor) []types.RelatedLink {
	return h.gatedRelated([]relatedLinkSpec{acquiredEntitySpec}, actor)
}

func (h *ArchitectureModelingLinks) VendorXRelatedForActor(actor sharedctx.Actor) []types.RelatedLink {
	return h.gatedRelated([]relatedLinkSpec{vendorSpec}, actor)
}

func (h *ArchitectureModelingLinks) InternalTeamXRelatedForActor(actor sharedctx.Actor) []types.RelatedLink {
	return h.gatedRelated([]relatedLinkSpec{internalTeamSpec}, actor)
}

func (h *ArchitectureModelingLinks) RelationLinks(id string) sharedAPI.Links {
	links := h.Crud("/relations/" + id)
	links["describedby"] = h.Get("/reference/relations/generic")
	links["collection"] = h.Get("/relations")
	return links
}

func (h *ArchitectureModelingLinks) RelationTypeLinks(relationType valueobjects.RelationType) sharedAPI.Links {
	doc := "relations/generic"
	switch relationType {
	case valueobjects.RelationTypeTriggers:
		doc = "relations/triggering"
	case valueobjects.RelationTypeServes:
		doc = "relations/serving"
	}
	return sharedAPI.Links{"describedby": h.Get("/reference/" + doc)}
}

func (h *ArchitectureModelingLinks) originEntityLinksForActor(cfg originResourceConfig, id string, actor sharedctx.Actor) sharedAPI.Links {
	p := cfg.Path + "/" + id
	links := sharedAPI.Links{
		"self":       h.Get(p),
		"collection": h.Get(cfg.Collection),
	}
	h.AddEditOrGrantLink(links, actor, sharedAPI.EditGrantParams{
		Permission:   cfg.Permission,
		ArtifactType: cfg.ArtifactType,
		ArtifactID:   id,
		EditLink:     h.Put(p),
	})
	if actor.CanDelete(cfg.Permission) {
		links["delete"] = h.Del(p)
	}
	h.AddEditGrantsLink(links, actor, cfg.Permission)
	return links
}

func (h *ArchitectureModelingLinks) AcquiredEntityLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	return h.originEntityLinksForActor(acquiredEntityConfig, id, actor)
}

func (h *ArchitectureModelingLinks) VendorLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	return h.originEntityLinksForActor(vendorConfig, id, actor)
}

func (h *ArchitectureModelingLinks) InternalTeamLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	return h.originEntityLinksForActor(internalTeamConfig, id, actor)
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
