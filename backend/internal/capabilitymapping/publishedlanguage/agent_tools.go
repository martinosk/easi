package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	var specs []pl.AgentToolSpec
	specs = append(specs, capabilityTools()...)
	specs = append(specs, capabilityMetadataTools()...)
	specs = append(specs, businessDomainTools()...)
	specs = append(specs, dependencyTools()...)
	specs = append(specs, strategyTools()...)
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

func capabilityMetadataTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "get_capability_metadata_index", Description: "Get the metadata reference index for capabilities. Returns links to available metadata categories: maturity levels, statuses, and ownership models. Use to discover what metadata options exist.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/metadata",
		},
		{
			Name: "get_capability_maturity_levels", Description: "Get available maturity levels for capabilities. Maturity levels describe how mature a capability is (e.g. Initial, Defined, Managed, Optimizing). Configured in the MetaModel maturity scale.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/metadata/maturity-levels",
		},
		{
			Name: "get_capability_statuses", Description: "Get available status values for capabilities. Statuses track the lifecycle state of a capability (e.g. Active, Planned, Retiring).",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/metadata/statuses",
		},
		{
			Name: "get_capability_ownership_models", Description: "Get available ownership model values for capabilities. Ownership models describe how a capability is governed (e.g. Centralized, Federated, Shared).",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/metadata/ownership-models",
		},
		{
			Name: "get_capability_expert_roles", Description: "Get available expert role values for capabilities. Expert roles define the types of subject matter experts that can be assigned to capabilities (e.g. Business Owner, Technical Lead).",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/expert-roles",
		},
		{
			Name: "update_capability_metadata", Description: "Update operational metadata of a capability: maturity level, status, ownership model, and owners. Does not change the capability's name, description, hierarchy position, or realizations.",
			Access: pl.AccessUpdate, Permission: "capabilities:write",
			Method: "PUT", Path: "/capabilities/{id}/metadata",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("status", "Capability status (e.g. Active, Planned, Retiring)", true),
				pl.StringParam("maturityLevel", "Maturity level name", false),
				pl.StringParam("ownershipModel", "Ownership model (e.g. Centralized, Federated)", false),
				pl.StringParam("primaryOwner", "Primary owner name", false),
				pl.StringParam("eaOwner", "EA owner name", false),
			},
		},
		{
			Name: "get_capability_realizations", Description: "Get the application components (IT systems) that realize a specific capability. Shows which systems support this capability and at what realization level (Full, Partial, Planned).",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/{id}/systems",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
		},
		{
			Name: "get_capability_business_domains", Description: "Get the business domains that a capability belongs to. For L1 capabilities this shows direct assignments; for L2-L4 it shows inherited domain membership through the parent hierarchy.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/{id}/business-domains",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
		},
		{
			Name: "get_domain_importance_overview", Description: "Get strategic importance ratings for all capabilities within a business domain. Shows how critical each capability is to each strategy pillar in this domain context. Use to compare strategic significance across a domain's capabilities.",
			Access: pl.AccessRead, Permission: "domains:read",
			Method: "GET", Path: "/business-domains/{id}/importance",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Business domain ID (UUID)")},
		},
		{
			Name: "get_fit_scores_by_pillar", Description: "Get all application fit scores for a specific strategy pillar. Shows how well each application component supports this strategic dimension. Use to compare application fitness across the portfolio for a given strategy pillar.",
			Access: pl.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/strategy-pillars/{pillarId}/fit-scores",
			PathParams: []pl.ParamSpec{pl.UUIDParam("pillarId", "Strategy pillar ID (UUID)")},
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

func dependencyTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_capability_dependencies", Description: "List all capability dependencies. Dependencies are directed links between capabilities showing that one capability depends on another to deliver its function. Use to map upstream/downstream capability relationships.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capability-dependencies",
		},
		{
			Name: "create_capability_dependency", Description: "Create a dependency from one capability to another, indicating the source capability depends on the target to function. Dependencies are directional: source → target means source depends on target.",
			Access: pl.AccessCreate, Permission: "capabilities:write",
			Method: "POST", Path: "/capability-dependencies",
			BodyParams: []pl.ParamSpec{
				{Name: "sourceCapabilityId", Type: "uuid", Description: "Source capability ID (UUID) — the capability that depends", Required: true},
				{Name: "targetCapabilityId", Type: "uuid", Description: "Target capability ID (UUID) — the capability depended upon", Required: true},
				pl.StringParam("description", "Dependency description", false),
			},
		},
		{
			Name: "delete_capability_dependency", Description: "Remove a dependency between two capabilities. Does not affect the capabilities themselves.",
			Access: pl.AccessDelete, Permission: "capabilities:write",
			Method: "DELETE", Path: "/capability-dependencies/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Dependency ID (UUID)")},
		},
		{
			Name: "get_capability_children", Description: "Get the direct children of a capability in the hierarchy. L1 capabilities have L2 children, L2 have L3, L3 have L4. L4 capabilities have no children. Use to navigate the capability tree.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/{id}/children",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Parent capability ID (UUID)")},
		},
	}
}

func strategyTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "get_strategy_importance", Description: "Get strategic importance ratings for a capability. Importance is rated per strategy pillar (defined in MetaModel) for a capability within a business domain context. Shows how critical this capability is to each strategic dimension.",
			Access: pl.AccessRead, Permission: "capabilities:read",
			Method: "GET", Path: "/capabilities/{id}/importance",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Capability ID (UUID)")},
		},
		{
			Name: "set_strategy_importance", Description: "Set the strategic importance of a capability for a specific strategy pillar within a business domain context. Importance levels: Critical, High, Medium, Low. Each capability+domain pair can have one importance rating per pillar.",
			Access: pl.AccessCreate, Permission: "domains:write",
			Method: "POST", Path: "/business-domains/{id}/capabilities/{capabilityId}/importance",
			PathParams: []pl.ParamSpec{
				pl.UUIDParam("id", "Business domain ID (UUID)"),
				pl.UUIDParam("capabilityId", "Capability ID (UUID)"),
			},
			BodyParams: []pl.ParamSpec{
				{Name: "pillarId", Type: "uuid", Description: "Strategy pillar ID (UUID)", Required: true},
				pl.StringParam("importance", "Importance level: Critical, High, Medium, or Low", true),
			},
		},
		{
			Name: "get_application_fit_scores", Description: "Get fit scores for an application component across all strategy pillars. Fit scores rate how well an application supports each strategic dimension (e.g. excellent, adequate, poor). Use to assess an application's strategic alignment.",
			Access: pl.AccessRead, Permission: "components:read",
			Method: "GET", Path: "/components/{id}/fit-scores",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Application component ID (UUID)")},
		},
		{
			Name: "set_application_fit_score", Description: "Set or update the fit score of an application for a specific strategy pillar. Fit scores rate how well the application supports that strategic dimension. The score is a numeric value where higher means better fit.",
			Access: pl.AccessUpdate, Permission: "components:write",
			Method: "PUT", Path: "/components/{id}/fit-scores/{pillarId}",
			PathParams: []pl.ParamSpec{
				pl.UUIDParam("id", "Application component ID (UUID)"),
				pl.UUIDParam("pillarId", "Strategy pillar ID (UUID)"),
			},
			BodyParams: []pl.ParamSpec{
				pl.IntParam("score", "Fit score value (higher = better fit)"),
			},
		},
		{
			Name: "get_strategic_fit_analysis", Description: "Get strategic fit analysis for a specific strategy pillar. Shows capabilities ranked by importance with their realizing applications and fit scores, revealing gaps where important capabilities lack well-fitting systems. Use for gap analysis and investment prioritization.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/strategic-fit-analysis/{pillarId}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("pillarId", "Strategy pillar ID (UUID)")},
		},
	}
}
