package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	var specs []pl.AgentToolSpec
	specs = append(specs, capabilityTools()...)
	specs = append(specs, businessDomainTools()...)
	return specs
}

func capabilityTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_capabilities", Description: "List business capabilities. Optionally filter by name.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("name", "Filter by capability name (partial match)", false),
				pl.IntParam("limit", "Max results (1-50, default 20)"),
			},
		},
		{
			Name: "get_capability_details", Description: "Get full details of a capability including realizations",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
		},
		{
			Name: "create_capability", Description: "Create a new business capability. Capabilities form a hierarchy: L1 (top-level, no parent) → L2 (child of L1) → L3 (child of L2) → L4 (child of L3). The level must match the parent depth.",
			Access: pl.AccessCreate, Permission: "capabilities:write",
			Method: "POST", Path: "/capabilities",
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "Capability name", true),
				pl.StringParam("level", "Hierarchy level: L1 (no parent), L2 (parent is L1), L3 (parent is L2), or L4 (parent is L3)", true),
				pl.StringParam("parentId", "Parent capability ID (UUID). Required for L2/L3/L4, omit for L1.", false),
				pl.StringParam("description", "Capability description", false),
			},
		},
		{
			Name: "update_capability", Description: "Update an existing capability's properties",
			Access: pl.AccessUpdate, Permission: "capabilities:write",
			Method: "PUT", Path: "/capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "New capability name", false),
				pl.StringParam("description", "New capability description", false),
			},
		},
		{
			Name: "delete_capability", Description: "Delete a capability",
			Access: pl.AccessDelete, Permission: "capabilities:write",
			Method: "DELETE", Path: "/capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
		},
		{
			Name: "realize_capability", Description: "Link an application to a capability (realize it)",
			Access: pl.AccessCreate, Permission: "capabilities:write",
			Method: "POST", Path: "/capabilities/{id}/systems",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				{Name: "componentId", Type: "uuid", Description: "Application component ID (UUID)", Required: true},
				pl.StringParam("realizationLevel", "Realization level", false),
			},
		},
		{
			Name: "unrealize_capability", Description: "Unlink an application from a capability",
			Access: pl.AccessDelete, Permission: "capabilities:write",
			Method: "DELETE", Path: "/capability-realizations/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Realization ID (UUID)")},
		},
	}
}

func businessDomainTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_business_domains", Description: "List all business domains",
			Access: pl.AccessRead, Permission: "domains:read",
			Method: "GET", Path: "/business-domains",
		},
		{
			Name: "get_business_domain_details", Description: "Get details of a business domain with its capabilities",
			Access: pl.AccessRead, Permission: "domains:read",
			Method: "GET", Path: "/business-domains/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Business domain ID (UUID)")},
		},
		{
			Name: "create_business_domain", Description: "Create a new business domain",
			Access: pl.AccessCreate, Permission: "domains:write",
			Method: "POST", Path: "/business-domains",
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "Business domain name", true),
				pl.StringParam("description", "Business domain description", false),
			},
		},
		{
			Name: "update_business_domain", Description: "Update an existing business domain's properties",
			Access: pl.AccessUpdate, Permission: "domains:write",
			Method: "PUT", Path: "/business-domains/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Business domain ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "New business domain name", false),
				pl.StringParam("description", "New business domain description", false),
			},
		},
		{
			Name: "assign_capability_to_domain", Description: "Assign an L1 capability to a business domain",
			Access: pl.AccessCreate, Permission: "domains:write",
			Method: "POST", Path: "/business-domains/{domainId}/capabilities",
			PathParams: []pl.ParamSpec{pl.UUIDParam("domainId", "Business domain ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				{Name: "capabilityId", Type: "uuid", Description: "Capability ID (UUID) — must be an L1 capability", Required: true},
			},
		},
		{
			Name: "remove_capability_from_domain", Description: "Remove a capability assignment from a business domain",
			Access: pl.AccessDelete, Permission: "domains:write",
			Method: "DELETE", Path: "/business-domains/{domainId}/capabilities/{capabilityId}",
			PathParams: []pl.ParamSpec{
				pl.UUIDParam("domainId", "Business domain ID (UUID)"),
				pl.UUIDParam("capabilityId", "Capability ID (UUID)"),
			},
		},
	}
}
