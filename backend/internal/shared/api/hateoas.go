package api

import (
	"net/url"

	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type Links = types.Links

type HATEOASLinks struct {
	base string
}

func NewHATEOASLinks(baseURL string) *HATEOASLinks {
	if baseURL == "" {
		baseURL = APIVersionPrefix
	}
	return &HATEOASLinks{base: baseURL}
}

func (h *HATEOASLinks) Base() string { return h.base }

func (h *HATEOASLinks) L(href, method string) types.Link {
	return types.Link{Href: href, Method: method}
}

func (h *HATEOASLinks) Get(path string) types.Link   { return h.L(h.base+path, "GET") }
func (h *HATEOASLinks) Put(path string) types.Link   { return h.L(h.base+path, "PUT") }
func (h *HATEOASLinks) Post(path string) types.Link  { return h.L(h.base+path, "POST") }
func (h *HATEOASLinks) Del(path string) types.Link   { return h.L(h.base+path, "DELETE") }
func (h *HATEOASLinks) Patch(path string) types.Link { return h.L(h.base+path, "PATCH") }

func (h *HATEOASLinks) Crud(path string) Links {
	return Links{"self": h.Get(path), "edit": h.Put(path), "delete": h.Del(path)}
}

type ResourceConfig struct {
	Path       string
	Collection string
	Permission string
}

func (h *HATEOASLinks) SimpleResourceLinks(cfg ResourceConfig, id string, actor sharedctx.Actor) Links {
	p := cfg.Path + "/" + id
	links := Links{"self": h.Get(p), "collection": h.Get(cfg.Collection)}
	if actor.CanWrite(cfg.Permission) {
		links["edit"] = h.Put(p)
	}
	if actor.CanDelete(cfg.Permission) {
		links["delete"] = h.Del(p)
	}
	return links
}

type ExpertParams struct {
	ResourcePath string
	ExpertName   string
	ExpertRole   string
	ContactInfo  string
}

func (h *HATEOASLinks) ExpertRemoveLink(p ExpertParams, actor sharedctx.Actor, permission string) Links {
	links := Links{}
	if actor.CanDelete(permission) {
		links["x-remove"] = h.Del(p.ResourcePath + "/experts?name=" + url.QueryEscape(p.ExpertName) + "&role=" + url.QueryEscape(p.ExpertRole) + "&contact=" + url.QueryEscape(p.ContactInfo))
	}
	return links
}

func (h *HATEOASLinks) AddEditOrGrantLink(links Links, actor sharedctx.Actor, permission, artifactType, artifactID string, editLink types.Link, extraWriteLinks map[string]types.Link) {
	if actor.CanWrite(permission) {
		links["edit"] = editLink
		for k, v := range extraWriteLinks {
			links[k] = v
		}
	} else if actor.HasEditGrant(artifactType, artifactID) {
		links["edit"] = editLink
	}
}

func (h *HATEOASLinks) AddEditGrantsLink(links Links, actor sharedctx.Actor, permission string) {
	if actor.CanWrite(permission) || actor.HasPermission("edit-grants:manage") {
		links["x-edit-grants"] = h.Post("/edit-grants")
	}
}

func (h *HATEOASLinks) ReferenceDocLink(resourceType string) string {
	return h.base + "/reference/" + resourceType
}

func NewLink(href, method string) types.Link {
	return types.Link{Href: href, Method: method}
}
