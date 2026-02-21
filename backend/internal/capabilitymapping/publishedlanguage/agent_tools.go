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
			Name: "list_capabilities", Description: "List business capabilities. Capabilities form an L1→L4 hierarchy representing what the business does (not how). L1 are top-level strategic capabilities and the only level assignable to Business Domains. Each can be realized by application components. Filter by name substring. Returns up to limit results.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("name", "Filter by capability name (partial match)", false),
				pl.IntParam("limit", "Max results (1-50, default 20)"),
			},
		},
		{
			Name: "get_capability_details", Description: "Get full details of a capability including its level, parent, children, and realizations (linked application components). Use to inspect hierarchy position and which systems support this capability.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
		},
		{
			Name: "create_capability", Description: "Create a new business capability. Capabilities form a strict hierarchy: L1 (top-level, no parent) → L2 (child of L1) → L3 (child of L2) → L4 (child of L3). The level must match the parent depth. Only L1 capabilities can later be assigned to Business Domains.",
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
			Name: "update_capability", Description: "Update a capability's name or description. Does not change its level, parent, or realizations.",
			Access: pl.AccessUpdate, Permission: "capabilities:write",
			Method: "PUT", Path: "/capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "New capability name", false),
				pl.StringParam("description", "New capability description", false),
			},
		},
		{
			Name: "delete_capability", Description: "Delete a business capability. Fails if the capability has children — remove children first. Also removes any realizations and domain assignments.",
			Access: pl.AccessDelete, Permission: "capabilities:write",
			Method: "DELETE", Path: "/capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
		},
		{
			Name: "realize_capability", Description: "Record that an application component (IT system) realizes a business capability. Realization level: Full (complete support), Partial (some aspects), Planned (future). One capability can have multiple realizing systems. One system can realize multiple capabilities.",
			Access: pl.AccessCreate, Permission: "capabilities:write",
			Method: "POST", Path: "/capabilities/{id}/systems",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				{Name: "componentId", Type: "uuid", Description: "Application component ID (UUID)", Required: true},
				pl.StringParam("realizationLevel", "Realization level: Full, Partial, or Planned", false),
			},
		},
		{
			Name: "unrealize_capability", Description: "Remove a realization link between an application and a capability. Does not affect the capability or the application themselves.",
			Access: pl.AccessDelete, Permission: "capabilities:write",
			Method: "DELETE", Path: "/capability-realizations/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Realization ID (UUID)")},
		},
	}
}

func businessDomainTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_business_domains", Description: "List all business domains. Domains are organizational groupings of L1 capabilities (e.g. Finance, Customer Experience). One L1 capability can belong to multiple domains.",
			Access: pl.AccessRead, Permission: "domains:read",
			Method: "GET", Path: "/business-domains",
		},
		{
			Name: "get_business_domain_details", Description: "Get details of a business domain including its assigned L1 capabilities. Use to see which top-level capabilities are grouped under this domain.",
			Access: pl.AccessRead, Permission: "domains:read",
			Method: "GET", Path: "/business-domains/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Business domain ID (UUID)")},
		},
		{
			Name: "create_business_domain", Description: "Create a new business domain. Domains group L1 capabilities into organizational areas. After creation, assign L1 capabilities using assign_capability_to_domain.",
			Access: pl.AccessCreate, Permission: "domains:write",
			Method: "POST", Path: "/business-domains",
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "Business domain name", true),
				pl.StringParam("description", "Business domain description", false),
			},
		},
		{
			Name: "update_business_domain", Description: "Update a business domain's name or description. Does not affect its capability assignments.",
			Access: pl.AccessUpdate, Permission: "domains:write",
			Method: "PUT", Path: "/business-domains/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Business domain ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "New business domain name", false),
				pl.StringParam("description", "New business domain description", false),
			},
		},
		{
			Name: "assign_capability_to_domain", Description: "Assign an L1 capability to a business domain. Only L1 (top-level) capabilities can be assigned — L2-L4 are included implicitly via their parent. One L1 can belong to multiple domains.",
			Access: pl.AccessCreate, Permission: "domains:write",
			Method: "POST", Path: "/business-domains/{domainId}/capabilities",
			PathParams: []pl.ParamSpec{pl.UUIDParam("domainId", "Business domain ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				{Name: "capabilityId", Type: "uuid", Description: "Capability ID (UUID) — must be an L1 capability", Required: true},
			},
		},
		{
			Name: "remove_capability_from_domain", Description: "Remove an L1 capability assignment from a business domain. Does not delete the capability itself — only removes the domain grouping.",
			Access: pl.AccessDelete, Permission: "domains:write",
			Method: "DELETE", Path: "/business-domains/{domainId}/capabilities/{capabilityId}",
			PathParams: []pl.ParamSpec{
				pl.UUIDParam("domainId", "Business domain ID (UUID)"),
				pl.UUIDParam("capabilityId", "Capability ID (UUID)"),
			},
		},
	}
}
