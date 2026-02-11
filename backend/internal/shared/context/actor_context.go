package context

import (
	"context"
)

const ActorContextKey contextKey = "actor"

type ResourceName string

type Role string

const (
	RoleAdmin       Role = "admin"
	RoleArchitect   Role = "architect"
	RoleStakeholder Role = "stakeholder"
)

func (r Role) String() string { return string(r) }

func (r Role) Permissions() map[string]bool {
	permList, ok := rolePermissions[r]
	if !ok {
		return make(map[string]bool)
	}
	perms := make(map[string]bool, len(permList))
	for _, p := range permList {
		perms[p] = true
	}
	return perms
}

type Actor struct {
	ID          string
	Email       string
	Role        Role
	Permissions map[string]bool
	editGrants  map[ResourceName]map[string]bool
}

func (a Actor) HasPermission(perm string) bool {
	if a.Permissions == nil {
		return false
	}
	return a.Permissions[perm]
}

func (a Actor) CanWrite(resource ResourceName) bool {
	return a.HasPermission(string(resource) + ":write")
}

func (a Actor) CanDelete(resource ResourceName) bool {
	return a.HasPermission(string(resource) + ":delete")
}

func (a Actor) CanRead(resource ResourceName) bool {
	return a.HasPermission(string(resource) + ":read")
}

func (a Actor) HasEditGrant(artifactType ResourceName, artifactID string) bool {
	if a.editGrants == nil {
		return false
	}
	ids, ok := a.editGrants[artifactType]
	if !ok {
		return false
	}
	return ids[artifactID]
}

func (a Actor) EditGrantIDs(artifactType ResourceName) map[string]bool {
	if a.editGrants == nil {
		return nil
	}
	return a.editGrants[artifactType]
}

var resourceAliases = map[string]ResourceName{
	"capability":      "capabilities",
	"component":       "components",
	"view":            "views",
	"domain":          "domains",
	"vendor":          "vendors",
	"internal_team":   "internal_teams",
	"acquired_entity": "acquired_entities",
}

func PluralResourceName(singular string) ResourceName {
	if plural, ok := resourceAliases[singular]; ok {
		return plural
	}
	return ResourceName(singular)
}

func (a Actor) WithEditGrants(grants map[string]map[string]bool) Actor {
	normalized := make(map[ResourceName]map[string]bool, len(grants))
	for key, ids := range grants {
		normalized[PluralResourceName(key)] = ids
	}
	a.editGrants = normalized
	return a
}

func NewActor(id, email string, role Role) Actor {
	return Actor{
		ID:          id,
		Email:       email,
		Role:        role,
		Permissions: role.Permissions(),
	}
}

var rolePermissions = map[Role][]string{
	RoleAdmin: {
		"components:read", "components:write", "components:delete",
		"views:read", "views:write", "views:delete",
		"capabilities:read", "capabilities:write", "capabilities:delete",
		"domains:read", "domains:write", "domains:delete",
		"users:read", "users:manage",
		"invitations:manage",
		"metamodel:read", "metamodel:write",
		"audit:read",
		"enterprise-arch:read", "enterprise-arch:write", "enterprise-arch:delete",
		"edit-grants:manage",
		"valuestreams:read", "valuestreams:write", "valuestreams:delete",
	},
	RoleArchitect: {
		"components:read", "components:write", "components:delete",
		"views:read", "views:write", "views:delete",
		"capabilities:read", "capabilities:write", "capabilities:delete",
		"domains:read", "domains:write", "domains:delete",
		"users:read",
		"metamodel:read",
		"audit:read",
		"enterprise-arch:read", "enterprise-arch:write", "enterprise-arch:delete",
		"edit-grants:manage",
		"valuestreams:read", "valuestreams:write", "valuestreams:delete",
	},
	RoleStakeholder: {
		"components:read",
		"views:read",
		"capabilities:read",
		"domains:read",
		"metamodel:read",
		"audit:read",
		"enterprise-arch:read",
		"valuestreams:read",
	},
}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, ActorContextKey, actor)
}

func GetActor(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(ActorContextKey).(Actor)
	return actor, ok
}
