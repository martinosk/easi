package tools

import (
	"context"
	"errors"
	"fmt"
)

type AccessClass string

const (
	AccessRead  AccessClass = "read"
	AccessWrite AccessClass = "write"
)

var (
	ErrToolNotFound     = errors.New("tool not found")
	ErrPermissionDenied = errors.New("permission denied for tool")
)

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []ParameterDef
	Permission  string
	Access      AccessClass
}

type ParameterDef struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

type ToolResult struct {
	Content string
	IsError bool
}

type ToolExecutor interface {
	Execute(ctx context.Context, args map[string]interface{}) ToolResult
}

type PermissionChecker interface {
	HasPermission(permission string) bool
}

type LLMToolDef struct {
	Type     string         `json:"type"`
	Function LLMFunctionDef `json:"function"`
}

type LLMFunctionDef struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Parameters  LLMParametersDef `json:"parameters"`
}

type LLMParametersDef struct {
	Type       string                    `json:"type"`
	Properties map[string]LLMPropertyDef `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

type LLMPropertyDef struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type registeredTool struct {
	definition ToolDefinition
	executor   ToolExecutor
}

type Registry struct {
	tools map[string]*registeredTool
	order []string
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]*registeredTool),
	}
}

func (r *Registry) Register(def ToolDefinition, executor ToolExecutor) {
	r.tools[def.Name] = &registeredTool{
		definition: def,
		executor:   executor,
	}
	r.order = append(r.order, def.Name)
}

func (r *Registry) AvailableTools(permissions PermissionChecker, allowWriteOperations bool) []ToolDefinition {
	var result []ToolDefinition
	for _, name := range r.order {
		tool := r.tools[name]
		if !permissions.HasPermission(tool.definition.Permission) {
			continue
		}
		if tool.definition.Access == AccessWrite && !allowWriteOperations {
			continue
		}
		result = append(result, tool.definition)
	}
	return result
}

func (r *Registry) Execute(ctx context.Context, permissions PermissionChecker, name string, args map[string]interface{}) (ToolResult, error) {
	tool, exists := r.tools[name]
	if !exists {
		return ToolResult{}, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}
	if !permissions.HasPermission(tool.definition.Permission) {
		return ToolResult{}, fmt.Errorf("%w: %s", ErrPermissionDenied, name)
	}
	return tool.executor.Execute(ctx, args), nil
}

func (r *Registry) FormatForLLM(permissions PermissionChecker, allowWriteOperations bool) []LLMToolDef {
	available := r.AvailableTools(permissions, allowWriteOperations)
	result := make([]LLMToolDef, len(available))
	for i, def := range available {
		result[i] = toLLMToolDef(def)
	}
	return result
}

func toLLMToolDef(def ToolDefinition) LLMToolDef {
	properties := make(map[string]LLMPropertyDef, len(def.Parameters))
	var required []string
	for _, p := range def.Parameters {
		properties[p.Name] = LLMPropertyDef{
			Type:        p.Type,
			Description: p.Description,
		}
		if p.Required {
			required = append(required, p.Name)
		}
	}
	return LLMToolDef{
		Type: "function",
		Function: LLMFunctionDef{
			Name:        def.Name,
			Description: def.Description,
			Parameters: LLMParametersDef{
				Type:       "object",
				Properties: properties,
				Required:   required,
			},
		},
	}
}
