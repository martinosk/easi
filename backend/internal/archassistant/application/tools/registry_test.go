package tools_test

import (
	"context"
	"testing"

	"easi/backend/internal/archassistant/application/tools"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPermissions struct {
	permissions map[string]bool
}

func (m *mockPermissions) HasPermission(perm string) bool {
	return m.permissions[perm]
}

type mockExecutor struct {
	result tools.ToolResult
}

func (m *mockExecutor) Execute(_ context.Context, _ map[string]interface{}) tools.ToolResult {
	return m.result
}

func newTool(name, permission string, access tools.AccessClass) tools.ToolDefinition {
	return tools.ToolDefinition{
		Name:        name,
		Description: name + " description",
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "identifier", Required: true},
		},
		Permission: permission,
		Access:     access,
	}
}

func successExecutor() *mockExecutor {
	return &mockExecutor{result: tools.ToolResult{Content: "ok"}}
}

func permsFor(allowed ...string) *mockPermissions {
	m := &mockPermissions{permissions: make(map[string]bool, len(allowed))}
	for _, p := range allowed {
		m.permissions[p] = true
	}
	return m
}

func mixedToolRegistry() *tools.Registry {
	registry := tools.NewRegistry()
	registry.Register(newTool("list_components", "components:read", tools.AccessRead), successExecutor())
	registry.Register(newTool("create_component", "components:write", tools.AccessWrite), successExecutor())
	registry.Register(newTool("list_vendors", "vendors:read", tools.AccessRead), successExecutor())
	return registry
}

func toolNames(defs []tools.ToolDefinition) []string {
	names := make([]string, len(defs))
	for i, d := range defs {
		names[i] = d.Name
	}
	return names
}

func TestRegistry_AvailableTools_ReadOnlyMode(t *testing.T) {
	registry := mixedToolRegistry()
	allPerms := permsFor("components:read", "components:write", "vendors:read")

	available := registry.AvailableTools(allPerms, false)

	names := toolNames(available)
	assert.ElementsMatch(t, []string{"list_components", "list_vendors"}, names)
}

func TestRegistry_AvailableTools_WriteMode(t *testing.T) {
	registry := mixedToolRegistry()
	allPerms := permsFor("components:read", "components:write", "vendors:read")

	available := registry.AvailableTools(allPerms, true)

	names := toolNames(available)
	assert.ElementsMatch(t, []string{"list_components", "create_component", "list_vendors"}, names)
}

func TestRegistry_AvailableTools_FiltersUnpermittedTools(t *testing.T) {
	registry := mixedToolRegistry()
	limitedPerms := permsFor("components:read")

	available := registry.AvailableTools(limitedPerms, true)

	names := toolNames(available)
	assert.ElementsMatch(t, []string{"list_components"}, names)
}

func TestRegistry_AvailableTools_EmptyWhenNoPermissions(t *testing.T) {
	registry := mixedToolRegistry()

	available := registry.AvailableTools(permsFor(), true)

	assert.Empty(t, available)
}

func TestRegistry_Execute_Success(t *testing.T) {
	registry := tools.NewRegistry()
	expected := tools.ToolResult{Content: `{"id":"comp-1"}`}
	registry.Register(newTool("list_components", "components:read", tools.AccessRead), &mockExecutor{result: expected})

	result, err := registry.Execute(context.Background(), permsFor("components:read"), "list_components", map[string]interface{}{"id": "comp-1"})

	require.NoError(t, err)
	assert.Equal(t, expected.Content, result.Content)
	assert.False(t, result.IsError)
}

func TestRegistry_Execute_ToolNotFound(t *testing.T) {
	registry := tools.NewRegistry()

	_, err := registry.Execute(context.Background(), permsFor(), "nonexistent", nil)

	assert.ErrorIs(t, err, tools.ErrToolNotFound)
}

func TestRegistry_Execute_PermissionDenied(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(newTool("list_components", "components:read", tools.AccessRead), successExecutor())

	_, err := registry.Execute(context.Background(), permsFor(), "list_components", nil)

	assert.ErrorIs(t, err, tools.ErrPermissionDenied)
}

func TestRegistry_FormatForLLM_CorrectFormat(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(tools.ToolDefinition{
		Name:        "get_component",
		Description: "Get a component by ID",
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Component ID", Required: true},
			{Name: "include_details", Type: "boolean", Description: "Include details"},
		},
		Permission: "components:read",
		Access:     tools.AccessRead,
	}, successExecutor())

	formatted := registry.FormatForLLM(permsFor("components:read"), false)

	require.Len(t, formatted, 1)
	fn := formatted[0]
	assert.Equal(t, "function", fn.Type)
	assert.Equal(t, "get_component", fn.Function.Name)
	assert.Equal(t, "Get a component by ID", fn.Function.Description)
	assert.Equal(t, "object", fn.Function.Parameters.Type)
	assert.Len(t, fn.Function.Parameters.Properties, 2)
	assert.Equal(t, tools.LLMPropertyDef{Type: "string", Description: "Component ID"}, fn.Function.Parameters.Properties["id"])
	assert.Equal(t, tools.LLMPropertyDef{Type: "boolean", Description: "Include details"}, fn.Function.Parameters.Properties["include_details"])
	assert.Equal(t, []string{"id"}, fn.Function.Parameters.Required)
}

func TestRegistry_Register_MultipleTools(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(newTool("tool_a", "perm:a", tools.AccessRead), successExecutor())
	registry.Register(newTool("tool_b", "perm:b", tools.AccessRead), successExecutor())
	registry.Register(newTool("tool_c", "perm:c", tools.AccessWrite), successExecutor())

	available := registry.AvailableTools(permsFor("perm:a", "perm:b", "perm:c"), true)

	assert.ElementsMatch(t, []string{"tool_a", "tool_b", "tool_c"}, toolNames(available))
}
