package tools_test

import (
	"testing"

	"easi/backend/internal/archassistant/application/tools"

	"github.com/stretchr/testify/assert"
)

func TestAgentScopedPermissions_BlocksNonCeilingPermissions(t *testing.T) {
	actor := permsFor(
		"components:read", "components:write",
		"users:manage", "invitations:manage", "audit:read",
	)

	scoped := tools.NewAgentScopedPermissions(actor)

	assert.True(t, scoped.HasPermission("components:read"))
	assert.True(t, scoped.HasPermission("components:write"))
	assert.False(t, scoped.HasPermission("users:manage"))
	assert.False(t, scoped.HasPermission("invitations:manage"))
	assert.False(t, scoped.HasPermission("audit:read"))
}

func TestAgentScopedPermissions_BlocksWhenUserLacksPermission(t *testing.T) {
	actor := permsFor("components:read")

	scoped := tools.NewAgentScopedPermissions(actor)

	assert.True(t, scoped.HasPermission("components:read"))
	assert.False(t, scoped.HasPermission("components:write"), "ceiling allows but user lacks")
}

func TestAgentScopedPermissions_AllowsAllCeilingPermissions(t *testing.T) {
	allowed := []string{
		"components:read", "components:write",
		"capabilities:read", "capabilities:write",
		"domains:read", "domains:write",
		"valuestreams:read", "valuestreams:write",
		"enterprise-arch:read", "enterprise-arch:write",
		"views:read",
		"metamodel:read",
		"assistant:use",
	}
	actor := permsFor(allowed...)

	scoped := tools.NewAgentScopedPermissions(actor)

	for _, perm := range allowed {
		assert.True(t, scoped.HasPermission(perm), "ceiling should allow %s", perm)
	}
}

func TestAgentScopedPermissions_BlocksAllExcludedPermissions(t *testing.T) {
	excluded := []string{
		"users:manage",
		"invitations:manage",
		"edit-grants:manage",
		"audit:read",
		"metamodel:write",
		"importing:write",
	}
	actor := permsFor(excluded...)

	scoped := tools.NewAgentScopedPermissions(actor)

	for _, perm := range excluded {
		assert.False(t, scoped.HasPermission(perm), "ceiling should block %s", perm)
	}
}

func TestHasAnyWritePermission(t *testing.T) {
	assert.True(t, tools.HasAnyWritePermission(permsFor("components:read", "components:write")))
	assert.True(t, tools.HasAnyWritePermission(permsFor("capabilities:write")))
	assert.False(t, tools.HasAnyWritePermission(permsFor(
		"components:read", "capabilities:read", "domains:read",
		"enterprise-arch:read", "architecture-direction:read",
		"valuestreams:read", "views:read", "metamodel:read", "assistant:use",
	)), "read-only actor has no write permission")
	assert.False(t, tools.HasAnyWritePermission(permsFor("metamodel:write")),
		"write permission outside the ceiling does not count")
}
