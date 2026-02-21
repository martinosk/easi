package toolimpls

import (
	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
)

func RegisterSpecTools(registry *tools.Registry, client *agenthttp.Client) {
	for _, spec := range AllContextToolSpecs() {
		allParams := collectAllParams(spec)
		registry.Register(tools.ToolDefinition{
			Name:        spec.Name,
			Description: spec.Description,
			Permission:  spec.Permission,
			Access:      spec.Access,
			Parameters:  allParams,
		}, NewGenericExecutor(spec, client))
	}
}

func collectAllParams(spec AgentToolSpec) []tools.ParameterDef {
	var params []tools.ParameterDef
	for _, groups := range [][]ParamSpec{spec.PathParams, spec.QueryParams, spec.BodyParams} {
		for _, p := range groups {
			paramType := p.Type
			if paramType == "uuid" {
				paramType = "string"
			}
			params = append(params, tools.ParameterDef{
				Name:        p.Name,
				Type:        paramType,
				Description: p.Description,
				Required:    p.Required,
			})
		}
	}
	return params
}
