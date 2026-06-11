package publishedlanguage

import (
	pl "easi/backend/internal/archassistant/publishedlanguage"
)

func AgentTools() []pl.AgentToolSpec {
	var specs []pl.AgentToolSpec
	specs = append(specs, enterpriseCapabilityTools()...)
	specs = append(specs, enterpriseAnalysisTools()...)
	return specs
}

func enterpriseCapabilityTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "list_enterprise_capabilities", Description: "List enterprise capabilities. Enterprise capabilities group domain-level capabilities across business domains into cross-cutting strategic themes (e.g. Digital Customer Engagement). An enterprise capability is composed of the source capabilities of its active direction plus their subtrees, and can carry strategic importance ratings and maturity targets.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities",
		},
		{
			Name: "get_enterprise_capability_details", Description: "Get details of an enterprise capability including its included-capability count, domain count, strategic importance ratings, and target maturity. Use get_enterprise_capability_composition to see which domain capabilities roll up into this strategic theme.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "create_enterprise_capability", Description: "Create a new enterprise capability. Enterprise capabilities are cross-domain strategic groupings composed of domain-level capabilities. After creation, capture a direction with source capabilities to compose it.",
			Access: pl.AccessCreate, Permission: "enterprise-arch:write",
			Method: "POST", Path: "/enterprise-capabilities",
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "Enterprise capability name", true),
				pl.StringParam("description", "Enterprise capability description", false),
			},
		},
		{
			Name: "update_enterprise_capability", Description: "Update an enterprise capability's name or description. Does not affect its composition, importance ratings, or maturity targets.",
			Access: pl.AccessUpdate, Permission: "enterprise-arch:write",
			Method: "PUT", Path: "/enterprise-capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				pl.StringParam("name", "New enterprise capability name", false),
				pl.StringParam("description", "New enterprise capability description", false),
			},
		},
		{
			Name: "delete_enterprise_capability", Description: "Delete an enterprise capability. Its active direction (if any) is rejected as part of deletion, releasing all source capabilities.",
			Access: pl.AccessDelete, Permission: "enterprise-arch:write",
			Method: "DELETE", Path: "/enterprise-capabilities/{id}",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "get_enterprise_capability_composition", Description: "Get the composition of an enterprise capability: every domain capability included via its active direction's sources and their subtrees, grouped by business domain, with carve-out attribution where a more specific source on another enterprise capability owns a subtree.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}/composition",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "search_direction_source_candidates", Description: "Search domain capabilities by name as candidate sources for an enterprise capability's direction, with per-candidate eligibility (a capability may be the explicit source of at most one active direction).",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/capabilities/source-candidates",
			QueryParams: []pl.ParamSpec{
				pl.StringParam("q", "Search term (case-insensitive substring match on capability name)", true),
				{Name: "ecId", Type: "uuid", Description: "Enterprise capability the sources are searched for", Required: true},
				{Name: "domainId", Type: "uuid", Description: "Filter to capabilities in this business domain"},
				pl.IntParam("limit", "Max results to return (default 20)"),
			},
		},
		{
			Name: "get_enterprise_strategic_importance", Description: "Get strategic importance ratings for an enterprise capability. Importance is rated per strategy pillar (defined in MetaModel) using levels like Critical, High, Medium, Low. Shows how strategically significant this enterprise capability is across each strategic dimension.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}/strategic-importance",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "set_enterprise_strategic_importance", Description: "Set the strategic importance of an enterprise capability for a specific strategy pillar. Importance levels: Critical, High, Medium, Low. Each enterprise capability can have one importance rating per pillar.",
			Access: pl.AccessCreate, Permission: "enterprise-arch:write",
			Method: "POST", Path: "/enterprise-capabilities/{id}/strategic-importance",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
			BodyParams: []pl.ParamSpec{
				{Name: "pillarId", Type: "uuid", Description: "Strategy pillar ID (UUID)", Required: true},
				pl.StringParam("importance", "Importance level: Critical, High, Medium, or Low", true),
			},
		},
	}
}

func enterpriseAnalysisTools() []pl.AgentToolSpec {
	return []pl.AgentToolSpec{
		{
			Name: "get_time_suggestions", Description: "Get TIME classification suggestions for enterprise capabilities. TIME (Tolerate, Invest, Migrate, Eliminate) is an investment categorization framework. Suggestions are computed from the strategic importance, maturity gaps, and fit scores of included capabilities.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/time-suggestions",
		},
		{
			Name: "get_maturity_analysis", Description: "Get maturity analysis candidates — enterprise capabilities where the current maturity of included domain capabilities falls below the target maturity level. Use to identify strategic themes that need maturity investment.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/maturity-analysis",
		},
		{
			Name: "get_maturity_gap_detail", Description: "Get detailed maturity gap analysis for a specific enterprise capability. Shows each included domain capability's current maturity versus the enterprise capability's target maturity, highlighting where gaps exist.",
			Access: pl.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}/maturity-gap",
			PathParams: []pl.ParamSpec{pl.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
	}
}
