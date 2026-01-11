package context

import (
	"context"
)

const ActorContextKey contextKey = "actor"

type Actor struct {
	ID          string
	Email       string
	Role        string
	Permissions map[string]bool
}

func (a Actor) HasPermission(perm string) bool {
	if a.Permissions == nil {
		return false
	}
	return a.Permissions[perm]
}

func (a Actor) CanWrite(resource string) bool {
	return a.HasPermission(resource + ":write")
}

func (a Actor) CanDelete(resource string) bool {
	return a.HasPermission(resource + ":delete")
}

func (a Actor) CanRead(resource string) bool {
	return a.HasPermission(resource + ":read")
}

func NewActor(id, email, role string) Actor {
	return Actor{
		ID:          id,
		Email:       email,
		Role:        role,
		Permissions: PermissionsForRole(role),
	}
}

func PermissionsForRole(role string) map[string]bool {
	perms := make(map[string]bool)

	switch role {
	case "admin":
		perms["components:read"] = true
		perms["components:write"] = true
		perms["components:delete"] = true
		perms["views:read"] = true
		perms["views:write"] = true
		perms["views:delete"] = true
		perms["capabilities:read"] = true
		perms["capabilities:write"] = true
		perms["capabilities:delete"] = true
		perms["domains:read"] = true
		perms["domains:write"] = true
		perms["domains:delete"] = true
		perms["users:read"] = true
		perms["users:manage"] = true
		perms["invitations:manage"] = true
		perms["metamodel:read"] = true
		perms["metamodel:write"] = true
		perms["audit:read"] = true
		perms["enterprise-arch:read"] = true
		perms["enterprise-arch:write"] = true
		perms["enterprise-arch:delete"] = true

	case "architect":
		perms["components:read"] = true
		perms["components:write"] = true
		perms["components:delete"] = true
		perms["views:read"] = true
		perms["views:write"] = true
		perms["views:delete"] = true
		perms["capabilities:read"] = true
		perms["capabilities:write"] = true
		perms["capabilities:delete"] = true
		perms["domains:read"] = true
		perms["domains:write"] = true
		perms["domains:delete"] = true
		perms["metamodel:read"] = true
		perms["audit:read"] = true
		perms["enterprise-arch:read"] = true
		perms["enterprise-arch:write"] = true
		perms["enterprise-arch:delete"] = true

	case "stakeholder":
		perms["components:read"] = true
		perms["views:read"] = true
		perms["capabilities:read"] = true
		perms["domains:read"] = true
		perms["metamodel:read"] = true
		perms["audit:read"] = true
		perms["enterprise-arch:read"] = true
	}

	return perms
}

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, ActorContextKey, actor)
}

func GetActor(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(ActorContextKey).(Actor)
	return actor, ok
}
