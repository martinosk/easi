package publishedlanguage

import (
	"easi/backend/internal/shared/agenttools"
)

func enterpriseCapabilityTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "list_enterprise_capabilities", Description: "List enterprise capabilities. Enterprise capabilities group domain-level capabilities across business domains into cross-cutting strategic themes (e.g. Digital Customer Engagement). An enterprise capability is composed of the source capabilities of its active direction plus their subtrees, and can carry strategic importance ratings and maturity targets.",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities",
		},
		{
			Name: "get_enterprise_capability_details", Description: "Get details of an enterprise capability including its included-capability count, domain count, strategic importance ratings, and target maturity. Use get_enterprise_capability_composition to see which domain capabilities roll up into this strategic theme.",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "create_enterprise_capability", Description: "Create a new enterprise capability. Enterprise capabilities are cross-domain strategic groupings composed of domain-level capabilities. After creation, capture a direction with source capabilities to compose it.",
			Access: agenttools.AccessCreate, Permission: "enterprise-arch:write",
			Method: "POST", Path: "/enterprise-capabilities",
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "Enterprise capability name", true),
				agenttools.StringParam("description", "Enterprise capability description", false),
			},
		},
		{
			Name: "update_enterprise_capability", Description: "Update an enterprise capability's name or description. Does not affect its composition, importance ratings, or maturity targets.",
			Access: agenttools.AccessUpdate, Permission: "enterprise-arch:write",
			Method: "PUT", Path: "/enterprise-capabilities/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				agenttools.StringParam("name", "New enterprise capability name", false),
				agenttools.StringParam("description", "New enterprise capability description", false),
			},
		},
		{
			Name: "delete_enterprise_capability", Description: "Delete an enterprise capability. Its active direction (if any) is rejected as part of deletion, releasing all source capabilities.",
			Access: agenttools.AccessDelete, Permission: "enterprise-arch:write",
			Method: "DELETE", Path: "/enterprise-capabilities/{id}",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "get_enterprise_strategic_importance", Description: "Get strategic importance ratings for an enterprise capability. Importance is rated per strategy pillar (defined in MetaModel) using levels like Critical, High, Medium, Low. Shows how strategically significant this enterprise capability is across each strategic dimension.",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/enterprise-capabilities/{id}/strategic-importance",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
		},
		{
			Name: "set_enterprise_strategic_importance", Description: "Set the strategic importance of an enterprise capability for a specific strategy pillar. Importance levels: Critical, High, Medium, Low. Each enterprise capability can have one importance rating per pillar.",
			Access: agenttools.AccessCreate, Permission: "enterprise-arch:write",
			Method: "POST", Path: "/enterprise-capabilities/{id}/strategic-importance",
			PathParams: []agenttools.ParamSpec{agenttools.UUIDParam("id", "Enterprise capability ID (UUID)")},
			BodyParams: []agenttools.ParamSpec{
				{Name: "pillarId", Type: "uuid", Description: "Strategy pillar ID (UUID)", Required: true},
				agenttools.StringParam("importance", "Importance level: Critical, High, Medium, or Low", true),
			},
		},
	}
}

func enterpriseAnalysisTools() []agenttools.AgentToolSpec {
	return []agenttools.AgentToolSpec{
		{
			Name: "get_time_suggestions", Description: "Get TIME classification suggestions for enterprise capabilities. TIME (Tolerate, Invest, Migrate, Eliminate) is an investment categorization framework. Suggestions are computed from the strategic importance, maturity gaps, and fit scores of included capabilities.",
			Access: agenttools.AccessRead, Permission: "enterprise-arch:read",
			Method: "GET", Path: "/time-suggestions",
		},
	}
}
