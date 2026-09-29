package api

import (
	"net/url"

	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/stewardship/application/readmodels"
)

const (
	stewardshipsPath   sharedAPI.ResourcePath = "/stewardships"
	domainsPermission  sharedctx.ResourceName = "domains"
	businessDomainPath sharedAPI.ResourcePath = "/business-domains"
	activeUsersPath    sharedAPI.ResourcePath = "/users?status=active"
)

type StewardshipLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewStewardshipLinks(h *sharedAPI.HATEOASLinks) *StewardshipLinks {
	return &StewardshipLinks{HATEOASLinks: h}
}

func collectionPath(domainID string) string {
	return string(stewardshipsPath) + "?domainId=" + url.QueryEscape(domainID)
}

func itemPath(domainID, concern string) string {
	return string(stewardshipsPath) + "/" + url.PathEscape(domainID) + "/" + url.PathEscape(concern)
}

func (h *StewardshipLinks) ItemLinks(domainID string, item readmodels.ConcernStewardshipDTO, actor sharedctx.Actor) sharedAPI.Links {
	p := itemPath(domainID, item.Concern)
	links := sharedAPI.Links{
		"self":       h.Get(p),
		"collection": h.Get(collectionPath(domainID)),
	}
	if !actor.CanWrite(domainsPermission) {
		return links
	}
	links["x-assign"] = h.Put(p)
	if item.Steward != nil {
		links["x-release"] = h.Del(p)
	}
	return links
}

func (h *StewardshipLinks) CollectionLinks(domainID string, actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{
		"self":     h.Get(collectionPath(domainID)),
		"x-domain": h.Get(string(businessDomainPath) + "/" + url.PathEscape(domainID)),
	}
	if actor.CanWrite(domainsPermission) {
		links["x-candidates"] = h.Get(string(activeUsersPath))
	}
	return links
}
