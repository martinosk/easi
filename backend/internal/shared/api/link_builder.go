package api

import (
	"fmt"

	"easi/backend/internal/shared/types"
)

const APIVersionPrefix = "/api/v1"

type ResourcePath string

func (p ResourcePath) String() string {
	return string(p)
}

type ResourceID string

func (id ResourceID) String() string {
	return string(id)
}

type LinkRelation string

func (r LinkRelation) String() string {
	return string(r)
}

const (
	RelSelf       LinkRelation = "self"
	RelCollection LinkRelation = "collection"
	RelEdit       LinkRelation = "edit"
	RelDelete     LinkRelation = "delete"
)

type LinkBuilder struct {
	resourcePath ResourcePath
}

func NewLinkBuilder(resourcePath ResourcePath) *LinkBuilder {
	return &LinkBuilder{resourcePath: resourcePath}
}

func (b *LinkBuilder) Self(id ResourceID) string {
	return fmt.Sprintf("%s%s/%s", APIVersionPrefix, b.resourcePath, id)
}

func (b *LinkBuilder) Collection() string {
	return fmt.Sprintf("%s%s", APIVersionPrefix, b.resourcePath)
}

func (b *LinkBuilder) Update(id ResourceID) string {
	return b.Self(id)
}

func (b *LinkBuilder) Delete(id ResourceID) string {
	return b.Self(id)
}

func (b *LinkBuilder) SubResource(id ResourceID, subPath ResourcePath) string {
	return BuildSubResourceLink(b.resourcePath, id, subPath)
}

func (b *LinkBuilder) Related(relatedPath ResourcePath, relatedID ResourceID) string {
	return fmt.Sprintf("%s%s/%s", APIVersionPrefix, relatedPath, relatedID)
}

func BuildLink(path ResourcePath) string {
	return fmt.Sprintf("%s%s", APIVersionPrefix, path)
}

func BuildResourceLink(resourcePath ResourcePath, id ResourceID) string {
	return fmt.Sprintf("%s%s/%s", APIVersionPrefix, resourcePath, id)
}

func BuildSubResourceLink(resourcePath ResourcePath, id ResourceID, subPath ResourcePath) string {
	return fmt.Sprintf("%s%s/%s%s", APIVersionPrefix, resourcePath, id, subPath)
}

type ResourceLinks struct {
	links types.Links
}

func NewResourceLinks() *ResourceLinks {
	return &ResourceLinks{links: make(types.Links)}
}

func (r *ResourceLinks) Add(rel LinkRelation, href string, method string) *ResourceLinks {
	r.links[rel.String()] = types.Link{Href: href, Method: method}
	return r
}

func (r *ResourceLinks) Self(path ResourcePath) *ResourceLinks {
	return r.Add(RelSelf, BuildLink(path), "GET")
}

func (r *ResourceLinks) SelfWithID(resourcePath ResourcePath, id ResourceID) *ResourceLinks {
	return r.Add(RelSelf, BuildResourceLink(resourcePath, id), "GET")
}

func (r *ResourceLinks) Collection(resourcePath ResourcePath) *ResourceLinks {
	return r.Add(RelCollection, BuildLink(resourcePath), "GET")
}

func (r *ResourceLinks) Edit(resourcePath ResourcePath, id ResourceID) *ResourceLinks {
	return r.Add(RelEdit, BuildResourceLink(resourcePath, id), "PUT")
}

func (r *ResourceLinks) Delete(resourcePath ResourcePath, id ResourceID) *ResourceLinks {
	return r.Add(RelDelete, BuildResourceLink(resourcePath, id), "DELETE")
}

func (r *ResourceLinks) Related(rel LinkRelation, resourcePath ResourcePath, id ResourceID) *ResourceLinks {
	return r.Add(rel, BuildResourceLink(resourcePath, id), "GET")
}

func (r *ResourceLinks) SubResource(rel LinkRelation, resourcePath ResourcePath, id ResourceID, subPath ResourcePath) *ResourceLinks {
	return r.Add(rel, BuildSubResourceLink(resourcePath, id, subPath), "GET")
}

func (r *ResourceLinks) SelfSubResource(resourcePath ResourcePath, id ResourceID, subPath ResourcePath) *ResourceLinks {
	return r.Add(RelSelf, BuildSubResourceLink(resourcePath, id, subPath), "GET")
}

func (r *ResourceLinks) Build() types.Links {
	return r.links
}
