package api

import (
	sharedAPI "easi/backend/internal/shared/api"
)

type MetaModelLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewMetaModelLinks(h *sharedAPI.HATEOASLinks) *MetaModelLinks {
	return &MetaModelLinks{HATEOASLinks: h}
}

func (h *MetaModelLinks) MaturityScaleLinks(isDefault bool) sharedAPI.Links {
	return h.maturityLinks("/meta-model/maturity-scale", isDefault)
}

func (h *MetaModelLinks) MetaModelConfigLinks(id string, isDefault bool) sharedAPI.Links {
	return h.maturityLinks("/meta-model/configurations/"+id, isDefault)
}

func (h *MetaModelLinks) maturityLinks(selfPath string, isDefault bool) sharedAPI.Links {
	links := sharedAPI.Links{
		"self":              h.Get(selfPath),
		"edit":              h.Put("/meta-model/maturity-scale"),
		"x-maturity-levels": h.Get("/capabilities/metadata/maturity-levels"),
	}
	if !isDefault {
		links["x-reset"] = h.Post("/meta-model/maturity-scale/reset")
	}
	return links
}

func (h *MetaModelLinks) StrategyPillarLinks(id string, isActive bool) sharedAPI.Links {
	p := "/meta-model/strategy-pillars/" + id
	links := sharedAPI.Links{"self": h.Get(p), "collection": h.Get("/meta-model/strategy-pillars")}
	if isActive {
		links["edit"] = h.Put(p)
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *MetaModelLinks) StrategyPillarsCollectionLinks() sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/meta-model/strategy-pillars"), "create": h.Post("/meta-model/strategy-pillars")}
}
