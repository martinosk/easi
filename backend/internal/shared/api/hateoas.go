package api

import (
	"net/url"

	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type Links = types.Links

type Method string

const (
	MethodGet    Method = "GET"
	MethodPut    Method = "PUT"
	MethodPost   Method = "POST"
	MethodDelete Method = "DELETE"
	MethodPatch  Method = "PATCH"
)

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

func (h *HATEOASLinks) Get(path string) types.Link   { return NewLink(h.base+path, MethodGet) }
func (h *HATEOASLinks) Put(path string) types.Link   { return NewLink(h.base+path, MethodPut) }
func (h *HATEOASLinks) Post(path string) types.Link  { return NewLink(h.base+path, MethodPost) }
func (h *HATEOASLinks) Del(path string) types.Link   { return NewLink(h.base+path, MethodDelete) }
func (h *HATEOASLinks) Patch(path string) types.Link { return NewLink(h.base+path, MethodPatch) }

func (h *HATEOASLinks) Crud(path string) Links {
	return Links{"self": h.Get(path), "edit": h.Put(path), "delete": h.Del(path)}
}

type ResourceConfig struct {
	Path       string
	Collection string
	Permission sharedctx.ResourceName
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

func (h *HATEOASLinks) ExpertRemoveLink(p ExpertParams, actor sharedctx.Actor, permission sharedctx.ResourceName) Links {
	links := Links{}
	if actor.CanDelete(permission) {
		links["x-remove"] = h.Del(p.ResourcePath + "/experts?name=" + url.QueryEscape(p.ExpertName) + "&role=" + url.QueryEscape(p.ExpertRole) + "&contact=" + url.QueryEscape(p.ContactInfo))
	}
	return links
}

type EditGrantParams struct {
	Permission   sharedctx.ResourceName
	ArtifactType sharedctx.ResourceName
	ArtifactID   string
	EditLink     types.Link
	ExtraWrite   map[string]types.Link
}

func (h *HATEOASLinks) AddEditOrGrantLink(links Links, actor sharedctx.Actor, params EditGrantParams) {
	if actor.CanWrite(params.Permission) {
		links["edit"] = params.EditLink
		for k, v := range params.ExtraWrite {
			links[k] = v
		}
	} else if actor.HasEditGrant(params.ArtifactType, params.ArtifactID) {
		links["edit"] = params.EditLink
	}
}

func (h *HATEOASLinks) AddEditGrantsLink(links Links, actor sharedctx.Actor, permission sharedctx.ResourceName) {
	if actor.CanWrite(permission) || actor.HasPermission("edit-grants:manage") {
		links["x-edit-grants"] = h.Post("/edit-grants")
	}
}

func (h *HATEOASLinks) ReferenceDocLink(resourceType string) string {
	return h.base + "/reference/" + resourceType
}

func NewLink(href string, method Method) types.Link {
	return types.Link{Href: href, Method: string(method)}
}
