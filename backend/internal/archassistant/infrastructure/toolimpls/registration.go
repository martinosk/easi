package toolimpls

import (
	"easi/backend/internal/archassistant/application/tools"
	"easi/backend/internal/archassistant/infrastructure/agenthttp"
)

func RegisterMutationTools(registry *tools.Registry, client *agenthttp.Client) {
	registerComponentTools(registry, client)
	registerCapabilityTools(registry, client)
	registerDomainTools(registry, client)
}

func registerComponentTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name:        "create_application",
		Description: "Create a new application in the architecture portfolio",
		Permission:  "components:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "name", Type: "string", Description: "Application name", Required: true},
			{Name: "description", Type: "string", Description: "Application description"},
		},
	}, &createApplicationTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "update_application",
		Description: "Update an existing application's properties",
		Permission:  "components:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Application ID (UUID)", Required: true},
			{Name: "name", Type: "string", Description: "New application name"},
			{Name: "description", Type: "string", Description: "New application description"},
		},
	}, &updateApplicationTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "delete_application",
		Description: "Delete an application from the portfolio",
		Permission:  "components:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Application ID (UUID)", Required: true},
		},
	}, &deleteApplicationTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "create_application_relation",
		Description: "Create a relation between two applications",
		Permission:  "components:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "sourceId", Type: "string", Description: "Source application ID (UUID)", Required: true},
			{Name: "targetId", Type: "string", Description: "Target application ID (UUID)", Required: true},
			{Name: "type", Type: "string", Description: "Relation type (e.g. depends_on)", Required: true},
			{Name: "description", Type: "string", Description: "Relation description"},
		},
	}, &createApplicationRelationTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "delete_application_relation",
		Description: "Delete a relation between applications",
		Permission:  "components:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "componentId", Type: "string", Description: "Application ID (UUID)", Required: true},
			{Name: "relationId", Type: "string", Description: "Relation ID (UUID)", Required: true},
		},
	}, &deleteApplicationRelationTool{client: client})
}

func registerCapabilityTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name:        "create_capability",
		Description: "Create a new business capability. Capabilities form a hierarchy: L1 (top-level, no parent) → L2 (child of L1) → L3 (child of L2) → L4 (child of L3). The level must match the parent depth.",
		Permission:  "capabilities:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "name", Type: "string", Description: "Capability name", Required: true},
			{Name: "level", Type: "string", Description: "Hierarchy level: L1 (no parent), L2 (parent is L1), L3 (parent is L2), or L4 (parent is L3)", Required: true},
			{Name: "parentId", Type: "string", Description: "Parent capability ID (UUID). Required for L2/L3/L4, omit for L1."},
			{Name: "description", Type: "string", Description: "Capability description"},
		},
	}, &createCapabilityTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "update_capability",
		Description: "Update an existing capability's properties",
		Permission:  "capabilities:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Capability ID (UUID)", Required: true},
			{Name: "name", Type: "string", Description: "New capability name"},
			{Name: "description", Type: "string", Description: "New capability description"},
		},
	}, &updateCapabilityTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "delete_capability",
		Description: "Delete a capability",
		Permission:  "capabilities:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Capability ID (UUID)", Required: true},
		},
	}, &deleteCapabilityTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "realize_capability",
		Description: "Link an application to a capability (realize it)",
		Permission:  "capabilities:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "capabilityId", Type: "string", Description: "Capability ID (UUID)", Required: true},
			{Name: "applicationId", Type: "string", Description: "Application ID (UUID)", Required: true},
		},
	}, &realizeCapabilityTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "unrealize_capability",
		Description: "Unlink an application from a capability",
		Permission:  "capabilities:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "capabilityId", Type: "string", Description: "Capability ID (UUID)", Required: true},
			{Name: "realizationId", Type: "string", Description: "Realization ID (UUID)", Required: true},
		},
	}, &unrealizeCapabilityTool{client: client})
}

func registerDomainTools(registry *tools.Registry, client *agenthttp.Client) {
	registry.Register(tools.ToolDefinition{
		Name:        "create_business_domain",
		Description: "Create a new business domain",
		Permission:  "domains:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "name", Type: "string", Description: "Business domain name", Required: true},
			{Name: "description", Type: "string", Description: "Business domain description"},
		},
	}, &createBusinessDomainTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "update_business_domain",
		Description: "Update an existing business domain's properties",
		Permission:  "domains:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "id", Type: "string", Description: "Business domain ID (UUID)", Required: true},
			{Name: "name", Type: "string", Description: "New business domain name"},
			{Name: "description", Type: "string", Description: "New business domain description"},
		},
	}, &updateBusinessDomainTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "assign_capability_to_domain",
		Description: "Assign an L1 capability to a business domain",
		Permission:  "domains:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "domainId", Type: "string", Description: "Business domain ID (UUID)", Required: true},
			{Name: "capabilityId", Type: "string", Description: "Capability ID (UUID) — must be an L1 capability", Required: true},
		},
	}, &assignCapabilityToDomainTool{client: client})

	registry.Register(tools.ToolDefinition{
		Name:        "remove_capability_from_domain",
		Description: "Remove a capability assignment from a business domain",
		Permission:  "domains:write",
		Access:      tools.AccessWrite,
		Parameters: []tools.ParameterDef{
			{Name: "domainId", Type: "string", Description: "Business domain ID (UUID)", Required: true},
			{Name: "capabilityId", Type: "string", Description: "Capability ID (UUID)", Required: true},
		},
	}, &removeCapabilityFromDomainTool{client: client})
}
