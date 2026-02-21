package toolimpls

import (
	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
)

func uuidParam(name, description string) ParamSpec {
	return ParamSpec{Name: name, Type: "uuid", Description: description, Required: true}
}

func stringParam(name, description string, required bool) ParamSpec {
	return ParamSpec{Name: name, Type: "string", Description: description, Required: required}
}

func intParam(name, description string) ParamSpec {
	return ParamSpec{Name: name, Type: "integer", Description: description}
}

var toolSpecs = []AgentToolSpec{
	{
		Name: "list_applications", Description: "List applications in the architecture portfolio. Optionally filter by name.",
		Access: tools.AccessRead, Permission: "components:read",
		Method: "GET", Path: "/components",
		QueryParams: []ParamSpec{
			stringParam("name", "Filter by application name (partial match)", false),
			intParam("limit", "Max results (1-50, default 20)"),
		},
	},
	{
		Name: "get_application_details", Description: "Get full details of an application by ID",
		Access: tools.AccessRead, Permission: "components:read",
		Method: "GET", Path: "/components/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Application ID (UUID)")},
	},
	{
		Name: "list_capabilities", Description: "List business capabilities. Optionally filter by name.",
		Access: tools.AccessRead, Permission: "capabilities:read",
		Method: "GET", Path: "/capabilities",
		QueryParams: []ParamSpec{
			stringParam("name", "Filter by capability name (partial match)", false),
			intParam("limit", "Max results (1-50, default 20)"),
		},
	},
	{
		Name: "get_capability_details", Description: "Get full details of a capability including realizations",
		Access: tools.AccessRead, Permission: "capabilities:read",
		Method: "GET", Path: "/capabilities/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Capability ID (UUID)")},
	},
	{
		Name: "list_business_domains", Description: "List all business domains",
		Access: tools.AccessRead, Permission: "domains:read",
		Method: "GET", Path: "/business-domains",
	},
	{
		Name: "get_business_domain_details", Description: "Get details of a business domain with its capabilities",
		Access: tools.AccessRead, Permission: "domains:read",
		Method: "GET", Path: "/business-domains/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Business domain ID (UUID)")},
	},
	{
		Name: "list_value_streams", Description: "List all value streams",
		Access: tools.AccessRead, Permission: "valuestreams:read",
		Method: "GET", Path: "/value-streams",
	},
	{
		Name: "get_value_stream_details", Description: "Get value stream details including stages and mapped capabilities",
		Access: tools.AccessRead, Permission: "valuestreams:read",
		Method: "GET", Path: "/value-streams/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Value stream ID (UUID)")},
	},

	{
		Name: "create_application", Description: "Create a new application in the architecture portfolio",
		Access: tools.AccessCreate, Permission: "components:write",
		Method: "POST", Path: "/components",
		BodyParams: []ParamSpec{
			stringParam("name", "Application name", true),
			stringParam("description", "Application description", false),
		},
	},
	{
		Name: "update_application", Description: "Update an existing application's properties",
		Access: tools.AccessUpdate, Permission: "components:write",
		Method: "PUT", Path: "/components/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Application ID (UUID)")},
		BodyParams: []ParamSpec{
			stringParam("name", "New application name", false),
			stringParam("description", "New application description", false),
		},
	},
	{
		Name: "delete_application", Description: "Delete an application from the portfolio",
		Access: tools.AccessDelete, Permission: "components:write",
		Method: "DELETE", Path: "/components/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Application ID (UUID)")},
	},
	{
		Name: "create_application_relation", Description: "Create a relation between two applications",
		Access: tools.AccessCreate, Permission: "components:write",
		Method: "POST", Path: "/components/{sourceId}/relations",
		PathParams: []ParamSpec{uuidParam("sourceId", "Source application ID (UUID)")},
		BodyParams: []ParamSpec{
			{Name: "targetId", Type: "uuid", Description: "Target application ID (UUID)", Required: true},
			stringParam("type", "Relation type (e.g. depends_on)", true),
			stringParam("description", "Relation description", false),
		},
	},
	{
		Name: "delete_application_relation", Description: "Delete a relation between applications",
		Access: tools.AccessDelete, Permission: "components:write",
		Method: "DELETE", Path: "/components/{componentId}/relations/{relationId}",
		PathParams: []ParamSpec{
			uuidParam("componentId", "Application ID (UUID)"),
			uuidParam("relationId", "Relation ID (UUID)"),
		},
	},

	{
		Name: "create_capability", Description: "Create a new business capability. Capabilities form a hierarchy: L1 (top-level, no parent) → L2 (child of L1) → L3 (child of L2) → L4 (child of L3). The level must match the parent depth.",
		Access: tools.AccessCreate, Permission: "capabilities:write",
		Method: "POST", Path: "/capabilities",
		BodyParams: []ParamSpec{
			stringParam("name", "Capability name", true),
			stringParam("level", "Hierarchy level: L1 (no parent), L2 (parent is L1), L3 (parent is L2), or L4 (parent is L3)", true),
			stringParam("parentId", "Parent capability ID (UUID). Required for L2/L3/L4, omit for L1.", false),
			stringParam("description", "Capability description", false),
		},
	},
	{
		Name: "update_capability", Description: "Update an existing capability's properties",
		Access: tools.AccessUpdate, Permission: "capabilities:write",
		Method: "PUT", Path: "/capabilities/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Capability ID (UUID)")},
		BodyParams: []ParamSpec{
			stringParam("name", "New capability name", false),
			stringParam("description", "New capability description", false),
		},
	},
	{
		Name: "delete_capability", Description: "Delete a capability",
		Access: tools.AccessDelete, Permission: "capabilities:write",
		Method: "DELETE", Path: "/capabilities/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Capability ID (UUID)")},
	},
	{
		Name: "realize_capability", Description: "Link an application to a capability (realize it)",
		Access: tools.AccessCreate, Permission: "capabilities:write",
		Method: "POST", Path: "/capabilities/{capabilityId}/realizations",
		PathParams: []ParamSpec{uuidParam("capabilityId", "Capability ID (UUID)")},
		BodyParams: []ParamSpec{
			{Name: "applicationId", Type: "uuid", Description: "Application ID (UUID)", Required: true},
		},
	},
	{
		Name: "unrealize_capability", Description: "Unlink an application from a capability",
		Access: tools.AccessDelete, Permission: "capabilities:write",
		Method: "DELETE", Path: "/capabilities/{capabilityId}/realizations/{realizationId}",
		PathParams: []ParamSpec{
			uuidParam("capabilityId", "Capability ID (UUID)"),
			uuidParam("realizationId", "Realization ID (UUID)"),
		},
	},

	{
		Name: "create_business_domain", Description: "Create a new business domain",
		Access: tools.AccessCreate, Permission: "domains:write",
		Method: "POST", Path: "/business-domains",
		BodyParams: []ParamSpec{
			stringParam("name", "Business domain name", true),
			stringParam("description", "Business domain description", false),
		},
	},
	{
		Name: "update_business_domain", Description: "Update an existing business domain's properties",
		Access: tools.AccessUpdate, Permission: "domains:write",
		Method: "PUT", Path: "/business-domains/{id}",
		PathParams: []ParamSpec{uuidParam("id", "Business domain ID (UUID)")},
		BodyParams: []ParamSpec{
			stringParam("name", "New business domain name", false),
			stringParam("description", "New business domain description", false),
		},
	},
	{
		Name: "assign_capability_to_domain", Description: "Assign an L1 capability to a business domain",
		Access: tools.AccessCreate, Permission: "domains:write",
		Method: "POST", Path: "/business-domains/{domainId}/capabilities",
		PathParams: []ParamSpec{uuidParam("domainId", "Business domain ID (UUID)")},
		BodyParams: []ParamSpec{
			{Name: "capabilityId", Type: "uuid", Description: "Capability ID (UUID) — must be an L1 capability", Required: true},
		},
	},
	{
		Name: "remove_capability_from_domain", Description: "Remove a capability assignment from a business domain",
		Access: tools.AccessDelete, Permission: "domains:write",
		Method: "DELETE", Path: "/business-domains/{domainId}/capabilities/{capabilityId}",
		PathParams: []ParamSpec{
			uuidParam("domainId", "Business domain ID (UUID)"),
			uuidParam("capabilityId", "Capability ID (UUID)"),
		},
	},
}

func RegisterSpecTools(registry *tools.Registry, client *agenthttp.Client) {
	for _, spec := range toolSpecs {
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
