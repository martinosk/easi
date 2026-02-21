package tools

var agentPermissionCeiling = map[string]bool{
	"components:read":       true,
	"components:write":      true,
	"capabilities:read":     true,
	"capabilities:write":    true,
	"domains:read":          true,
	"domains:write":         true,
	"valuestreams:read":     true,
	"valuestreams:write":    true,
	"enterprise-arch:read":  true,
	"enterprise-arch:write": true,
	"views:read":            true,
	"metamodel:read":        true,
	"assistant:use":         true,
}

type AgentScopedPermissions struct {
	inner PermissionChecker
}

func NewAgentScopedPermissions(inner PermissionChecker) *AgentScopedPermissions {
	return &AgentScopedPermissions{inner: inner}
}

func (a *AgentScopedPermissions) HasPermission(permission string) bool {
	return agentPermissionCeiling[permission] && a.inner.HasPermission(permission)
}
